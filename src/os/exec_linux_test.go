//go: build linux && !baremetal && !tinygo.wasm

package os_test

import (
	"errors"
	. "os"
	"runtime"
	"testing"
)

// Test the functionality of the forkExec function, which is used to fork and exec a new process.
// This test is not run on Windows, as forkExec is not supported on Windows.
// This test is not run on Plan 9, as forkExec is not supported on Plan 9.
func TestForkExec(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Logf("skipping test on %s", runtime.GOOS)
		return
	}

	proc, err := StartProcess("echo", []string{"/bin/echo", "hello", "world"}, &ProcAttr{})
	if !errors.Is(err, nil) {
		t.Fatalf("forkExec failed: %v", err)
	}

	if proc.Pid == 0 {
		t.Fatalf("forkExec failed: new process has pid 0")
	}
	t.Logf("forkExec succeeded: new process has pid %d", proc)

	proc, err = StartProcess("invalid", []string{"/bin/nonexistent"}, &ProcAttr{})
	if errors.Is(err, nil) {
		t.Fatalf("wanted err, got nil")
	}

	if proc.Pid != 0 {
		t.Fatalf("wanted 0, got %v", proc.Pid)
	}
}
