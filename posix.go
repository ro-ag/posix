package posix

import (
	"unsafe"
)

//ShmOpen
//Create and open a new object, or open an existing object.
//This is analogous to open.  The call returns a file
//descriptor for use by the other interfaces listed below.
func ShmOpen(name string, oflag int, mode uint32) (fd int, err error) {
	return shmOpen(name, oflag, mode)
}

//Ftruncate
//Set the size of the shared memory object.  (A newly
//created shared memory object has a length of zero.)
func Ftruncate(fd int, length int) error {
	return ftruncate(fd, length)
}

//Madvise
//This system call is used to give advice or directions to
//the kernel about the address range beginning at address addr and
//with size length bytes In most cases, the goal of such advice is
//to improve system or application performance.
func Madvise(b []byte, behav int) error {
	return madvise(b, behav)
}

//MmapAt
//Map the shared memory object into the virtual address
//space of the calling process.
//MmapAt has the same parameters as the C mmap implementation where the address is exposed.
func MmapAt(address unsafe.Pointer, length uintptr, prot accFlags, flags int, fd int, offset int64) (data []byte, add uintptr, err error) {
	return mapper.Mmap(address, length, int(prot), flags, fd, offset)
}

//Munmap
//Unmap the shared memory object from the virtual address
//space of the calling process.
func Munmap(b []byte) error {
	return mapper.Munmap(b)
}

//Mprotect
//changes the access protections for the calling
//process's memory pages containing any part of the address range
//in the interval [addr, addr+len-1].  addr must be aligned to a
//page boundary.
func Mprotect(b []byte, prot accFlags) error {
	return mprotect(b, int(prot))
}

//Mlock
//lock part of the calling process's virtual address space into RAM,
//preventing that memory from being paged to the swap area.
func Mlock(b []byte, size int) error {
	return mlock(b, size)
}

//Munlock
//perform the converse operation,
//unlocking part of the calling process's virtual address
//space, so that pages in the specified virtual address range may
//once more to be swapped out if required by the kernel memory
//manager.
func Munlock(b []byte, size int) error {
	return munlock(b, size)
}

//Mlockall
//lock all calling process's virtual address space into RAM,
//preventing that memory from being paged to the swap area.
func Mlockall(flags int) error {
	return mlockall(flags)
}

//Munlockall
//perform the converse operation,
//unlocking all calling process's virtual address
//space, so that pages in the specified virtual address range may
//once more to be swapped out if required by the kernel memory
//manager.
func Munlockall() error {
	return munlockall()
}

//Msync flushes changes made to the in-core copy of a file that
//was mapped into memory using mmap(2) back to the filesystem.
//Without use of this call, there is no guarantee that changes are
//written back before munmap(2) is called.  To be more precise, the
//part of the file that corresponds to the memory area starting at
//addr and having length 'length' is updated.
func Msync(b []byte, flags int) error {
	return msync(b, flags)
}

//ShmUnlink
//Remove a shared memory object shmName.
func ShmUnlink(path string) (err error) {
	return shmUnlink(path)
}

//Close
//the file descriptor allocated by shm_open(3) when it
//is no longer needed.
func Close(fd int) error {
	return closeFd(fd)
}

//Fstat
//Obtain a stat structure that describes the shared memory
//object.  Among the information returned by this call are
//the object's size (st_size), permissions (st_mode), owner
//(st_uid), and group (st_gid).
func Fstat(fd int, stat *Stat_t) error {
	return fstat(fd, stat)
}

//Fchown
//To change the ownership of a shared memory object.
func Fchown(fd int, uid int, gid int) error {
	return fchown(fd, uid, gid)
}

//Fchmod
//To change the permissions of a shared memory object.
func Fchmod(fd int, mode int) error {
	return fchmod(fd, mode)
}

//Fcntl performs one of the operations described below on the
//open file descriptor fd.  The operation is determined by cmd
func Fcntl(fd int, cmd int, arg int) (val int, err error) {
	return fcntl(fd, cmd, arg)
}

// Single-word zero for use when we need a valid pointer to 0 bytes.
var _zero uintptr
