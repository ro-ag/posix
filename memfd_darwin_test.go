package posix_test

import (
	"os"
	"testing"

	"gopkg.in/ro-ag/posix.v1"
)

// These tests pin the macOS-specific behavior of the shm_open-based memfd
// emulation. The _darwin_test.go filename keeps them macOS-only.

// TestMacOSShmTruncateOnce: a macOS shm object's size is set once — a second
// Ftruncate returns an error. (Linux allows repeated ftruncate.)
func TestMacOSShmTruncateOnce(t *testing.T) {
	fd, err := posix.MemfdCreate("once", posix.MFD_ALLOW_SEALING)
	if err != nil {
		t.Fatalf("MemfdCreate: %v", err)
	}
	defer func() { _ = posix.Close(fd) }()

	pg := posix.Getpagesize()
	if err := posix.Ftruncate(fd, pg); err != nil {
		t.Fatalf("first Ftruncate: %v", err)
	}
	if err := posix.Ftruncate(fd, 2*pg); err == nil {
		t.Error("second Ftruncate on a macOS shm object: want error, got nil")
	}
}

// TestMacOSShmSizePageRounds: macOS rounds a shm object's size up to a whole
// page, so a sub-page request reads back as one page.
func TestMacOSShmSizePageRounds(t *testing.T) {
	fd, err := posix.MemfdCreate("round", posix.MFD_ALLOW_SEALING)
	if err != nil {
		t.Fatalf("MemfdCreate: %v", err)
	}
	defer func() { _ = posix.Close(fd) }()

	pg := posix.Getpagesize()
	req := pg/2 + 123 // deliberately not a page multiple
	if err := posix.Ftruncate(fd, req); err != nil {
		t.Fatalf("Ftruncate: %v", err)
	}
	var st posix.Stat_t
	if err := posix.Fstat(fd, &st); err != nil {
		t.Fatalf("Fstat: %v", err)
	}
	if st.Size != int64(pg) {
		t.Errorf("Ftruncate(%d) -> Size %d, want %d (rounded up to a page)", req, st.Size, pg)
	}
}

// TestMacOSShmNoMapPrivate: macOS shm objects cannot be mapped MAP_PRIVATE.
func TestMacOSShmNoMapPrivate(t *testing.T) {
	fd, err := posix.MemfdCreate("priv", posix.MFD_ALLOW_SEALING)
	if err != nil {
		t.Fatalf("MemfdCreate: %v", err)
	}
	defer func() { _ = posix.Close(fd) }()
	if err := posix.Ftruncate(fd, posix.Getpagesize()); err != nil {
		t.Fatalf("Ftruncate: %v", err)
	}
	if _, _, err := posix.Mmap(nil, posix.Getpagesize(), posix.PROT_READ, posix.MAP_PRIVATE, fd, 0); err == nil {
		t.Error("MAP_PRIVATE mapping of a macOS shm object: want error, got nil")
	}
}

// TestMacOSShmFchmodFchownRejected: macOS rejects changing a shm object's
// permissions or ownership after creation (the mode passed to ShmOpen is the
// only chance to set them).
func TestMacOSShmFchmodFchownRejected(t *testing.T) {
	fd, err := posix.MemfdCreate("perm", posix.MFD_ALLOW_SEALING)
	if err != nil {
		t.Fatalf("MemfdCreate: %v", err)
	}
	defer func() { _ = posix.Close(fd) }()
	if err := posix.Fchmod(fd, 0o600); err == nil {
		t.Error("Fchmod on a macOS shm object: want error, got nil")
	}
	if err := posix.Fchown(fd, os.Geteuid(), os.Getgid()); err == nil {
		t.Error("Fchown on a macOS shm object: want error, got nil")
	}
}

// TestShmAnonymousUsable: ShmAnonymous (the macOS-native anonymous shm helper)
// returns a descriptor that can be sized, mapped, and read/written.
func TestShmAnonymousUsable(t *testing.T) {
	fd, err := posix.ShmAnonymous()
	if err != nil {
		t.Fatalf("ShmAnonymous: %v", err)
	}
	defer func() { _ = posix.Close(fd) }()
	pg := posix.Getpagesize()
	if err := posix.Ftruncate(fd, pg); err != nil {
		t.Fatalf("Ftruncate: %v", err)
	}
	buf, _, err := posix.Mmap(nil, pg, posix.PROT_READ|posix.PROT_WRITE, posix.MAP_SHARED, fd, 0)
	if err != nil {
		t.Fatalf("Mmap: %v", err)
	}
	copy(buf, "anon works")
	if got := string(buf[:10]); got != "anon works" {
		t.Errorf("read back %q, want %q", got, "anon works")
	}
	if err := posix.Munmap(buf); err != nil {
		t.Errorf("Munmap: %v", err)
	}
}
