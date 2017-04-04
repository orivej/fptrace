package main

type SysState struct {
	FS *FS
}

type FS struct {
	seq       int
	inodePath map[int]string
	pathInode map[string]int
}

func NewSysState() *SysState {
	return &SysState{FS: NewFS()}
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
