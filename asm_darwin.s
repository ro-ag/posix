#include "textflag.h"

TEXT libc_shm_open_trampoline<>(SB),NOSPLIT,$0-0
	JMP	libc_shm_open(SB)

GLOBL	·libc_shm_open_trampoline_addr(SB), RODATA, $8
DATA	·libc_shm_open_trampoline_addr(SB)/8, $libc_shm_open_trampoline<>(SB)


TEXT libc_shm_unlink_trampoline<>(SB),NOSPLIT,$0-0
	JMP	libc_shm_unlink(SB)

GLOBL	·libc_shm_unlink_trampoline_addr(SB), RODATA, $8
DATA	·libc_shm_unlink_trampoline_addr(SB)/8, $libc_shm_unlink_trampoline<>(SB)

TEXT libc_fcntl_trampoline<>(SB),NOSPLIT,$0-0
	JMP	libc_fcntl(SB)

GLOBL	·libc_fcntl_trampoline_addr(SB), RODATA, $8
DATA	·libc_fcntl_trampoline_addr(SB)/8, $libc_fcntl_trampoline<>(SB)

TEXT libc_close_trampoline<>(SB),NOSPLIT,$0-0
	JMP	libc_close(SB)

GLOBL	·libc_close_trampoline_addr(SB), RODATA, $8
DATA	·libc_close_trampoline_addr(SB)/8, $libc_close_trampoline<>(SB)

TEXT libc_mmap_trampoline<>(SB),NOSPLIT,$0-0
	JMP	libc_mmap(SB)

GLOBL	·libc_mmap_trampoline_addr(SB), RODATA, $8
DATA	·libc_mmap_trampoline_addr(SB)/8, $libc_mmap_trampoline<>(SB)

TEXT libc_munmap_trampoline<>(SB),NOSPLIT,$0-0
	JMP	libc_munmap(SB)

GLOBL	·libc_munmap_trampoline_addr(SB), RODATA, $8
DATA	·libc_munmap_trampoline_addr(SB)/8, $libc_munmap_trampoline<>(SB)

TEXT libc_ftruncate_trampoline<>(SB),NOSPLIT,$0-0
	JMP	libc_ftruncate(SB)

GLOBL	·libc_ftruncate_trampoline_addr(SB), RODATA, $8
DATA	·libc_ftruncate_trampoline_addr(SB)/8, $libc_ftruncate_trampoline<>(SB)

TEXT libc_madvise_trampoline<>(SB),NOSPLIT,$0-0
	JMP	libc_madvise(SB)

GLOBL	·libc_madvise_trampoline_addr(SB), RODATA, $8
DATA	·libc_madvise_trampoline_addr(SB)/8, $libc_madvise_trampoline<>(SB)

TEXT libc_mlock_trampoline<>(SB),NOSPLIT,$0-0
	JMP	libc_mlock(SB)

GLOBL	·libc_mlock_trampoline_addr(SB), RODATA, $8
DATA	·libc_mlock_trampoline_addr(SB)/8, $libc_mlock_trampoline<>(SB)

TEXT libc_mlockall_trampoline<>(SB),NOSPLIT,$0-0
	JMP	libc_mlockall(SB)

GLOBL	·libc_mlockall_trampoline_addr(SB), RODATA, $8
DATA	·libc_mlockall_trampoline_addr(SB)/8, $libc_mlockall_trampoline<>(SB)

TEXT libc_mprotect_trampoline<>(SB),NOSPLIT,$0-0
	JMP	libc_mprotect(SB)

GLOBL	·libc_mprotect_trampoline_addr(SB), RODATA, $8
DATA	·libc_mprotect_trampoline_addr(SB)/8, $libc_mprotect_trampoline<>(SB)

TEXT libc_msync_trampoline<>(SB),NOSPLIT,$0-0
	JMP	libc_msync(SB)

GLOBL	·libc_msync_trampoline_addr(SB), RODATA, $8
DATA	·libc_msync_trampoline_addr(SB)/8, $libc_msync_trampoline<>(SB)

TEXT libc_munlock_trampoline<>(SB),NOSPLIT,$0-0
	JMP	libc_munlock(SB)

GLOBL	·libc_munlock_trampoline_addr(SB), RODATA, $8
DATA	·libc_munlock_trampoline_addr(SB)/8, $libc_munlock_trampoline<>(SB)

TEXT libc_munlockall_trampoline<>(SB),NOSPLIT,$0-0
	JMP	libc_munlockall(SB)

GLOBL	·libc_munlockall_trampoline_addr(SB), RODATA, $8
DATA	·libc_munlockall_trampoline_addr(SB)/8, $libc_munlockall_trampoline<>(SB)

TEXT libc_fchmod_trampoline<>(SB),NOSPLIT,$0-0
	JMP	libc_fchmod(SB)

GLOBL	·libc_fchmod_trampoline_addr(SB), RODATA, $8
DATA	·libc_fchmod_trampoline_addr(SB)/8, $libc_fchmod_trampoline<>(SB)

TEXT libc_fstat64_trampoline<>(SB),NOSPLIT,$0-0
	JMP	libc_fstat64(SB)

GLOBL	·libc_fstat64_trampoline_addr(SB), RODATA, $8
DATA	·libc_fstat64_trampoline_addr(SB)/8, $libc_fstat64_trampoline<>(SB)

TEXT libc_fchown_trampoline<>(SB),NOSPLIT,$0-0
	JMP	libc_fchown(SB)

GLOBL	·libc_fchown_trampoline_addr(SB), RODATA, $8
DATA	·libc_fchown_trampoline_addr(SB)/8, $libc_fchown_trampoline<>(SB)

