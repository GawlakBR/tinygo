// +build scheduler.tasks,tinygo.riscv

package task

import "unsafe"

const stackSize = 1024

// calleeSavedRegs is the list of registers that must be saved and restored when
// switching between tasks. Also see scheduler_riscv.S that relies on the
// exact layout of this struct.
type calleeSavedRegs struct {
	s0  uintptr // x8 (fp)
	s1  uintptr // x9
	s2  uintptr // x18
	s3  uintptr // x19
	s4  uintptr // x20
	s5  uintptr // x21
	s6  uintptr // x22
	s7  uintptr // x23
	s8  uintptr // x24
	s9  uintptr // x25
	s10 uintptr // x26
	s11 uintptr // x27

	pc uintptr
}

// registers gets a pointer to the registers stored at the top of the stack.
func (s *state) registers() *calleeSavedRegs {
	return (*calleeSavedRegs)(unsafe.Pointer(s.sp))
}

// startTask is a small wrapper function that sets up the first (and only)
// argument to the new goroutine and makes sure it is exited when the goroutine
// finishes.
//go:extern tinygo_startTask
var startTask [0]uint8

// archInit runs architecture-specific setup for the goroutine startup.
func (s *state) archInit(stack []uintptr, fn uintptr, args unsafe.Pointer) {
	// Set up the stack canary, a random number that should be checked when
	// switching from the task back to the scheduler. The stack canary pointer
	// points to the first word of the stack. If it has changed between now and
	// the next stack switch, there was a stack overflow.
	s.canaryPtr = &stack[0]
	*s.canaryPtr = stackCanary

	// Store the initial sp for the startTask function (implemented in assembly).
	s.sp = uintptr(unsafe.Pointer(&stack[uintptr(len(stack))-(unsafe.Sizeof(calleeSavedRegs{})/unsafe.Sizeof(uintptr(0)))]))

	// Initialize the registers.
	// These will be popped off of the stack on the first resume of the goroutine.
	r := s.registers()

	// Start the function at tinygo_startTask (defined in scheduler_riscv.S).
	// This assembly code calls a function (passed in s0) with a single argument
	// (passed in s1). After the function returns, it calls Pause().
	r.pc = uintptr(unsafe.Pointer(&startTask))

	// Pass the function to call in s0.
	// This function is a compiler-generated wrapper which loads arguments out
	// of a struct pointer. See createGoroutineStartWrapper (defined in
	// compiler/goroutine.go) for more information.
	r.s0 = fn

	// Pass the pointer to the arguments struct in s1.
	r.s1 = uintptr(args)
}

func (s *state) resume() {
	switchToTask(s.sp)
}

//export tinygo_switchToTask
func switchToTask(uintptr)

//export tinygo_switchToScheduler
func switchToScheduler(*uintptr)

func (s *state) pause() {
	switchToScheduler(&s.sp)
}

//export tinygo_pause
func pause() {
	Pause()
}
