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

var (
	flEnv      = flag.Bool("e", false, "record environment variables")
	flUndelete = flag.Bool("u", false, "undelete files")
)

func main() {
	flTrace := flag.String("t", "/dev/null", "trace output file")
	flTracee := flag.String("tracee", tracee, "tracee command")
	flDeps := flag.String("d", "", "deps output file")
	flDepsWithOutput := flag.Bool("do", false, "output only deps with outputs")
	flScripts := flag.String("s", "", "scripts output dir")
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
	resume(pid, 0)

	if *flScripts != "" {
		err := os.MkdirAll(*flScripts, os.ModePerm)
		e.Exit(err)
	}
	sys := NewSysState()
	records := []Record{}
	recorder := func(p *ProcState) {
		r := p.Record(sys)
		no := len(r.Outputs)
		noOutputs := no == 0 || (no == 1 && r.Outputs[0] == "/dev/tty")
		if *flDepsWithOutput && noOutputs {
			return
		}
		if *flScripts != "" {
			writeScript(*flScripts, r)
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
	terminated := map[int]bool{}
	term := func(pid int) {
		if !terminated[pid] {
			terminate(pid, pstates[pid], recorder)
			terminated[pid] = true
		}
	}
	for {
		pid, wstatus, ok := waitForSyscall()
		if !ok {
			// Linux may fail to report PTRACE_EVENT_EXIT.
			term(pid)

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
			delete(terminated, newpid)
			fmt.Println(pid, wstatusText[wstatus], newpid)
			// Resume suspended.
			if newstatus, ok := suspended[newpid]; ok {
				delete(suspended, newpid)
				resume(pid, 0)
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
			term(oldpid)
			delete(terminated, pid)
			sys.Proc.Exec(pstate)
			pstate.SysEnter = true
			pstates[pid] = pstate
			fmt.Println(oldpid, "_exec", pid)
		case syscall.PTRACE_EVENT_EXIT:
			term(pid)
			fmt.Println(pid, "_exit")
		case 0:
			// Toggle edge.
			pstate.SysEnter = !pstate.SysEnter

			var ok bool
			if pstate.SysEnter {
				ok = sysenter(pid, pstate)
			} else {
				ok = sysexit(pid, pstate, sys)
			}

			if !ok {
				term(pid)
				fmt.Println(pid, "_vanish")
				continue
			}
		default:
			panic("unexpected wstatus")
		}
		resume(pid, 0)
	}
}

func terminate(pid int, pstate *ProcState, recorder func(p *ProcState)) {
	if pstate.IOs.Cnt == 1 && pstate.CurCmd.ID != 0 {
		recorder(pstate)
		fmt.Println(pid, "record", pstate.CurCmd)
	}
	pstate.ResetIOs()
}

func sysenter(pid int, pstate *ProcState) bool {
	regs, ok := getRegs(pid)
	if !ok {
		return false
	}
	pstate.Syscall = int(regs.Orig_rax)
	switch pstate.Syscall {
	case syscall.SYS_EXECVE:
		pstate.NextCmd = Cmd{
			Path: pstate.Abs(readString(pid, regs.Rdi)),
			Args: readStrings(pid, regs.Rsi),
			Dir:  pstate.CurDir,
		}
		if *flEnv {
			pstate.NextCmd.Env = readStrings(pid, regs.Rdx)
		}
		fmt.Println(pid, "execve", pstate.NextCmd)
	case syscall.SYS_UNLINK, syscall.SYS_RMDIR:
		if *flUndelete {
			regs.Orig_rax = syscall.SYS_ACCESS
			regs.Rsi = syscall.F_OK
			err := syscall.PtraceSetRegs(pid, &regs)
			e.Exit(err)
		}
	case syscall.SYS_UNLINKAT:
		if *flUndelete {
			regs.Orig_rax = syscall.SYS_FACCESSAT
			regs.R10 = regs.Rdx
			regs.Rdx = syscall.F_OK
			err := syscall.PtraceSetRegs(pid, &regs)
			e.Exit(err)
		}
	}
	return true
}

func sysexit(pid int, pstate *ProcState, sys *SysState) bool {
	regs, ok := getRegs(pid)
	if !ok {
		return false
	}
	ret := int(regs.Rax)
	if ret < 0 {
		return true
	}
	switch pstate.Syscall {
	case syscall.SYS_OPEN:
		path := pstate.Abs(readString(pid, regs.Rdi))
		flags := regs.Rsi
		write := flags&(syscall.O_WRONLY|syscall.O_RDWR) != 0
		inode := sys.FS.Inode(path)
		pstate.FDs[ret] = inode
		fmt.Println(pid, "open", write, path)
		if pstate.IOs.Map[true][inode] {
			break // Treat reads after writes as writes only.
		}
		fi, err := os.Stat(path)
		e.Exit(err)
		if fi.IsDir() {
			break // Do not record directories.
		}
		pstate.IOs.Map[write][inode] = true
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
	return true
}
