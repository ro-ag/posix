//go:build darwin || linux

package posix_test

import (
	"testing"

	"gopkg.in/ro-ag/posix.v1"
)

// TestSealConstantValues pins the F_SEAL_* bits to their kernel values. They are
// hard-coded (not derived from syscall), so a wrong edit would silently break
// sealing on Linux; this locks them.
func TestSealConstantValues(t *testing.T) {
	for _, c := range []struct {
		name      string
		got, want int
	}{
		{"F_SEAL_SEAL", posix.F_SEAL_SEAL, 0x0001},
		{"F_SEAL_SHRINK", posix.F_SEAL_SHRINK, 0x0002},
		{"F_SEAL_GROW", posix.F_SEAL_GROW, 0x0004},
		{"F_SEAL_WRITE", posix.F_SEAL_WRITE, 0x0008},
		{"F_SEAL_FUTURE_WRITE", posix.F_SEAL_FUTURE_WRITE, 0x0010},
	} {
		if c.got != c.want {
			t.Errorf("%s = %#x, want %#x", c.name, c.got, c.want)
		}
	}
}

// TestMfdConstantValues pins the memfd flag and huge-page constants. These are
// hard-coded too — MFD_HUGE_16GB had a wrong value once — so they get a lock.
func TestMfdConstantValues(t *testing.T) {
	for _, c := range []struct {
		name      string
		got, want int
	}{
		{"MFD_CLOEXEC", posix.MFD_CLOEXEC, 0x0001},
		{"MFD_ALLOW_SEALING", posix.MFD_ALLOW_SEALING, 0x0002},
		{"MFD_HUGETLB", posix.MFD_HUGETLB, 0x0004},
		{"MFD_HUGE_64KB", posix.MFD_HUGE_64KB, 16 << 26},
		{"MFD_HUGE_2MB", posix.MFD_HUGE_2MB, 21 << 26},
		{"MFD_HUGE_1GB", posix.MFD_HUGE_1GB, 30 << 26},
		{"MFD_HUGE_16GB", posix.MFD_HUGE_16GB, 34 << 26},
	} {
		if c.got != c.want {
			t.Errorf("%s = %#x, want %#x", c.name, c.got, c.want)
		}
	}
}

// TestSealFutureWrite: F_SEAL_FUTURE_WRITE blocks a new shared writable mapping.
// Unlike F_SEAL_WRITE it does not require unmapping existing mappings first, so
// no live mapping is needed here.
func TestSealFutureWrite(t *testing.T) {
	pg := posix.Getpagesize()
	fd, err := posix.MemfdCreate("futurewrite", posix.MFD_ALLOW_SEALING)
	if err != nil {
		t.Fatalf("MemfdCreate: %v", err)
	}
	defer func() { _ = posix.Close(fd) }()
	if err := posix.Ftruncate(fd, pg); err != nil {
		t.Fatalf("Ftruncate: %v", err)
	}
	if err := posix.AddSeals(fd, posix.F_SEAL_FUTURE_WRITE); err != nil {
		t.Fatalf("AddSeals(FUTURE_WRITE): %v", err)
	}
	if _, _, err := posix.Mmap(nil, pg, posix.PROT_READ|posix.PROT_WRITE, posix.MAP_SHARED, fd, 0); err == nil {
		t.Error("PROT_WRITE shared mapping after F_SEAL_FUTURE_WRITE: want error, got nil")
	}
	buf, _, err := posix.Mmap(nil, pg, posix.PROT_READ, posix.MAP_SHARED, fd, 0)
	if err != nil {
		t.Fatalf("PROT_READ mapping should still be allowed: %v", err)
	}
	_ = posix.Munmap(buf)
}
