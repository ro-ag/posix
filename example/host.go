package main

import (
	"bytes"
	"gopkg.in/ro-ag/posix.v0"
	"log"
	"unsafe"
)

const FixedAddress uintptr = 0x20000000000

func main() {
	fd, err := posix.MemfdCreate("test-mem", posix.MFD_ALLOW_SEALING)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = posix.Close(fd)
	}()

	if err = posix.Ftruncate(fd, posix.Getpagesize()); err != nil {
		log.Fatal(err)
	}

	mem, addr, err := posix.Mmap(unsafe.Pointer(FixedAddress), posix.Getpagesize(), posix.PROT_WRITE, posix.MAP_ANONYMOUS|posix.MAP_SHARED, fd, 0)
	if err != nil {
		log.Fatal(err)
	}
	if addr != FixedAddress {
		log.Fatalf("Expected Address %p but got %p\n", unsafe.Pointer(addr), unsafe.Pointer(FixedAddress))
	}

	buf := bytes.NewBuffer(mem)

	buf.WriteString("writing shared memory")
	if err = posix.Munmap(mem); err != nil {
		log.Fatal(err)
	}
}
