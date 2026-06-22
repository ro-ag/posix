package posix_test

import (
	"testing"

	"gopkg.in/ro-ag/posix.v1"
)

// These tests assert real Linux kernel behavior — the semantics the macOS side
// only emulates. The _linux_test.go filename keeps them Linux-only.

// TestMemfdSealingRequiresAllowFlag: on Linux a memfd is sealable only if created
// with MFD_ALLOW_SEALING. Without it, AddSeals must fail; with it, it succeeds.
// Validates the MFD_ALLOW_SEALING value and the kernel's sealing gate.
func TestMemfdSealingRequiresAllowFlag(t *testing.T) {
	noSeal, err := posix.MemfdCreate("nosealing", 0)
	if err != nil {
		t.Fatalf("MemfdCreate(0): %v", err)
	}
	defer func() { _ = posix.Close(noSeal) }()
	if err := posix.AddSeals(noSeal, posix.F_SEAL_WRITE); err == nil {
		t.Error("AddSeals on a memfd created without MFD_ALLOW_SEALING: want error, got nil")
	}

	sealable, err := posix.MemfdCreate("sealing", posix.MFD_ALLOW_SEALING)
	if err != nil {
		t.Fatalf("MemfdCreate(MFD_ALLOW_SEALING): %v", err)
	}
	defer func() { _ = posix.Close(sealable) }()
	if err := posix.AddSeals(sealable, posix.F_SEAL_WRITE); err != nil {
		t.Errorf("AddSeals with MFD_ALLOW_SEALING: %v", err)
	}
}

// TestMemfdNonSealableReportsSealSeal: on Linux a memfd created without
// MFD_ALLOW_SEALING starts with exactly F_SEAL_SEAL set (sealed against further
// seals). Validates F_GET_SEALS plumbing and the F_SEAL_SEAL value.
func TestMemfdNonSealableReportsSealSeal(t *testing.T) {
	fd, err := posix.MemfdCreate("nosealing", 0)
	if err != nil {
		t.Fatalf("MemfdCreate(0): %v", err)
	}
	defer func() { _ = posix.Close(fd) }()
	seals, err := posix.Seals(fd)
	if err != nil {
		t.Fatalf("Seals: %v", err)
	}
	if seals != posix.F_SEAL_SEAL {
		t.Errorf("Seals on a non-sealable memfd = %#x, want F_SEAL_SEAL (%#x)", seals, posix.F_SEAL_SEAL)
	}
}
