package posix

import "unsafe"

type Timespec struct {
	Sec  int64
	Nsec int64
}

type Stat_t struct {
	Dev     int32
	Mode    uint16
	Nlink   uint16
	Ino     uint64
	Uid     uint32
	Gid     uint32
	Rdev    int32
	Atim    Timespec
	Mtim    Timespec
	Ctim    Timespec
	Btim    Timespec
	Size    int64
	Blocks  int64
	Blksize int32
	Flags   uint32
	Gen     uint32
	Lspare  int32
	Qspare  [2]int64
}

/* -------------------------------------------------------------------------------------------------------------------*/
func madvise(b []byte, behav int) (err error) {
	var _p0 unsafe.Pointer
	if len(b) > 0 {
		_p0 = unsafe.Pointer(&b[0])
	} else {
		_p0 = unsafe.Pointer(&_zero)
	}
	_, _, e1 := syscall_syscall(libc_madvise_trampoline_addr, uintptr(_p0), uintptr(len(b)), uintptr(behav))
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

var libc_madvise_trampoline_addr uintptr

//go:cgo_import_dynamic libc_madvise madvise "/usr/lib/libSystem.B.dylib"

/* -------------------------------------------------------------------------------------------------------------------*/

func mlock(b []byte, size int) (err error) {
	var _p0 unsafe.Pointer
	if len(b) > 0 {
		_p0 = unsafe.Pointer(&b[0])
	} else {
		_p0 = unsafe.Pointer(&_zero)
	}
	_, _, e1 := syscall_syscall(libc_mlock_trampoline_addr, uintptr(_p0), uintptr(size), 0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

var libc_mlock_trampoline_addr uintptr

//go:cgo_import_dynamic libc_mlock mlock "/usr/lib/libSystem.B.dylib"

/* -------------------------------------------------------------------------------------------------------------------*/

func mlockall(flags int) (err error) {
	_, _, e1 := syscall_syscall(libc_mlockall_trampoline_addr, uintptr(flags), 0, 0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

var libc_mlockall_trampoline_addr uintptr

//go:cgo_import_dynamic libc_mlockall mlockall "/usr/lib/libSystem.B.dylib"

/* -------------------------------------------------------------------------------------------------------------------*/

func munlock(b []byte, size int) (err error) {
	var _p0 unsafe.Pointer
	if len(b) > 0 {
		_p0 = unsafe.Pointer(&b[0])
	} else {
		_p0 = unsafe.Pointer(&_zero)
	}
	_, _, e1 := syscall_syscall(libc_munlock_trampoline_addr, uintptr(_p0), uintptr(size), 0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

var libc_munlock_trampoline_addr uintptr

//go:cgo_import_dynamic libc_munlock munlock "/usr/lib/libSystem.B.dylib"

/* -------------------------------------------------------------------------------------------------------------------*/

func munlockall() (err error) {
	_, _, e1 := syscall_syscall(libc_munlockall_trampoline_addr, 0, 0, 0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

var libc_munlockall_trampoline_addr uintptr

//go:cgo_import_dynamic libc_munlockall munlockall "/usr/lib/libSystem.B.dylib"

/* -------------------------------------------------------------------------------------------------------------------*/
func mprotect(b []byte, prot int) (err error) {
	var _p0 unsafe.Pointer
	if len(b) > 0 {
		_p0 = unsafe.Pointer(&b[0])
	} else {
		_p0 = unsafe.Pointer(&_zero)
	}
	_, _, e1 := syscall_syscall(libc_mprotect_trampoline_addr, uintptr(_p0), uintptr(len(b)), uintptr(prot))
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

var libc_mprotect_trampoline_addr uintptr

//go:cgo_import_dynamic libc_mprotect mprotect "/usr/lib/libSystem.B.dylib"

/* -------------------------------------------------------------------------------------------------------------------*/
func mmap(addr uintptr, length uintptr, prot int, flag int, fd int, pos int64) (ret uintptr, err error) {
	r0, _, e1 := syscall_syscall6(libc_mmap_trampoline_addr,
		uintptr(addr), uintptr(length), uintptr(prot),
		uintptr(flag), uintptr(fd), uintptr(pos))
	ret = uintptr(r0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

var libc_mmap_trampoline_addr uintptr

//go:cgo_import_dynamic libc_mmap mmap "/usr/lib/libSystem.B.dylib"

/* -------------------------------------------------------------------------------------------------------------------*/
func munmap(addr uintptr, length uintptr) (err error) {
	_, _, e1 := syscall_syscall(libc_munmap_trampoline_addr, uintptr(addr), uintptr(length), 0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

var libc_munmap_trampoline_addr uintptr

//go:cgo_import_dynamic libc_munmap munmap "/usr/lib/libSystem.B.dylib"

/* -------------------------------------------------------------------------------------------------------------------*/
func closeFd(fd int) (err error) {
	_, _, e1 := syscall_syscall(libc_close_trampoline_addr, uintptr(fd), 0, 0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

var libc_close_trampoline_addr uintptr

//go:cgo_import_dynamic libc_close close "/usr/lib/libSystem.B.dylib"

/* -------------------------------------------------------------------------------------------------------------------*/

func fcntl(fd int, cmd int, arg int) (val int, err error) {
	r0, _, e1 := syscall_syscall(libc_fcntl_trampoline_addr, uintptr(fd), uintptr(cmd), uintptr(arg))
	val = int(r0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

var libc_fcntl_trampoline_addr uintptr

//go:cgo_import_dynamic libc_fcntl fcntl "/usr/lib/libSystem.B.dylib"

/* -------------------------------------------------------------------------------------------------------------------*/

func ftruncate(fd int, length int) (err error) {
	_, _, e1 := syscall_syscall(libc_ftruncate_trampoline_addr, uintptr(fd), uintptr(length), 0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

var libc_ftruncate_trampoline_addr uintptr

//go:cgo_import_dynamic libc_ftruncate ftruncate "/usr/lib/libSystem.B.dylib"

/* -------------------------------------------------------------------------------------------------------------------*/

func fstat(fd int, stat *Stat_t) (err error) {
	_, _, e1 := syscall_syscall(libc_fstat64_trampoline_addr, uintptr(fd), uintptr(unsafe.Pointer(stat)), 0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

var libc_fstat64_trampoline_addr uintptr

//go:cgo_import_dynamic libc_fstat64 fstat64 "/usr/lib/libSystem.B.dylib"

/* -------------------------------------------------------------------------------------------------------------------*/

func fchown(fd int, uid int, gid int) (err error) {
	_, _, e1 := syscall_syscall(libc_fchown_trampoline_addr, uintptr(fd), uintptr(uid), uintptr(gid))
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

var libc_fchown_trampoline_addr uintptr

//go:cgo_import_dynamic libc_fchown fchown "/usr/lib/libSystem.B.dylib"

/* -------------------------------------------------------------------------------------------------------------------*/

func fchmod(fd int, mode int) (err error) {
	_, _, e1 := syscall_syscall(libc_fchmod_trampoline_addr, uintptr(fd), uintptr(mode), 0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

var libc_fchmod_trampoline_addr uintptr

//go:cgo_import_dynamic libc_fchmod fchmod "/usr/lib/libSystem.B.dylib"

/* -------------------------------------------------------------------------------------------------------------------*/

// Implemented in the runtime package (runtime/sys_darwin.go)
func syscall_syscall(fn, a1, a2, a3 uintptr) (r1, r2 uintptr, err Errno)
func syscall_syscall6(fn, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err Errno)
func syscall_syscall6X(fn, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err Errno)
func syscall_syscall9(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9 uintptr) (r1, r2 uintptr, err Errno) // 32-bit only
func syscall_rawSyscall(fn, a1, a2, a3 uintptr) (r1, r2 uintptr, err Errno)
func syscall_rawSyscall6(fn, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err Errno)
func syscall_syscallPtr(fn, a1, a2, a3 uintptr) (r1, r2 uintptr, err Errno)

//go:linkname syscall_syscall syscall.syscall
//go:linkname syscall_syscall6 syscall.syscall6
//go:linkname syscall_syscall6X syscall.syscall6X
//go:linkname syscall_syscall9 syscall.syscall9
//go:linkname syscall_rawSyscall syscall.rawSyscall
//go:linkname syscall_rawSyscall6 syscall.rawSyscall6
//go:linkname syscall_syscallPtr syscall.syscallPtr
