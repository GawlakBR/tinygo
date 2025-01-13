//go:build scheduler.cores

package task

import (
	"runtime/interrupt"
	"unsafe"
)

import "C" // dummy import, to make sure task_cores.c is included in the build

type runState uint8

const (
	runStateRunning runState = iota
	runStateResuming
	runStatePaused
)

type state struct {
	// Which state the task is currently in.
	// The state is protected by the scheduler lock, and must only be
	// read/modified with that lock held.
	runState runState

	// The stack pointer while the task is switched away.
	sp unsafe.Pointer

	// canaryPtr points to the top word of the stack (the lowest address).
	// This is used to detect stack overflows.
	// When initializing the goroutine, the stackCanary constant is stored there.
	// If the stack overflowed, the word will likely no longer equal stackCanary.
	canaryPtr *uintptr
}

var (
	runQueue   *Task
	sleepQueue *Task
)

//go:linkname runtimeCurrentTask runtime.currentTask
func runtimeCurrentTask() *Task

// Current returns the current task, or nil if we're in the scheduler.
func Current() *Task {
	return runtimeCurrentTask()
}

func Init(mainTask *Task, canaryPtr *uintptr) {
	// The topmost word of the default stack is used as a stack canary.
	*canaryPtr = stackCanary
	mainTask.state.canaryPtr = canaryPtr
}

func Pause() {
	// Check whether the canary (the lowest address of the stack) is still
	// valid. If it is not, a stack overflow has occurred.
	if *Current().state.canaryPtr != stackCanary {
		runtimePanic("goroutine stack overflow")
	}
	if interrupt.In() {
		runtimePanic("blocked inside interrupt")
	}

	// Note: Pause() must be called with the scheduler lock locked!
	schedulerLock()
	pauseLocked()
	schedulerUnlock()
}

var schedulerIsRunning bool

func pauseLocked() {
	t := Current()
	for {
		if t.state.runState == runStateResuming {
			t.state.runState = runStateRunning
			return
		}

		// Make sure only one core is calling sleepTicks etc.
		if schedulerIsRunning {
			schedulerUnlock()
			waitForEvents()
			schedulerLock()
			continue
		}

		if runnable := runQueue; runnable != nil {
			// Resume it now.
			runQueue = runQueue.Next
			runnable.Next = nil
			if t == runnable {
				// We're actually the task that's supposed to be resumed, so we
				// are ready!
			} else {
				// It's not us that's ready, so switch to this other task.
				setCurrentTask(runnable)
				t.state.runState = runStatePaused

				// Switch away!
				switchTask(&t.state.sp, runnable.state.sp)

				// We got back from the switch, so another task resumed us.
				t.state.runState = runStateRunning
			}
			return
		}

		// Check whether there's a sleeping task that is ready to run.
		if sleepingTask := sleepQueue; sleepingTask != nil {
			now := runtimeTicks()
			if now >= sleepingTask.Data {
				// This task is done sleeping.
				// Resume it now.
				sleepQueue = sleepQueue.Next
				sleepingTask.Next = nil
				if t == sleepingTask {
					// We're actually the task that's sleeping, so we are ready!
				} else {
					// It's not us that's ready, so switch to this other task.
					setCurrentTask(sleepingTask)
					t.state.runState = runStatePaused

					// Switch away!
					switchTask(&t.state.sp, sleepingTask.state.sp)

					// We got back from the switch, so another task resumed us.
					t.state.runState = runStateRunning
				}
				return
			} else {
				// Sleep for a bit until the next task is ready to run.
				schedulerIsRunning = true
				schedulerUnlock()
				delay := sleepingTask.Data - now
				runtimeSleepTicks(delay)
				schedulerLock()
				schedulerIsRunning = false
				continue
			}
		}
	}
}

func start(fn uintptr, args unsafe.Pointer, stackSize uintptr) {
	t := &Task{}
	stack := runtime_alloc(stackSize, nil)
	stackTop := unsafe.Add(stack, stackSize-16)
	topRegs := unsafe.Slice((*uintptr)(stackTop), 4)
	topRegs[0] = uintptr(unsafe.Pointer(&startTask))
	topRegs[1] = uintptr(args)
	topRegs[2] = fn
	t.state.sp = stackTop

	canaryPtr := (*uintptr)(stack)
	*canaryPtr = stackCanary
	t.state.canaryPtr = canaryPtr

	schedulerLock()
	addToRunqueue(t)
	schedulerUnlock()
}

func GCScan() {
	panic("todo: task.GCScan")
}

func StackTop() uintptr {
	println("todo: task.StackTop")
	for {
	}
}

func Sleep(wakeup uint64) {
	schedulerLock()
	addSleepTask(Current(), wakeup)
	pauseLocked()
	schedulerUnlock()
}

func Resume(t *Task) {
	schedulerLock()
	switch t.state.runState {
	case runStatePaused:
		// Paused, state is saved on the stack.
		addToRunqueue(t)
	case runStateRunning:
		// Going to pause soon, so let the Pause() function know it can resume
		// immediately.
		t.state.runState = runStateResuming
	default:
		println("unknown run state??")
		for {
		}
	}
	schedulerUnlock()
}

// May only be called with the scheduler lock held!
func addToRunqueue(t *Task) {
	t.Next = runQueue
	runQueue = t
}

func addSleepTask(t *Task, wakeup uint64) {
	// Save the timestamp when the task should be woken up.
	t.Data = wakeup

	// Find the position where we should insert this task in the queue.
	q := &sleepQueue
	for {
		if *q == nil {
			// Found the end of the time queue. Insert it here, at the end.
			break
		}
		if (*q).Data > t.Data {
			// Found a task in the queue that has a timeout before the
			// to-be-sleeping task. Insert our task right before.
			break
		}
		q = &(*q).Next
	}

	// Insert the task into the queue (this could be at the end, if *q is nil).
	t.Next = *q
	*q = t
}

//go:linkname schedulerLock runtime.schedulerLock
func schedulerLock()

//go:linkname schedulerUnlock runtime.schedulerUnlock
func schedulerUnlock()

//go:linkname runtimeTicks runtime.runtimeTicks
func runtimeTicks() uint64

//go:linkname runtimeSleepTicks runtime.runtimeSleepTicks
func runtimeSleepTicks(duration uint64)

// startTask is a small wrapper function that sets up the first (and only)
// argument to the new goroutine and makes sure it is exited when the goroutine
// finishes.
//
//go:extern tinygo_cores_startTask
var startTask [0]uint8

//export tinygo_exitTask
func exitTask() {
	Pause()
}

//export tinygo_schedulerUnlock
func tinygo_schedulerUnlock() {
	schedulerUnlock()
}

//export tinygo_switchTask
func switchTask(oldStack *unsafe.Pointer, newStack unsafe.Pointer)

//go:linkname waitForEvents runtime.waitForEvents
func waitForEvents()

//go:linkname setCurrentTask runtime.setCurrentTask
func setCurrentTask(task *Task)
