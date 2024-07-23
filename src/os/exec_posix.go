// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build aix || darwin || dragonfly || freebsd || (js && wasm) || linux || netbsd || openbsd || solaris || wasip1 || wasip2 || windows

package os

import (
	"runtime"
	"syscall"
	"unsafe"
)

// The only signal values guaranteed to be present in the os package on all
// systems are os.Interrupt (send the process an interrupt) and os.Kill (force
// the process to exit). On Windows, sending os.Interrupt to a process with
// os.Process.Signal is not implemented; it will return an error instead of
// sending a signal.
var (
	Interrupt Signal = syscall.SIGINT
	Kill      Signal = syscall.SIGKILL
)

// Keep compatible with golang and always succeed and return new proc with pid on Linux.
func findProcess(pid int) (*Process, error) {
	return &Process{Pid: pid}, nil
}

func (p *Process) release() error {
	// NOOP for unix.
	p.Pid = -1
	// no need for a finalizer anymore
	runtime.SetFinalizer(p, nil)
	return nil
}

// This function is a wrapper around the forkExec function, which is a wrapper around the fork and execve system calls.
// The StartProcess function creates a new process by forking the current process and then calling execve to replace the current process with the new process.
// It thereby replaces the newly created process with the specified command and arguments.
// Differences to upstream golang implementation (https://cs.opensource.google/go/go/+/master:src/syscall/exec_unix.go;l=143):
// * No setting of Process Attributes
// * Ignoring Ctty
// * No ForkLocking (might be introduced by #4273)
// * No parent-child communication via pipes (TODO)
// * No waiting for crashes child processes to prohibit zombie process accumulation / Wait status checking (TODO)
func forkExec(argv0 string, argv []string, attr *ProcAttr) (pid int, err error) {
	var (
		ret uintptr
	)

	argv0p, err := syscall.BytePtrFromString(argv0)
	if err != nil {
		return 0, err
	}
	argvp, err := syscall.SlicePtrFromStrings(argv)
	if err != nil {
		return 0, err
	}
	envvp, err := syscall.SlicePtrFromStrings(attr.Env)
	if err != nil {
		return 0, err
	}

	if (runtime.GOOS == "freebsd" || runtime.GOOS == "dragonfly") && len(argv) > 0 && len(argv[0]) > len(argv0) {
		argvp[0] = argv0p
	}

	ret, _, _ = syscall.Syscall(syscall.SYS_FORK, 0, 0, 0)
	if ret != 0 {
		// if fd == 0 code runs in parent
		return int(ret), nil
	} else {
		// else code runs in child, which then should exec the new process
		ret, _, _ = syscall.Syscall6(syscall.SYS_EXECVE, uintptr(unsafe.Pointer(argv0p)), uintptr(unsafe.Pointer(&argvp[0])), uintptr(unsafe.Pointer(&envvp[0])), 0, 0, 0)
		if ret != 0 {
			// exec failed
			syscall.Exit(1)
		}
		// 3. TODO: use pipes to communicate back child status
		return int(ret), nil
	}
}

// In Golang, the idiomatic way to create a new process is to use the StartProcess function.
// Since the Model of operating system processes in tinygo differs from the one in Golang, we need to implement the StartProcess function differently.
// The startProcess function is a wrapper around the forkExec function, which is a wrapper around the fork and execve system calls.
// The StartProcess function creates a new process by forking the current process and then calling execve to replace the current process with the new process.
// It thereby replaces the newly created process with the specified command and arguments.
func startProcess(name string, argv []string, attr *ProcAttr) (p *Process, err error) {
	pid, err := ForkExec(name, argv, attr)
	if err != nil {
		return nil, err
	}

	return findProcess(pid)
}
