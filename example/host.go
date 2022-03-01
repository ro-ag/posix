package main

import (
	"gopkg.in/ro-ag/posix.v0"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"
)

const ApproximatedAddress uintptr = 0x20000000000

type Head struct {
	Addr     unsafe.Pointer
	MemSize  uintptr
	TextSize uintptr
	TextPtr  unsafe.Pointer
}

type Holder struct {
	*Head
	Data []byte
}

const exeName = "./child_exe"

func main() {

	fd, err := posix.MemfdCreate("test-mem", posix.MFD_ALLOW_SEALING)
	CheckErr(err)

	MemSize := posix.Getpagesize() * 4
	err = posix.Ftruncate(fd, MemSize)
	CheckErr(err)

	buf, addr, err := posix.Mmap(unsafe.Pointer(ApproximatedAddress), MemSize, posix.PROT_WRITE, posix.MAP_SHARED, fd, 0)
	CheckErr(err)

	log.Printf("Got Address %p\n", unsafe.Pointer(addr))

	text := "This works as Mmap in C"
	offset := unsafe.Sizeof(Head{})

	hdr := (*Head)(unsafe.Pointer(&buf[0]))
	hdr.Addr = unsafe.Pointer(addr)
	hdr.MemSize = uintptr(MemSize)
	hdr.TextSize = uintptr(len(text))
	hdr.TextPtr = unsafe.Pointer(&buf[offset])

	holder := &Holder{
		Head: hdr,
		Data: unsafe.Slice(&buf[offset], uintptr(len(text))),
	}

	log.Println("bytes written:", copy(holder.Data, text))
	log.Println("buf: ", buf[:offset+holder.TextSize])
	log.Printf("%T = %+v", holder.Head, holder.Head)

	/* Execute External Program which needs a pointer */
	stdout, stderr := RunChild(exeName, uintptr(fd))
	log.Printf("stderr: %s", stderr)
	log.Printf("stdout: \n%s", stdout)

	CheckErr(posix.Munmap(buf)) // unmap memory
	CheckErr(posix.Close(fd))   // close anonymous file

	log.Println("- example done -")
}

func RunChild(name string, fd uintptr) (out, err string) {
	file := os.NewFile(fd, "anonymous-fd")
	cmd := exec.Command(name)
	cmd.ExtraFiles = []*os.File{file}
	var stderr, stdout strings.Builder
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	return stdout.String(), stderr.String()
}

func init() {
	log.Println("compile C program")
	compile := exec.Command("gcc", "-o", exeName, "_child.c")
	compile.Env = append(os.Environ())
	compile.Stderr = os.Stderr
	compile.Stdout = os.Stdout
	CheckErr(compile.Run())
	log.Println("compile success")
}

func CheckErr(err error) {
	if err != nil {
		no := err.(syscall.Errno)
		log.Fatalf("%s(%d): %v, msg: %s\nhelp: %s", posix.ErrnoName(no), no, err, posix.ErrnoString(no), posix.ErrnoHelp(no))
	}
}
