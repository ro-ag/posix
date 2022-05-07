//go:build darwin || linux

package posix

import (
	"syscall"
	"unsafe"
)

//goland:noinspection GoSnakeCaseUsage
const (
	EINVAL     = syscall.EINVAL
	EAGAIN     = syscall.EAGAIN
	ENOENT     = syscall.ENOENT
	EFAULT     = syscall.EFAULT
	O_RDWR     = syscall.O_RDWR     // open for reading and writing
	O_CREAT    = syscall.O_CREAT    // create if nonexistent
	O_EXCL     = syscall.O_EXCL     // error if already exists
	O_NOFOLLOW = syscall.O_NOFOLLOW // don't follow symlinks
	O_RDONLY   = syscall.O_RDONLY   // open for reading only
	O_WRONLY   = syscall.O_WRONLY   // open for writing only
	O_ACCMODE  = syscall.O_ACCMODE  // mask for modes O_RDONLY & O_WRONLY
	O_CLOEXEC  = syscall.O_CLOEXEC
)

//goland:noinspection GoSnakeCaseUsage
const (
	PROT_NONE  = syscall.PROT_NONE  // The memory cannot be accessed at all.
	PROT_READ  = syscall.PROT_READ  // The memory can be read.
	PROT_WRITE = syscall.PROT_WRITE // The memory can be modified.
	PROT_EXEC  = syscall.PROT_EXEC  // The memory can be executed.
	PROT_RDWR  = PROT_READ | PROT_WRITE
)

/*
The Mode Bits for Access Permission
The file mode, stored in the st_mode field of the file attributes, contains two kinds of information: the file type code, and the access permission bits. This section discusses only the access permission bits, which control who can read or write the file. See Testing File Type, for information about the file type code.
All the symbols listed in this section are defined in the header file sys/stat.h.
These symbolic constants are defined for the file mode bits that control access permission for the file:
*/
//goland:noinspection GoSnakeCaseUsage
const (
	S_IRUSR    = syscall.S_IRUSR  // Read permission bit for the owner of the file. On many systems this bit is 0400.
	S_IREAD    = syscall.S_IREAD  // Deprecated: S_IREAD is an Obsolete synonym provided for BSD compatibility
	S_IWUSR    = syscall.S_IWUSR  // Write permission bit for the owner of the file. Usually 0200.
	S_IWRITE   = syscall.S_IWRITE // Deprecated: S_IWRITE is an obsolete synonym provided for BSD compatibility.
	S_IXUSR    = syscall.S_IXUSR  // Execute (for ordinary files) or search (for directories) permission bit for the owner of the file. Usually 0100.
	S_IEXEC    = syscall.S_IEXEC  // Deprecated: S_IEXEC is an obsolete synonym provided for BSD compatibility.
	S_IRWXU    = syscall.S_IRWXU  // This is equivalent to ‘(S_IRUSR | S_IWUSR | S_IXUSR)’.
	S_IRGRP    = syscall.S_IRGRP  // Read permission bit for the group owner of the file. Usually 040.
	S_IWGRP    = syscall.S_IWGRP  // Write permission bit for the group owner of the file. Usually 020.
	S_IXGRP    = syscall.S_IXGRP  // Execute or search permission bit for the group owner of the file. Usually 010.
	S_IRWXG    = syscall.S_IRWXG  // This is equivalent to ‘(S_IRGRP | S_IWGRP | S_IXGRP)’.
	S_IROTH    = syscall.S_IROTH  // Read permission bit for other users. Usually 04.
	S_IWOTH    = syscall.S_IWOTH  // Write permission bit for other users. Usually 02.
	S_IXOTH    = syscall.S_IXOTH  // Execute or search permission bit for other users. Usually 01.
	S_IRWXO    = syscall.S_IRWXO  // This is equivalent to ‘(S_IROTH | S_IWOTH | S_IXOTH)’.
	S_ISUID    = syscall.S_ISUID  // This is the set-user-ID on execute bit, usually 04000. See How Change Persona.
	S_ISGID    = syscall.S_ISGID  // This is the set-group-ID on execute bit, usually 02000. See How Change Persona.
	S_ISVTX    = syscall.S_ISVTX  // This is the sticky bit, usually 01000.
	FP_SPECIAL = 1
)

//goland:noinspection GoSnakeCaseUsage
const (
	S_IFMT   = syscall.S_IFMT   // [XSI] type of file mask
	S_IFIFO  = syscall.S_IFIFO  // [XSI] named pipe (fifo)
	S_IFCHR  = syscall.S_IFCHR  // [XSI] character special
	S_IFDIR  = syscall.S_IFDIR  // [XSI] directory
	S_IFBLK  = syscall.S_IFBLK  // [XSI] block special
	S_IFREG  = syscall.S_IFREG  // [XSI] regular
	S_IFLNK  = syscall.S_IFLNK  // [XSI] symbolic link
	S_IFSOCK = syscall.S_IFSOCK // [XSI] socket
)

//goland:noinspection GoSnakeCaseUsage
const (
	MFD_ALLOW_SEALING = 0x2
	MFD_CLOEXEC       = 0x1
	MFD_HUGETLB       = 0x4
	MFD_HUGE_16GB     = -0x78000000
	MFD_HUGE_16MB     = 0x60000000
	MFD_HUGE_1GB      = 0x78000000
	MFD_HUGE_1MB      = 0x50000000
	MFD_HUGE_256MB    = 0x70000000
	MFD_HUGE_2GB      = 0x7c000000
	MFD_HUGE_2MB      = 0x54000000
	MFD_HUGE_32MB     = 0x64000000
	MFD_HUGE_512KB    = 0x4c000000
	MFD_HUGE_512MB    = 0x74000000
	MFD_HUGE_64KB     = 0x40000000
	MFD_HUGE_8MB      = 0x5c000000
	MFD_HUGE_MASK     = 0x3f
	MFD_HUGE_SHIFT    = 0x1a
	MCL_CURRENT       = syscall.MCL_CURRENT
	MCL_FUTURE        = syscall.MCL_FUTURE
	MADV_DONTNEED     = syscall.MADV_DONTNEED
	MADV_NORMAL       = syscall.MADV_NORMAL
	MADV_RANDOM       = syscall.MADV_RANDOM
	MADV_SEQUENTIAL   = syscall.MADV_SEQUENTIAL
	MADV_WILLNEED     = syscall.MADV_WILLNEED
	MAP_ANON          = syscall.MAP_ANON // Don't use a file.
	MAP_ANONYMOUS     = syscall.MAP_ANON // Don't use a file.
	MAP_FILE          = syscall.MAP_FILE
	MAP_FIXED         = syscall.MAP_FIXED // Interpret address exactly.
	MAP_NORESERVE     = syscall.MAP_NORESERVE
	MAP_PRIVATE       = syscall.MAP_PRIVATE // Changes are private.
	MAP_SHARED        = syscall.MAP_SHARED  // Share changes.
	MS_ASYNC          = syscall.MS_ASYNC    // Perform synchronous page faults for the mapping.
	MS_INVALIDATE     = syscall.MS_INVALIDATE
	MS_SYNC           = syscall.MS_SYNC
	NAME_MAX          = syscall.NAME_MAX
)

// File Fcntl
//goland:noinspection GoSnakeCaseUsage
const (
	F_GETFD    = syscall.F_GETFD
	F_SETFD    = syscall.F_SETFD
	FD_CLOEXEC = syscall.FD_CLOEXEC
)

var (
	errEAGAIN error = syscall.EAGAIN
	errEINVAL error = syscall.EINVAL
	errENOENT error = syscall.ENOENT
)

func errnoErr(e syscall.Errno) error {
	switch e {
	case 0:
		return nil
	case EAGAIN:
		return errEAGAIN
	case EINVAL:
		return errEINVAL
	case ENOENT:
		return errENOENT
	}
	return e
}

type Errno = syscall.Errno

// _unsafeSlice is the runtime representation of a slice.
type _unsafeSlice struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}
