package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"syscall"

	"github.com/orivej/e"
	"golang.org/x/sys/unix"
)

const PTRACE_O_EXITKILL = 1 << 20 // since Linux 3.8

var tracee = "tracee"

var wstatusText = map[int]string{
	syscall.PTRACE_EVENT_FORK:       "fork",
	syscall.PTRACE_EVENT_VFORK:      "vfork",
	syscall.PTRACE_EVENT_VFORK_DONE: "vforke",
	syscall.PTRACE_EVENT_CLONE:      "clone",
}

func main() {
	flTrace := flag.String("t", "/dev/null", "trace output file")
	flTracee := flag.String("tracee", tracee, "tracee command")
	flDeps := flag.String("d", "", "deps output file")
	flDepsWithOutput := flag.Bool("do", false, "output only deps with outputs")
	flag.Parse()

	args := flag.Args()
	runtime.LockOSThread()
	proc, err := trace(*flTracee, args)
	e.Exit(err)
	pid := proc.Pid
	_, err = syscall.Wait4(pid, nil, 0, nil)
	e.Exit(err)

	f, err := os.Create(*flTrace)
	e.Exit(err)
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

	sys := NewSysState()
	records := []Record{}
	recorder := func(p *ProcState) {
		r := p.Record(sys)
		no := len(r.Outputs)
		noOutputs := no == 0 || (no == 1 && r.Outputs[0] == "/dev/tty")
		if *flDepsWithOutput && noOutputs {
			return
		}
		records = append(records, r)
	}
	mainLoop(sys, pid, recorder)

	if *flDeps != "" {
		f, err := os.Create(*flDeps)
		e.Exit(err)
		defer e.CloseOrPrint(f)
		err = json.NewEncoder(f).Encode(records)
		e.Exit(err)
	}
}

func mainLoop(sys *SysState, mainPID int, recorder func(p *ProcState)) {
	var err error
	pstates := map[int]*ProcState{}
	pstates[mainPID] = NewProcState()
	pstates[mainPID].CurDir, err = os.Getwd()
	e.Exit(err)

	suspended := map[int]int{}
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
			suspended[pid] = wstatus
			fmt.Println(pid, "_suspend")
			continue
		}

	wstatusSwitch:
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
			fmt.Println(pid, wstatusText[wstatus], newpid)
			// Resume suspended.
			if newstatus, ok := suspended[newpid]; ok {
				delete(suspended, newpid)
				resume(pid)
				fmt.Println(newpid, "_resume")
				pid, wstatus, pstate = newpid, newstatus, pstates[newpid]
				goto wstatusSwitch
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
			sys.Proc.Exec(pstate)
			pstate.SysEnter = true
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
				sysexit(pid, pstate, sys)
			}
		default:
			panic("unexpected wstatus")
		}
		resume(pid)
	}
}

func terminate(pid int, pstate *ProcState, recorder func(p *ProcState)) {
	if pstate.IOs.Cnt == 1 && pstate.CurCmd.ID != 0 {
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
		pstate.NextCmd = Cmd{
			Path: pstate.Abs(readString(pid, regs.Rdi)),
			Args: readStrings(pid, regs.Rsi),
			Dir:  pstate.CurDir,
		}
		fmt.Println(pid, "execve", pstate.NextCmd)
	}
}

func sysexit(pid int, pstate *ProcState, sys *SysState) {
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
		inode := sys.FS.Inode(path)
		pstate.FDs[ret] = inode
		if !pstate.IOs.Map[true][inode] {
			// Treat reads after writes as writes only.
			pstate.IOs.Map[write][inode] = true
		}
		fmt.Println(pid, "open", write, path)
	case syscall.SYS_CHDIR:
		path := pstate.Abs(readString(pid, regs.Rdi))
		pstate.CurDir = path
		fmt.Println(pid, "chdir", path)
	case syscall.SYS_FCHDIR:
		path := sys.FS.Path(pstate.FDs[int(regs.Rdi)])
		pstate.CurDir = path
		fmt.Println(pid, "fchdir", path)
	case syscall.SYS_RENAME:
		oldpath := pstate.Abs(readString(pid, regs.Rdi))
		newpath := pstate.Abs(readString(pid, regs.Rsi))
		sys.FS.Rename(oldpath, newpath)
		fmt.Println(pid, "rename", oldpath, newpath)
	case syscall.SYS_RENAMEAT, unix.SYS_RENAMEAT2:
		panic("renameat unimplemented")
	}
}
