package posix

import (
	"crypto/rand"
	"fmt"
	"sync/atomic"
	"syscall"
	"unsafe"
)

/* -------------------------------------------------------------------------------------------------------------------*/

// nameCounter makes generated object names unique even if two concurrent
// callers happen to read identical random bytes.
var nameCounter atomic.Uint64

// newName returns a process-unique macOS shared-memory object name that fits in
// the platform name-length limit. It mixes crypto/rand entropy with a global
// counter, so concurrent callers never collide and uniqueness does not depend on
// wall-clock resolution (the previous time-based scheme produced frequent
// duplicates under load). The caller's base name is only length-validated by
// memfdCreate; the emulated object is unlinked immediately, so its on-disk name
// is irrelevant.
func newName(string) string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("%s%x%08x", _MFD_NAME_PREFIX, b[:], uint32(nameCounter.Add(1)))
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
	_HUGETLB_FLAG_ENCODE_MASK  = 0x3f
	_MFD_HUGE_SHIFT            = _HUGETLB_FLAG_ENCODE_SHIFT
	_MFD_HUGE_MASK             = _HUGETLB_FLAG_ENCODE_MASK
)

/* -------------------------------------------------------------------------------------------------------------------*/

// ShmAnonymous creates an anonymous shared-memory object and returns a file
// descriptor for it. It is the macOS analogue of MemfdCreate: the object is
// created, unlinked immediately, and has FD_CLOEXEC cleared so the descriptor
// can be inherited across exec. Available on macOS only.
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
// remCloseOnExec clears the FD_CLOEXEC flag on fd so the descriptor survives
// exec. It does not close fd on failure; the caller owns fd's lifecycle (every
// caller already closes fd on its error path, so closing here would double-close
// and risk closing an unrelated descriptor that reused the number).
func remCloseOnExec(fd int) (err error) {
	var arg int
	if arg, err = Fcntl(fd, F_GETFD, 0); err != nil {
		return
	}

	arg &^= FD_CLOEXEC // Clear the close-on-exec flag.

	if _, err = Fcntl(fd, F_SETFD, arg); err != nil {
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
