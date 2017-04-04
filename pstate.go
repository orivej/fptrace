package main

import "path"

type IOs struct {
	Cnt int          // IOs referenc count
	Map map[int]bool // inodes; inputs are false, outputs are true
}

type Cmd struct {
	Path string
	Args []string
	Dir  string
}

type ProcState struct {
	SysEnter bool        // true on enter to syscall
	Syscall  int         // call number on exit from syscall
	CurDir   string      // working directory
	CurCmd   Cmd         // current command
	NextCmd  Cmd         // command after return from execve
	FDs      map[int]int // map fds to inodes

	IOs *IOs
}

type Record struct {
	Cmd     Cmd
	Inputs  []string
	Outputs []string
}

func NewProcState() *ProcState {
	return &ProcState{
		FDs: make(map[int]int),
		IOs: &IOs{1, make(map[int]bool)},
	}
}

func (ps *ProcState) ResetIOs() {
	ps.IOs.Cnt--
	ps.IOs = &IOs{1, make(map[int]bool)}
}

func (ps *ProcState) Abs(p string) string {
	if path.IsAbs(p) {
		return p
	}
	return path.Join(ps.CurDir, p)
}

func (ps *ProcState) Clone() *ProcState {
	newps := NewProcState()
	newps.IOs = ps.IOs // IOs are shared until exec
	ps.IOs.Cnt++
	newps.CurDir = ps.CurDir
	newps.CurCmd = ps.CurCmd
	for n, s := range ps.FDs {
		newps.FDs[n] = s
	}
	return newps
}

func (ps *ProcState) Record(sys *SysState) Record {
	r := Record{Cmd: ps.CurCmd}
	paths := map[string]bool{}
	// Deduplicate paths after renames.
	for inode, output := range ps.IOs.Map {
		s := sys.FS.Path(inode)
		paths[s] = paths[s] || output
	}
	for s, output := range paths {
		if output {
			r.Outputs = append(r.Outputs, s)
		} else {
			r.Inputs = append(r.Inputs, s)
		}
	}
	return r
}
