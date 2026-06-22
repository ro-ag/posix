//go:build darwin || linux

package posix_test

import (
	"fmt"
	"os"
	"runtime/debug"
	"testing"

	"gopkg.in/ro-ag/posix.v1"
)

// TestFstatSizeAfterFtruncate validates the hand-rolled, per-platform Stat_t
// layout end to end: truncate an object to a known size, then check that Fstat
// reads back that size and the owner. Size and Uid sit at different struct
// offsets (Size late, Uid early), so a wrong offset on either platform shifts
// one of them off its field and fails here — something nothing else in the suite
// catches (every other Fstat call only prints the result).
//
// Sizes are page multiples on purpose: macOS rounds a shared-memory object's
// size up to a page, so a page multiple reads back exactly on both platforms.
func TestFstatSizeAfterFtruncate(t *testing.T) {
	pg := posix.Getpagesize()
	for _, size := range []int{0, pg, pg * 2, pg * 5} {
		fd, err := posix.MemfdCreate("fstat-size", posix.MFD_ALLOW_SEALING)
		if err != nil {
			t.Fatalf("MemfdCreate: %v", err)
		}
		if err := posix.Ftruncate(fd, size); err != nil {
			_ = posix.Close(fd)
			t.Fatalf("Ftruncate(%d): %v", size, err)
		}
		var st posix.Stat_t
		if err := posix.Fstat(fd, &st); err != nil {
			_ = posix.Close(fd)
			t.Fatalf("Fstat: %v", err)
		}
		if st.Size != int64(size) {
			t.Errorf("size %d: Stat_t.Size = %d, want %d — Stat_t layout likely wrong", size, st.Size, size)
		}
		if st.Uid != uint32(os.Geteuid()) {
			t.Errorf("size %d: Stat_t.Uid = %d, want %d — Stat_t layout likely wrong", size, st.Uid, os.Geteuid())
		}
		_ = posix.Close(fd)
	}
}

// TestMmapRoundTripData proves the []byte returned by Mmap is correctly formed
// (data pointer, len, cap) and genuinely backed by the object: a pattern written
// through one mapping is still there after unmap + remap.
func TestMmapRoundTripData(t *testing.T) {
	fd, err := posix.MemfdCreate("roundtrip-data", posix.MFD_ALLOW_SEALING)
	if err != nil {
		t.Fatalf("MemfdCreate: %v", err)
	}
	defer func() { _ = posix.Close(fd) }()

	size := posix.Getpagesize() * 2
	if err := posix.Ftruncate(fd, size); err != nil {
		t.Fatalf("Ftruncate: %v", err)
	}

	offsets := []int{0, posix.Getpagesize() - 1, posix.Getpagesize(), size - 1}

	buf, _, err := posix.Mmap(nil, size, posix.PROT_RDWR, posix.MAP_SHARED, fd, 0)
	if err != nil {
		t.Fatalf("Mmap: %v", err)
	}
	for i, off := range offsets {
		buf[off] = byte(0x41 + i)
	}
	if err := posix.Munmap(buf); err != nil {
		t.Fatalf("Munmap: %v", err)
	}

	// Remap: the writes must have reached the backing object, not just RAM the
	// first mapping happened to touch.
	buf2, _, err := posix.Mmap(nil, size, posix.PROT_RDWR, posix.MAP_SHARED, fd, 0)
	if err != nil {
		t.Fatalf("Mmap (remap): %v", err)
	}
	for i, off := range offsets {
		if got, want := buf2[off], byte(0x41+i); got != want {
			t.Errorf("offset %d: got %#x, want %#x — data did not survive remap", off, got, want)
		}
	}
	if err := posix.Munmap(buf2); err != nil {
		t.Errorf("Munmap (remap): %v", err)
	}
}

// TestMmapMunmapErrors locks down the safety invariants of the mapping registry:
// invalid lengths are rejected, and Munmap of a nil / foreign / already-unmapped
// slice fails instead of unmapping the wrong region.
func TestMmapMunmapErrors(t *testing.T) {
	if _, _, err := posix.Mmap(nil, -1, posix.PROT_RDWR, posix.MAP_ANON|posix.MAP_SHARED, 0, 0); err == nil {
		t.Error("Mmap(length=-1): want error, got nil")
	}
	if _, _, err := posix.Mmap(nil, 0, posix.PROT_RDWR, posix.MAP_ANON|posix.MAP_SHARED, 0, 0); err == nil {
		t.Error("Mmap(length=0): want error, got nil")
	}
	if err := posix.Munmap(nil); err == nil {
		t.Error("Munmap(nil): want error, got nil")
	}
	if err := posix.Munmap(make([]byte, 16)); err == nil {
		t.Error("Munmap(non-mapping): want error, got nil")
	}

	buf, _, err := posix.Mmap(nil, posix.Getpagesize(), posix.PROT_RDWR, posix.MAP_ANON|posix.MAP_SHARED, 0, 0)
	if err != nil {
		t.Fatalf("Mmap: %v", err)
	}
	if err := posix.Munmap(buf); err != nil {
		t.Fatalf("first Munmap: %v", err)
	}
	if err := posix.Munmap(buf); err == nil {
		t.Error("double Munmap: want error, got nil")
	}
}

// TestShmOpenErrors exercises the named-object error paths production code relies
// on: re-creating with O_EXCL, opening a missing object, and an invalid name.
func TestShmOpenErrors(t *testing.T) {
	name := fmt.Sprintf("/posix-err-%d", os.Getpid())
	fd, err := posix.ShmOpen(name, posix.O_RDWR|posix.O_CREAT|posix.O_EXCL, posix.S_IRUSR|posix.S_IWUSR)
	if err != nil {
		t.Fatalf("ShmOpen create: %v", err)
	}
	defer func() { _ = posix.ShmUnlink(name) }()
	defer func() { _ = posix.Close(fd) }()

	if fd2, err := posix.ShmOpen(name, posix.O_RDWR|posix.O_CREAT|posix.O_EXCL, posix.S_IRUSR|posix.S_IWUSR); err == nil {
		_ = posix.Close(fd2)
		t.Error("ShmOpen O_EXCL on existing object: want EEXIST, got nil")
	}
	if fd2, err := posix.ShmOpen(fmt.Sprintf("/posix-missing-%d", os.Getpid()), posix.O_RDWR, 0); err == nil {
		_ = posix.Close(fd2)
		t.Error("ShmOpen on missing object without O_CREAT: want ENOENT, got nil")
	}
	if fd2, err := posix.ShmOpen("", posix.O_RDWR|posix.O_CREAT, posix.S_IRUSR|posix.S_IWUSR); err == nil {
		_ = posix.Close(fd2)
		t.Error(`ShmOpen(""): want error, got nil`)
	}
}

// protectSink defeats the compiler eliding the reads in TestMprotectEnforced.
var protectSink byte

// TestMprotectEnforced verifies Mprotect actually changes page permissions: a
// write to a PROT_READ page faults, a read from a PROT_NONE page faults. It runs
// on every platform (it previously lived in a _darwin_test.go file and so never
// ran on Linux), which also exercises the mprotect syscall on linux/arm64.
func TestMprotectEnforced(t *testing.T) {
	size := posix.Getpagesize()
	buf, _, err := posix.Mmap(nil, size, posix.PROT_READ|posix.PROT_WRITE, posix.MAP_ANON|posix.MAP_PRIVATE, 0, 0)
	if err != nil {
		t.Fatalf("Mmap: %v", err)
	}
	defer func() { _ = posix.Munmap(buf) }()

	buf[0] = 1 // writable now

	if err := posix.Mprotect(buf, posix.PROT_READ); err != nil {
		t.Fatalf("Mprotect PROT_READ: %v", err)
	}
	if !faults(func() { buf[0] = 2 }) {
		t.Error("write to a PROT_READ mapping did not fault")
	}
	protectSink = buf[0] // reads are still allowed

	if err := posix.Mprotect(buf, posix.PROT_NONE); err != nil {
		t.Fatalf("Mprotect PROT_NONE: %v", err)
	}
	if !faults(func() { protectSink = buf[0] }) {
		t.Error("read from a PROT_NONE mapping did not fault")
	}

	if err := posix.Mprotect(buf, posix.PROT_READ|posix.PROT_WRITE); err != nil {
		t.Fatalf("Mprotect restore: %v", err)
	}
}

// TestFcntlCloexecRoundTrip drives F_GETFD/F_SETFD to set and clear FD_CLOEXEC.
// Clearing it is exactly what ShmAnonymous and MemfdCreate rely on so the
// descriptor survives exec; this also exercises the fcntl syscall on
// linux/arm64.
func TestFcntlCloexecRoundTrip(t *testing.T) {
	fd, err := posix.MemfdCreate("cloexec", posix.MFD_ALLOW_SEALING)
	if err != nil {
		t.Fatalf("MemfdCreate: %v", err)
	}
	defer func() { _ = posix.Close(fd) }()

	flags, err := posix.Fcntl(fd, posix.F_GETFD, 0)
	if err != nil {
		t.Fatalf("F_GETFD: %v", err)
	}
	if _, err := posix.Fcntl(fd, posix.F_SETFD, flags|posix.FD_CLOEXEC); err != nil {
		t.Fatalf("F_SETFD set: %v", err)
	}
	got, err := posix.Fcntl(fd, posix.F_GETFD, 0)
	if err != nil {
		t.Fatalf("F_GETFD after set: %v", err)
	}
	if got&posix.FD_CLOEXEC == 0 {
		t.Error("FD_CLOEXEC not set after F_SETFD")
	}

	if _, err := posix.Fcntl(fd, posix.F_SETFD, got&^posix.FD_CLOEXEC); err != nil {
		t.Fatalf("F_SETFD clear: %v", err)
	}
	if got, err = posix.Fcntl(fd, posix.F_GETFD, 0); err != nil {
		t.Fatalf("F_GETFD after clear: %v", err)
	}
	if got&posix.FD_CLOEXEC != 0 {
		t.Error("FD_CLOEXEC still set after clearing it")
	}
}

// faults reports whether fn triggered a memory-protection fault, using
// debug.SetPanicOnFault to turn the SIGSEGV/SIGBUS into a recoverable panic.
func faults(fn func()) (faulted bool) {
	defer debug.SetPanicOnFault(false)
	debug.SetPanicOnFault(true)
	defer func() {
		if recover() != nil {
			faulted = true
		}
	}()
	fn()
	return false
}
