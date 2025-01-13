//go:build baremetal && !tinygo.unicore

package runtime

import "runtime/interrupt"

func atomicLock() interrupt.State {
	return atomicLockImpl()
}

func atomicUnlock(mask interrupt.State) {
	atomicUnlockImpl(mask)
}
