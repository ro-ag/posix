//go:build darwin || linux
// +build darwin linux

package posix

import (
	"strings"
	"syscall"
)

//2.2 Error Codes
//The error code macros are defined in the header file errno.h. All of them expand into integer constant values. Some of these error codes can' t occur on GNU systems, but they can occur using the GNU C Library on other systems.

// ErrnoName returns the error name for error number e.
func ErrnoName(e syscall.Errno) string {

	if index := getIndex(e); index != -1 {
		return errDescription[index].Name
	}
	return ""
}

// ErrnoString returns the message for error number e.
func ErrnoString(e syscall.Errno) string {
	if index := getIndex(e); index != -1 {
		return errDescription[index].Text
	}
	return ""
}

// ErrnoHelp returns descriptive messaging for error number e.
func ErrnoHelp(e syscall.Errno) string {
	if index := getIndex(e); index != -1 {
		return wrap(errDescription[index].Helper, 78)
	}
	return ""
}

var errDescription = [...]struct {
	Num    syscall.Errno
	Name   string
	Text   string
	Helper string
}{
	{syscall.EPERM, "EPERM", "Operation not permitted.", "Only the owner of the file (or other resource) or processes with special privileges can perform the operation."},
	{syscall.ENOENT, "ENOENT", "No such file or directory.", "This is a “file doesn't exist” error for ordinary files that are referenced in contexts where they are expected to already exist."},
	{syscall.ESRCH, "ESRCH", "No such process.", "No process matches the specified process ID."},
	{syscall.EINTR, "EINTR", "Interrupted system call.", "An asynchronous signal occurred and prevented completion of the call. When this happens, you should try the call again."},
	// You can choose to have functions resume after a signal that is handled, rather than failing with EINTR; see Interrupted Primitives.
	{syscall.EIO, "EIO", "Input/output error.", "Usually used for physical read or write errors."},
	{syscall.ENXIO, "ENXIO", "No such device or address.", "The system tried to use the device represented by a file you specified, and it couldn't find the device. This can mean that the device file was installed incorrectly, or that the physical device is missing or not correctly attached to the computer."},
	{syscall.E2BIG, "E2BIG", "Argument list too long.", "Used when the arguments passed to a new program being executed with one of the exec functions (see Executing a File) occupy too much memory space. This condition never arises on GNU/Hurd systems."},
	{syscall.ENOEXEC, "ENOEXEC", "Exec format error.", "Invalid executable file format. This condition is detected by the exec functions; see Executing a File."},
	{syscall.EBADF, "EBADF", "Bad file descriptor.", "For example, I/O on a descriptor that has been closed or reading from a descriptor open only for writing (or vice versa)."},
	{syscall.ECHILD, "ECHILD", "No child processes.", "This error happens on operations that are supposed to manipulate child processes, when there aren' t any processes to manipulate."},
	{syscall.EDEADLK, "EDEADLK", "Resource deadlock avoided.", "Allocating a system resource would have resulted in a deadlock situation. The system does not guarantee that it will notice all such situations. This error means you got lucky and the system noticed; it might just hang. See File Locks, for an example."},
	{syscall.ENOMEM, "ENOMEM", "Cannot allocate memory.", "The system cannot allocate more virtual memory because its capacity is full." + "- mlockall: (Linux 2.6.9 and later) the caller had a nonzero\n" +
		"RLIMIT_MEMLOCK soft resource limit, but tried to lock more\n" +
		"memory than the limit permitted.  This limit is not\n" +
		"enforced if the process is privileged (CAP_IPC_LOCK)."},
	{syscall.EACCES, "EACCES", "Permission denied.", "The file permissions do not allow the attempted operation."},
	{syscall.EFAULT, "EFAULT", "Bad address.", "An invalid pointer was detected. On GNU/Hurd systems, this error never happens; you get a signal instead."},
	{syscall.ENOTBLK, "ENOTBLK", "Block device required.", "A file that isn' t a block special file was given in a situation that requires one. For example, trying to mount an ordinary file as a file system in Unix gives this error."},
	{syscall.EBUSY, "EBUSY", "Device or resource busy.", "A system resource that can' t be shared is already in use. For example, if you try to delete a file that is the root of a currently mounted filesystem, you get this error."},
	{syscall.EEXIST, "EEXIST", "File exists.", "An existing file was specified in a context where it only makes sense to specify a new file."},
	{syscall.EXDEV, "EXDEV", "Invalid cross-device link.", "An attempt to make an improper link across file systems was detected. This happens not only when you use link (see Hard Links) but also when you rename a file with rename (see Renaming Files)."},
	{syscall.ENODEV, "ENODEV", "No such device.", "The wrong type of device was given to a function that expects a particular sort of device."},
	{syscall.ENOTDIR, "ENOTDIR", "Not a directory.", "A file that isn' t a directory was specified when a directory is required."},
	{syscall.EISDIR, "EISDIR", "Is a directory.", "You cannot open a directory for writing, or create or remove hard links to it."},
	{syscall.EINVAL, "EINVAL", "Invalid argument.", "This is used to indicate various kinds of problems with passing the wrong argument to a library function."},
	{syscall.EMFILE, "EMFILE", "Too many open files.", "The current process has too many files open and can' t open any more. Duplicate descriptors do count toward this limit.\n" +
		"In BSD and GNU, the number of open files is controlled by a resource limit that can usually be increased. If you get this error, you might want to increase the RLIMIT_NOFILE limit or make it unlimited; see Limits on Resources."},
	{syscall.ENFILE, "ENFILE", "Too many open files in system.", "There are too many distinct file openings in the entire system. Note that any number of linked channels count as just one file opening; see Linked Channels. This error never occurs on GNU/Hurd systems."},
	{syscall.ENOTTY, "ENOTTY", "Inappropriate ioctl for device.", "Inappropriate I/O control operation, such as trying to set terminal modes on an ordinary file."},
	{syscall.ETXTBSY, "ETXTBSY", "Text file busy.", "An attempt to execute a file that is currently open for writing, or write to a file that is currently being executed. Often using a debugger to run a program is considered having it open for writing and will cause this error. (The name stands for “text file busy”.) This is not an error on GNU/Hurd systems; the text is copied as necessary."},
	{syscall.EFBIG, "EFBIG", "File too large.", "The size of a file would be larger than allowed by the system."},
	{syscall.ENOSPC, "ENOSPC", "No space left on device.", "Write operation on a file failed because the disk is full."},
	{syscall.ESPIPE, "ESPIPE", "Illegal seek.", "Invalid seek operation (such as on a pipe)."},
	{syscall.EROFS, "EROFS", "Read-only file system.", "An attempt was made to modify something on a read-only file system."},
	{syscall.EMLINK, "EMLINK", "Too many links.", "The link count of a single file would become too large. rename can cause this error if the file being renamed already has as many links as it can take (see Renaming Files)."},
	{syscall.EPIPE, "EPIPE", "Broken pipe.", "There is no process reading from the other end of a pipe. Every library function that returns this error code also generates a SIGPIPE signal; this signal terminates the program if not handled or blocked. Thus, your program will never actually see EPIPE unless it has handled or blocked SIGPIPE."},
	{syscall.EDOM, "EDOM", "Numerical argument out of domain.", "Used by mathematical functions when an argument value does not fall into the domain over which the function is defined."},
	{syscall.ERANGE, "ERANGE", "Numerical result out of range.", "Used by mathematical functions when the result value is not representable because of overflow or underflow."},
	{syscall.EAGAIN, "EAGAIN", "Resource temporarily unavailable.", "The call might work if you try again later. The macro EWOULDBLOCK is another name for EAGAIN; they are always the same in the GNU C Library." +
		"This error can happen in a few different situations:\n" +
		"An operation that would block was attempted on an object that has non-blocking mode selected. Trying the same operation again will block until some external condition makes it possible to read, write, or connect (whatever the operation). You can use select to find out when the operation will be possible; see Waiting for I/O.\n" +
		"Portability Note: In many older Unix systems, this condition was indicated by EWOULDBLOCK, which was a distinct error code different from EAGAIN. To make your program portable, you should check for both codes and treat them the same.\n" +
		"A temporary resource shortage made an operation impossible. fork can return this error. It indicates that the shortage is expected to pass, so your program can try the call again later and it may succeed. It is probably a good idea to delay for a few seconds before trying it again, to allow time for other processes to release scarce resources. Such shortages are usually fairly serious and affect the whole system, so usually an interactive program should report the error to the user and return to its command loop."},
	{syscall.EWOULDBLOCK, "EWOULDBLOCK", "Operation would block.", "In the GNU C Library, this is another name for EAGAIN (above). The values are always the same, on every operating system." +
		"C libraries in many older Unix systems have EWOULDBLOCK as a separate error code."},
	{syscall.EINPROGRESS, "EINPROGRESS", "Operation now in progress.", "An operation that cannot complete immediately was initiated on an object that has non-blocking mode selected. Some functions that must always block (such as connect; see Connecting) never return EAGAIN. Instead, they return EINPROGRESS to indicate that the operation has begun and will take some time. Attempts to manipulate the object before the call completes return EALREADY. You can use the select function to find out when the pending operation has completed; see Waiting for I/O."},
	{syscall.EALREADY, "EALREADY", "Operation already in progress.", "An operation is already in progress on an object that has non-blocking mode selected."},
	{syscall.ENOTSOCK, "ENOTSOCK", "Socket operation on non-socket.", "A file that isn' t a socket was specified when a socket is required."},
	{syscall.EMSGSIZE, "EMSGSIZE", "Message too long.", "The size of a message sent on a socket was larger than the supported maximum size."},
	{syscall.EPROTOTYPE, "EPROTOTYPE", "Protocol wrong type for socket.", "The socket type does not support the requested communications protocol."},
	{syscall.ENOPROTOOPT, "ENOPROTOOPT", "Protocol not available.", "You specified a socket option that doesn't make sense for the particular protocol being used by the socket. See Socket Options."},
	{syscall.EPROTONOSUPPORT, "EPROTONOSUPPORT", "Protocol not supported.", "The socket domain does not support the requested communications protocol (perhaps because the requested protocol is completely invalid). See Creating a Socket."},
	{syscall.ESOCKTNOSUPPORT, "ESOCKTNOSUPPORT", "Socket type not supported.", "The socket type is not supported."},
	{syscall.EOPNOTSUPP, "EOPNOTSUPP", "Operation not supported.", "The operation you requested is not supported. Some socket functions don' t make sense for all types of sockets, and others may not be implemented for all communications protocols. On GNU/Hurd systems, this error can happen for many calls when the object does not support the particular operation; it is a generic indication that the server knows nothing to do for that call."},
	{syscall.EPFNOSUPPORT, "EPFNOSUPPORT", "Protocol family not supported.", "The socket communications protocol family you requested is not supported."},
	{syscall.EAFNOSUPPORT, "EAFNOSUPPORT", "Address family not supported by protocol.", "The address family specified for a socket is not supported; it is inconsistent with the protocol being used on the socket. See Sockets."},
	{syscall.EADDRINUSE, "EADDRINUSE", "Address already in use.", "The requested socket address is already in use. See Socket Addresses."},
	{syscall.EADDRNOTAVAIL, "EADDRNOTAVAIL", "Cannot assign requested address.", "The requested socket address is not available; for example, you tried to give a socket a name that doesn't match the local host name. See Socket Addresses."},
	{syscall.ENETDOWN, "ENETDOWN", "Network is down.", "A socket operation failed because the network was down."},
	{syscall.ENETUNREACH, "ENETUNREACH", "Network is unreachable.", "A socket operation failed because the subnet containing the remote host was unreachable."},
	{syscall.ENETRESET, "ENETRESET", "Network dropped connection on reset.", "A network connection was reset because the remote host crashed."},
	{syscall.ECONNABORTED, "ECONNABORTED", "Software caused connection abort.", "A network connection was aborted locally."},
	{syscall.ECONNRESET, "ECONNRESET", "Connection reset by peer.", "A network connection was closed for reasons outside the control of the local host, such as by the remote machine rebooting or an unrecoverable protocol violation."},
	{syscall.ENOBUFS, "ENOBUFS", "No buffer space available.", "The kernel' s buffers for I/O operations are all in use. In GNU, this error is always synonymous with ENOMEM; you may get one or the other from network operations."},
	{syscall.EISCONN, "EISCONN", "Transport endpoint is already connected.", "You tried to connect a socket that is already connected. See Connecting."},
	{syscall.ENOTCONN, "ENOTCONN", "Transport endpoint is not connected.", "The socket is not connected to anything. You get this error when you try to transmit data over a socket, without first specifying a destination for the data. For a connectionless socket (for datagram protocols, such as UDP), you get EDESTADDRREQ instead."},
	{syscall.EDESTADDRREQ, "EDESTADDRREQ", "Destination address required.", "No default destination address was set for the socket. You get this error when you try to transmit data over a connectionless socket, without first specifying a destination for the data with connect."},
	{syscall.ESHUTDOWN, "ESHUTDOWN", "Cannot send after transport endpoint shutdown.", "The socket has already been shut down."},
	{syscall.ETOOMANYREFS, "ETOOMANYREFS", "Too many references: cannot splice.", ""},
	{syscall.ETIMEDOUT, "ETIMEDOUT", "“Connection timed out.” A socket operation with a specified timeout received no response during the timeout period.", ""},
	{syscall.ECONNREFUSED, "ECONNREFUSED", "Connection refused.", "A remote host refused to allow the network connection (typically because it is not running the requested service)."},
	{syscall.ELOOP, "ELOOP", "Too many levels of symbolic links.", "Too many levels of symbolic links were encountered in looking up a file name. This often indicates a cycle of symbolic links."},
	{syscall.ENAMETOOLONG, "ENAMETOOLONG", "File name too long.", "Filename too long (longer than PATH_MAX; see Limits for Files) or host name too long (in gethostname or sethostname; see Host Identification)."},
	{syscall.EHOSTDOWN, "EHOSTDOWN", "Host is down.", "The remote host for a requested network connection is down."},
	{syscall.EHOSTUNREACH, "EHOSTUNREACH", "No route to host.", "The remote host for a requested network connection is not reachable."},
	{syscall.ENOTEMPTY, "ENOTEMPTY", "Directory not empty.", "Directory not empty, where an empty directory was expected. Typically, this error occurs when you are trying to delete a directory."},
	{syscall.EUSERS, "EUSERS", "Too many users.", "The file quota system is confused because there are too many users."},
	{syscall.EDQUOT, "EDQUOT", "Disk quota exceeded.", "The user' s disk quota was exceeded."},
	{syscall.ESTALE, "ESTALE", "Stale file handle.", "This indicates an internal confusion in the file system which is due to file system rearrangements on the server host for NFS file systems or corruption in other file systems. Repairing this condition usually requires unmounting, possibly repairing and remounting the file system."},
	{syscall.EREMOTE, "EREMOTE", "Object is remote.", "An attempt was made to NFS-mount a remote file system with a file name that already specifies an NFS-mounted file. (This is an error on some operating systems, but we expect it to work properly on GNU/Hurd systems, making this error code impossible.)"},
	{syscall.ENOLCK, "ENOLCK", "“No locks available.” This is used by the file locking facilities; see File Locks. This error is never generated by GNU/Hurd systems, but it can result from an operation to an NFS server running another operating system.", ""},
	{syscall.ENOSYS, "ENOSYS", "Function not implemented.", "This indicates that the function called is not implemented at all, either in the C library itself or in the operating system. When you get this error, you can be sure that this particular function will always fail with ENOSYS unless you install a new version of the C library or the operating system."},
	{syscall.ENOTSUP, "ENOTSUP", "Not supported.", "A function returns this error when certain parameter values are valid, but the functionality they request is not available. This can mean that the function does not implement a particular command or option value or flag bit at all. For functions that operate on some object given in a parameter, such as a file descriptor or a port, it might instead mean that only that specific object (file descriptor, port, etc.) is unable to support the other parameters given; different file descriptors might support different ranges of parameter values." +
		"If the entire function is not available at all in the implementation, it returns ENOSYS instead."},
	{syscall.EILSEQ, "EILSEQ", "Invalid or incomplete multibyte or wide character.", "While decoding a multibyte character the function came along an invalid or an incomplete sequence of bytes or the given wide character is invalid."},
	{syscall.EBADMSG, "EBADMSG", "Bad message.", ""},
	{syscall.EIDRM, "EIDRM", "Identifier removed.", ""},
	{syscall.EMULTIHOP, "EMULTIHOP", "Multihop attempted.", ""},
	{syscall.ENODATA, "ENODATA", "“No data available.”", ""},
	{syscall.ENOLINK, "ENOLINK", "Link has been severed.", ""},
	{syscall.ENOMSG, "ENOMSG", "“No message of desired type.”", ""},
	{syscall.ENOSR, "ENOSR", "Out of streams resources.", ""},
	{syscall.ENOSTR, "ENOSTR", "“Device not a stream.”", ""},
	{syscall.EOVERFLOW, "EOVERFLOW", "Value too large for defined data type.", ""},
	{syscall.EPROTO, "EPROTO", "“Protocol error.”", ""},
	{syscall.ETIME, "ETIME", "Timer expired.", ""},
	{syscall.ECANCELED, "ECANCELED", "“Operation canceled.” An asynchronous operation was canceled before it completed. See Asynchronous I/O. When you call aio_cancel, the normal result is for the operations affected to complete with this error; see Cancel AIO Operations.", ""},
	{syscall.EOWNERDEAD, "EOWNERDEAD", "Owner died.", ""},
	{syscall.ENOTRECOVERABLE, "ENOTRECOVERABLE", "“State not recoverable.”", ""},
}

func wrap(text string, lineWidth int) string {
	words := strings.Fields(strings.TrimSpace(text))
	if len(words) == 0 {
		return text
	}
	wrapped := words[0]
	spaceLeft := lineWidth - len(wrapped)
	for _, word := range words[1:] {
		if len(word)+1 > spaceLeft {
			wrapped += "\n" + word
			spaceLeft = lineWidth - len(word)
		} else {
			wrapped += " " + word
			spaceLeft -= 1 + len(word)
		}
	}
	return wrapped
}

func getIndex(e syscall.Errno) int {
	for i := range errDescription {
		if errDescription[i].Num == e {
			return i
		}
	}
	return -1
}
