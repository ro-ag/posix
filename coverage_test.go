//go:build darwin || linux

package posix_test

import (
	"os"
	"testing"

	"gopkg.in/ro-ag/posix.v1"
)

// TestMfdHugePageConstants pins every MFD_HUGE_* to log2(size) << 26 — the
// kernel's hugetlb size encoding. These are hand-coded; MFD_HUGE_16GB was wrong
// once, so the whole set gets a lock.
func TestMfdHugePageConstants(t *testing.T) {
	for _, c := range []struct {
		name      string
		got, log2 int
	}{
		{"MFD_HUGE_64KB", posix.MFD_HUGE_64KB, 16},
		{"MFD_HUGE_512KB", posix.MFD_HUGE_512KB, 19},
		{"MFD_HUGE_1MB", posix.MFD_HUGE_1MB, 20},
		{"MFD_HUGE_2MB", posix.MFD_HUGE_2MB, 21},
		{"MFD_HUGE_8MB", posix.MFD_HUGE_8MB, 23},
		{"MFD_HUGE_16MB", posix.MFD_HUGE_16MB, 24},
		{"MFD_HUGE_32MB", posix.MFD_HUGE_32MB, 25},
		{"MFD_HUGE_256MB", posix.MFD_HUGE_256MB, 28},
		{"MFD_HUGE_512MB", posix.MFD_HUGE_512MB, 29},
		{"MFD_HUGE_1GB", posix.MFD_HUGE_1GB, 30},
		{"MFD_HUGE_2GB", posix.MFD_HUGE_2GB, 31},
		{"MFD_HUGE_16GB", posix.MFD_HUGE_16GB, 34},
	} {
		if want := c.log2 << 26; c.got != want {
			t.Errorf("%s = %#x, want %#x (%d << 26)", c.name, c.got, want, c.log2)
		}
	}
}

// TestDerivedConstants checks the convenience aliases match their parts.
func TestDerivedConstants(t *testing.T) {
	if posix.PROT_RDWR != posix.PROT_READ|posix.PROT_WRITE {
		t.Errorf("PROT_RDWR = %#x, want PROT_READ|PROT_WRITE = %#x",
			posix.PROT_RDWR, posix.PROT_READ|posix.PROT_WRITE)
	}
	if posix.MAP_ANONYMOUS != posix.MAP_ANON {
		t.Errorf("MAP_ANONYMOUS = %#x, want MAP_ANON = %#x", posix.MAP_ANONYMOUS, posix.MAP_ANON)
	}
}

// TestGetpagesize: the page size is positive and a power of two.
func TestGetpagesize(t *testing.T) {
	pg := posix.Getpagesize()
	if pg <= 0 {
		t.Fatalf("Getpagesize() = %d, want > 0", pg)
	}
	if pg&(pg-1) != 0 {
		t.Errorf("Getpagesize() = %d, want a power of two", pg)
	}
}

// TestMadviseHints: the portable advice values are accepted (madvise returns 0)
// on an anonymous mapping.
func TestMadviseHints(t *testing.T) {
	pg := posix.Getpagesize()
	buf, _, err := posix.Mmap(nil, pg, posix.PROT_READ|posix.PROT_WRITE, posix.MAP_ANON|posix.MAP_PRIVATE, 0, 0)
	if err != nil {
		t.Fatalf("Mmap: %v", err)
	}
	defer func() { _ = posix.Munmap(buf) }()
	for _, a := range []struct {
		name  string
		value int
	}{
		{"MADV_NORMAL", posix.MADV_NORMAL},
		{"MADV_RANDOM", posix.MADV_RANDOM},
		{"MADV_SEQUENTIAL", posix.MADV_SEQUENTIAL},
		{"MADV_WILLNEED", posix.MADV_WILLNEED},
		{"MADV_DONTNEED", posix.MADV_DONTNEED},
	} {
		if err := posix.Madvise(buf, a.value); err != nil {
			t.Errorf("Madvise(%s): %v", a.name, err)
		}
	}
}

// TestMsyncModes: MS_SYNC and MS_ASYNC are each accepted on a file-backed shared
// mapping.
func TestMsyncModes(t *testing.T) {
	pg := posix.Getpagesize()
	fd, err := posix.MemfdCreate("msync", posix.MFD_ALLOW_SEALING)
	if err != nil {
		t.Fatalf("MemfdCreate: %v", err)
	}
	defer func() { _ = posix.Close(fd) }()
	if err := posix.Ftruncate(fd, pg); err != nil {
		t.Fatalf("Ftruncate: %v", err)
	}
	buf, _, err := posix.Mmap(nil, pg, posix.PROT_READ|posix.PROT_WRITE, posix.MAP_SHARED, fd, 0)
	if err != nil {
		t.Fatalf("Mmap: %v", err)
	}
	defer func() { _ = posix.Munmap(buf) }()
	buf[0] = 0x5a
	if err := posix.Msync(buf, posix.MS_SYNC); err != nil {
		t.Errorf("Msync(MS_SYNC): %v", err)
	}
	if err := posix.Msync(buf, posix.MS_ASYNC); err != nil {
		t.Errorf("Msync(MS_ASYNC): %v", err)
	}
}

// TestMmapOffset: a page-aligned offset maps the right slice of the object — a
// marker written at page 1 is visible through a mapping that starts at page 1.
func TestMmapOffset(t *testing.T) {
	pg := posix.Getpagesize()
	fd, err := posix.MemfdCreate("offset", posix.MFD_ALLOW_SEALING)
	if err != nil {
		t.Fatalf("MemfdCreate: %v", err)
	}
	defer func() { _ = posix.Close(fd) }()
	if err := posix.Ftruncate(fd, 2*pg); err != nil {
		t.Fatalf("Ftruncate: %v", err)
	}

	full, _, err := posix.Mmap(nil, 2*pg, posix.PROT_READ|posix.PROT_WRITE, posix.MAP_SHARED, fd, 0)
	if err != nil {
		t.Fatalf("Mmap(full): %v", err)
	}
	full[pg] = 0xab
	if err := posix.Munmap(full); err != nil {
		t.Fatalf("Munmap(full): %v", err)
	}

	part, _, err := posix.Mmap(nil, pg, posix.PROT_READ, posix.MAP_SHARED, fd, int64(pg))
	if err != nil {
		t.Fatalf("Mmap(offset=%d): %v", pg, err)
	}
	defer func() { _ = posix.Munmap(part) }()
	if part[0] != 0xab {
		t.Errorf("byte 0 of the offset mapping = %#x, want 0xab", part[0])
	}
}

// TestMlockUnlock: lock then unlock a small region. Mlock can fail under
// RLIMIT_MEMLOCK on a constrained runner — tolerate that rather than flake.
func TestMlockUnlock(t *testing.T) {
	pg := posix.Getpagesize()
	buf, _, err := posix.Mmap(nil, pg, posix.PROT_READ|posix.PROT_WRITE, posix.MAP_ANON|posix.MAP_PRIVATE, 0, 0)
	if err != nil {
		t.Fatalf("Mmap: %v", err)
	}
	defer func() { _ = posix.Munmap(buf) }()
	if err := posix.Mlock(buf, len(buf)); err != nil {
		t.Skipf("Mlock unavailable (RLIMIT_MEMLOCK?): %v", err)
	}
	if err := posix.Munlock(buf, len(buf)); err != nil {
		t.Errorf("Munlock after a successful Mlock: %v", err)
	}
}

// TestModeTHelpers: Fstat a regular file and check the Stat_t mode classifiers.
func TestModeTHelpers(t *testing.T) {
	f, err := os.CreateTemp("", "posix-mode")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer func() { _ = os.Remove(f.Name()); _ = f.Close() }()

	var st posix.Stat_t
	if err := posix.Fstat(int(f.Fd()), &st); err != nil {
		t.Fatalf("Fstat: %v", err)
	}
	if !st.Mode.S_ISREG() {
		t.Errorf("S_ISREG() = false for a regular file (mode %#o)", st.Mode)
	}
	if st.Mode.S_ISDIR() {
		t.Error("S_ISDIR() = true for a regular file")
	}
}

// TestErrnoHelpers: the errno name/description table resolves known codes.
func TestErrnoHelpers(t *testing.T) {
	if got := posix.ErrnoName(posix.EINVAL); got != "EINVAL" {
		t.Errorf("ErrnoName(EINVAL) = %q, want EINVAL", got)
	}
	if got := posix.ErrnoName(posix.EPERM); got != "EPERM" {
		t.Errorf("ErrnoName(EPERM) = %q, want EPERM", got)
	}
	if posix.ErrnoString(posix.EINVAL) == "" {
		t.Error("ErrnoString(EINVAL) is empty")
	}
}
