//go:build linux && arm64
// +build linux,arm64

package syscalls

import "golang.org/x/sys/unix"

const (
	// specifies max syscall id that doesn't exist in this cpu architecture, but does in other ones
	NOT_ON_THIS_ARCH_MAX = -1

	SYS_CHDIR     = unix.SYS_CHDIR
	SYS_CLOSE     = unix.SYS_CLOSE
	SYS_DUP       = unix.SYS_DUP
	SYS_DUP2      = NOT_ON_THIS_ARCH_MAX - iota
	SYS_DUP3      = unix.SYS_DUP3
	SYS_EXECVE    = unix.SYS_EXECVE
	SYS_EXECVEAT  = unix.SYS_EXECVEAT
	SYS_FCHDIR    = unix.SYS_FCHDIR
	SYS_FCNTL     = unix.SYS_FCNTL
	SYS_LINK      = NOT_ON_THIS_ARCH_MAX - iota
	SYS_LINKAT    = unix.SYS_LINKAT
	SYS_OPEN      = NOT_ON_THIS_ARCH_MAX - iota
	SYS_OPENAT    = unix.SYS_OPENAT
	SYS_PIPE      = NOT_ON_THIS_ARCH_MAX - iota
	SYS_PIPE2     = unix.SYS_PIPE2
	SYS_PREAD64   = unix.SYS_PREAD64
	SYS_PREADV    = unix.SYS_PREADV
	SYS_PREADV2   = unix.SYS_PREADV2
	SYS_PWRITE64  = unix.SYS_PWRITE64
	SYS_PWRITEV   = unix.SYS_PWRITEV
	SYS_PWRITEV2  = unix.SYS_PWRITEV2
	SYS_READ      = unix.SYS_READ
	SYS_READV     = unix.SYS_READV
	SYS_RENAME    = NOT_ON_THIS_ARCH_MAX - iota
	SYS_RENAMEAT  = unix.SYS_RENAMEAT
	SYS_RENAMEAT2 = unix.SYS_RENAMEAT2
	SYS_RMDIR     = NOT_ON_THIS_ARCH_MAX - iota
	SYS_UNLINK    = NOT_ON_THIS_ARCH_MAX - iota
	SYS_UNLINKAT  = unix.SYS_UNLINKAT
	SYS_WRITE     = unix.SYS_WRITE
	SYS_WRITEV    = unix.SYS_WRITEV

	// syscalls fptrace refers to but doesn't trace on
	SYS_ACCESS    = NOT_ON_THIS_ARCH_MAX - iota
	SYS_FACCESSAT = unix.SYS_FACCESSAT
)
