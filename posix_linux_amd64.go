package posix

import "syscall"

// MAP_32BIT requests a mapping in the first 2 GB of the address space. It is
// x86-64 only and is intentionally not defined on linux/arm64.
//
//goland:noinspection GoSnakeCaseUsage
const MAP_32BIT = syscall.MAP_32BIT

// Stat_t mirrors the linux/amd64 kernel struct stat returned by the fstat
// syscall.
//
//goland:noinspection GoSnakeCaseUsage
type Stat_t struct {
	Dev     DevT
	Ino     uint64
	Nlink   uint64
	Mode    ModeT
	Uid     uint32
	Gid     uint32
	_       int32
	Rdev    DevT
	Size    int64
	Blksize int64
	Blocks  int64
	Atim    Timespec
	Mtim    Timespec
	Ctim    Timespec
	_       [3]int64
}

// Linux/amd64 syscall numbers.
//
//goland:noinspection GoSnakeCaseUsage
const (
	_SYS_FTRUNCATE    = 77
	_SYS_MEMFD_CREATE = 319
	_SYS_MADVISE      = 28
	_SYS_MMAP         = 9
	_SYS_MUNMAP       = 11
	_SYS_MPROTECT     = 10
	_SYS_MLOCK        = 149
	_SYS_MUNLOCK      = 150
	_SYS_MLOCKALL     = 151
	_SYS_MUNLOCKALL   = 152
	_SYS_MSYNC        = 26
	_SYS_CLOSE        = 3
	_SYS_FCHOWN       = 93
	_SYS_FSTAT        = 5
	_SYS_FCHMOD       = 91
	_SYS_FCNTL        = 72
	_SYS_OPENAT       = 257
	_SYS_UNLINKAT     = 263
)
