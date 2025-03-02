package runtime

// This file implements Go interfaces.
//
// Interfaces are represented as a pair of {typecode, value}, where value can be
// anything (including non-pointers).

import (
	"unsafe"
)

type _interface struct {
	typecode unsafe.Pointer
	value    unsafe.Pointer
}

//go:inline
func composeInterface(typecode, value unsafe.Pointer) _interface {
	return _interface{typecode, value}
}

//go:inline
func decomposeInterface(i _interface) (unsafe.Pointer, unsafe.Pointer) {
	return i.typecode, i.value
}

// Return true iff both interfaces are equal.
//
//go:linkname interfaceEqual reflect.interfaceEqual
func interfaceEqual(x, y interface{}) bool

// interfaceTypeAssert is called when a type assert without comma-ok still
// returns false.
func interfaceTypeAssert(ok bool) {
	if !ok {
		runtimePanic("type assert failed")
	}
}

// The following declarations are only used during IR construction. They are
// lowered to inline IR in the interface lowering pass.
// See compiler/interface-lowering.go for details.

type structField struct {
	typecode unsafe.Pointer // type of this struct field
	data     *uint8         // pointer to byte array containing name, tag, varint-encoded offset, and some flags
}

// Pseudo function call used during a type assert. It is used during interface
// lowering, to assign the lowest type numbers to the types with the most type
// asserts. Also, it is replaced with const false if this type assert can never
// happen.
func typeAssert(actualType unsafe.Pointer, assertedType *uint8) bool
