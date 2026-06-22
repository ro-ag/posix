package posix

import (
	"syscall"
	"unsafe"
)

func memfdCreate(name string, flags int) (fd int, err error) {
	var _p0 *byte
	_p0, err = syscall.BytePtrFromString(name)
	if err != nil {
		return
	}
	r0, _, e1 := _Syscall(_SYS_MEMFD_CREATE, uintptr(unsafe.Pointer(_p0)), uintptr(flags), 0)
	fd = int(r0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

/* -------------------------------------------------------------------------------------------------------------------*/
func ftruncate(fd int, length int) (err error) {
	_, _, e1 := _Syscall(_SYS_FTRUNCATE, uintptr(fd), uintptr(length), 0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

/* -------------------------------------------------------------------------------------------------------------------*/

func mmap(addr uintptr, length uintptr, prot int, flags int, fd int, offset int64) (addr2 uintptr, err error) {
	r0, _, e1 := _Syscall6(_SYS_MMAP, addr, length, uintptr(prot), uintptr(flags), uintptr(fd), uintptr(offset))
	addr2 = r0
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

/* -------------------------------------------------------------------------------------------------------------------*/

func munmap(addr uintptr, length uintptr) (err error) {
	_, _, e1 := _Syscall(_SYS_MUNMAP, addr, length, 0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

/*--------------------------------------------------------------------------------------------------------------------*/

func madvise(b []byte, advice int) (err error) {
	var _p0 unsafe.Pointer
	if len(b) > 0 {
		_p0 = unsafe.Pointer(&b[0])
	} else {
		_p0 = unsafe.Pointer(&_zero)
	}
	_, _, e1 := _Syscall(_SYS_MADVISE, uintptr(_p0), uintptr(len(b)), uintptr(advice))
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

/*--------------------------------------------------------------------------------------------------------------------*/
func mprotect(b []byte, prot int) (err error) {
	var _p0 unsafe.Pointer
	if len(b) > 0 {
		_p0 = unsafe.Pointer(&b[0])
	} else {
		_p0 = unsafe.Pointer(&_zero)
	}
	_, _, e1 := _Syscall(_SYS_MPROTECT, uintptr(_p0), uintptr(len(b)), uintptr(prot))
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

/*--------------------------------------------------------------------------------------------------------------------*/
func mlock(b []byte, size int) (err error) {
	var _p0 unsafe.Pointer
	if len(b) > 0 {
		_p0 = unsafe.Pointer(&b[0])
	} else {
		_p0 = unsafe.Pointer(&_zero)
	}
	_, _, e1 := _Syscall(_SYS_MLOCK, uintptr(_p0), uintptr(size), 0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

/*--------------------------------------------------------------------------------------------------------------------*/
func mlockall(flags int) (err error) {
	_, _, e1 := _Syscall(_SYS_MLOCKALL, uintptr(flags), 0, 0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

/*--------------------------------------------------------------------------------------------------------------------*/
func munlock(b []byte, size int) (err error) {
	var _p0 unsafe.Pointer
	if len(b) > 0 {
		_p0 = unsafe.Pointer(&b[0])
	} else {
		_p0 = unsafe.Pointer(&_zero)
	}
	_, _, e1 := _Syscall(_SYS_MUNLOCK, uintptr(_p0), uintptr(size), 0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

/*--------------------------------------------------------------------------------------------------------------------*/
func munlockall() (err error) {
	_, _, e1 := _Syscall(_SYS_MUNLOCKALL, 0, 0, 0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

/*--------------------------------------------------------------------------------------------------------------------*/

func msync(b []byte, flags int) (err error) {
	var _p0 unsafe.Pointer
	if len(b) > 0 {
		_p0 = unsafe.Pointer(&b[0])
	} else {
		_p0 = unsafe.Pointer(&_zero)
	}
	_, _, e1 := _Syscall(_SYS_MSYNC, uintptr(_p0), uintptr(len(b)), uintptr(flags))
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

/*--------------------------------------------------------------------------------------------------------------------*/
func closeFd(fd int) (err error) {
	_, _, e1 := _Syscall(_SYS_CLOSE, uintptr(fd), 0, 0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

/*--------------------------------------------------------------------------------------------------------------------*/
func fchown(fd int, uid int, gid int) (err error) {
	_, _, e1 := _Syscall(_SYS_FCHOWN, uintptr(fd), uintptr(uid), uintptr(gid))
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

/*--------------------------------------------------------------------------------------------------------------------*/
func fstat(fd int, stat *Stat_t) (err error) {
	_, _, e1 := _Syscall(_SYS_FSTAT, uintptr(fd), uintptr(unsafe.Pointer(stat)), 0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

type Timespec struct {
	Sec  int64
	Nsec int64
}

type ModeT uint32
type DevT uint64

/*--------------------------------------------------------------------------------------------------------------------*/
func fchmod(fd int, mode int) (err error) {
	_, _, e1 := _Syscall(_SYS_FCHMOD, uintptr(fd), uintptr(mode), 0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

/*--------------------------------------------------------------------------------------------------------------------*/

// fcntl64Syscall is usually SYS_FCNTL, but is overridden on 32-bit Linux
// systems by fcntl_linux_32bit.go to be SYS_FCNTL64.
var fcntl64Syscall uintptr = _SYS_FCNTL

func fcntl(fd int, cmd, arg int) (int, error) {
	valptr, _, errno := _Syscall(fcntl64Syscall, uintptr(fd), uintptr(cmd), uintptr(arg))
	var err error
	if errno != 0 {
		err = errnoErr(errno)
	}
	return int(valptr), err
}

/*--------------------------------------------------------------------------------------------------------------------*/

func _Syscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno)
func _Syscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err syscall.Errno)

/*--------------------------------------------------------------------------------------------------------------------*/
// Stat_t, the _SYS_* syscall numbers, and MAP_32BIT live in the per-arch files
// posix_linux_amd64.go and posix_linux_arm64.go.
