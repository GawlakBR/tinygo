//go:build scheduler.cores

package task

import "runtime/interrupt"

// A futex is a way for userspace to wait with the pointer as the key, and for
// another thread to wake one or all waiting threads keyed on the same pointer.
//
// A futex does not change the underlying value, it only reads it before to prevent
// lost wake-ups.
type Futex struct {
	Uint32

	waiters Stack
}

// Atomically check for cmp to still be equal to the futex value and if so, go
// to sleep. Return true if we were definitely awoken by a call to Wake or
// WakeAll, and false if we can't be sure of that.
func (f *Futex) Wait(cmp uint32) (awoken bool) {
	mask := futexLock()

	if f.Uint32.Load() != cmp {
		futexUnlock(mask)
		return false
	}

	// Push the current goroutine onto the waiter stack.
	f.waiters.Push(Current())

	futexUnlock(mask)

	// Pause until this task is awoken by Wake/WakeAll.
	Pause()

	// We were awoken by a call to Wake or WakeAll. There is no chance for
	// spurious wakeups.
	return true
}

// Wake a single waiter.
func (f *Futex) Wake() {
	mask := futexLock()
	if t := f.waiters.Pop(); t != nil {
		scheduleTask(t)
	}
	futexUnlock(mask)
}

// Wake all waiters.
func (f *Futex) WakeAll() {
	mask := futexLock()
	for t := f.waiters.Pop(); t != nil; t = f.waiters.Pop() {
		scheduleTask(t)
	}
	futexUnlock(mask)
}

//go:linkname futexLock runtime.futexLock
func futexLock() interrupt.State

//go:linkname futexUnlock runtime.futexUnlock
func futexUnlock(interrupt.State)
