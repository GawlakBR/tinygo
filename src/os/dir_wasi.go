// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file was derived from src/os/dir_darwin.go since the logic for WASI is
// fairly similar: we use fdopendir, fdclosedir, and readdir from wasi-libc in
// a similar way that the darwin code uses functions from libc.

//go:build wasip1 || wasip2

package os

import (
	"internal/goarch"
	"io"
	"runtime"
	"sync"
	"syscall"
	"unsafe"
)

//// opaque DIR* returned by fdopendir
////
//// We add an unused field so it is not the empty struct, which is usually
//// a special case in Go.
//type dirInfo struct{ _ int32 }

// Auxiliary information if the File describes a directory
type dirInfo struct {
	mu   sync.Mutex
	buf  *[]byte // buffer for directory I/O
	nbuf int     // length of buf; return value from Getdirentries
	bufp int     // location of next record in buf.
}

const (
	// More than 5760 to work around https://golang.org/issue/24015.
	blockSize = 8192
)

var dirBufPool = sync.Pool{
	New: func() any {
		// The buffer must be at least a block long.
		buf := make([]byte, blockSize)
		return &buf
	},
}

func (d *dirInfo) close() {
	syscall.Fdclosedir(uintptr(unsafe.Pointer(d)))
}

func (f *File) readdir(n int, mode readdirMode) (names []string, dirents []DirEntry, infos []FileInfo, err error) {
	// If this file has no dirInfo, create one.
	d := f.dirinfo
	if d == nil {
		d = new(dirInfo)
		f.dirinfo = d
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.buf == nil {
		d.buf = dirBufPool.Get().(*[]byte)
	}

	// Change the meaning of n for the implementation below.
	//
	// The n above was for the public interface of "if n <= 0,
	// Readdir returns all the FileInfo from the directory in a
	// single slice".
	//
	// But below, we use only negative to mean looping until the
	// end and positive to mean bounded, with positive
	// terminating at 0.
	if n == 0 {
		n = -1
	}

	for n != 0 {
		// Refill the buffer if necessary
		if d.bufp >= d.nbuf {
			d.bufp = 0
			var errno error
			d.nbuf, errno = f.pfd.ReadDirent(*d.buf)
			runtime.KeepAlive(f)
			if errno != nil {
				return names, dirents, infos, &PathError{Op: "readdirent", Path: f.name, Err: errno}
			}
			if d.nbuf <= 0 {
				// Optimization: we can return the buffer to the pool, there is nothing else to read.
				dirBufPool.Put(d.buf)
				d.buf = nil
				break // EOF
			}
		}

		// Drain the buffer
		buf := (*d.buf)[d.bufp:d.nbuf]
		reclen, ok := direntReclen(buf)
		if !ok || reclen > uint64(len(buf)) {
			break
		}
		rec := buf[:reclen]
		d.bufp += int(reclen)
		ino, ok := direntIno(rec)
		if !ok {
			break
		}
		// When building to wasip1, the host runtime might be running on Windows
		// or might expose a remote file system which does not have the concept
		// of inodes. Therefore, we cannot make the assumption that it is safe
		// to skip entries with zero inodes.
		if ino == 0 && runtime.GOOS != "wasip1" {
			continue
		}
		const namoff = uint64(unsafe.Offsetof(syscall.Dirent{}.Name))
		namlen, ok := direntNamlen(rec)
		if !ok || namoff+namlen > uint64(len(rec)) {
			break
		}
		name := rec[namoff : namoff+namlen]
		for i, c := range name {
			if c == 0 {
				name = name[:i]
				break
			}
		}
		// Check for useless names before allocating a string.
		if string(name) == "." || string(name) == ".." {
			continue
		}
		if n > 0 { // see 'n == 0' comment above
			n--
		}
		if mode == readdirName {
			names = append(names, string(name))
		} else if mode == readdirDirEntry {
			de, err := newUnixDirent(f.name, string(name), direntType(rec))
			if IsNotExist(err) {
				// File disappeared between readdir and stat.
				// Treat as if it didn't exist.
				continue
			}
			if err != nil {
				return nil, dirents, nil, err
			}
			dirents = append(dirents, de)
		} else {
			info, err := lstat(f.name + "/" + string(name))
			if IsNotExist(err) {
				// File disappeared between readdir + stat.
				// Treat as if it didn't exist.
				continue
			}
			if err != nil {
				return nil, nil, infos, err
			}
			infos = append(infos, info)
		}
	}

	if n > 0 && len(names)+len(dirents)+len(infos) == 0 {
		return nil, nil, nil, io.EOF
	}
	return names, dirents, infos, nil
}

//func (f *File) readdir(n int, mode readdirMode) (names []string, dirents []DirEntry, infos []FileInfo, err error) {
//	if f.dirinfo == nil {
//		dir, errno := syscall.Fdopendir(syscallFd(f.handle.(unixFileHandle)))
//		if errno != nil {
//			return nil, nil, nil, &PathError{Op: "fdopendir", Path: f.name, Err: errno}
//		}
//		f.dirinfo = (*dirInfo)(unsafe.Pointer(dir))
//	}
//	d := uintptr(unsafe.Pointer(f.dirinfo))
//
//	// see src/os/dir_unix.go
//	if n == 0 {
//		n = -1
//	}
//
//	for n != 0 {
//		dirent, errno := syscall.Readdir(d)
//		if errno != nil {
//			if errno == syscall.EINTR {
//				continue
//			}
//			return names, dirents, infos, &PathError{Op: "readdir", Path: f.name, Err: errno}
//		}
//		if dirent == nil { // EOF
//			break
//		}
//		name := dirent.Name()
//		// Check for useless names before allocating a string.
//		if string(name) == "." || string(name) == ".." {
//			continue
//		}
//		if n > 0 {
//			n--
//		}
//		if mode == readdirName {
//			names = append(names, string(name))
//		} else if mode == readdirDirEntry {
//			de, err := newUnixDirent(f.name, string(name), dtToType(dirent.Type))
//			if IsNotExist(err) {
//				// File disappeared between readdir and stat.
//				// Treat as if it didn't exist.
//				continue
//			}
//			if err != nil {
//				return nil, dirents, nil, err
//			}
//			dirents = append(dirents, de)
//		} else {
//			info, err := lstat(f.name + "/" + string(name))
//			if IsNotExist(err) {
//				// File disappeared between readdir + stat.
//				// Treat as if it didn't exist.
//				continue
//			}
//			if err != nil {
//				return nil, nil, infos, err
//			}
//			infos = append(infos, info)
//		}
//		runtime.KeepAlive(f)
//	}
//
//	if n > 0 && len(names)+len(dirents)+len(infos) == 0 {
//		return nil, nil, nil, io.EOF
//	}
//	return names, dirents, infos, nil
//}

func dtToType(typ uint8) FileMode {
	switch typ {
	case syscall.DT_BLK:
		return ModeDevice
	case syscall.DT_CHR:
		return ModeDevice | ModeCharDevice
	case syscall.DT_DIR:
		return ModeDir
	case syscall.DT_FIFO:
		return ModeNamedPipe
	case syscall.DT_LNK:
		return ModeSymlink
	case syscall.DT_REG:
		return 0
	}
	return ^FileMode(0)
}

// readInt returns the size-bytes unsigned integer in native byte order at offset off.
func readInt(b []byte, off, size uintptr) (u uint64, ok bool) {
	if len(b) < int(off+size) {
		return 0, false
	}
	if goarch.BigEndian {
		return readIntBE(b[off:], size), true
	}
	return readIntLE(b[off:], size), true
}

func readIntBE(b []byte, size uintptr) uint64 {
	switch size {
	case 1:
		return uint64(b[0])
	case 2:
		return uint64(BEUint16(b))
	case 4:
		return uint64(BEUint32(b))
	case 8:
		return uint64(BEUint64(b))
	default:
		panic("syscall: readInt with unsupported size")
	}
}

func readIntLE(b []byte, size uintptr) uint64 {
	switch size {
	case 1:
		return uint64(b[0])
	case 2:
		return uint64(LEUint16(b))
	case 4:
		return uint64(LEUint32(b))
	case 8:
		return uint64(LEUint64(b))
	default:
		panic("syscall: readInt with unsupported size")
	}
}
