package main

import "path"

type IOs struct {
	Cnt int // IOs reference count

	Map [2]IntSliceSet // input(false)/output(true) inodes
}

type Cmd struct {
	Parent int // parent Cmd ID
	ID     int // Cmd ID, changes only with execve

	Dir  string
	Path string
	Args []string
	Env  []string `json:",omitempty"`
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
	FDs     map[int]string
}

func NewIOs() *IOs {
	return &IOs{1, [2]IntSliceSet{
		NewIntSliceSet(),
		NewIntSliceSet(),
	}}
}

func NewProcState() *ProcState {
	return &ProcState{
		FDs: make(map[int]int),
		IOs: NewIOs(),
	}
}

func (ps *ProcState) ResetIOs() {
	ps.IOs.Cnt--
	ps.IOs = NewIOs()
}

func (ps *ProcState) Abs(p string) string {
	return ps.AbsAt(ps.CurDir, p)
}

func (ps *ProcState) AbsAt(dir, p string) string {
	if !path.IsAbs(p) {
		p = path.Join(dir, p)
	}
	return path.Clean(p)
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
	r := Record{Cmd: ps.CurCmd, Inputs: []string{}, Outputs: []string{}}
	for output, inodes := range ps.IOs.Map {
		paths := &r.Inputs
		if output == W {
			paths = &r.Outputs
		}
		for _, inode := range inodes.Slice {
			s := sys.FS.Path(inode)
			*paths = append(*paths, s)
		}
	}
	return r
}
