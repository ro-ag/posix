package posix

import "sync"

// macOS has no kernel file sealing, so seals are emulated in-process. The state
// is per descriptor; the package's Mmap and Ftruncate consult it. It is
// advisory only — see the note in seal.go.
var (
	sealMu    sync.Mutex
	sealState = make(map[int]int)
)

func addSeals(fd, seals int) error {
	// Match Linux: F_SEAL_WRITE is refused while a writable mapping is live.
	if seals&F_SEAL_WRITE != 0 && mapper.hasWritableMapping(fd) {
		return EBUSY
	}
	sealMu.Lock()
	defer sealMu.Unlock()
	if sealState[fd]&F_SEAL_SEAL != 0 {
		return EPERM
	}
	sealState[fd] |= seals
	return nil
}

func getSeals(fd int) (int, error) {
	sealMu.Lock()
	defer sealMu.Unlock()
	return sealState[fd], nil
}

func sealsOf(fd int) int {
	sealMu.Lock()
	defer sealMu.Unlock()
	return sealState[fd]
}

// sealCheckMmap blocks a writable shared mapping of a write-sealed object.
// A private (copy-on-write) mapping is allowed, matching Linux: its writes do
// not reach the object.
func sealCheckMmap(fd, prot, flags int) error {
	if prot&PROT_WRITE == 0 || flags&MAP_SHARED == 0 {
		return nil
	}
	if sealsOf(fd)&(F_SEAL_WRITE|F_SEAL_FUTURE_WRITE) != 0 {
		return EPERM
	}
	return nil
}

// sealCheckTruncate blocks shrinking or growing a size-sealed object.
func sealCheckTruncate(fd, length int) error {
	s := sealsOf(fd)
	if s&(F_SEAL_SHRINK|F_SEAL_GROW) == 0 {
		return nil
	}
	var st Stat_t
	if err := fstat(fd, &st); err != nil {
		return err
	}
	cur := int(st.Size)
	if length < cur && s&F_SEAL_SHRINK != 0 {
		return EPERM
	}
	if length > cur && s&F_SEAL_GROW != 0 {
		return EPERM
	}
	return nil
}

// sealForget drops the emulated seal state when fd is closed, so a reused fd
// number does not inherit stale seals.
func sealForget(fd int) {
	sealMu.Lock()
	delete(sealState, fd)
	sealMu.Unlock()
}
