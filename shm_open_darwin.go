package posix

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

/* -------------------------------------------------------------------------------------------------------------------*/

func newName(name string) string {
	unique := ""
	n := fmt.Sprintf("%d.%s", time.Now().UnixNano(), name)
	x := md5.Sum([]byte(n))
	s := hex.EncodeToString(x[:])
	unique = _MFD_NAME_PREFIX + s[len(_MFD_NAME_PREFIX)+1:]
	return unique
}

func memfdCreate(name string, flags int) (fd int, err error) {

	fd = -1
	if flags&MFD_HUGETLB == 0 {
		if flags & ^_MFD_ALL_FLAGS != 0 {
			return fd, EINVAL
		}
	} else {
		/* Allow huge page size encoding in flags. */
		if flags & ^(_MFD_ALL_FLAGS|(_MFD_HUGE_MASK<<_MFD_HUGE_SHIFT)) != 0 {
			return fd, EINVAL
		}
	}

	/* length includes terminating zero */
	nameLen := len(name)
	if nameLen <= 0 {
		return fd, EFAULT
	}
	if nameLen > _MFD_NAME_MAX_LEN {
		return fd, EINVAL
	}

	unique := newName(name)
	for {
		if fd, err = shmOpen(unique, O_RDWR|O_CREAT|O_EXCL|O_NOFOLLOW, S_IRUSR|S_IWUSR|S_IRGRP); err != nil {
			if err != syscall.EEXIST {
				return -1, err
			} else {
				unique = newName(name)
			}
		} else {
			break
		}
	}

	if flags&MFD_CLOEXEC == 0 {
		/* Remove Flag FD_CLOEXEC which deletes the file if it needs to be passed to another process */
		if err = remCloseOnExec(fd); err != nil {
			goto unlinking
		}
	}

	/* Delete shmName but keep the file descriptor */
	if err = shmUnlink(unique); err != nil {
		goto closing
	}

	return
unlinking:
	_ = shmUnlink(unique)
closing:
	_ = Close(fd)
	fd = -1
	return
}

//goland:noinspection GoSnakeCaseUsage
const (
	_MFD_NAME_PREFIX           = "memfd:"
	_MFD_NAME_PREFIX_LEN       = 10
	_MFD_NAME_MAX_LEN          = syscall.NAME_MAX - _MFD_NAME_PREFIX_LEN
	_MFD_ALL_FLAGS             = MFD_CLOEXEC | MFD_ALLOW_SEALING | MFD_HUGETLB
	_HUGETLB_FLAG_ENCODE_SHIFT = 26
	_HUGETLB_FLAG_ENCODE_MASK  = 0x3
	_MFD_HUGE_SHIFT            = _HUGETLB_FLAG_ENCODE_SHIFT
	_MFD_HUGE_MASK             = _HUGETLB_FLAG_ENCODE_MASK
)

/* -------------------------------------------------------------------------------------------------------------------*/

func ShmAnonymous() (fd int, err error) {

	fd = -1

	name := "shm_anon"
	unique := newName(name)
	for {
		if fd, err = shmOpen(unique, O_RDWR|O_CREAT|O_EXCL|O_NOFOLLOW, S_IRUSR|S_IWUSR|S_IRGRP); err != nil {
			if err != syscall.EEXIST {
				return -1, err
			} else {
				unique = newName(name)
			}
		} else {
			break
		}
	}

	if err != nil {
		return
	}

	/* Delete shmName but keep the file descriptor */
	if err = shmUnlink(unique); err != nil {
		goto closing
	}

	/* Remove Flag FD_CLOEXEC which deletes the file if it needs to be passed to another process */
	if err = remCloseOnExec(fd); err != nil {
		goto closing
	}

	return
closing:
	Close(fd)
	fd = -1
	return
}

/* -------------------------------------------------------------------------------------------------------------------*/
func remCloseOnExec(fd int) (err error) {
	var arg int
	if arg, err = Fcntl(fd, F_GETFD, 0); err != nil {
		_ = Close(fd)
		return
	}

	arg &^= FD_CLOEXEC // Clear the close-on-exec flag.

	if arg, err = Fcntl(fd, F_SETFD, arg); err != nil {
		_ = Close(fd)
		return
	}
	return
}

/* -------------------------------------------------------------------------------------------------------------------*/
func shmOpen(path string, oflag int, mode uint32) (fd int, err error) {
	var _p0 *byte
	_p0, err = syscall.BytePtrFromString(path)
	if err != nil {
		return
	}
	r0, _, e1 := syscall_syscall(libc_shm_open_trampoline_addr,
		uintptr(unsafe.Pointer(_p0)), uintptr(oflag), uintptr(mode))
	fd = int(r0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

var libc_shm_open_trampoline_addr uintptr

//go:cgo_import_dynamic libc_shm_open shm_open "/usr/lib/libSystem.B.dylib"

/* -------------------------------------------------------------------------------------------------------------------*/

func shmUnlink(path string) (err error) {
	var _p0 *byte
	_p0, err = syscall.BytePtrFromString(path)
	if err != nil {
		return
	}
	_, _, e1 := syscall_syscall(libc_shm_unlink_trampoline_addr, uintptr(unsafe.Pointer(_p0)), 0, 0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

var libc_shm_unlink_trampoline_addr uintptr

//go:cgo_import_dynamic libc_shm_unlink shm_unlink "/usr/lib/libSystem.B.dylib"

/* -------------------------------------------------------------------------------------------------------------------*/
