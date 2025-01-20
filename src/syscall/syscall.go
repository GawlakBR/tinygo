package syscall

import (
	"errors"
)

func Exit(code int)

func Setrlimit(resource int, rlim *Rlimit) error {
	return errors.New("Setrlimit not implemented")
}
