// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"internal/task"
	"sync"
	"sync/atomic"
)

// This file contains stub implementations for internal/poll.
// The official golang implementation states:
//
// "That is, don't think of these as semaphores.
// Think of them as a way to implement sleep and wakeup
// such that every sleep is paired with a single wakeup,
// even if, due to races, the wakeup happens before the sleep."
//
// This is an experimental and probably incomplete implementation of the
// semaphore system, tailed to the network use case. That means, that it does not
// implement the modularity that the semacquire/semacquire1 implementation model
// offers, which in fact is emitted here entirely.
// This means we assume the following constant settings from the golang standard
// library: lifo=false,profile=semaBlock,skipframe=0,reason=waitReasonSemaquire

// The global state of the semaphore table.
// Semaphores are identified by their address.
// The table maps the address to the task that is currently holding the semaphore.
// The table is protected by a mutex.
// When a task acquires a semaphore, the mapping is added to the map.
// When a task releases a semaphore, the mapping is removed from the map.
//
// The table is used to implement the cansemacquire function.
// The cansemacquire function is called by the semacquire function.
// The cansemacquire function checks if the semaphore is available.
// If the semaphore is available, the function returns true.
// If the semaphore is not available, the function returns false.
type semTable struct {
	table map[*uint32]*task.Task
	lock  sync.Mutex
}

var semtable semTable

func init() {
	semtable.table = make(map[*uint32]*task.Task)
}

func (s *semTable) Lock() {
	s.lock.Lock()
}

func (s *semTable) Unlock() {
	s.lock.Unlock()
}

//go:linkname semacquire internal/poll.runtime_Semacquire
func semacquire(sema *uint32) {
	if cansemacquire(sema) {
		return
	}
}

// Copied from src/runtime/sema.go
func cansemacquire(addr *uint32) bool {
	// Busy Looping until a lookup to the global semaphore table can be made
	semtable.Lock()

	if _, ok := semtable.table[addr]; !ok {
		semtable.table[addr] = task.Current()
		semtable.Unlock()
		return true
	}

	v := atomic.LoadUint32(addr)
	if v == 0 {
		semtable.Unlock()
		return false
	}
	if atomic.CompareAndSwapUint32(addr, v, v-1) {
		semtable.Unlock()
		return true
	}
	return true
}

//go:linkname semrelease internal/poll.runtime_Semrelease
func semrelease(sema *uint32) {
	// Check if the semaphore is in the table
	semtable.Lock()
	if _, ok := semtable.table[sema]; !ok {
		panic("invalid semaphore")
	}

	atomic.AddUint32(sema, 1)
	semtable.Unlock()

	Gosched()
}
