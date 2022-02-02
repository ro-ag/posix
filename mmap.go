package posix

import (
	"sync"
	"unsafe"
)

// Mmap manager, for use by operating system-specific implementations.

type mmapper struct {
	sync.Mutex
	active map[*byte][]byte // active mappings; key is last byte in mapping
	mmap   func(addr, length uintptr, prot, flags, fd int, offset int64) (uintptr, error)
	munmap func(addr uintptr, length uintptr) error
}

func (m *mmapper) Mmap(address unsafe.Pointer, length uintptr, prot int, flags int, fd int, offset int64) (data []byte, addr uintptr, err error) {
	if length <= 0 {
		return nil, 0x00, EINVAL
	}

	// Map the requested memory.
	addr, err = m.mmap(uintptr(address), length, prot, flags, fd, offset)
	if err != nil {
		return nil, 0x00, err
	}

	// Use unsafe to convert addr into a []byte.
	var b []byte
	hdr := (*_unsafeSlice)(unsafe.Pointer(&b))
	hdr.Data = unsafe.Pointer(addr)
	hdr.Cap = int(length)
	hdr.Len = int(length)

	// Register mapping in m and return it.
	p := &b[cap(b)-1]
	m.Lock()
	defer m.Unlock()
	m.active[p] = b
	return b, addr, nil
}

func (m *mmapper) Munmap(data []byte) (err error) {
	if len(data) == 0 || len(data) != cap(data) {
		return EINVAL
	}

	// Find the base of the mapping.
	p := &data[cap(data)-1]
	m.Lock()
	defer m.Unlock()
	b := m.active[p]
	if b == nil || &b[0] != &data[0] {
		return EINVAL
	}

	// Unmap the memory and update m.
	if errno := m.munmap(uintptr(unsafe.Pointer(&b[0])), uintptr(len(b))); errno != nil {
		return errno
	}
	delete(m.active, p)
	return nil
}

var mapper = &mmapper{
	active: make(map[*byte][]byte),
	mmap:   mmap,
	munmap: munmap,
}
