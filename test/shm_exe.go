package main

import (
	"fmt"
	"gopkg.in/ro-ag/posix.v1"
	"log"
	"os"
	"unsafe"
)

const ShmFile uintptr = 3

func main() {
	fmt.Println("»»» run from extern program")
	f := os.NewFile(ShmFile, "external memory")
	bts, addr, err := posix.Mmap(nil, posix.Getpagesize(), posix.PROT_WRITE, posix.MAP_SHARED, int(f.Fd()), 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("»»» got address", unsafe.Pointer(addr))
	textToWrite := fmt.Sprintf("PID: %.10d", os.Getpid())
	copy(bts, textToWrite)
	if err = posix.Munmap(bts); err != nil {
		log.Fatal(err)
	}
}
