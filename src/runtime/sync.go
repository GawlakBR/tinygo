// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"sync/atomic"
	"unsafe"
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

type semaRoot struct {
	nwait atomic.Uint32
}

var semtable semTable

// Prime to not correlate with any user patterns.
const semTabSize = 251

type semTable [semTabSize]struct {
	root semaRoot
	pad  [64 - unsafe.Sizeof(semaRoot{})]byte // only 64 x86_64, make this variable
}

func (t *semTable) rootFor(addr *uint32) *semaRoot {
	return &t[(uintptr(unsafe.Pointer(addr))>>3)%semTabSize].root
}

//go:linkname semacquire internal/poll.runtime_Semacquire
func semacquire(sema *uint32) {
	if cansemacquire(sema) {
		return
	}
}

// Copied from src/runtime/sema.go
func cansemacquire(addr *uint32) bool {
	for {
		v := atomic.LoadUint32(addr)
		if v == 0 {
			return false
		}
		if atomic.CompareAndSwapUint32(addr, v, v-1) {
			return true
		}
	}
}

//go:linkname semrelease internal/poll.runtime_Semrelease
func semrelease(sema *uint32) {
	root := semtable.rootFor(sema)
	atomic.AddUint32(sema, 1)
	if root.nwait.Load() == 0 {
		return
	}
}
