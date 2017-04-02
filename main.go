package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/orivej/e"
)

func main() {
	flOut := flag.String("o", "", "output file")
	flag.Parse()
	log.SetFlags(0)

	args := flag.Args()
	proc, err := trace(args)
	e.Exit(err)
	pid := proc.Pid
	_, err = syscall.Wait4(pid, nil, 0, nil)
	e.Exit(err)

	if *flOut != "" {
		f, err2 := os.Create(*flOut)
		e.Exit(err2)
		os.Stdout = f
	}

	err = syscall.PtraceSetOptions(pid, syscall.PTRACE_O_TRACESYSGOOD|
		syscall.PTRACE_O_TRACEEXEC|
		syscall.PTRACE_O_TRACECLONE|
		syscall.PTRACE_O_TRACEFORK|
		syscall.PTRACE_O_TRACEVFORK)
	e.Exit(err)

	resume(pid)
	mainLoop(pid)
}

func mainLoop(mainPID int) {
	var err error
	pstates := map[int]*ProcState{}
	pstates[mainPID] = NewProcState()
	pstates[mainPID].CurDir, err = os.Getwd()
	e.Exit(err)

	var regs syscall.PtraceRegs
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

		err := syscall.PtraceGetRegs(pid, &regs)
		e.Exit(err)
		ret := int(regs.Rax)

		// New proc.
		switch wstatus {
		case syscall.PTRACE_EVENT_FORK,
			syscall.PTRACE_EVENT_VFORK,
			syscall.PTRACE_EVENT_VFORK_DONE,
			syscall.PTRACE_EVENT_CLONE:

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
			goto resume
		default:
			// PTRACE_EVENT_EXEC
			goto resume
		case 0:
		}

		// Toggle edge.
		pstate.SysEnter = !pstate.SysEnter

		// SysEnter.
		if pstate.SysEnter {
			pstate.Syscall = int(regs.Orig_rax)
			switch pstate.Syscall {
			case syscall.SYS_EXECVE:
				if regs.Rdi != 0 {
					pstate.ExecPath = readString(pid, regs.Rdi)
				}
			}
			goto resume
		}

		// SysExit.
		if ret < 0 {
			goto resume
		}

		switch pstate.Syscall {
		case syscall.SYS_OPEN:
			path := pstate.Abs(readString(pid, regs.Rdi))
			pstate.FDs[ret] = path
			flags := regs.Rsi
			write := flags&(syscall.O_WRONLY|syscall.O_RDWR) != 0
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
			fmt.Println(pid, "execve", pstate.ExecPath)
		}
	resume:
		resume(pid)
	}
}
