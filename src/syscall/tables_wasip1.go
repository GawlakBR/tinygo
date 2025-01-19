// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build wasip1

package syscall

import "runtime"

// TODO: Auto-generate some day. (Hard-coded in binaries so not likely to change.)
var errorstr = [...]string{
	E2BIG:         "Argument list too long",
	EACCES:        "Permission denied",
	EADDRINUSE:    "Address already in use",
	EADDRNOTAVAIL: "Address not available",
	EAFNOSUPPORT:  "Address family not supported by protocol family",
	EAGAIN:        "Try again",
	EALREADY:      "Socket already connected",
	EBADF:         "Bad file number",
	//EBADFD:          "file descriptor in bad state",
	EBADMSG:         "Trying to read unreadable message",
	EBUSY:           "Device or resource busy",
	ECANCELED:       "Operation canceled.",
	ECHILD:          "No child processes",
	ECONNABORTED:    "Connection aborted",
	ECONNREFUSED:    "Connection refused",
	ECONNRESET:      "Connection reset by peer",
	EDEADLK:         "Deadlock condition",
	EDESTADDRREQ:    "Destination address required",
	EDOM:            "Math arg out of domain of func",
	EDQUOT:          "Quota exceeded",
	EEXIST:          "File exists",
	EFAULT:          "Bad address",
	EFBIG:           "File too large",
	EHOSTUNREACH:    "Host is unreachable",
	EIDRM:           "Identifier removed",
	EILSEQ:          "EILSEQ",
	EINPROGRESS:     "Connection already in progress",
	EINTR:           "Interrupted system call",
	EINVAL:          "Invalid argument",
	EIO:             "I/O error",
	EISCONN:         "Socket is already connected",
	EISDIR:          "Is a directory",
	ELOOP:           "Too many symbolic links",
	EMFILE:          "Too many open files",
	EMLINK:          "Too many links",
	EMSGSIZE:        "Message too long",
	EMULTIHOP:       "Multihop attempted",
	ENAMETOOLONG:    "File name too long",
	ENETDOWN:        "Network interface is not configured",
	ENETRESET:       "Network dropped connection on reset",
	ENETUNREACH:     "Network is unreachable",
	ENFILE:          "File table overflow",
	ENOBUFS:         "No buffer space available",
	ENODEV:          "No such device",
	ENOENT:          "No such file or directory",
	ENOEXEC:         "Exec format error",
	ENOLCK:          "No record locks available",
	ENOLINK:         "The link has been severed",
	ENOMEM:          "Out of memory",
	ENOMSG:          "No message of desired type",
	ENOPROTOOPT:     "Protocol not available",
	ENOSPC:          "No space left on device",
	ENOSYS:          "Not implemented on " + runtime.GOOS,
	ENOTCONN:        "Socket is not connected",
	ENOTDIR:         "Not a directory",
	ENOTEMPTY:       "Directory not empty",
	ENOTRECOVERABLE: "State not recoverable",
	ENOTSOCK:        "Socket operation on non-socket",
	ENOTSUP:         "Not supported",
	ENOTTY:          "Not a typewriter",
	ENXIO:           "No such device or address",
	EOVERFLOW:       "Value too large for defined data type",
	//EOWNERDEAD:      "Owner died",
	EPERM:           "Operation not permitted",
	EPIPE:           "Broken pipe",
	EPROTO:          "Protocol error",
	EPROTONOSUPPORT: "Unknown protocol",
	EPROTOTYPE:      "Protocol wrong type for socket",
	ERANGE:          "Math result not representable",
	EROFS:           "Read-only file system",
	ESPIPE:          "Illegal seek",
	ESRCH:           "No such process",
	ESTALE:          "Stale file handle",
	ETIMEDOUT:       "Connection timed out",
	ETXTBSY:         "Text file busy",
	EXDEV:           "Cross-device link",
	ENOTCAPABLE:     "Capabilities insufficient",
}

// Do the interface allocations only once for common
// Errno values.
var (
	errEAGAIN error = EAGAIN
	errEINVAL error = EINVAL
	errENOENT error = ENOENT
)

// errnoErr returns common boxed Errno values, to prevent
// allocations at runtime.
//
// We set both noinline and nosplit to reduce code size, this function has many
// call sites in the syscall package, inlining it causes a significant increase
// of the compiled code; the function call ultimately does not make a difference
// in the performance of syscall functions since the time is dominated by calls
// to the imports and path resolution.
//
//go:noinline
//go:nosplit
func errnoErr(e Errno) error {
	switch e {
	case 0:
		return nil
	case EAGAIN:
		return errEAGAIN
	case EINVAL:
		return errEINVAL
	case ENOENT:
		return errENOENT
	}
	return e
}
