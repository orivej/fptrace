package main

type SysState struct {
	FS   *FS
	Proc *Proc
}

type FS struct {
	seq       int
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
	ps.NextCmd.Parent = ps.CurCmd.ID
	p.lastID++
	ps.NextCmd.ID = p.lastID
	ps.CurCmd = ps.NextCmd
}