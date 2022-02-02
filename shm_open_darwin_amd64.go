package posix

import (
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"
)

func ShmAnonymous(name string) (fd int, err error) {

	uniName := fmt.Sprintf("%s-%d", name, time.Now().Nanosecond())
	if fd, err = shmOpen(uniName, O_RDWR|O_CREAT|O_EXCL|O_NOFOLLOW, S_IRUSR|S_IWUSR); err != nil {
		return
	}

	/* Delete name but keep the file descriptor */
	if err = shmUnlink(uniName); err != nil {
		err = &os.PathError{
			Op:   "ShmUnlink",
			Path: uniName,
			Err:  err,
		}
		return
	}

	/* Remove Flag FD_CLOEXEC which deletes the file if it needs to be passed to another process */

	var arg int
	if arg, err = Fcntl(fd, F_GETFD, 0); err != nil {
		_ = Close(fd)
		return -1, err
	}
	arg &^= FD_CLOEXEC // Clear the close-on-exec flag.

	if arg, err = Fcntl(fd, F_SETFD, arg); err != nil {
		_ = Close(fd)
		return -1, err
	}

	return
}

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
