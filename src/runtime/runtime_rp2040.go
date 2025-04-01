//go:build rp2040

package runtime

import (
	"device/arm"
	"device/rp"
	"internal/task"
	"machine"
	"machine/usb/cdc"
	"reflect"
	"runtime/interrupt"
	"unsafe"
)

// machineTicks is provided by package machine.
func machineTicks() uint64

// machineLightSleep is provided by package machine.
func machineLightSleep(uint64)

type timeUnit int64

const numCPU = 2

// ticks returns the number of ticks (microseconds) elapsed since power up.
func ticks() timeUnit {
	t := machineTicks()
	return timeUnit(t)
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks) * 1000
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns / 1000)
}

func sleepTicks(d timeUnit) {
	if hasScheduler {
		// With scheduler, sleepTicks may return early if an interrupt or
		// event fires - so scheduler can schedule any go routines now
		// eligible to run
		machineLightSleep(uint64(d))
		return
	}

	// Busy loop
	sleepUntil := ticks() + d
	for ticks() < sleepUntil {
	}
}

func waitForEvents() {
	arm.Asm("wfe")
}

func putchar(c byte) {
	//mask := serialLock()
	machine.Serial.WriteByte(c)
	//serialUnlock(mask)
}

func getchar() byte {
	mask := serialLock()
	for machine.Serial.Buffered() == 0 {
		Gosched()
	}
	v, _ := machine.Serial.ReadByte()
	serialUnlock(mask)
	return v
}

func buffered() int {
	return machine.Serial.Buffered()
}

// machineInit is provided by package machine.
func machineInit()

func init() {
	machineInit()

	mask := serialLock()
	cdc.EnableUSBCDC()
	machine.USBDev.Configure(machine.UARTConfig{})
	machine.InitSerial()
	serialUnlock(mask)
}

//export Reset_Handler
func main() {
	preinit()
	run()
	exit(0)
}

func multicore_fifo_rvalid() bool {
	return rp.SIO.FIFO_ST.Get()&rp.SIO_FIFO_ST_VLD != 0
}

func multicore_fifo_wready() bool {
	return rp.SIO.FIFO_ST.Get()&rp.SIO_FIFO_ST_RDY != 0
}

func multicore_fifo_drain() {
	for multicore_fifo_rvalid() {
		rp.SIO.FIFO_RD.Get()
	}
}

func multicore_fifo_push_blocking(data uint32) {
	for !multicore_fifo_wready() {
	}
	rp.SIO.FIFO_WR.Set(data)
	arm.Asm("sev")
}

func multicore_fifo_pop_blocking() uint32 {
	for !multicore_fifo_rvalid() {
		arm.Asm("wfe")
	}

	return rp.SIO.FIFO_RD.Get()
}

//go:extern __isr_vector
var __isr_vector [0]uint32

//go:extern _stack1_top
var _stack1_top [0]uint32

var core1StartSequence = [...]uint32{
	0, 0, 1,
	uint32(uintptr(unsafe.Pointer(&__isr_vector))),
	uint32(uintptr(unsafe.Pointer(&_stack1_top))),
	uint32(uintptr(reflect.ValueOf(runCore1).Pointer())),
}

func startSecondaryCores() {
	// Start the second core of the RP2040.
	// See section 2.8.2 in the datasheet.
	seq := 0
	for {
		cmd := core1StartSequence[seq]
		if cmd == 0 {
			multicore_fifo_drain()
			arm.Asm("sev")
		}
		multicore_fifo_push_blocking(cmd)
		response := multicore_fifo_pop_blocking()
		if cmd != response {
			seq = 0
			continue
		}
		seq = seq + 1
		if seq >= len(core1StartSequence) {
			break
		}
	}
}

var core1Task task.Task

func runCore1() {
	//until := ticks() + nanosecondsToTicks(1900e6)
	//for ticks() < until {
	//}
	println("starting core 1")

	runSecondary(1, &core1Task)

	// Just blink a LED to show that this core is running.
	// TODO: use a real scheduler.
	//led := machine.GP16
	//led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	//const cycles = 7000_000
	//for {
	//	for i := 0; i < cycles; i++ {
	//		led.Low()
	//	}

	//	for i := 0; i < cycles; i++ {
	//		led.High()
	//	}
	//}
}

func currentCPU() uint32 {
	return rp.SIO.CPUID.Get()
}

const (
	spinlockAtomic = iota
	spinlockFutex
	spinlockScheduler
)

func atomicLockImpl() interrupt.State {
	mask := interrupt.Disable()
	for rp.SIO.SPINLOCK0.Get() == 0 {
	}
	return mask
}

func atomicUnlockImpl(mask interrupt.State) {
	rp.SIO.SPINLOCK0.Set(0)
	interrupt.Restore(mask)
}

func futexLock() interrupt.State {
	// Disable interrupts.
	// This is necessary since we might do some futex operations (like Wake)
	// inside an interrupt and we don't want to deadlock with a non-interrupt
	// goroutine that has taken the spinlock at the same time.
	mask := interrupt.Disable()

	// Acquire the spinlock.
	for rp.SIO.SPINLOCK1.Get() == 0 {
		// Spin, until the lock is released.
	}

	return mask
}

func futexUnlock(mask interrupt.State) {
	// Release the spinlock.
	rp.SIO.SPINLOCK1.Set(0)

	// Restore interrupts.
	interrupt.Restore(mask)
}

var schedulerLockMasks [numCPU]interrupt.State
var schedulerLocked = false

// WARNING: doesn't check for deadlocks!
func schedulerLock() {
	//schedulerLockMasks[currentCPU()] = interrupt.Disable()
	for rp.SIO.SPINLOCK2.Get() == 0 {
	}
	schedulerLocked = true
}

func schedulerUnlock() {
	if !schedulerLocked {
		println("!!! not locked at unlock")
		for {
		}
	}
	schedulerLocked = false
	rp.SIO.SPINLOCK2.Set(0)
	//interrupt.Restore(schedulerLockMasks[currentCPU()])
}

func serialLock() interrupt.State {
	//mask := interrupt.Disable()
	for rp.SIO.SPINLOCK3.Get() == 0 {
	}
	//return mask
	return 0
}

func serialUnlock(mask interrupt.State) {
	rp.SIO.SPINLOCK3.Set(0)
	//interrupt.Restore(mask)
}
