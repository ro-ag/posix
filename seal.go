//go:build darwin || linux

package posix

// File seals restrict the operations allowed on a shared-memory object.
//
// On Linux they are kernel memfd seals: the object must be created with
// MFD_ALLOW_SEALING (as MemfdCreate does), and once set they are enforced
// against every process, irreversibly.
//
// macOS has no kernel sealing, so seals are emulated best-effort and are
// ADVISORY: this package's own Mmap and Ftruncate honor them within the
// process, but they are NOT a security boundary — another process, or a raw
// syscall, can ignore them. For a hard cross-process guarantee use Linux, or
// restrict access through the permission bits passed to ShmOpen (a reader that
// opens the object O_RDONLY cannot map it PROT_WRITE on either platform).
//
//goland:noinspection GoSnakeCaseUsage
const (
	F_SEAL_SEAL         = 0x0001 // prevent any further seals from being set
	F_SEAL_SHRINK       = 0x0002 // prevent the object from being shrunk
	F_SEAL_GROW         = 0x0004 // prevent the object from being grown
	F_SEAL_WRITE        = 0x0008 // prevent writes via a shared mapping
	F_SEAL_FUTURE_WRITE = 0x0010 // prevent future writes; existing mappings keep working
)

// AddSeals adds seals (a bitmask of F_SEAL_*) to fd. Seals are additive and
// cannot be removed; once F_SEAL_SEAL is set, no further seals may be added.
// Adding F_SEAL_WRITE fails with EBUSY while a writable shared mapping of fd is
// still live.
//
// On Linux this is kernel-enforced (fcntl F_ADD_SEALS). On macOS it is an
// in-process, advisory emulation — see the file-seal note above.
func AddSeals(fd int, seals int) error {
	return addSeals(fd, seals)
}

// Seals returns the seals currently set on fd.
func Seals(fd int) (int, error) {
	return getSeals(fd)
}
