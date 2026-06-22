package posix

// Stat_t mirrors the linux/arm64 kernel struct stat. arm64 uses the asm-generic
// layout, which differs from amd64: Mode precedes Nlink (both uint32) and the
// padding around Rdev/Blksize is different.
//
//goland:noinspection GoSnakeCaseUsage
type Stat_t struct {
	Dev     DevT
	Ino     uint64
	Mode    ModeT
	Nlink   uint32
	Uid     uint32
	Gid     uint32
	Rdev    DevT
	_       uint64
	Size    int64
	Blksize int32
	_       int32
	Blocks  int64
	Atim    Timespec
	Mtim    Timespec
	Ctim    Timespec
	_       [2]int32
}

// Linux/arm64 syscall numbers (the asm-generic table; all differ from amd64).
//
//goland:noinspection GoSnakeCaseUsage
const (
	_SYS_FCNTL        = 25
	_SYS_UNLINKAT     = 35
	_SYS_FTRUNCATE    = 46
	_SYS_FCHMOD       = 52
	_SYS_FCHOWN       = 55
	_SYS_OPENAT       = 56
	_SYS_CLOSE        = 57
	_SYS_FSTAT        = 80
	_SYS_MUNMAP       = 215
	_SYS_MMAP         = 222
	_SYS_MPROTECT     = 226
	_SYS_MSYNC        = 227
	_SYS_MLOCK        = 228
	_SYS_MUNLOCK      = 229
	_SYS_MLOCKALL     = 230
	_SYS_MUNLOCKALL   = 231
	_SYS_MADVISE      = 233
	_SYS_MEMFD_CREATE = 279
)
