package reflect

import "unsafe"

// Cause a runtime panic, which is (currently) always a string.
//
//go:linkname runtimePanic runtime.runtimePanic
func runtimePanic(msg string)

type _interface struct {
	typecode unsafe.Pointer
	value    unsafe.Pointer
}

// Return true iff both interfaces are equal.
func interfaceEqual(x, y interface{}) bool {
	return reflectValueEqual(ValueOf(x), ValueOf(y))
}

func reflectValueEqual(x, y Value) bool {
	// Note: doing a x.Type() == y.Type() comparison would not work here as that
	// would introduce an infinite recursion: comparing two Type values
	// is done with this reflectValueEqual runtime call.
	if x.RawType() == nil || y.RawType() == nil {
		// One of them is nil.
		return x.RawType() == y.RawType()
	}

	if x.RawType() != y.RawType() {
		// The type is not the same, which means the interfaces are definitely
		// not the same.
		return false
	}

	switch x.RawType().Kind() {
	case Bool:
		return x.Bool() == y.Bool()
	case Int, Int8, Int16, Int32, Int64:
		return x.Int() == y.Int()
	case Uint, Uint8, Uint16, Uint32, Uint64, Uintptr:
		return x.Uint() == y.Uint()
	case Float32, Float64:
		return x.Float() == y.Float()
	case Complex64, Complex128:
		return x.Complex() == y.Complex()
	case String:
		return x.String() == y.String()
	case Chan, Ptr, UnsafePointer:
		return x.UnsafePointer() == y.UnsafePointer()
	case Array:
		for i := 0; i < x.Len(); i++ {
			if !reflectValueEqual(x.Index(i), y.Index(i)) {
				return false
			}
		}
		return true
	case Struct:
		for i := 0; i < x.NumField(); i++ {
			if !reflectValueEqual(x.Field(i), y.Field(i)) {
				return false
			}
		}
		return true
	case Interface:
		return reflectValueEqual(x.Elem(), y.Elem())
	default:
		runtimePanic("comparing un-comparable type")
		return false // unreachable
	}
}
