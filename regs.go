package main

import "syscall"

type Regs struct {
	syscall int64
	arg0    uint64
	arg1    uint64
	arg2    uint64
	arg3    uint64
	arg4    uint64
	arg5    uint64
	ret     uint64

	platform syscall.PtraceRegs
}

func getPlatformRegs(pid int) (syscall.PtraceRegs, bool) {
	var regs syscall.PtraceRegs
	err := syscall.PtraceGetRegs(pid, &regs)
	if err != nil && err.(syscall.Errno) == syscall.ESRCH {
		return regs, false
	}
	return regs, true
}
