package main

import (
	"bytes"
	"os"
	"os/exec"
	"path"
	"syscall"

	"github.com/orivej/e"
)

func getRegs(pid int) syscall.PtraceRegs {
	var regs syscall.PtraceRegs
	err := syscall.PtraceGetRegs(pid, &regs)
	e.Exit(err)
	return regs
}

func readString(pid int, addr uint64) string {
	var buf [1024]byte
	n, err := syscall.PtracePeekData(pid, uintptr(addr), buf[:])
	e.Print(err)
	res := buf[:n]
	end := bytes.IndexByte(res, 0)
	return string(res[:end])
}

func resume(pid int) {
	err := syscall.PtraceSyscall(pid, 0)
	e.Exit(err)
}

func waitForSyscall() (int, int, bool) {
	var wstatus syscall.WaitStatus
	for {
		wpid, err := syscall.Wait4(-1, &wstatus, syscall.WALL, nil)
		e.Exit(err)
		switch {
		case wstatus.Stopped() && wstatus.StopSignal()&0x80 != 0:
			return wpid, 0, true
		case wstatus.Stopped() && wstatus.TrapCause() > 0:
			return wpid, wstatus.TrapCause(), true
		case wstatus.Exited():
			return wpid, 0, false
		}
		resume(wpid)
	}
}

func trace(argv []string) (*os.Process, error) {
	tracee, err := findTracee()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(tracee, argv...) //#nosec
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	return cmd.Process, err
}

func findTracee() (string, error) {
	tracee := "tracee"
	exe, err := os.Executable()
	if err == nil {
		tracee2, err := exec.LookPath(path.Join(path.Dir(exe), tracee))
		if err == nil {
			return tracee2, nil
		}
	}
	return exec.LookPath(tracee)
}
