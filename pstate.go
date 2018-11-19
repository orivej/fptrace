package main

import (
	"fmt"
	"path"
)

type IOs struct {
	Cnt int // IOs reference count

	Map [2]IntSliceSet // input(false)/output(true) inodes
}

type Cmd struct {
	Parent int // parent Cmd ID
	ID     int // Cmd ID, changes only with execve
	Exit   int `json:",omitempty"` // Exit code of the first process (or 0-signal),
	Exec   int `json:",omitempty"` // or ID of the Cmd executed by the first process

	Dir  string
	Path string
	Args []string
	Env  []string `json:",omitempty"`
}

type ProcState struct {
	SysEnter bool           // true on enter to syscall
	Syscall  int            // call number on exit from syscall
	CurDir   string         // working directory
	CurCmd   *Cmd           // current command
	NextCmd  Cmd            // command after return from execve
	FDs      map[int32]int  // map fds to inodes
	FDCX     map[int32]bool // cloexec fds

	IOs *IOs
}

type Record struct {
	Cmd     Cmd
	Inputs  []string
	Outputs []string
	FDs     map[int32]string
}

func NewIOs() *IOs {
	return &IOs{1, [2]IntSliceSet{
		NewIntSliceSet(),
		NewIntSliceSet(),
	}}
}

func NewProcState() *ProcState {
	return &ProcState{
		FDs:  make(map[int32]int),
		FDCX: make(map[int32]bool),
		IOs:  NewIOs(),
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
		if !path.IsAbs(dir) {
			panic(fmt.Sprintf("dir is not absolute: %q", dir))
		}
		p = path.Join(dir, p)
	}
	return path.Clean(p)
}

func (ps *ProcState) Clone(cloneFiles bool) *ProcState {
	newps := NewProcState()
	newps.IOs = ps.IOs // IOs are shared until exec
	ps.IOs.Cnt++
	newps.CurDir = ps.CurDir
	newps.CurCmd = ps.CurCmd
	if cloneFiles {
		newps.FDs = ps.FDs
		newps.FDCX = ps.FDCX
	} else {
		for n, s := range ps.FDs {
			newps.FDs[n] = s
		}
		for n, b := range ps.FDCX {
			if b {
				newps.FDCX[n] = b
			}
		}
	}
	return newps
}

func (ps *ProcState) Record(sys *SysState) Record {
	r := Record{Cmd: *ps.CurCmd, Inputs: []string{}, Outputs: []string{}}
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
