package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"os/exec"
	"syscall"

	"github.com/orivej/e"
	"golang.org/x/sys/unix"
)

var errTraceeExited = errors.New("tracee failed to start")

func getRegs(pid int) (syscall.PtraceRegs, bool) {
	var regs syscall.PtraceRegs
	err := syscall.PtraceGetRegs(pid, &regs)
	if err != nil && err.(syscall.Errno) == syscall.ESRCH {
		return regs, false
	}
	e.Exit(err)
	return regs, true
}

func readString(pid int, addr uint64) string {
	var chunk [64]byte
	var buf []byte
	for {
		n, err := syscall.PtracePeekData(pid, uintptr(addr), chunk[:])
		if err != syscall.EIO {
			e.Print(err)
		}
		end := bytes.IndexByte(chunk[:n], 0)
		if end != -1 {
			buf = append(buf, chunk[:end]...)
			return string(buf)
		}
		buf = append(buf, chunk[:n]...)
		addr += uint64(n)
	}
}

func readStrings(pid int, addr uint64) []string {
	var buf [8]byte
	var res []string
	for {
		n, err := syscall.PtracePeekData(pid, uintptr(addr), buf[:])
		e.Exit(err)
		saddr := binary.LittleEndian.Uint64(buf[:])
		if saddr == 0 {
			return res
		}
		res = append(res, readString(pid, saddr))
		addr += uint64(n)
	}
}

func resume(pid, signal int, duringSyscall bool) {
	if duringSyscall || !withSeccomp {
		err := syscall.PtraceSyscall(pid, signal)
		e.Print(err)
	} else {
		err := syscall.PtraceCont(pid, signal)
		e.Print(err)
	}
}

func waitForSyscall() (pid, trapcause int, alive bool) {
	var wstatus syscall.WaitStatus
	for {
		pid, err := syscall.Wait4(-1, &wstatus, syscall.WALL, nil)
		e.Exit(err)
		switch {
		case wstatus.Exited(): // Normal exit.
			return pid, wstatus.ExitStatus(), false
		case wstatus.Signaled(): // Signal exit.
			return pid, -int(wstatus.Signal()), false
		case wstatus.StopSignal()&0x80 != 0: // Ptrace stop.
			return pid, 0, true
		case wstatus.TrapCause() > 0: // SIGTRAP stop.
			return pid, wstatus.TrapCause(), true
		default: // Another signal stop (e.g. SIGSTOP).
			resume(pid, int(wstatus.StopSignal()), false)
		}
	}
}

func trace(tracee string, argv []string) (int, error) {
	argv = append([]string{"--"}, argv...)
	if withSeccomp {
		argv = append([]string{"-seccomp"}, argv...)
	}
	cmd := exec.Command(tracee, argv...) //#nosec
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return 0, err
	}

	pid := cmd.Process.Pid
	var wstatus syscall.WaitStatus
	_, err = syscall.Wait4(pid, &wstatus, 0, nil)
	if err != nil {
		return 0, err
	}
	if wstatus.Exited() {
		return 0, errTraceeExited
	}

	err = syscall.PtraceSetOptions(pid, 0|
		unix.PTRACE_O_EXITKILL|
		unix.PTRACE_O_TRACESECCOMP|
		syscall.PTRACE_O_TRACESYSGOOD|
		syscall.PTRACE_O_TRACEEXEC|
		syscall.PTRACE_O_TRACECLONE|
		syscall.PTRACE_O_TRACEFORK|
		syscall.PTRACE_O_TRACEVFORK)
	if err != nil {
		return 0, err
	}

	resume(pid, 0, false)
	return pid, nil
}
