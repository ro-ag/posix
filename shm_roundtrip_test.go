//go:build darwin || linux

package posix_test

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"

	"gopkg.in/ro-ag/posix.v1"
)

// TestShmRoundTrip proves cross-process shared memory end to end. It builds the
// example program and runs it; the example creates a named shared-memory object
// via ShmOpen, writes a struct, re-execs itself as a *separate process* that
// opens the same object by name and writes a reply back, then verifies the
// reply. A green run means a struct made the full round trip between two
// processes through shared memory.
func TestShmRoundTrip(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "roundtrip")

	build := exec.Command("go", "build", "-o", bin, "./example/roundtrip")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build example: %v\n%s", err, out)
	}

	out, err := exec.Command(bin).CombinedOutput()
	if err != nil {
		t.Fatalf("run example: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "round-trip OK") {
		t.Fatalf("missing success marker; output:\n%s", out)
	}
	t.Logf("example output:\n%s", out)
}

// TestMmapFixedAddress proves the package's headline feature: mapping at a
// caller-chosen virtual address. We request a high, normally-unused address with
// MAP_FIXED, which requires the kernel to honor it exactly, and confirm the
// returned address matches and the region is usable. This is the capability
// syscall.Mmap and golang.org/x/sys do not expose.
func TestMmapFixedAddress(t *testing.T) {
	const want = 0x30000000000 // distinct from the address other tests use
	size := posix.Getpagesize()

	buf, got, err := posix.Mmap(unsafe.Pointer(uintptr(want)), size,
		posix.PROT_READ|posix.PROT_WRITE, posix.MAP_ANON|posix.MAP_SHARED|posix.MAP_FIXED, 0, 0)
	if err != nil {
		t.Fatalf("fixed-address Mmap: %v", err)
	}
	defer func() {
		if err := posix.Munmap(buf); err != nil {
			t.Errorf("Munmap: %v", err)
		}
	}()

	if got != uintptr(want) {
		t.Fatalf("MAP_FIXED not honored: requested %#x, got %#x", uintptr(want), got)
	}

	// The mapping must be readable and writable.
	buf[0], buf[size-1] = 0xAB, 0xCD
	if buf[0] != 0xAB || buf[size-1] != 0xCD {
		t.Fatal("fixed-address mapping is not usable")
	}
}
