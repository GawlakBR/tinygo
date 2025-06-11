//go:build scheduler.cores

package runtime

import (
	"internal/task"
	"sync/atomic"
)

// Normally 0. During a GC scan it has various purposes for signalling between
// the core running the GC and the other cores in the system.
var gcScanState atomic.Uint32

// Start GC scan by pausing the world (all other cores) and scanning their
// stacks. It doesn't resume the world.
func gcMarkReachable() {
	core := currentCPU()

	// Interrupt all other cores.
	gcScanState.Store(1)
	for i := uint32(0); i < numCPU; i++ {
		if i == core {
			continue
		}
		gcPauseCore(i)
	}

	// Scan the stack(s) of the current core.
	scanCurrentStack()
	if !task.OnSystemStack() {
		// Mark system stack.
		markRoots(task.SystemStack(), coreStackTop(core))
	}

	// Scan globals.
	findGlobals(markRoots)

	// Busy-wait until all the other cores are ready. They certainly should be,
	// after the scanning we did above.
	for gcScanState.Load() != numCPU {
		spinLoopHint()
	}
	gcScanState.Store(0)

	// Signal each core in turn that they can scan the stack.
	for i := uint32(0); i < numCPU; i++ {
		if i == core {
			continue
		}

		// Wake up the core to scan the stack.
		gcSignalCore(i)

		// Busy-wait until this core finished scanning.
		for gcScanState.Load() == 0 {
			spinLoopHint()
		}
		gcScanState.Store(0)
	}

	// All the stack are now scanned.
}

//go:export tinygo_scanCurrentStack
func scanCurrentStack()

//go:export tinygo_scanstack
func scanstack(sp uintptr) {
	// Mark the current stack.
	// This function is called by scanCurrentStack, after pushing all registers
	// onto the stack.
	if task.OnSystemStack() {
		// This is the system stack.
		// Scan all words on the stack.
		markRoots(sp, coreStackTop(currentCPU()))
	} else {
		// This is a goroutine stack.
		markCurrentGoroutineStack(sp)
	}
}

// Resume the world after a call to gcMarkReachable.
func gcResumeWorld() {
	// Signal each core that they can resume.
	hartID := currentCPU()
	for i := uint32(0); i < numCPU; i++ {
		if i == hartID {
			continue
		}

		// Signal the core.
		gcSignalCore(i)
	}

	// Busy-wait until the core acknowledges the signal (and is going to return
	// from the interrupt handler).
	for gcScanState.Load() != numCPU-1 {
		spinLoopHint()
	}
	gcScanState.Store(0)
}
