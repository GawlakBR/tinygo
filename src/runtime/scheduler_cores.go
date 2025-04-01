//go:build scheduler.cores

package runtime

import (
	"device/arm"
	"internal/task"
)

const hasScheduler = true

const hasParallelism = true

const coresVerbose = false

var (
	mainTask   task.Task
	cpuTasks   [numCPU]*task.Task
	sleepQueue *task.Task
	runQueue   *task.Task
)

func deadlock() {
	// Call yield without requesting a wakeup.
	task.Pause()
	trap()
}

func scheduleTask(t *task.Task) {
	schedulerLock()
	switch t.RunState {
	case task.RunStatePaused:
		// Paused, state is saved on the stack.

		if coresVerbose {
			println("## schedule: add to runQueue")
		}
		addToRunQueue(t)
		arm.Asm("sev")
	case task.RunStateRunning:
		// Not yet paused (probably going to pause very soon), so let the
		// Pause() function know it can resume immediately.
		t.RunState = task.RunStateResuming
		if coresVerbose {
			println("## schedule: mark as resuming")
		}
	default:
		println("Unknown run state??")
		for {
		}
	}
	schedulerUnlock()
}

// Add task to runQueue.
// Scheduler lock must be held when calling this function.
func addToRunQueue(t *task.Task) {
	t.Next = runQueue
	runQueue = t
}

func addSleepTask(t *task.Task, wakeup timeUnit) {
	// Save the timestamp when the task should be woken up.
	t.Data = uint64(wakeup)

	// Find the position where we should insert this task in the queue.
	q := &sleepQueue
	for {
		if *q == nil {
			// Found the end of the time queue. Insert it here, at the end.
			break
		}
		if timeUnit((*q).Data) > timeUnit(t.Data) {
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

func Gosched() {
	// TODO
}

// NumCPU returns the number of logical CPUs usable by the current process.
func NumCPU() int {
	// Return the hardcoded number of physical CPU cores.
	return numCPU
}

func addTimer(tn *timerNode) {
	runtimePanic("todo: timers")
}

func removeTimer(t *timer) bool {
	runtimePanic("todo: timers")
	return false
}

func schedulerRunQueue() *task.Queue {
	println("todo: schedulerRunQueue")
	for {
	}
	return nil
}

// Pause the current task for a given time.
//
//go:linkname sleep time.Sleep
func sleep(duration int64) {
	if duration <= 0 {
		return
	}

	wakeup := ticks() + nanosecondsToTicks(duration)

	schedulerLock()
	addSleepTask(task.Current(), wakeup)
	task.PauseLocked()
}

func run() {
	initHeap()
	cpuTasks[0] = &mainTask
	initAll() // TODO: move into main goroutine!

	until := ticks() + nanosecondsToTicks(200e6)
	for ticks() < until {
	}
	println("\n\n=====")

	go func() {
		//initAll()
		startSecondaryCores()
		callMain()
		mainExited = true
	}()
	schedulerLock()
	scheduler()
}

func runSecondary(core uint32, t *task.Task) {
	println("-- runSecondary")
	cpuTasks[core] = t
	println("-- locking for 2nd core")
	schedulerLock()
	println("--   locked!")
	scheduler()
}

var schedulerIsRunning = false

func scheduler() {
	if coresVerbose {
		println("** scheduler on core:", currentCPU())
	}
	for {
		//until := ticks() + nanosecondsToTicks(100e6)
		//for ticks() < until {
		//}

		// Check for ready-to-run tasks.
		if runnable := runQueue; runnable != nil {
			if coresVerbose {
				println("** scheduler", currentCPU(), "run runnable")
			}
			// Pop off the run queue.
			runQueue = runnable.Next
			runnable.Next = nil

			// Resume it now.
			setCurrentTask(runnable)
			schedulerUnlock()
			runnable.Resume()
			if coresVerbose {
				println("** scheduler", currentCPU(), "  returned (from runqueue resume)")
			}
			setCurrentTask(nil)

			continue
		}

		// If another core is using the clock, let it handle the sleep queue.
		if schedulerIsRunning {
			if coresVerbose {
				println("** scheduler", currentCPU(), "wait for other core")
			}
			schedulerUnlock()
			waitForEvents()
			schedulerLock()
			continue
		}

		if sleepingTask := sleepQueue; sleepingTask != nil {
			now := ticks()
			if now >= timeUnit(sleepingTask.Data) {
				if coresVerbose {
					println("** scheduler", currentCPU(), "run sleeping")
				}
				// This task is done sleeping.
				// Resume it now.
				sleepQueue = sleepQueue.Next
				sleepingTask.Next = nil

				setCurrentTask(sleepingTask)
				schedulerUnlock()
				sleepingTask.Resume()
				if coresVerbose {
					println("** scheduler", currentCPU(), "  returned (from sleepQueue resume)")
				}
				setCurrentTask(nil)
				continue
			}

			delay := timeUnit(sleepingTask.Data) - now
			if coresVerbose {
				println("** scheduler", currentCPU(), "sleep", ticksToNanoseconds(delay)/1e6)
			}

			// Sleep for a bit until the next task is ready to run.
			schedulerIsRunning = true
			schedulerUnlock()
			sleepTicks(delay)
			schedulerLock()
			schedulerIsRunning = false
			continue
		}

		if coresVerbose {
			println("** scheduler", currentCPU(), "wait for events")
		}
		schedulerUnlock()
		waitForEvents()
		schedulerLock()
	}
}

func currentTask() *task.Task {
	return cpuTasks[currentCPU()]
}

func setCurrentTask(task *task.Task) {
	cpuTasks[currentCPU()] = task
}

func runtimeTicks() uint64 {
	return uint64(ticks())
}

func runtimeSleepTicks(delay uint64) {
	sleepTicks(timeUnit(delay))
}

//export tinygo_schedulerUnlock
func tinygo_schedulerUnlock() {
	schedulerUnlock()
}
