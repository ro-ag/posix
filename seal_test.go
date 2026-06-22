//go:build darwin || linux

package posix_test

import (
	"fmt"
	"os"
	"runtime"
	"testing"

	"gopkg.in/ro-ag/posix.v1"
)

func sealableFd(t *testing.T) int {
	t.Helper()
	fd, err := posix.MemfdCreate("seal", posix.MFD_ALLOW_SEALING)
	if err != nil {
		t.Fatalf("MemfdCreate: %v", err)
	}
	return fd
}

// TestSealWriteBlocksMapping: once F_SEAL_WRITE is set (no live writable
// mapping), a PROT_WRITE shared mapping is refused and a read-only one is
// allowed. Kernel-enforced on Linux, advisory (this package) on macOS — same
// observable result.
func TestSealWriteBlocksMapping(t *testing.T) {
	pg := posix.Getpagesize()
	fd := sealableFd(t)
	defer func() { _ = posix.Close(fd) }()
	if err := posix.Ftruncate(fd, pg); err != nil {
		t.Fatalf("Ftruncate: %v", err)
	}
	if err := posix.AddSeals(fd, posix.F_SEAL_WRITE); err != nil {
		t.Fatalf("AddSeals(WRITE): %v", err)
	}
	if s, _ := posix.Seals(fd); s&posix.F_SEAL_WRITE == 0 {
		t.Errorf("Seals = %#x, want F_SEAL_WRITE set", s)
	}
	if _, _, err := posix.Mmap(nil, pg, posix.PROT_READ|posix.PROT_WRITE, posix.MAP_SHARED, fd, 0); err == nil {
		t.Error("PROT_WRITE MAP_SHARED on a write-sealed object: want error, got nil")
	}
	buf, _, err := posix.Mmap(nil, pg, posix.PROT_READ, posix.MAP_SHARED, fd, 0)
	if err != nil {
		t.Fatalf("PROT_READ mapping should still be allowed: %v", err)
	}
	_ = posix.Munmap(buf)
}

// TestSealWriteBusyWithMapping: F_SEAL_WRITE is refused while a writable shared
// mapping is live (EBUSY-like), and accepted once it is unmapped. Matches Linux.
func TestSealWriteBusyWithMapping(t *testing.T) {
	pg := posix.Getpagesize()
	fd := sealableFd(t)
	defer func() { _ = posix.Close(fd) }()
	if err := posix.Ftruncate(fd, pg); err != nil {
		t.Fatalf("Ftruncate: %v", err)
	}
	buf, _, err := posix.Mmap(nil, pg, posix.PROT_READ|posix.PROT_WRITE, posix.MAP_SHARED, fd, 0)
	if err != nil {
		t.Fatalf("Mmap: %v", err)
	}
	if err := posix.AddSeals(fd, posix.F_SEAL_WRITE); err == nil {
		t.Error("AddSeals(WRITE) with a live writable mapping: want error, got nil")
	}
	if err := posix.Munmap(buf); err != nil {
		t.Fatalf("Munmap: %v", err)
	}
	if err := posix.AddSeals(fd, posix.F_SEAL_WRITE); err != nil {
		t.Errorf("AddSeals(WRITE) after unmap: %v", err)
	}
}

// TestSealShrinkGrow: size seals reject the matching ftruncate. Page-multiple
// sizes are used so macOS's page-rounding does not blur the comparison.
func TestSealShrinkGrow(t *testing.T) {
	pg := posix.Getpagesize()
	fd := sealableFd(t)
	defer func() { _ = posix.Close(fd) }()
	if err := posix.Ftruncate(fd, 2*pg); err != nil {
		t.Fatalf("Ftruncate: %v", err)
	}
	if err := posix.AddSeals(fd, posix.F_SEAL_SHRINK|posix.F_SEAL_GROW); err != nil {
		t.Fatalf("AddSeals(SHRINK|GROW): %v", err)
	}
	if err := posix.Ftruncate(fd, pg); err == nil {
		t.Error("shrink of a SHRINK-sealed object: want error, got nil")
	}
	if err := posix.Ftruncate(fd, 3*pg); err == nil {
		t.Error("grow of a GROW-sealed object: want error, got nil")
	}
	// A no-op truncate to the current size is allowed on Linux. macOS rejects any
	// second ftruncate on a shm object (its size is set once), so skip it there.
	if runtime.GOOS != "darwin" {
		if err := posix.Ftruncate(fd, 2*pg); err != nil {
			t.Errorf("truncate to the current size should be allowed: %v", err)
		}
	}
}

// TestSealSeal: F_SEAL_SEAL blocks any further seals.
func TestSealSeal(t *testing.T) {
	fd := sealableFd(t)
	defer func() { _ = posix.Close(fd) }()
	if err := posix.AddSeals(fd, posix.F_SEAL_SEAL); err != nil {
		t.Fatalf("AddSeals(SEAL): %v", err)
	}
	if err := posix.AddSeals(fd, posix.F_SEAL_WRITE); err == nil {
		t.Error("AddSeals after F_SEAL_SEAL: want error, got nil")
	}
}

// TestSealsAdditive: seals accumulate and read back via Seals.
func TestSealsAdditive(t *testing.T) {
	fd := sealableFd(t)
	defer func() { _ = posix.Close(fd) }()
	if err := posix.AddSeals(fd, posix.F_SEAL_SHRINK); err != nil {
		t.Fatalf("AddSeals(SHRINK): %v", err)
	}
	if err := posix.AddSeals(fd, posix.F_SEAL_GROW); err != nil {
		t.Fatalf("AddSeals(GROW): %v", err)
	}
	s, err := posix.Seals(fd)
	if err != nil {
		t.Fatalf("Seals: %v", err)
	}
	if want := posix.F_SEAL_SHRINK | posix.F_SEAL_GROW; s&want != want {
		t.Errorf("Seals = %#x, want SHRINK|GROW (%#x) set", s, want)
	}
}

// TestSealWritePrivateMappingAllowed: a write seal blocks shared writable maps
// but not a private (copy-on-write) one — its writes never reach the object.
func TestSealWritePrivateMappingAllowed(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("macOS shared-memory objects do not support MAP_PRIVATE")
	}
	pg := posix.Getpagesize()
	fd := sealableFd(t)
	defer func() { _ = posix.Close(fd) }()
	if err := posix.Ftruncate(fd, pg); err != nil {
		t.Fatalf("Ftruncate: %v", err)
	}
	if err := posix.AddSeals(fd, posix.F_SEAL_WRITE); err != nil {
		t.Fatalf("AddSeals(WRITE): %v", err)
	}
	buf, _, err := posix.Mmap(nil, pg, posix.PROT_READ|posix.PROT_WRITE, posix.MAP_PRIVATE, fd, 0)
	if err != nil {
		t.Fatalf("private writable mapping of a write-sealed object should be allowed: %v", err)
	}
	buf[0] = 1 // copy-on-write, local only
	_ = posix.Munmap(buf)
}

// TestReadOnlyFdRejectsWriteMapping: independent of seals, a descriptor opened
// O_RDONLY cannot be mapped PROT_WRITE. This is the kernel-enforced, genuinely
// cross-process way to make a region read-only to a peer — it works on both OSes
// and is the real backstop on macOS, where seals are only advisory.
func TestReadOnlyFdRejectsWriteMapping(t *testing.T) {
	pg := posix.Getpagesize()
	name := fmt.Sprintf("/posix-ro-%d", os.Getpid())
	_ = posix.ShmUnlink(name)
	w, err := posix.ShmOpen(name, posix.O_RDWR|posix.O_CREAT|posix.O_EXCL, 0o600)
	if err != nil {
		t.Fatalf("ShmOpen create: %v", err)
	}
	defer func() { _ = posix.ShmUnlink(name); _ = posix.Close(w) }()
	if err := posix.Ftruncate(w, pg); err != nil {
		t.Fatalf("Ftruncate: %v", err)
	}

	ro, err := posix.ShmOpen(name, posix.O_RDONLY, 0)
	if err != nil {
		t.Fatalf("ShmOpen O_RDONLY: %v", err)
	}
	defer func() { _ = posix.Close(ro) }()

	if _, _, err := posix.Mmap(nil, pg, posix.PROT_READ|posix.PROT_WRITE, posix.MAP_SHARED, ro, 0); err == nil {
		t.Error("PROT_WRITE mapping of an O_RDONLY fd: want error, got nil")
	}
	buf, _, err := posix.Mmap(nil, pg, posix.PROT_READ, posix.MAP_SHARED, ro, 0)
	if err != nil {
		t.Fatalf("PROT_READ mapping of an O_RDONLY fd should work: %v", err)
	}
	_ = posix.Munmap(buf)
}
