// CMSIS abstraction functions.
//
// Original copyright:
//
//     Copyright (c) 2009 - 2015 ARM LIMITED
//
//     All rights reserved.
//     Redistribution and use in source and binary forms, with or without
//     modification, are permitted provided that the following conditions are met:
//     - Redistributions of source code must retain the above copyright
//       notice, this list of conditions and the following disclaimer.
//     - Redistributions in binary form must reproduce the above copyright
//       notice, this list of conditions and the following disclaimer in the
//       documentation and/or other materials provided with the distribution.
//     - Neither the name of ARM nor the names of its contributors may be used
//       to endorse or promote products derived from this software without
//       specific prior written permission.
//
//     THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
//     AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
//     IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
//     ARE DISCLAIMED. IN NO EVENT SHALL COPYRIGHT HOLDERS AND CONTRIBUTORS BE
//     LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
//     CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
//     SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
//     INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
//     CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
//     ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
//     POSSIBILITY OF SUCH DAMAGE.
package arm

import (
	"unsafe"
)

// Run the given assembly code. The code will be marked as having side effects,
// as it doesn't produce output and thus would normally be eliminated by the
// optimizer.
func Asm(asm string)

// Run the given inline assembly. The code will be marked as having side
// effects, as it would otherwise be optimized away. The inline assembly string
// recognizes template values in the form {name}, like so:
//
//     arm.AsmFull(
//         "str {value}, {result}",
//         map[string]interface{}{
//             "value":  1
//             "result": &dest,
//         })
func AsmFull(asm string, regs map[string]interface{})

// ReadRegister returns the contents of the specified register. The register
// must be a processor register, reachable with the "mov" instruction.
func ReadRegister(name string) uintptr

//go:volatile
type RegValue uint32

const (
	SCS_BASE  = 0xE000E000
	NVIC_BASE = SCS_BASE + 0x0100
)

// Nested Vectored Interrupt Controller (NVIC).
type NVIC_Type struct {
	ISER [8]RegValue
}

var NVIC = (*NVIC_Type)(unsafe.Pointer(uintptr(NVIC_BASE)))

// Enable the given interrupt number.
func EnableIRQ(irq uint32) {
	NVIC.ISER[irq>>5] = 1 << (irq & 0x1F)
}
