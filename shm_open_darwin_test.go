package posix_test

import (
	"gopkg.in/ro-ag/posix.v0"
	"log"
	"testing"
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
			gotFd, err := posix.ShmAnonymous(tt.args.name)
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
