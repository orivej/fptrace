//go:build linux && arm64
// +build linux,arm64

package main

import (
	"syscall"
)

func getSysenterRegs(pstate *ProcState, pid int) (regs Regs, ok bool) {
	regs.platform, ok = getPlatformRegs(pid)
	if !ok {
		return regs, false
	}

	regs.syscall = int64(regs.platform.Regs[8])
	regs.arg0 = regs.platform.Regs[0]
	regs.arg1 = regs.platform.Regs[1]
	regs.arg2 = regs.platform.Regs[2]
	regs.arg3 = regs.platform.Regs[3]
	regs.arg4 = regs.platform.Regs[4]
	regs.arg5 = regs.platform.Regs[5]

	// on arm64 Linux overrides x0 with the sys return value, and doesn't provide
	// a way to retrieve the original value like with orig_rax on x64
	pstate.Arg0 = regs.arg0

	return regs, true
}

func getSysexitRegs(pstate *ProcState, pid int) (regs Regs, ok bool) {
	regs.platform, ok = getPlatformRegs(pid)
	if !ok {
		return regs, false
	}

	regs.syscall = int64(regs.platform.Regs[8])
	regs.arg0 = pstate.Arg0
	regs.arg1 = regs.platform.Regs[1]
	regs.arg2 = regs.platform.Regs[2]
	regs.arg3 = regs.platform.Regs[3]
	regs.arg4 = regs.platform.Regs[4]
	regs.arg5 = regs.platform.Regs[5]
	regs.ret = regs.platform.Regs[0]

	return regs, true
}

func ptraceSetSysenterRegs(pstate *ProcState, pid int, regs Regs) error {
	pstate.Arg0 = regs.arg0

	ptraceRegs := regs.platform
	ptraceRegs.Regs[8] = uint64(regs.syscall)
	ptraceRegs.Regs[0] = regs.arg0
	ptraceRegs.Regs[1] = regs.arg1
	ptraceRegs.Regs[2] = regs.arg2
	ptraceRegs.Regs[3] = regs.arg3
	ptraceRegs.Regs[4] = regs.arg4
	ptraceRegs.Regs[5] = regs.arg5

	return syscall.PtraceSetRegs(pid, &ptraceRegs)
}
