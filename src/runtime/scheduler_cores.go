//go:build scheduler.cores

package runtime

import (
	"internal/task"
	"unsafe"
)

const hasScheduler = true

const hasParallelism = true

var (
	mainTask task.Task
	cpuTasks [numCPU]*task.Task
)

func deadlock() {
	// Call yield without requesting a wakeup.
	task.Pause()
	trap()
}

func scheduleTask(t *task.Task) {
	task.Resume(t)
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
	task.Sleep(uint64(wakeup))
}

func run() {
	initHeap()
	cpuTasks[0] = &mainTask
	task.Init(&mainTask, (*uintptr)(unsafe.Pointer(&stackTopSymbol)))
	initAll()
	startOtherCores()
	callMain()
	mainExited = true
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
