//go:build darwin || linux

// Package posix is a golang implementation for mmap, shm_open for MacOSx and Linux without CGO !
// MacOSx has an emulation for memfd_create using shm_open.
package posix

import (
	"syscall"
	"unsafe"
)

// ShmOpen creates and opens a new shared-memory object, or opens an existing
// one. It is analogous to open(2) and returns a file descriptor for use by the
// other calls here.
//
// When creating (O_CREAT), mode sets the object's permission bits, and those
// bits are the access control: another process may ShmOpen the same name only
// if they allow its uid/gid (0600 keeps it private to the owner, 0660 shares it
// with a group, and so on). On macOS those permissions are fixed at creation —
// Fchmod and Fchown do not work on shared memory there — so set mode correctly
// up front; on Linux it can be changed afterward.
func ShmOpen(name string, oflag int, mode uint32) (fd int, err error) {
	return shmOpen(name, oflag, mode)
}

// Ftruncate sets the size of the shared-memory object. (A newly created object
// has length zero.) When a size seal is set, the matching change is rejected:
// F_SEAL_SHRINK blocks shrinking and F_SEAL_GROW blocks growing.
//
// On macOS a shm object's size can be set only once; a later Ftruncate returns
// EINVAL, and the size is rounded up to a page.
func Ftruncate(fd int, length int) error {
	if err := sealCheckTruncate(fd, length); err != nil {
		return err
	}
	return ftruncate(fd, length)
}

// Madvise
// This system call is used to give advice or directions to
// the kernel about the address range beginning at address addr and
// with size length bytes In most cases, the goal of such advice is
// to improve system or application performance.
func Madvise(b []byte, behav int) error {
	return madvise(b, behav)
}

// Mmap maps length bytes of the object referred to by fd (or anonymous memory)
// into the calling process's address space, starting at the given offset.
//
// Unlike syscall.Mmap and golang.org/x/sys, the destination address is exposed:
// pass it as a hint, or add MAP_FIXED to place the mapping at exactly that
// address. Mmap returns the mapped slice and the address the kernel chose.
//
// The caller must release the mapping with Munmap; it is not garbage-collected.
func Mmap(address unsafe.Pointer, length int, prot int, flags int, fd int, offset int64) (data []byte, add uintptr, err error) {
	if length <= 0 {
		return nil, 0, EINVAL
	}
	if err := sealCheckMmap(fd, prot, flags); err != nil {
		return nil, 0, err
	}
	return mapper.Mmap(address, uintptr(length), prot, flags, fd, offset)
}

// Munmap
// Unmap the shared memory object from the virtual address
// space of the calling process.
func Munmap(b []byte) error {
	return mapper.Munmap(b)
}

// Mprotect
// changes the access protections for the calling
// process's memory pages containing any part of the address range
// in the interval [addr, addr+len-1].  addr must be aligned to a
// page boundary.
func Mprotect(b []byte, prot int) error {
	return mprotect(b, prot)
}

// Mlock
// lock part of the calling process's virtual address space into RAM,
// preventing that memory from being paged to the swap area.
func Mlock(b []byte, size int) error {
	return mlock(b, size)
}

// Munlock
// perform the converse operation,
// unlocking part of the calling process's virtual address
// space, so that pages in the specified virtual address range may
// once more to be swapped out if required by the kernel memory
// manager.
func Munlock(b []byte, size int) error {
	return munlock(b, size)
}

// Mlockall
// lock all calling process's virtual address space into RAM,
// preventing that memory from being paged to the swap area.
func Mlockall(flags int) error {
	return mlockall(flags)
}

// Munlockall
// perform the converse operation,
// unlocking all calling process's virtual address
// space, so that pages in the specified virtual address range may
// once more to be swapped out if required by the kernel memory
// manager.
func Munlockall() error {
	return munlockall()
}

// Msync flushes changes made to the in-core copy of a file that
// was mapped into memory using mmap(2) back to the filesystem.
// Without use of this call, there is no guarantee that changes are
// written back before munmap(2) is called.  To be more precise, the
// part of the file that corresponds to the memory area starting at
// addr and having length 'length' is updated.
func Msync(b []byte, flags int) error {
	return msync(b, flags)
}

// ShmUnlink
// Remove a shared memory object shmName.
func ShmUnlink(path string) (err error) {
	return shmUnlink(path)
}

// Close
// the file descriptor allocated by shm_open(3) when it
// is no longer needed.
func Close(fd int) error {
	err := closeFd(fd)
	sealForget(fd)
	return err
}

// Fstat
// Obtain a stat structure that describes the shared memory
// object.  Among the information returned by this call are
// the object's size (st_size), permissions (st_mode), owner
// (st_uid), and group (st_gid).
func Fstat(fd int, stat *Stat_t) error {
	return fstat(fd, stat)
}

// Fchown changes the ownership of a shared-memory object. It works on Linux but
// returns EINVAL on macOS, where shm ownership is fixed at creation (the object
// is owned by the process that created it).
func Fchown(fd int, uid int, gid int) error {
	return fchown(fd, uid, gid)
}

// Fchmod changes the permission bits of a shared-memory object. It works on
// Linux but returns EINVAL on macOS, where shm permissions are fixed at
// creation; pass the desired mode to ShmOpen instead.
func Fchmod(fd int, mode int) error {
	return fchmod(fd, mode)
}

// Fcntl performs one of the operations described below on the
// open file descriptor fd.  The operation is determined by cmd
func Fcntl(fd int, cmd int, arg int) (val int, err error) {
	return fcntl(fd, cmd, arg)
}

// Getpagesize
// The function returns the number of bytes in a memory page,
// where "page" is a fixed-length block, the unit for
// memory allocation and file mapping performed by mmap
func Getpagesize() int {
	return syscall.Getpagesize()
}

// MemfdCreate creates an anonymous, in-memory file and returns a descriptor for
// it. The descriptor can be Ftruncate'd, Mmap'd, and inherited by a child
// process; the object is released once all descriptors are closed.
//
// On Linux this is the native memfd_create syscall. macOS has no memfd_create,
// so it is emulated: a uniquely named shm_open object is created and its name is
// unlinked immediately, leaving an anonymous descriptor. The emulation covers
// the create/size/map/share path, but differs from a Linux memfd in two ways —
// the object size is rounded up to a page, and kernel sealing (MFD_ALLOW_SEALING)
// has no effect. For macOS-native code, ShmAnonymous is the direct equivalent.
func MemfdCreate(name string, flags int) (fd int, err error) {
	return memfdCreate(name, flags)
}

// Single-word zero for use when we need a valid pointer to 0 bytes.
var _zero uintptr
