package posix

import (
	_ "golang.org/x/sys/unix"
	"log"
	"reflect"
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
		{"regular", args{"test-1", O_RDWR | O_CREAT | O_EXCL | O_NOFOLLOW, S_IRUSR | S_IWUSR}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFd, err := ShmOpen(tt.args.shmName, tt.args.oflag, tt.args.mode)
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
				if err := ShmUnlink(tt.args.shmName); err != nil {
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
				fd, err = MemfdCreate("test", MFD_ALLOW_SEALING)
				if err != nil {
					t.Errorf("ShmAnonymous = %v", err)
				} else {
					tt.args.fd = fd
				}
			}
			if err := Close(tt.args.fd); (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := Close(tt.args.fd); (err != nil) != tt.reErr {
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
		{"Zero", args{0, S_IRUSR}, false, true},
		{"Fail", args{50, S_IRUSR}, false, true},
		{"Normal", args{0, S_IWGRP}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.create {
				fd, err := MemfdCreate("test-anon", MFD_ALLOW_SEALING)
				var fs Stat_t
				if err = Fstat(fd, &fs); err != nil {
					t.Errorf("Fstat() error = %v", err)
				}
				fs.DisplayStatInfo()
				log.Println("fd:", fd, FilePermStr(ModeT(fs.Mode), 0))
				if err != nil {
					t.Errorf("TempFile = %v", err)
				}
				defer func() {
					_ = Close(fd)
				}()
				tt.args.fd = fd
			}
			if err := Fchmod(tt.args.fd, tt.args.mode); (err != nil) != tt.wantErr {
				t.Errorf("Fchmod() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFchown(t *testing.T) {
	type args struct {
		fd  int
		uid int
		gid int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		//{"Zero", args{0, S_IRUSR}, false, true},
		//{"Fail", args{50, S_IRUSR}, false, true},
		//{"Normal", args{0, S_IWGRP}, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Fchown(tt.args.fd, tt.args.uid, tt.args.gid); (err != nil) != tt.wantErr {
				t.Errorf("Fchown() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFcntl(t *testing.T) {
	type args struct {
		fd  int
		cmd int
		arg int
	}
	tests := []struct {
		name    string
		args    args
		wantVal int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, err := Fcntl(tt.args.fd, tt.args.cmd, tt.args.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Fcntl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotVal != tt.wantVal {
				t.Errorf("Fcntl() gotVal = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestFstat(t *testing.T) {
	type args struct {
		fd   int
		stat *Stat_t
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Fstat(tt.args.fd, tt.args.stat); (err != nil) != tt.wantErr {
				t.Errorf("Fstat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFtruncate(t *testing.T) {
	type args struct {
		fd     int
		length int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Ftruncate(tt.args.fd, tt.args.length); (err != nil) != tt.wantErr {
				t.Errorf("Ftruncate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMadvice(t *testing.T) {
	type args struct {
		b     []byte
		behav int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Madvice(tt.args.b, tt.args.behav); (err != nil) != tt.wantErr {
				t.Errorf("Madvice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMlock(t *testing.T) {
	type args struct {
		b    []byte
		size int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Mlock(tt.args.b, tt.args.size); (err != nil) != tt.wantErr {
				t.Errorf("Mlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMlockall(t *testing.T) {
	type args struct {
		flags int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Mlockall(tt.args.flags); (err != nil) != tt.wantErr {
				t.Errorf("Mlockall() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMmap(t *testing.T) {
	type args struct {
		address unsafe.Pointer
		length  uintptr
		prot    int
		flags   int
		fd      int
		offset  int64
	}
	tests := []struct {
		name     string
		args     args
		wantData []byte
		wantAdd  uintptr
		wantErr  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotData, gotAdd, err := Mmap(tt.args.address, tt.args.length, tt.args.prot, tt.args.flags, tt.args.fd, tt.args.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("Mmap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotData, tt.wantData) {
				t.Errorf("Mmap() gotData = %v, want %v", gotData, tt.wantData)
			}
			if gotAdd != tt.wantAdd {
				t.Errorf("Mmap() gotAdd = %v, want %v", gotAdd, tt.wantAdd)
			}
		})
	}
}

func TestMprotect(t *testing.T) {
	type args struct {
		b    []byte
		prot accFlags
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Mprotect(tt.args.b, tt.args.prot); (err != nil) != tt.wantErr {
				t.Errorf("Mprotect() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMunlock(t *testing.T) {
	type args struct {
		b    []byte
		size int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Munlock(tt.args.b, tt.args.size); (err != nil) != tt.wantErr {
				t.Errorf("Munlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMunlockall(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Munlockall(); (err != nil) != tt.wantErr {
				t.Errorf("Munlockall() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMunmap(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Munmap(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("Munmap() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestShmOpen1(t *testing.T) {
	type args struct {
		name  string
		oflag int
		mode  uint32
	}
	tests := []struct {
		name    string
		args    args
		wantFd  int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFd, err := ShmOpen(tt.args.name, tt.args.oflag, tt.args.mode)
			if (err != nil) != tt.wantErr {
				t.Errorf("ShmOpen() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotFd != tt.wantFd {
				t.Errorf("ShmOpen() gotFd = %v, want %v", gotFd, tt.wantFd)
			}
		})
	}
}

func TestShmUnlink(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ShmUnlink(tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("ShmUnlink() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
