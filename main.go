package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"syscall"

	"github.com/orivej/e"
)

const (
	PTRACE_O_EXITKILL = 1 << 20 // since Linux 3.8
)

func main() {
	flTrace := flag.String("t", "/dev/null", "trace output file")
	flDeps := flag.String("d", "", "deps output file")
	flDepsWithOutput := flag.Bool("do", false, "output only deps with outputs")
	flag.Parse()

	args := flag.Args()
	runtime.LockOSThread()
	proc, err := trace(args)
	e.Exit(err)
	pid := proc.Pid
	_, err = syscall.Wait4(pid, nil, 0, nil)
	e.Exit(err)

	f, err2 := os.Create(*flTrace)
	e.Exit(err2)
	defer e.CloseOrPrint(f)
	os.Stdout = f

	err = syscall.PtraceSetOptions(pid, PTRACE_O_EXITKILL|
		syscall.PTRACE_O_TRACESYSGOOD|
		syscall.PTRACE_O_TRACEEXEC|
		syscall.PTRACE_O_TRACEEXIT|
		syscall.PTRACE_O_TRACECLONE|
		syscall.PTRACE_O_TRACEFORK|
		syscall.PTRACE_O_TRACEVFORK)
	e.Exit(err)
	resume(pid)

	var records []Record
	recorder := func(p *ProcState) {
		r := p.Record()
		no := len(r.Outputs)
		noOutputs := no == 0 || (no == 1 && r.Outputs[0] == "/dev/tty")
		if *flDepsWithOutput && noOutputs {
			return
		}
		records = append(records, r)
	}
	mainLoop(pid, recorder)

	if *flDeps != "" {
		f, err := os.Create(*flDeps)
		e.Exit(err)
		defer e.CloseOrPrint(f)
		err = json.NewEncoder(f).Encode(records)
		e.Exit(err)
	}
}

func mainLoop(mainPID int, recorder func(p *ProcState)) {
	var err error
	pstates := map[int]*ProcState{}
	pstates[mainPID] = NewProcState()
	pstates[mainPID].CurDir, err = os.Getwd()
	e.Exit(err)

	suspended := map[int]bool{}
	for {
		pid, wstatus, ok := waitForSyscall()
		if !ok {
			if pid == mainPID {
				// Exit with the first child.
				break
			}
			continue
		}

		// Select PID state.
		pstate, ok := pstates[pid]
		if !ok {
			// Keep this PID suspended until we are notified of its creation.
			suspended[pid] = true
			continue
		}

		switch wstatus {
		case syscall.PTRACE_EVENT_FORK,
			syscall.PTRACE_EVENT_VFORK,
			syscall.PTRACE_EVENT_VFORK_DONE,
			syscall.PTRACE_EVENT_CLONE:
			// New proc.
			unewpid, err := syscall.PtraceGetEventMsg(pid)
			e.Exit(err)
			newpid := int(unewpid)
			pstates[newpid] = pstate.Clone()
			fmt.Println(pid, "clone", newpid)
			// Resume suspended.
			if suspended[newpid] {
				delete(suspended, newpid)
				resume(newpid)
			}
		case syscall.PTRACE_EVENT_EXEC:
			uoldpid, err := syscall.PtraceGetEventMsg(pid)
			e.Exit(err)
			oldpid := int(uoldpid)
			if oldpid != pid && pstate.IOs.Cnt != 1 {
				panic("lost pstate")
			}
			pstate = pstates[oldpid]
			terminate(oldpid, pstate, recorder)
			pstate.SysEnter = true
			pstate.CurCmd = pstate.NextCmd
			pstates[pid] = pstate
			fmt.Println(oldpid, "_exec", pid)
		case syscall.PTRACE_EVENT_EXIT:
			terminate(pid, pstate, recorder)
			fmt.Println(pid, "_exit")
		case 0:
			// Toggle edge.
			pstate.SysEnter = !pstate.SysEnter

			if pstate.SysEnter {
				sysenter(pid, pstate)
			} else {
				sysexit(pid, pstate)
			}
		default:
			panic("unexpected wstatus")
		}
		resume(pid)
	}
}

func terminate(pid int, pstate *ProcState, recorder func(p *ProcState)) {
	if pstate.IOs.Cnt == 1 && len(pstate.IOs.Map) != 0 {
		recorder(pstate)
		fmt.Println(pid, "record", pstate.CurCmd)
	}
	pstate.ResetIOs()
}

func sysenter(pid int, pstate *ProcState) {
	regs := getRegs(pid)
	pstate.Syscall = int(regs.Orig_rax)
	switch pstate.Syscall {
	case syscall.SYS_EXECVE:
		if regs.Rdi == 0 {
			break
		}
		pstate.NextCmd = Cmd{
			Path: readString(pid, regs.Rdi),
			Args: readStrings(pid, regs.Rsi),
		}
	}
}

func sysexit(pid int, pstate *ProcState) {
	regs := getRegs(pid)
	ret := int(regs.Rax)
	if ret < 0 {
		return
	}
	switch pstate.Syscall {
	case syscall.SYS_OPEN:
		path := pstate.Abs(readString(pid, regs.Rdi))
		flags := regs.Rsi
		write := flags&(syscall.O_WRONLY|syscall.O_RDWR) != 0
		pstate.FDs[ret] = path
		pstate.IOs.Map[path] = pstate.IOs.Map[path] || write
		fmt.Println(pid, "open", write, path)
	case syscall.SYS_CHDIR:
		path := pstate.Abs(readString(pid, regs.Rdi))
		pstate.CurDir = path
		fmt.Println(pid, "chdir", path)
	case syscall.SYS_FCHDIR:
		path := pstate.FDs[int(regs.Rdi)]
		pstate.CurDir = path
		fmt.Println(pid, "fchdir", path)
	case syscall.SYS_EXECVE:
		fmt.Println(pid, "execve", pstate.CurCmd)
	}
}
