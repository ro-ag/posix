// Command roundtrip is a runnable demonstration of cross-process shared memory
// with no cgo. The parent process creates a named POSIX shared-memory object,
// writes a struct into it, then re-executes itself as a separate child process
// that opens the same object by name and writes a reply back. The parent sees
// the child's writes through the shared mapping.
//
//	go run ./example/roundtrip
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"unsafe"

	"gopkg.in/ro-ag/posix.v1"
)

// payload is the fixed-layout struct both processes share. It holds no
// pointers, so its bytes mean the same thing in either process's address space.
type payload struct {
	Magic    uint64
	Seq      uint64
	ChildPID int64
	Reply    [96]byte
}

const (
	magic    = 0x504f53495872 // arbitrary sentinel ("POSIXr")
	childEnv = "POSIX_ROUNDTRIP_CHILD"
)

func main() {
	log.SetFlags(0)
	if name := os.Getenv(childEnv); name != "" {
		child(name)
		return
	}
	parent()
}

func parent() {
	size := int(unsafe.Sizeof(payload{}))
	name := fmt.Sprintf("/posix-rt-%d", os.Getpid())

	// Create the named shared-memory object and give it a size.
	fd, err := posix.ShmOpen(name, posix.O_RDWR|posix.O_CREAT|posix.O_EXCL, posix.S_IRUSR|posix.S_IWUSR)
	if err != nil {
		log.Fatalf("parent ShmOpen: %v", err)
	}
	defer func() { _ = posix.ShmUnlink(name) }()
	if err := posix.Ftruncate(fd, size); err != nil {
		log.Fatalf("parent Ftruncate: %v", err)
	}

	// Map it. The address argument is this package's differentiator: here we
	// pass a hint and report what the kernel actually returned.
	const hint = 0x20000000000
	buf, addr, err := posix.Mmap(unsafe.Pointer(uintptr(hint)), size, posix.PROT_RDWR, posix.MAP_SHARED, fd, 0)
	if err != nil {
		log.Fatalf("parent Mmap: %v", err)
	}
	log.Printf("parent: mapped %q at %#x (hint %#x)", name, addr, uintptr(hint))

	p := (*payload)(unsafe.Pointer(&buf[0]))
	p.Magic = magic
	p.Seq = 42
	log.Printf("parent: wrote Seq=%d", p.Seq)

	// Re-execute ourselves as the child, handing over the object's name.
	exe, err := os.Executable()
	if err != nil {
		log.Fatalf("parent Executable: %v", err)
	}
	cmd := exec.Command(exe)
	cmd.Env = append(os.Environ(), childEnv+"="+name)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("parent: child process failed: %v", err)
	}

	// The child's writes are visible here through the shared mapping.
	if p.Seq != 43 {
		log.Fatalf("parent: expected Seq=43 from child, got %d", p.Seq)
	}
	log.Printf("parent: read back Seq=%d ChildPID=%d Reply=%q", p.Seq, p.ChildPID, p.Reply[:clen(p.Reply[:])])

	if err := posix.Munmap(buf); err != nil {
		log.Fatalf("parent Munmap: %v", err)
	}
	if err := posix.Close(fd); err != nil {
		log.Fatalf("parent Close: %v", err)
	}
	fmt.Println("round-trip OK")
}

func child(name string) {
	size := int(unsafe.Sizeof(payload{}))

	// Open the same object by name and map it. The address hint is nil here:
	// the child only needs to see the bytes, not place them at a fixed address.
	fd, err := posix.ShmOpen(name, posix.O_RDWR, 0)
	if err != nil {
		log.Fatalf("child ShmOpen: %v", err)
	}
	buf, _, err := posix.Mmap(nil, size, posix.PROT_RDWR, posix.MAP_SHARED, fd, 0)
	if err != nil {
		log.Fatalf("child Mmap: %v", err)
	}

	p := (*payload)(unsafe.Pointer(&buf[0]))
	if p.Magic != magic {
		log.Fatalf("child: bad magic %#x", p.Magic)
	}
	if p.Seq != 42 {
		log.Fatalf("child: expected Seq=42 from parent, got %d", p.Seq)
	}
	p.Seq = 43
	p.ChildPID = int64(os.Getpid())
	copy(p.Reply[:], "hello from child")

	if err := posix.Munmap(buf); err != nil {
		log.Fatalf("child Munmap: %v", err)
	}
	if err := posix.Close(fd); err != nil {
		log.Fatalf("child Close: %v", err)
	}
}

// clen returns the length of the NUL-terminated prefix of b.
func clen(b []byte) int {
	for i, c := range b {
		if c == 0 {
			return i
		}
	}
	return len(b)
}
