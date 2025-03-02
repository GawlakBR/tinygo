package reflect

import (
	"unsafe"
)

//go:linkname hash32 runtime.hash32
func hash32(ptr unsafe.Pointer, n, seed uintptr) uint32

//go:linkname hashmapStringHash runtime.hashmapStringHash
func hashmapStringHash(s string, seed uintptr) uint32

func hashmapFloat32Hash(ptr unsafe.Pointer, seed uintptr) uint32 {
	f := *(*uint32)(ptr)
	if f == 0x80000000 {
		// convert -0 to 0 for hashing
		f = 0
	}
	return hash32(unsafe.Pointer(&f), 4, seed)
}

func hashmapFloat64Hash(ptr unsafe.Pointer, seed uintptr) uint32 {
	f := *(*uint64)(ptr)
	if f == 0x8000000000000000 {
		// convert -0 to 0 for hashing
		f = 0
	}
	return hash32(unsafe.Pointer(&f), 8, seed)
}

func hashmapInterfaceHash(itf interface{}, seed uintptr) uint32 {
	x := ValueOf(itf)
	if x.RawType() == nil {
		return 0 // nil interface
	}

	value := (*_interface)(unsafe.Pointer(&itf)).value
	ptr := value
	if x.RawType().Size() <= unsafe.Sizeof(uintptr(0)) {
		// Value fits in pointer, so it's directly stored in the pointer.
		ptr = unsafe.Pointer(&value)
	}

	switch x.RawType().Kind() {
	case Int, Int8, Int16, Int32, Int64:
		return hash32(ptr, x.RawType().Size(), seed)
	case Bool, Uint, Uint8, Uint16, Uint32, Uint64, Uintptr:
		return hash32(ptr, x.RawType().Size(), seed)
	case Float32:
		// It should be possible to just has the contents. However, NaN != NaN
		// so if you're using lots of NaNs as map keys (you shouldn't) then hash
		// time may become exponential. To fix that, it would be better to
		// return a random number instead:
		// https://research.swtch.com/randhash
		return hashmapFloat32Hash(ptr, seed)
	case Float64:
		return hashmapFloat64Hash(ptr, seed)
	case Complex64:
		rptr, iptr := ptr, unsafe.Add(ptr, 4)
		return hashmapFloat32Hash(rptr, seed) ^ hashmapFloat32Hash(iptr, seed)
	case Complex128:
		rptr, iptr := ptr, unsafe.Add(ptr, 8)
		return hashmapFloat64Hash(rptr, seed) ^ hashmapFloat64Hash(iptr, seed)
	case String:
		return hashmapStringHash(x.String(), seed)
	case Chan, Ptr, UnsafePointer:
		// It might seem better to just return the pointer, but that won't
		// result in an evenly distributed hashmap. Instead, hash the pointer
		// like most other types.
		return hash32(ptr, x.RawType().Size(), seed)
	case Array:
		var hash uint32
		for i := 0; i < x.Len(); i++ {
			hash ^= hashmapInterfaceHash(valueInterfaceUnsafe(x.Index(i)), seed)
		}
		return hash
	case Struct:
		var hash uint32
		for i := 0; i < x.NumField(); i++ {
			hash ^= hashmapInterfaceHash(valueInterfaceUnsafe(x.Field(i)), seed)
		}
		return hash
	default:
		runtimePanic("comparing un-comparable type")
		return 0 // unreachable
	}
}
