package main

import "unsafe"

const C.option2A = 20
const C.optionA = 0
const C.optionB = 1
const C.optionC = -5
const C.optionD = -4
const C.optionE = 10
const C.optionF = 11
const C.optionG = 12

type C.int16_t = int16
type C.int32_t = int32
type C.int64_t = int64
type C.int8_t = int8
type C.uint16_t = uint16
type C.uint32_t = uint32
type C.uint64_t = uint64
type C.uint8_t = uint8
type C.uintptr_t = uintptr
type C.char uint8
type C.int int32
type C.long int32
type C.longlong int64
type C.schar int8
type C.short int16
type C.uchar uint8
type C.uint uint32
type C.ulong uint32
type C.ulonglong uint64
type C.ushort uint16
type C.bitfield_t = C.struct_1
type C.myIntArray = [10]C.int
type C.myint = C.int
type C.option2_t = C.uint
type C.option_t = C.enum_option
type C.point2d_t = struct {
	x C.int
	y C.int
}
type C.point3d_t = C.struct_point3d
type C.types_t = struct {
	f   float32
	d   float64
	ptr *C.int
}

func (s *C.struct_1) bitfield_a() C.uchar          { return s.__bitfield_1 & 0x1f }
func (s *C.struct_1) set_bitfield_a(value C.uchar) { s.__bitfield_1 = s.__bitfield_1&^0x1f | value&0x1f<<0 }
func (s *C.struct_1) bitfield_b() C.uchar {
	return s.__bitfield_1 >> 5 & 0x1
}
func (s *C.struct_1) set_bitfield_b(value C.uchar) { s.__bitfield_1 = s.__bitfield_1&^0x20 | value&0x1<<5 }
func (s *C.struct_1) bitfield_c() C.uchar {
	return s.__bitfield_1 >> 6
}
func (s *C.struct_1) set_bitfield_c(value C.uchar,

) { s.__bitfield_1 = s.__bitfield_1&0x3f | value<<6 }

type C.struct_1 struct {
	start        C.uchar
	__bitfield_1 C.uchar

	d C.uchar
	e C.uchar
}
type C.struct_point3d struct {
	x C.int
	y C.int
	z C.int
}
type C.enum_option C.int
