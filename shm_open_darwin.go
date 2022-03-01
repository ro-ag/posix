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

type fileTable struct {
	Fd     int
	Name   string
	Unique string
	Seals  bool
}

type shmTable struct {
	sync.Mutex
	MapName   map[string]int
	MapActive map[int]fileTable
}

var shm = &shmTable{
	MapName:   make(map[string]int),
	MapActive: make(map[int]fileTable),
}

/* -------------------------------------------------------------------------------------------------------------------*/

func closeFd(fd int) error {
	return shm.Close(fd)
}

func (m *shmTable) Close(fd int) error {
	m.Lock()
	defer m.Unlock()
	if f, ok := m.MapActive[fd]; ok {
		delete(m.MapName, f.Unique)
		delete(m.MapActive, fd)
	}
	return fclose(fd)
}

/* -------------------------------------------------------------------------------------------------------------------*/

func memfdCreate(name string, flags int) (fd int, err error) {
	return shm.MemfdCreate(name, flags)
}

var rgxFm = regexp.MustCompile(`^memfd:(\d{3}):(.+)`)

/* -------------------------------------------------------------------------------------------------------------------*/

func (m *shmTable) NewName(name *string) error {
	// check if MapName exists
	if name == nil {
		return errnoErr(EFAULT)
	}
	unique := fmt.Sprintf(_MFD_NAME_PREFIX, 0, *name)
	for {
		if _, ok := m.MapName[unique]; ok {
			if o := rgxFm.FindSubmatch([]byte(unique)); o != nil {
				ind, _ := strconv.Atoi(string(o[1]))
				if ind == 999 {
					return EFAULT
				} else {
					unique = fmt.Sprintf(_MFD_NAME_PREFIX, ind+1, *name)
				}
			} else {
				return EFAULT
			}
		} else {
			break
		}
	}
	*name = unique
	return nil
}

func (m *shmTable) MemfdCreate(name string, flags int) (fd int, err error) {
	m.Lock()
	defer m.Unlock()
	var f fileTable
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

	uniName := name

	s := time.Now().UnixNano() / int64(time.Millisecond)
	t := s + int64(time.Millisecond)

	for i := s; i < t; i = time.Now().UnixNano() / int64(time.Millisecond) {
		if err = m.NewName(&uniName); err != nil {
			return
		}
		if fd, err = shmOpen(uniName, O_RDWR|O_CREAT|O_EXCL|O_NOFOLLOW, S_IRUSR|S_IWUSR|S_IRGRP); err != nil {
			if err != syscall.EEXIST {
				return
			}
		} else {
			break
		}
	}

	if err != nil {
		return
	}

	if _, ok := m.MapActive[fd]; ok {
		err = EFAULT
		goto unlinking
	}

	/* Delete shmName but keep the file descriptor */
	if err = shmUnlink(uniName); err != nil {
		goto closing
	}

	if flags&MFD_CLOEXEC == 0 {
		/* Remove Flag FD_CLOEXEC which deletes the file if it needs to be passed to another process */
		if err = remCloseOnExec(fd); err != nil {
			goto closing
		}
	}

	if flags&MFD_ALLOW_SEALING != 0 {
		f.Seals = true
	}
	f.Name = name
	f.Unique = uniName
	f.Fd = fd
	m.MapName[uniName] = fd
	m.MapActive[fd] = f
	return
unlinking:
	_ = shmUnlink(uniName)
closing:
	_ = Close(fd)
	fd = -1
	return
}

//goland:noinspection GoSnakeCaseUsage
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

func ShmAnonymous() (int, error) {
	return shm.ShmAnonymous()
}

func (m *shmTable) ShmAnonymous() (fd int, err error) {
	m.Lock()
	defer m.Unlock()
	var f fileTable
	fd = -1

	name := "shm_anon"
	uniName := name

	s := time.Now().UnixNano() / int64(time.Millisecond)
	t := s + int64(time.Millisecond)

	for i := s; i < t; i = time.Now().UnixNano() / int64(time.Millisecond) {
		if err = m.NewName(&uniName); err != nil {
			return
		}

		if fd, err = shmOpen(uniName, O_RDWR|O_CREAT|O_EXCL|O_NOFOLLOW, S_IRUSR|S_IWUSR); err != nil {
			if err != syscall.EEXIST {
				return
			}
		} else {
			break
		}
	}

	if err != nil {
		return
	}

	if _, ok := m.MapActive[fd]; ok {
		err = EFAULT
		goto unlinking
	}

	/* Delete shmName but keep the file descriptor */
	if err = shmUnlink(uniName); err != nil {
		goto closing
	}

	/* Remove Flag FD_CLOEXEC which deletes the file if it needs to be passed to another process */
	if err = remCloseOnExec(fd); err != nil {
		goto closing
	}

	f.Name = name
	f.Unique = uniName
	f.Fd = fd
	m.MapName[uniName] = fd
	m.MapActive[fd] = f
	return
unlinking:
	_ = shmUnlink(uniName)
closing:
	_ = Close(fd)
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
