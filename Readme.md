# Posix

This package contains the necessary functions for POSIX shared memory management, Linux and MacOSx The POSIX shared
memory API allows processes to communicate information by sharing a region of memory.

Golang doesn't have shm_open implementation and Mmap is not exposing the address parameter, needed for cross programming
maps.

> I highly recommend to use **golang.org/x/sys** (which this code is base from) if you don't require mmap with
fixed memory address or shm_open function.

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