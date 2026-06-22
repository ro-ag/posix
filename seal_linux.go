package posix

// Linux memfd sealing is kernel-enforced through fcntl. The object must have
// been created with MFD_ALLOW_SEALING (MemfdCreate sets it).
//
//goland:noinspection GoSnakeCaseUsage
const (
	_F_ADD_SEALS = 1033 // F_LINUX_SPECIFIC_BASE + 9
	_F_GET_SEALS = 1034 // F_LINUX_SPECIFIC_BASE + 10
)

func addSeals(fd, seals int) error {
	// Mirror the kernel's own EBUSY for a live writable mapping made through
	// this package; the fcntl below is the authoritative enforcement.
	if seals&F_SEAL_WRITE != 0 && mapper.hasWritableMapping(fd) {
		return EBUSY
	}
	_, err := fcntl(fd, _F_ADD_SEALS, seals)
	return err
}

func getSeals(fd int) (int, error) {
	return fcntl(fd, _F_GET_SEALS, 0)
}

// On Linux the kernel enforces seals directly, so the Mmap/Ftruncate/Close
// hooks are no-ops.
func sealCheckMmap(fd, prot, flags int) error { return nil }
func sealCheckTruncate(fd, length int) error  { return nil }
func sealForget(fd int)                       {}
