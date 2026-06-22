//go:build darwin || linux

package posix

import (
	"sync"
	"unsafe"
)

// mapping records one active mmap: the slice handed to the caller plus the fd
// and protection it was created with. The fd/prot are used by the seal logic
// (hasWritableMapping) to mirror the kernel's "F_SEAL_WRITE needs no live
// writable mapping" rule.
type mapping struct {
	data []byte
	fd   int
	prot int
}

// mmapper tracks active mappings so Munmap can recover each mapping's base
// address and length from the []byte the caller was handed. It is the OS-
// independent half of the implementation; the mmap/munmap fields are the
// per-platform syscalls.
type mmapper struct {
	sync.Mutex
	active map[*byte]mapping // active mappings, keyed by the first byte of each
	mmap   func(addr, length uintptr, prot, flags, fd int, offset int64) (uintptr, error)
	munmap func(addr uintptr, length uintptr) error
}

func (m *mmapper) Mmap(address unsafe.Pointer, length uintptr, prot int, flags int, fd int, offset int64) (data []byte, addr uintptr, err error) {
	if length == 0 {
		return nil, 0, EINVAL
	}

	// Map the requested memory.
	addr, err = m.mmap(uintptr(address), length, prot, flags, fd, offset)
	if err != nil {
		return nil, 0, err
	}

	// Expose the mapped region as a []byte without copying.
	b := unsafe.Slice((*byte)(unsafe.Pointer(addr)), length)

	// Register the mapping, keyed by its base byte (the kernel never returns two
	// live mappings starting at the same address), and return it.
	p := &b[0]
	m.Lock()
	defer m.Unlock()
	m.active[p] = mapping{data: b, fd: fd, prot: prot}
	return b, addr, nil
}

func (m *mmapper) Munmap(data []byte) (err error) {
	if len(data) == 0 || len(data) != cap(data) {
		return EINVAL
	}

	// Recover the mapping from its base byte and require an exact match, so a
	// resliced or unrelated slice cannot unmap a region it does not own.
	p := &data[0]
	m.Lock()
	defer m.Unlock()
	mp, ok := m.active[p]
	if !ok || &mp.data[0] != &data[0] || len(mp.data) != len(data) {
		return EINVAL
	}

	// Unmap the memory and drop the bookkeeping entry.
	if errno := m.munmap(uintptr(unsafe.Pointer(&mp.data[0])), uintptr(len(mp.data))); errno != nil {
		return errno
	}
	delete(m.active, p)
	return nil
}

// hasWritableMapping reports whether a writable mapping of fd made through this
// package is currently live. The seal logic uses it to refuse F_SEAL_WRITE while
// such a mapping exists, matching the Linux kernel.
func (m *mmapper) hasWritableMapping(fd int) bool {
	m.Lock()
	defer m.Unlock()
	for _, mp := range m.active {
		if mp.fd == fd && mp.prot&PROT_WRITE != 0 {
			return true
		}
	}
	return false
}

var mapper = &mmapper{
	active: make(map[*byte]mapping),
	mmap:   mmap,
	munmap: munmap,
}
