package main

import "fmt"

type SysState struct {
	FS   *FS
	Proc *Proc
}

type FS struct {
	seq       int
	pipe      int
	inodePath map[int]string
	pathInode map[string]int
}

type Proc struct {
	lastID int
}

func NewSysState() *SysState {
	return &SysState{FS: NewFS(), Proc: NewProc()}
}

func NewFS() *FS {
	return &FS{
		inodePath: map[int]string{},
		pathInode: map[string]int{},
	}
}

func (fs *FS) Inode(path string) int {
	if inode, ok := fs.pathInode[path]; ok {
		return inode
	}
	fs.seq++
	fs.inodePath[fs.seq] = path
	fs.pathInode[path] = fs.seq
	return fs.seq
}

func (fs *FS) Path(inode int) string {
	return fs.inodePath[inode]
}

func (fs *FS) Pipe() int {
	fs.pipe++
	return fs.Inode(fmt.Sprint("/dev/fptrace/pipe/", fs.pipe))
}

func (fs *FS) Rename(oldpath, newpath string) {
	if oldpath == newpath {
		return
	}
	oldInode := fs.pathInode[oldpath]
	delete(fs.pathInode, oldpath)
	fs.pathInode[newpath] = oldInode
	fs.inodePath[oldInode] = newpath
}

func NewProc() *Proc {
	return &Proc{}
}

func (p *Proc) Exec(ps *ProcState) {
	cmd := ps.NextCmd
	if ps.CurCmd != nil {
		cmd.Parent = ps.CurCmd.ID
	}
	p.lastID++
	cmd.ID = p.lastID
	ps.CurCmd = &cmd

	for n, b := range ps.FDCX {
		if b {
			delete(ps.FDs, n)
		}
	}
	ps.FDCX = make(map[int32]bool)
}
