//go:build baremetal && tinygo.unicore

package runtime

import "runtime/interrupt"

func atomicLock() interrupt.State {
	return interrupt.Disable()
}

func atomicUnlock(mask interrupt.State) {
	interrupt.Restore(mask)
}
