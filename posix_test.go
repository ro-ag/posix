package posix_test

import (
	"fmt"
	"golang.org/x/sys/unix"
	_ "golang.org/x/sys/unix"
	"io/ioutil"
	"os"
	"posix"
	"runtime"
	"syscall"
	"testing"
	"unsafe"
)

func TestShmOpen(t *testing.T) {
	type args struct {
		shmName string
		oflag   int
		mode    uint32
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		unlink  bool
	}{
		{"regular", args{"test-1", posix.O_RDWR | posix.O_CREAT | posix.O_EXCL | posix.O_NOFOLLOW, posix.S_IRUSR | posix.S_IWUSR}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFd, err := posix.ShmOpen(tt.args.shmName, tt.args.oflag, tt.args.mode)
			if (err != nil) != tt.wantErr {
				t.Errorf("ShmOpen() error = %v, wantErr %v", err, tt.wantErr)
				if err.(syscall.Errno) == syscall.EEXIST {
					goto unlink
				}
				return
			}
			if gotFd == -1 {
				t.Errorf("ShmOpen() gotFd = %v", gotFd)
			}
		unlink:
			if tt.unlink {
				if err := posix.ShmUnlink(tt.args.shmName); err != nil {
					t.Errorf("ShmUnlink() error = %v", err)
				}
				return
			}
		})
	}
}

func TestClose(t *testing.T) {
	type args struct {
		fd int
	}
	tests := []struct {
		name    string
		args    args
		create  bool
		wantErr bool
		reErr   bool
	}{
		{"normal", args{fd: 0}, true, false, true},
		{"failure", args{fd: 10}, false, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fd int
			var err error
			if tt.create {
				fd, err = posix.MemfdCreate("test", posix.MFD_ALLOW_SEALING)
				if err != nil {
					t.Errorf("ShmAnonymous = %v", err)
				} else {
					tt.args.fd = fd
				}
			}
			if err := posix.Close(tt.args.fd); (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := posix.Close(tt.args.fd); (err != nil) != tt.reErr {
				t.Errorf("Close() error = %v, reErr %v", err, tt.reErr)
			}

		})
	}
}

func TestFchmod(t *testing.T) {
	type args struct {
		fd   int
		mode int
	}
	tests := []struct {
		name    string
		args    args
		create  bool
		wantErr bool
	}{
		{"Zero", args{0, posix.S_IRUSR}, false, true},
		{"Fail", args{50, posix.S_IRUSR}, false, true},
		{"Normal", args{0, posix.S_IWGRP}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.create {
				tt.args.fd = createFd(t)
				defer func() {
					_ = posix.Close(tt.args.fd)
				}()
			}
			if err := posix.Fchmod(tt.args.fd, tt.args.mode); (err != nil) != tt.wantErr {
				t.Errorf("Fchmod() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFchown(t *testing.T) {
	uid := os.Geteuid()
	gid := os.Getgid()

	type args struct {
		fd  int
		uid int
		gid int
	}
	tests := []struct {
		name    string
		args    args
		create  bool
		wantErr bool
	}{
		{"Zero", args{0, 0, 0}, false, true},
		{"Fail", args{50, uid, uid}, false, true},
		{"Normal", args{0, uid, gid}, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.create {
				tt.args.fd = createFd(t)
				defer func() {
					_ = posix.Close(tt.args.fd)
				}()
			}
			if err := posix.Fchown(tt.args.fd, tt.args.uid, tt.args.gid); (err != nil) != tt.wantErr {
				t.Errorf("Fchown() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFcntl(t *testing.T) {
	t.Parallel()
	file, err := ioutil.TempFile("", "TestFcntlInt")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Remove(file.Name())
	}()
	defer func() {
		_ = file.Close()
	}()

	f := file.Fd()
	displayInfo(int(f), t)
	flags, err := posix.Fcntl(int(f), posix.F_GETFD, 0)
	if err != nil {
		t.Fatal(err)
	}
	if flags&posix.FD_CLOEXEC == 0 {
		t.Errorf("flags %#x do not include FD_CLOEXEC", flags)
	}
}

const FixedAddress uintptr = 0x20000000000

func TestMmap(t *testing.T) {
	b, a, err := posix.MmapAt(unsafe.Pointer(FixedAddress), uintptr(syscall.Getpagesize()), posix.PROT_NONE, unix.MAP_ANON|posix.MAP_PRIVATE, 0, 0)
	if err != nil {
		t.Fatalf("Mmap: %v", err)
	}
	if a != FixedAddress {
		t.Fatalf("Expecting address %p but have %p", unsafe.Pointer(FixedAddress), unsafe.Pointer(a))
	}
	if err := posix.Mprotect(b, posix.PROT_READ|posix.PROT_WRITE); err != nil {
		t.Fatalf("Mprotect: %v", err)
	}

	unix.Getpagesize()
	b[0] = 42

	if runtime.GOOS == "aix" {
		t.Skip("msync returns invalid argument for AIX, skipping msync test")
	} else {
		if err := posix.Msync(b, unix.MS_SYNC); err != nil {
			t.Fatalf("Msync: %v", err)
		}
	}

	if err := posix.Madvise(b, unix.MADV_DONTNEED); err != nil {
		t.Fatalf("Madvise: %v", err)
	}

	if err := posix.Mlock(b, len(b)-1); err != nil {
		t.Fatalf("Munlock: %v", err)
	}

	if err := posix.Mlockall(unix.MCL_CURRENT); err != nil {
		t.Fatalf("Mlockall: %v", err)
	}

	if err := posix.Munlock(b, len(b)-1); err != nil {
		t.Fatalf("Munlock: %v", err)
	}

	if err := posix.Munlockall(); err != nil {
		t.Fatalf("Munlockall: %v", err)
	}

	if err := posix.Munmap(b); err != nil {
		t.Fatalf("Munmap: %v", err)
	}
}

func createFd(t *testing.T) int {
	fd, err := posix.MemfdCreate("test-anon", posix.MFD_ALLOW_SEALING)
	if err != nil {
		t.Error(err)
		return -1
	}
	displayInfo(fd, t)
	return fd
}

func displayInfo(fd int, t *testing.T) {
	var fs posix.Stat_t
	if err := posix.Fstat(fd, &fs); err != nil {
		t.Errorf("Fstat() error = %v", err)
	}
	fmt.Printf("Descriptor %d\n", fd)
	fs.DisplayStatInfo()
}
