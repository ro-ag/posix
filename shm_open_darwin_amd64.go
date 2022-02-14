package posix

import (
	"fmt"
	"regexp"
	"strconv"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

type file struct {
	fd    int
	name  string
	seals bool
}

type mfd struct {
	sync.Mutex
	active map[string]file
}

var _mfd = &mfd{
	active: make(map[string]file),
}

//MemfdCreate creates an anonymous file and returns a file
//descriptor that refers to it.  The file behaves like a regular
//file, and so can be modified, truncated, memory-mapped, and so
//on.  However, unlike a regular file, it lives in RAM and has a
//volatile backing storage.  Once all references to the file are
//dropped, it is automatically released.
//
//NOTE: this is an Emulation of the original function in Linux
//      is made for testing only
func MemfdCreate(name string, flags int) (fd int, err error) {
	return _mfd.MemfdCreate(name, flags)
}

var rgxFm = regexp.MustCompile(`^memfd:(\d{3}):(.+)`)

/* -------------------------------------------------------------------------------------------------------------------*/

func (m *mfd) MemfdCreate(name string, flags int) (fd int, err error) {
	m.Lock()
	defer m.Unlock()
	var f file
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
	name_len := len(name)
	if name_len <= 0 {
		return fd, EFAULT
	}
	if name_len > _MFD_NAME_MAX_LEN {
		return fd, EINVAL
	}

	uniName := fmt.Sprintf(_MFD_NAME_PREFIX, 0, name)

	// check if name exists
	for {
		if _, ok := m.active[uniName]; ok {
			if o := rgxFm.FindSubmatch([]byte(uniName)); o != nil {
				ind, _ := strconv.Atoi(string(o[1]))
				if ind == 999 {
					return fd, EFAULT
				} else {
					uniName = fmt.Sprintf(_MFD_NAME_PREFIX, ind+1, name)
				}
			} else {
				return fd, EFAULT
			}
		} else {
			break
		}
	}

	if fd, err = shmOpen(uniName, O_RDWR|O_CREAT|O_EXCL|O_NOFOLLOW, S_IRUSR|S_IWUSR|S_IRGRP); err != nil {
		return
	}

	/* Delete shmName but keep the file descriptor */
	if err = shmUnlink(uniName); err != nil {
		goto closeall
	}

	if flags&MFD_CLOEXEC == 0 {
		/* Remove Flag FD_CLOEXEC which deletes the file if it needs to be passed to another process */
		if err = remCloseOnExec(fd); err != nil {
			goto closeall
		}
	}

	if flags&MFD_ALLOW_SEALING != 0 {
		f.seals = true
	}
	f.name = name
	f.fd = fd
	m.active[uniName] = f
	return
closeall:
	_ = Close(fd)
	fd = -1
	return
}

const (
	_MFD_NAME_PREFIX           = "memfd:%03d:%s"
	_MFD_NAME_PREFIX_LEN       = 10
	_MFD_NAME_MAX_LEN          = syscall.NAME_MAX - _MFD_NAME_PREFIX_LEN
	_MFD_ALL_FLAGS             = MFD_CLOEXEC | MFD_ALLOW_SEALING | MFD_HUGETLB
	_HUGETLB_FLAG_ENCODE_SHIFT = 26
	_HUGETLB_FLAG_ENCODE_MASK  = 0x3
	_MFD_HUGE_SHIFT            = _HUGETLB_FLAG_ENCODE_SHIFT
	_MFD_HUGE_MASK             = _HUGETLB_FLAG_ENCODE_MASK
)

/* -------------------------------------------------------------------------------------------------------------------*/

func ShmAnonymous(name string) (fd int, err error) {

	uniName := fmt.Sprintf("%s-%d", name, time.Now().Nanosecond())
	if fd, err = shmOpen(uniName, O_RDWR|O_CREAT|O_EXCL|O_NOFOLLOW, S_IRUSR|S_IWUSR); err != nil {
		return
	}

	/* Delete shmName but keep the file descriptor */
	if err = shmUnlink(uniName); err != nil {
		return
	}

	/* Remove Flag FD_CLOEXEC which deletes the file if it needs to be passed to another process */
	if err = remCloseOnExec(fd); err != nil {
		_ = Close(fd)
		return -1, err
	}
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
