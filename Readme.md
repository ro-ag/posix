# Posix

This package contains the necessary functions for POSIX shared memory management, Linux and MacOSx The POSIX shared
memory API allows processes to communicate information by sharing a region of memory.

Golang doesn't have shm_open implementation and Mmap is not exposing the address parameter, needed for cross programming
maps.

* Package Documentation https://pkg.go.dev/gopkg.in/ro-ag/posix.v1

> I highly recommend to use **golang.org/x/sys** (which this code is base from) if you don't require mmap with fixed memory address or shm_open function.

According to the Linux manual this package contains the following functions.

- **ShmOpen** Create and open a new object, or open an existing object. The call returns a file descriptor for use by
  the other interfaces listed below.


- **Ftruncate** Set the size of the shared memory object.  (A newly created shared memory object has a length of zero.)


- **Mmap** Map the shared memory object into the virtual address space of the calling process.


- **Munmap** Unmap the shared memory object from the virtual address space of the calling process.


- **ShmUnlink** Remove a shared memory object name.


- **Close** Close the file descriptor allocated by shm_open(3) when it is no longer needed.


- **Fstat** Obtain a stat structure that describes the shared memory object. Among the information returned by this call
  are the object's size (st_size), permissions (st_mode), owner (st_uid), and group (st_gid).


- **Fchown** To change the ownership of a shared memory object.


- **Fchmod** To change the permissions of a shared memory object.


- **MemfdCreate** This function is available since Linux 3.17, basically creates an anonymous memory file descriptor.

## Example

> host

```go
package main

import (
  "fmt"
  "gopkg.in/ro-ag/posix.v1"
  "log"
  "syscall"
  "unsafe"
)

const AppAddress uintptr = 0x20000000000 // Approximated Address 

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

func main() {
  // Create File Descriptor	
  fd, err := posix.MemfdCreate("From-Main", posix.MFD_ALLOW_SEALING)
  CheckErr(err)
  // Truncate File max len
  MemSize := posix.Getpagesize() * 4
  err = posix.Ftruncate(fd, MemSize)
  CheckErr(err)

  // mmap returns the map address from system call, 
  // because is not MAP_FIXED will return the closest memory if the address segment is used

  buf, addr, err := posix.Mmap(unsafe.Pointer(AppAddress), MemSize, posix.PROT_WRITE, posix.MAP_SHARED, fd, 0)
  CheckErr(err)

  text := "This works as Mmap in C"
  offset := unsafe.Sizeof(Head{})

  // Unsafe Structure, Easy for writing pointers
  hdr := (*Head)(unsafe.Pointer(&buf[0]))
  hdr.Addr = unsafe.Pointer(addr)
  hdr.MemSize = uintptr(MemSize)
  hdr.TextSize = uintptr(len(text))
  hdr.TextPtr = unsafe.Pointer(&buf[offset])

  holder := &Holder{
    Head: hdr,
    Data: unsafe.Slice(&buf[offset], uintptr(len(text))),
  }

  fmt.Println(holder)

  // Detailed code under ./example folder 

  CheckErr(posix.Munmap(buf)) // unmap memory
  CheckErr(posix.Close(fd))   // close anonymous file

}

// CheckErr wraps around error
func CheckErr(err error) {
  if err != nil {
    no := err.(syscall.Errno)
    log.Fatalf("%s(%d): %v, msg: %s\nhelp: %s", posix.ErrnoName(no), no, err, posix.ErrnoString(no), posix.ErrnoHelp(no))
  }
}
```