//go:build linux || unix

package syscall

const (
	MSG_DONTWAIT = 0x40
	AF_INET      = 0x2
	AF_INET6     = 0xa
)

type Rlimit struct {
	Cur uint64
	Max uint64
}

func Exec(argv0 string, argv []string, envv []string) (err error)

// The two SockaddrInet* structs have been copied from the Go source tree.

type SockaddrInet4 struct {
	Port int
	Addr [4]byte
	raw  RawSockaddrInet4
}

type SockaddrInet6 struct {
	Port   int
	ZoneId uint32
	Addr   [16]byte
	raw    RawSockaddrInet6
}
