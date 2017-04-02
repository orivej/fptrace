package main

import "path"

type ProcState struct {
	SysEnter bool           // true on enter to syscall
	Syscall  int            // call number on exit from syscall
	CurDir   string         // working directory
	ExecPath string         // path in the last call to execve
	FDs      map[int]string // map fds to abspaths
}

func NewProcState() *ProcState {
	return &ProcState{FDs: make(map[int]string)}
}

func (ps *ProcState) Abs(p string) string {
	if path.IsAbs(p) {
		return p
	}
	return path.Join(ps.CurDir, p)
}

func (ps *ProcState) Clone() *ProcState {
	newps := NewProcState()
	newps.CurDir = ps.CurDir
	for n, s := range ps.FDs {
		newps.FDs[n] = s
	}
	return newps
}
