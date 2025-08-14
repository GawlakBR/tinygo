package os

// Filesystem provides an interface for generic filesystem drivers mounted in
// the os package. The errors returned must be one of the os.Err* errors, or a
// custom error if one doesn't exist. It should not be a *PathError because
// errors will be wrapped with a *PathError by the filesystem abstraction.
//
// WARNING: this interface is not finalized and may change in a future version.
type Filesystem interface {
	// OpenFile opens the named file.
	OpenFile(name string, flag int, perm FileMode) (uintptr, error)

	// Mkdir creates a new directory with the specified permission (before
	// umask). Some filesystems may not support directories or permissions.
	Mkdir(name string, perm FileMode) error

	// Remove removes the named file or (empty) directory.
	Remove(name string) error
}

// FileHandle is an interface that should be implemented by filesystems
// implementing the Filesystem interface.
//
// WARNING: this interface is not finalized and may change in a future version.
type FileHandle interface {
	// Read reads up to len(b) bytes from the file.
	Read(b []byte) (n int, err error)

	// ReadAt reads up to len(b) bytes from the file starting at the given absolute offset
	ReadAt(b []byte, offset int64) (n int, err error)

	// Seek resets the file pointer relative to start, current position, or end
	Seek(offset int64, whence int) (newoffset int64, err error)

	// Sync blocks until buffered writes have been written to persistent storage
	Sync() (err error)

	// Write writes up to len(b) bytes to the file.
	Write(b []byte) (n int, err error)

	// WriteAt writes b to the file at the given absolute offset
	WriteAt(b []byte, offset int64) (n int, err error)

	// Close closes the file, making it unusable for further writes.
	Close() (err error)
}

type dummyFilesystem struct {
}

func (fs dummyFilesystem) Mkdir(path string, perm FileMode) error {
	return ErrNotImplemented
}

func (fs dummyFilesystem) Remove(path string) error {
	return ErrNotImplemented
}

func (fs dummyFilesystem) OpenFile(path string, flag int, perm FileMode) (uintptr, error) {
	return 0, ErrNotImplemented
}
