//go:build linux && amd64
// +build linux,amd64

package main

import (
	"syscall"
)

func getSysenterRegs(pstate *ProcState, pid int) (regs Regs, ok bool) {
	regs.platform, ok = getPlatformRegs(pid)
	if !ok {
		return regs, false
	}

	regs.syscall = int64(regs.platform.Orig_rax)
	regs.arg0 = regs.platform.Rdi
	regs.arg1 = regs.platform.Rsi
	regs.arg2 = regs.platform.Rdx
	regs.arg3 = regs.platform.R10
	regs.arg4 = regs.platform.R8
	regs.arg5 = regs.platform.R9

	return regs, true
}

func getSysexitRegs(pstate *ProcState, pid int) (regs Regs, ok bool) {
	regs.platform, ok = getPlatformRegs(pid)
	if !ok {
		return regs, false
	}

	regs.syscall = int64(regs.platform.Orig_rax)
	regs.arg0 = regs.platform.Rdi
	regs.arg1 = regs.platform.Rsi
	regs.arg2 = regs.platform.Rdx
	regs.arg3 = regs.platform.R10
	regs.arg4 = regs.platform.R8
	regs.arg5 = regs.platform.R9
	regs.ret = regs.platform.Rax

	return regs, true
}

func ptraceSetSysenterRegs(pstate *ProcState, pid int, regs Regs) error {
	ptraceRegs := regs.platform

	ptraceRegs.Orig_rax = uint64(regs.syscall)
	ptraceRegs.Rdi = regs.arg0
	ptraceRegs.Rsi = regs.arg1
	ptraceRegs.Rdx = regs.arg2
	ptraceRegs.R10 = regs.arg3
	ptraceRegs.R8 = regs.arg4
	ptraceRegs.R9 = regs.arg5
	ptraceRegs.Rax = regs.ret

	return syscall.PtraceSetRegs(pid, &ptraceRegs)
}
