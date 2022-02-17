package posix

import (
	"strings"
	"syscall"
	"unsafe"
)

func shmOpen(name string, oflag int, mode uint32) (fd int, err error) {
	fd = -1
	if name, err = shmName(name); err != nil {
		return
	}

	oflag |= O_NOFOLLOW | O_CLOEXEC

	return openat(_AT_FDCWD, name, oflag, mode)
}

func shmUnlink(name string) (err error) {
	if name, err = shmName(name); err != nil {
		return
	}
	return unlinkat(_AT_FDCWD, name, 0)
}

func shmName(name string) (string, error) {

	for len(name) != 0 && name[0] == '/' {
		name = name[1:]
	}

	nameLen := len(name)

	if nameLen == 0 || nameLen >= syscall.NAME_MAX || strings.Contains(name, "/") {
		return "", EINVAL
	}

	return prefix + name, nil
}

func openat(dirfd int, path string, flags int, mode uint32) (fd int, err error) {
	var _p0 *byte
	_p0, err = syscall.BytePtrFromString(path)
	if err != nil {
		return
	}
	r0, _, e1 := _Syscall6(_SYS_OPENAT, uintptr(dirfd), uintptr(unsafe.Pointer(_p0)), uintptr(flags), uintptr(mode), 0, 0)
	fd = int(r0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

func unlinkat(dirfd int, path string, flags int) (err error) {
	var _p0 *byte
	_p0, err = syscall.BytePtrFromString(path)
	if err != nil {
		return
	}
	_, _, e1 := _Syscall(_SYS_UNLINKAT, uintptr(dirfd), uintptr(unsafe.Pointer(_p0)), uintptr(flags))
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

//goland:noinspection GoSnakeCaseUsage
const (
	prefix        = "/dev/shmTable/"
	_SYS_OPENAT   = 257
	_SYS_UNLINKAT = 263
	_AT_FDCWD     = -0x64
)
