# posix

[![CI](https://github.com/ro-ag/posix/actions/workflows/ci.yml/badge.svg)](https://github.com/ro-ag/posix/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/gopkg.in/ro-ag/posix.v1.svg)](https://pkg.go.dev/gopkg.in/ro-ag/posix.v1)
[![Go Report Card](https://goreportcard.com/badge/gopkg.in/ro-ag/posix.v1)](https://goreportcard.com/report/gopkg.in/ro-ag/posix.v1)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)

**Pure-Go POSIX shared memory with fixed-address `mmap`, for Linux and macOS — no cgo.**

`posix` wraps the POSIX shared-memory and memory-mapping syscalls and exposes the
one knob the standard library and `golang.org/x/sys` keep hidden: the
**destination address** of an `mmap`. It needs no C compiler — `CGO_ENABLED=0`,
no `import "C"`.

```go
import "gopkg.in/ro-ag/posix.v1"
```

## Why it exists

- **Fixed-address `mmap`.** `Mmap` takes an `addr` argument. Pass it as a hint and
  the kernel maps near it; add `MAP_FIXED` and the region lands at *exactly* that
  virtual address. Map the same shared object at the same address in two
  processes and pointer-bearing structures in shared memory just work. The
  `Mmap` wrappers in `syscall` and `x/sys` don't expose `addr` at all.
- **No cgo.** The whole package builds with the cgo toolchain disabled (details
  below).

The combination — `shm_open` *and* fixed-address `mmap`, with no cgo — is rare. I
went looking for another Go package that does all three and couldn't find one;
if you know of one, I'd like to hear about it. (That's "the one I couldn't find,"
not "the only one that exists.")

## No cgo, precisely

Worth stating exactly, because "binds libc" usually implies cgo and here it
doesn't:

- `CGO_ENABLED=0`, and there is no `import "C"` anywhere in the package.
- On **Linux**, syscalls go through small hand-written amd64 assembly
  trampolines (`asm_linux_amd64.s`).
- On **macOS**, libc symbols (`shm_open`, `shm_unlink`, `madvise`, `mlock`, …)
  are bound with `//go:cgo_import_dynamic` — **the same mechanism
  `golang.org/x/sys` uses**. Despite the `cgo_` in the directive's name this is
  *not* cgo: no C compiler, no cgo preamble, and it still builds under
  `CGO_ENABLED=0`.

Check it yourself:

```sh
CGO_ENABLED=0 go build ./...
```

### Why no cgo is the point

- **Build speed** — no per-package C compiler invocation.
- **Static binaries** — `CGO_ENABLED=0` links a static binary with no libc
  dependency; it drops straight into a `scratch` or `distroless` image.
- **Clean cross-compilation** — `GOOS=… GOARCH=… go build` with no cross C
  toolchain to install.

## Install

```sh
go get gopkg.in/ro-ag/posix.v1
```

## Two processes sharing a region

A parent creates a named shared-memory object, writes a struct into it, and
re-executes itself as a separate child that opens the same object *by name* and
writes a reply back. The parent then reads that reply through the shared
mapping — a struct making the full round trip between two processes.

The complete, runnable program is in
[`example/roundtrip`](example/roundtrip/main.go). The essence is the symmetry
between the two processes:

```go
type payload struct { // no pointers: the bytes mean the same thing in both processes
	Magic    uint64
	Seq      uint64
	ChildPID int64
	Reply    [96]byte
}

// Parent — create, size, and map a named object, then write into it.
fd, _ := posix.ShmOpen("/demo", posix.O_RDWR|posix.O_CREAT|posix.O_EXCL, posix.S_IRUSR|posix.S_IWUSR)
posix.Ftruncate(fd, int(unsafe.Sizeof(payload{})))
buf, _, _ := posix.Mmap(nil, int(unsafe.Sizeof(payload{})), posix.PROT_RDWR, posix.MAP_SHARED, fd, 0)
p := (*payload)(unsafe.Pointer(&buf[0]))
p.Seq = 42

// Child (separate process) — open the same name and share the same bytes.
fd, _ := posix.ShmOpen("/demo", posix.O_RDWR, 0)
buf, _, _ := posix.Mmap(nil, size, posix.PROT_RDWR, posix.MAP_SHARED, fd, 0)
p := (*payload)(unsafe.Pointer(&buf[0])) // sees Seq == 42
p.Seq = 43                               // the parent sees this
```

Run it:

```sh
$ go run ./example/roundtrip
parent: mapped "/posix-rt-12345" at 0x20000000000 (hint 0x20000000000)
parent: wrote Seq=42
parent: read back Seq=43 ChildPID=12346 Reply="hello from child"
round-trip OK
```

## Supported platforms

| Target         | Status                                |
| -------------- | ------------------------------------- |
| `linux/amd64`  | ✅ built & tested in CI                |
| `linux/arm64`  | ✅ built & tested in CI                |
| `darwin/arm64` | ✅ built & tested in CI                |
| `darwin/amd64` | ✅ builds (CI's macOS runner is arm64) |

CI runs the full suite — including the cross-process round trip and a runtime
`Stat_t` ABI check — on `ubuntu-latest`, `ubuntu-24.04-arm`, and `macos-latest`.

## API

**Shared memory & files:** `ShmOpen`, `ShmUnlink`, `ShmAnonymous`, `Ftruncate`,
`Close`, `Fstat`, `Fchown`, `Fchmod`, `Fcntl`, `MemfdCreate`.

**Memory mapping:** `Mmap` (with `addr`), `Munmap`, `Mprotect`, `Msync`,
`Madvise`, `Mlock`, `Munlock`, `Mlockall`, `Munlockall`, `Getpagesize`.

`MemfdCreate` is native on Linux; on macOS it is emulated with `shm_open`
(convenient for portable tests). Full reference on
**[pkg.go.dev](https://pkg.go.dev/gopkg.in/ro-ag/posix.v1)**.

> If you don't need a fixed mmap address or `shm_open`, prefer
> [`golang.org/x/sys`](https://pkg.go.dev/golang.org/x/sys/unix) — this package's
> low-level plumbing is derived from it.

## License

[Apache-2.0](LICENSE). Portions of the syscall and mmap plumbing are derived from
the Go project and `golang.org/x/sys` (© The Go Authors, BSD-style license); see
[`THIRD-PARTY-NOTICES.txt`](THIRD-PARTY-NOTICES.txt).
