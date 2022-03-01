package posix_test

import (
	"gopkg.in/ro-ag/posix.v1"
	"log"
	"runtime/debug"
	"testing"
	"unsafe"
)

func TestShmAnonymous(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"regular", args{"regular"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFd, err := posix.ShmAnonymous()
			if (err != nil) != tt.wantErr {
				t.Errorf("ShmAnonymous() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotFd == -1 {
				t.Errorf("ShmAnonymous() got %v", gotFd)
			}
		})
	}
}

func TestMemfdCreate(t *testing.T) {
	type args struct {
		name  string
		flags int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"regular", args{"regular", posix.MFD_ALLOW_SEALING}, false},
		{"repeat-1", args{"regular", 0}, false},
		{"repeat-2", args{"regular", 0}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fd, err := posix.MemfdCreate(tt.args.name, tt.args.flags)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemfdCreate() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else {
				log.Println("Got ", fd)
			}
		})
	}
}

func TestMprotect(t *testing.T) {
	var buf []byte
	var err error
	bts := []byte("test to write")

	defer func() {
		_ = posix.Munmap(buf)
	}()
	protect := func(flags int) {
		if err = posix.Mprotect(buf, flags); err != nil {
			t.Log(unix.ErrnoName(err.(posix.Errno)))
			t.Errorf("Mprotect() error = %v", err)
			return
		}
	}

	write := func(name string, catch bool) {
		n := 0
		if catch {
			debug.SetPanicOnFault(true)
			defer debug.SetPanicOnFault(false)
			defer func() {
				if err := recover(); err != nil {
					if n != 0 {
						t.Errorf("%s(write-recoverd): bytes written = %v", name, n)
					}
				} else {
					t.Errorf("%s: Expects system failure", name)
				}
			}()
		}
		if n = copy(buf, bts); n <= 0 {
			t.Errorf("%s(write): bytes written = %v", name, n)
		}
	}

	read := func(name string, catch bool) {
		news := ""
		if catch {
			debug.SetPanicOnFault(true)
			defer debug.SetPanicOnFault(false)
			defer func() {
				if err := recover(); err != nil {
					if news != "" {
						t.Errorf("%s(write-recoverd): bytes read = %v", name, news)
					}
				} else {
					t.Errorf("%s: Expects system failure", name)
				}
			}()
		}

		if news = string(buf[:13]); news != string(bts) {
			t.Errorf("%s(write): bytes readfail", name)
		}
	}

	t.Run("mmap", func(t *testing.T) {
		if buf, _, err = posix.Mmap(unsafe.Pointer(uintptr(0)), posix.Getpagesize(), posix.PROT_NONE, posix.MAP_ANON|posix.MAP_PRIVATE, 0, 0); err != nil {
			t.Log(posix.ErrnoName(err.(posix.Errno)))
			t.Errorf("Mmap() error = %v", err)
			return
		}
	})

	t.Run("PROT_RDWR", func(t *testing.T) {
		protect(posix.PROT_RDWR)
		write("PROT_RDWR", false)
		read("PROT_RDWR", false)
	})

	t.Run("PROT_NONE", func(t *testing.T) {
		protect(posix.PROT_NONE)
		write("PROT_NONE", true)
		write("PROT_NONE", true)
	})

	t.Run("PROT_WRITE", func(t *testing.T) {
		protect(posix.PROT_WRITE)
		write("PROT_WRITE", false)
		read("PROT_WRITE", false)
	})

	t.Run("PROT_READ", func(t *testing.T) {
		protect(posix.PROT_READ)
		write("PROT_READ", true)
		read("PROT_READ", false)
	})
}
