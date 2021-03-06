//go:build darwin || linux

package posix_test

import (
	"bytes"
	"fmt"
	"gopkg.in/ro-ag/posix.v1"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"syscall"
	"testing"
	"time"
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
	if runtime.GOOS == "darwin" {
		t.Skip("fchmod doesn't work on shared memory in MacOSx")
		return
	}
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
	if runtime.GOOS == "darwin" {
		t.Skip("fchown doesn't work on shared memory in MacOSx")
		return
	}

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

func TestMmapParallel(t *testing.T) {
	for i := 0; i < 50; i++ {
		name := fmt.Sprintf("mmap-%.3d", i)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			b, a, err := posix.Mmap(unsafe.Pointer(FixedAddress), posix.Getpagesize()*500, posix.PROT_WRITE, posix.MAP_ANON|posix.MAP_SHARED, 0, 0)
			if err != nil {
				t.Errorf("Mmap: %v", err)
				return
			}
			if err = posix.Mlock(b, len(b)-1); err != nil {
				t.Errorf("Mlock: %v", err)
				return
			}
			t.Logf("name: %s, orig: %p, addr: %p, diff: %p", name, unsafe.Pointer(FixedAddress), unsafe.Pointer(a), unsafe.Pointer(a-FixedAddress))
			time.Sleep(50 * time.Microsecond)
			copy(b, name)
			if err = posix.Munmap(b); err != nil {
				t.Errorf("Munmap: %v", err)
			}
		})

	}
}

func TestMemory(t *testing.T) {
	var (
		b   []byte
		a   uintptr
		err error
	)

	t.Run("Mmap", func(t *testing.T) {
		b, a, err = posix.Mmap(unsafe.Pointer(FixedAddress), posix.Getpagesize(), posix.PROT_NONE, posix.MAP_ANON|posix.MAP_PRIVATE, 0, 0)
		if err != nil {
			t.Fatalf("Mmap: %v", err)
		}
		if a != FixedAddress {
			t.Fatalf("Expecting address %p but have %p", unsafe.Pointer(FixedAddress), unsafe.Pointer(a))
		}
	})

	t.Run("Mprotect", func(t *testing.T) {
		if err = posix.Mprotect(b, posix.PROT_READ|posix.PROT_WRITE); err != nil {
			t.Fatalf("Mprotect: %v", err)
		}
	})

	t.Run("write", func(t *testing.T) {
		b[0] = 42
	})

	t.Run("Msync", func(t *testing.T) {
		if err = posix.Msync(b, posix.MS_SYNC); err != nil {
			t.Fatalf("Msync: %v", err)
		}
	})

	t.Run("Madvise", func(t *testing.T) {
		if err = posix.Madvise(b, posix.MADV_DONTNEED); err != nil {
			t.Fatalf("Madvise: %v", err)
		}
	})

	t.Run("Mlock", func(t *testing.T) {
		if err := posix.Mlock(b, len(b)-1); err != nil {
			t.Fatalf("Munlock: %v", err)
		}
	})

	t.Run("Munlock", func(t *testing.T) {
		if err := posix.Munlock(b, len(b)-1); err != nil {
			t.Fatalf("Munlock: %v", err)
		}
	})

	t.Run("Munlockall", func(t *testing.T) {
		if err = posix.Munlockall(); err != nil {
			if err.(posix.Errno) == syscall.ENOSYS {
				t.Skip(err)
				return
			}
			t.Errorf("Munlockall: %v", err)
		}
	})

	t.Run("Mlockall", func(t *testing.T) {
		if err = posix.Mlockall(posix.MCL_CURRENT); err != nil {
			enum := err.(posix.Errno)
			if enum == syscall.ENOSYS {
				t.Skip(err)
				return
			} else if enum == syscall.ENOMEM {
				t.Skip(posix.ErrnoName(enum), posix.ErrnoString(enum), posix.ErrnoHelp(enum))
			}

			t.Errorf("Mlockall: %v - No: %d - %s", err, enum, posix.ErrnoName(enum))
		}
	})

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
	if err = posix.Ftruncate(fd, posix.Getpagesize()); err != nil {
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

func TestStress(t *testing.T) {
	t.Log("build external test")
	FileName := "./test/shm"
	cmd := exec.Command("go", "build", "-o", FileName, "./test")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Errorf("external test: %v", err)
		return
	}

	type segment struct {
		buf    []byte
		addr   uintptr
		fd     int
		f      *os.File
		cmd    *exec.Cmd
		pid    int
		stdout bytes.Buffer
		stderr bytes.Buffer
		err    error
	}

	var err error
	max := 600
	for x := 0; x < 10; x++ {
		name := fmt.Sprintf("(%d) create and delete %.3d maps", x, max)
		t.Run(name, func(t *testing.T) {
			var segments []*segment
			hasError := false
			var wg sync.WaitGroup

			for i := 0; i < max; i++ {
				seg := new(segment)
				if seg.fd, err = posix.MemfdCreate("testing", posix.MFD_ALLOW_SEALING); err != nil {
					t.Errorf("(%.3d) MemfdCreate: %v", i, err)
					hasError = true
					t.Fail()
					goto close
				} else {
					segments = append(segments, seg)
				}
			}

			for i, seg := range segments {
				if seg.err = posix.Ftruncate(seg.fd, 50); seg.err != nil {
					t.Errorf("(%.3d) Ftruncate fd=%d : %v", i, seg.fd, seg.err)
					hasError = true
				}
			}

			if hasError {
				goto close
			}

			for i, seg := range segments {
				seg.f = os.NewFile(uintptr(seg.fd), fmt.Sprintf("file-%d", seg.fd))
				if seg.f == nil {
					seg.err = fmt.Errorf("(%.3d) NewFile fd=%d is null, probably bad descriptior", i, seg.fd)
					t.Error(seg.err)
					hasError = true
				}
			}

			if hasError {
				goto close
			}

			for i, seg := range segments {
				if seg.buf, seg.addr, err = posix.Mmap(unsafe.Pointer(uintptr(0)), 50, posix.PROT_READ|posix.PROT_WRITE, posix.MAP_SHARED, seg.fd, 0); err != nil {
					t.Errorf("(%.3d) Mmap: %v", i, err)
					hasError = true
				}
			}

			if hasError {
				goto unmap
			}

			for i, seg := range segments {
				text := fmt.Sprintf("(%.3d)", i)
				if copy(seg.buf, text) != len(text) {
					seg.err = fmt.Errorf("(%.3d) NewFile fd=%d wrong number of bytes writen", i, seg.fd)
					t.Error(seg.err)
					hasError = true
				}
			}

			for i := range segments {
				seg := segments[i]
				seg.cmd = exec.Command(FileName)
				seg.cmd.Stdout = &seg.stdout
				seg.cmd.Stderr = &seg.stderr
				seg.cmd.ExtraFiles = []*os.File{seg.f}
			}

			t.Log("write from external program")

			for i := range segments {
				seg := segments[i]
				wg.Add(1)
				go func() {
					defer wg.Done()
					seg.err = seg.cmd.Run()
				}()
			}

			wg.Wait()

			for i, seg := range segments {
				if seg.err != nil {
					t.Errorf("(%d) fd=%d: %v", i, seg.fd, seg.err)
					hasError = true
				} else {
					seg.pid = seg.cmd.Process.Pid
				}
			}

			if hasError {
				goto unmap
			}

			t.Log("validate memory")
			for i, seg := range segments {
				if seg.err == nil {
					want := fmt.Sprintf("PID: %.10d", seg.pid)
					got := string(seg.buf[:len(want)])
					if got != want {
						fmt.Println(seg.buf)
						fmt.Println(seg.stdout.String())
						t.Errorf("(%d) fd=%d Got %s but want %s", i, seg.fd, got, want)
						t.Fail()
						hasError = true
					}
				} else {
					t.Errorf("(%d) fd=%d %v", i, seg.fd, err)
					hasError = true
				}
			}
			t.Log("done memory validation")
		unmap:
			t.Log("unmmap")
			for i, sg := range segments {
				if err = posix.Munmap(sg.buf); err != nil {
					t.Errorf("(%.3d) fd=%d Munmap(%p): %v", i, sg.fd, unsafe.Pointer(sg.addr), err)
					t.Fail()
				}
			}
		close:
			t.Log("closing")

			for i, sg := range segments {
				if sg.fd > 3 {
					err = sg.f.Close()
					if err != nil {
						t.Errorf("(%.3d) f=%p file Close(%d): %v", i, sg.f, sg.fd, err)
						t.Fail()
						return
					}
				}
			}
			t.Log("test done")
			return
		})

		if err != nil {
			t.Fail()
			return
		}
	}
}
