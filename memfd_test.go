//go:build darwin || linux

package posix_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

	"gopkg.in/ro-ag/posix.v1"
)

// These tests pin the behavior of the anonymous "memory fd" — native
// memfd_create on Linux, emulated with shm_open + immediate shm_unlink on macOS.
// They run on both platforms so the emulation is checked against the real thing.

// TestMemfdCloexecFlag checks the inheritance knob the whole fd-passing model
// depends on: without MFD_CLOEXEC the descriptor must survive exec (FD_CLOEXEC
// clear); with it, the descriptor must be close-on-exec.
func TestMemfdCloexecFlag(t *testing.T) {
	fd, err := posix.MemfdCreate("inherit", 0)
	if err != nil {
		t.Fatalf("MemfdCreate(0): %v", err)
	}
	flags, err := posix.Fcntl(fd, posix.F_GETFD, 0)
	_ = posix.Close(fd)
	if err != nil {
		t.Fatalf("F_GETFD: %v", err)
	}
	if flags&posix.FD_CLOEXEC != 0 {
		t.Error("MemfdCreate(0): FD_CLOEXEC is set; the fd must be inheritable for fd-passing")
	}

	fd2, err := posix.MemfdCreate("noinherit", posix.MFD_CLOEXEC)
	if err != nil {
		t.Fatalf("MemfdCreate(MFD_CLOEXEC): %v", err)
	}
	defer func() { _ = posix.Close(fd2) }()
	flags2, err := posix.Fcntl(fd2, posix.F_GETFD, 0)
	if err != nil {
		t.Fatalf("F_GETFD: %v", err)
	}
	if flags2&posix.FD_CLOEXEC == 0 {
		t.Error("MemfdCreate(MFD_CLOEXEC): FD_CLOEXEC is clear, want set")
	}
}

// TestMemfdConcurrentUnique creates many objects concurrently. On macOS the
// emulation derives a unique name per call; a collision would surface as EEXIST.
// This guards the crypto/rand naming against the old time-based scheme.
func TestMemfdConcurrentUnique(t *testing.T) {
	const n = 200
	fds := make([]int, n)
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			fds[i], errs[i] = posix.MemfdCreate("concurrent", posix.MFD_ALLOW_SEALING)
		}(i)
	}
	wg.Wait()
	for i := range n {
		if errs[i] != nil {
			t.Errorf("MemfdCreate #%d failed (name collision under concurrency?): %v", i, errs[i])
			continue
		}
		_ = posix.Close(fds[i])
	}
}

// TestMemfdNoLeak creates and closes many objects and checks the process's open
// descriptor count is stable — i.e. closing an anonymous object releases it.
func TestMemfdNoLeak(t *testing.T) {
	count := func() int {
		ents, err := os.ReadDir("/dev/fd")
		if err != nil {
			t.Skipf("/dev/fd unavailable: %v", err)
		}
		return len(ents)
	}
	before := count()
	for range 200 {
		fd, err := posix.MemfdCreate("leak", posix.MFD_ALLOW_SEALING)
		if err != nil {
			t.Fatalf("MemfdCreate: %v", err)
		}
		if err := posix.Close(fd); err != nil {
			t.Fatalf("Close: %v", err)
		}
	}
	if after := count(); after > before+20 {
		t.Errorf("open fd count grew %d -> %d; anonymous objects are leaking", before, after)
	}
}

// TestMemfdFdPassing proves the defining memfd property: an object with no name
// is shared with another process only by passing its descriptor. The parent
// creates an anonymous object and hands just that fd to a child (as fd 3); the
// child maps it and writes its PID; the parent reads it back through the shared
// mapping. No name is ever involved.
func TestMemfdFdPassing(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "shmexe")
	if out, err := exec.Command("go", "build", "-o", bin, "./test").CombinedOutput(); err != nil {
		t.Fatalf("build child: %v\n%s", err, out)
	}

	fd, err := posix.MemfdCreate("passing", posix.MFD_ALLOW_SEALING)
	if err != nil {
		t.Fatalf("MemfdCreate: %v", err)
	}
	defer func() { _ = posix.Close(fd) }()
	pg := posix.Getpagesize()
	if err := posix.Ftruncate(fd, pg); err != nil {
		t.Fatalf("Ftruncate: %v", err)
	}
	buf, _, err := posix.Mmap(nil, pg, posix.PROT_RDWR, posix.MAP_SHARED, fd, 0)
	if err != nil {
		t.Fatalf("Mmap: %v", err)
	}
	defer func() { _ = posix.Munmap(buf) }()

	f := os.NewFile(uintptr(fd), "memfd")
	cmd := exec.Command(bin)
	cmd.ExtraFiles = []*os.File{f} // child sees this as fd 3
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("child failed: %v\n%s", err, stderr.String())
	}

	// test/shm_exe.go writes "PID: <child pid>" into the shared object.
	want := fmt.Sprintf("PID: %.10d", cmd.Process.Pid)
	if got := string(buf[:len(want)]); got != want {
		t.Errorf("child wrote %q through the inherited fd, want %q", got, want)
	}
}
