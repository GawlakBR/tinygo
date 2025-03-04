package reflectlite

import (
	"internal/itoa"
	"unsafe"
)

type Kind uint8

// Copied from reflect/type.go
// https://golang.org/src/reflect/type.go?s=8302:8316#L217
// These constants must match basicTypes and the typeKind* constants in
// compiler/interface.go
const (
	Invalid Kind = iota
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Uintptr
	Float32
	Float64
	Complex64
	Complex128
	String
	UnsafePointer
	Chan
	Interface
	Pointer
	Slice
	Array
	Func
	Map
	Struct
)

// Ptr is the old name for the Pointer kind.
const Ptr = Pointer

func (k Kind) String() string {
	switch k {
	case Invalid:
		return "invalid"
	case Bool:
		return "bool"
	case Int:
		return "int"
	case Int8:
		return "int8"
	case Int16:
		return "int16"
	case Int32:
		return "int32"
	case Int64:
		return "int64"
	case Uint:
		return "uint"
	case Uint8:
		return "uint8"
	case Uint16:
		return "uint16"
	case Uint32:
		return "uint32"
	case Uint64:
		return "uint64"
	case Uintptr:
		return "uintptr"
	case Float32:
		return "float32"
	case Float64:
		return "float64"
	case Complex64:
		return "complex64"
	case Complex128:
		return "complex128"
	case String:
		return "string"
	case UnsafePointer:
		return "unsafe.Pointer"
	case Chan:
		return "chan"
	case Interface:
		return "interface"
	case Pointer:
		return "ptr"
	case Slice:
		return "slice"
	case Array:
		return "array"
	case Func:
		return "func"
	case Map:
		return "map"
	case Struct:
		return "struct"
	default:
		return "kind" + itoa.Itoa(int(int8(k)))
	}
}

// Copied from reflect/type.go
// https://go.dev/src/reflect/type.go?#L348

// ChanDir represents a channel type's direction.
type ChanDir int

const (
	RecvDir ChanDir             = 1 << iota // <-chan
	SendDir                                 // chan<-
	BothDir = RecvDir | SendDir             // chan
)

type Type interface {
	Name() string
	PkgPath() string
	Size() uintptr
	Kind() Kind
	Implements(u Type) bool
	AssignableTo(u Type) bool
	Comparable() bool
	String() string
	Elem() Type
}

// Constants for the 'meta' byte.
// These constants are also defined in the reflect package.
const (
	kindMask       = 31  // mask to apply to the meta byte to get the Kind value
	flagNamed      = 32  // flag that is set if this is a named type
	flagComparable = 64  // flag that is set if this type is comparable
	flagIsBinary   = 128 // flag that is set if this type uses the hashmap binary algorithm
)

// The below types (rawType, elemType, etc) are also defined in the reflect
// package and must match the compiler output.

type rawType struct {
	meta uint8
}

type elemType struct {
	rawType
	numMethod uint16
	ptrTo     *rawType
	elem      *rawType
}

type ptrType struct {
	rawType
	numMethod uint16
	elem      *rawType
}

type interfaceType struct {
	rawType
	ptrTo *rawType
}

type arrayType struct {
	rawType
	numMethod uint16
	ptrTo     *rawType
	elem      *rawType
	arrayLen  uintptr
	slicePtr  *rawType
}

type mapType struct {
	rawType
	numMethod uint16
	ptrTo     *rawType
	elem      *rawType
	key       *rawType
}

type namedType struct {
	rawType
	numMethod uint16
	ptrTo     *rawType
	elem      *rawType
	pkg       *byte
	name      [1]byte
}

type structType struct {
	rawType
	numMethod uint16
	ptrTo     *rawType
	pkgpath   *byte
	size      uint32
	numField  uint16
	fields    [1]structField // the remaining fields are all of type structField
}

type structField struct {
	fieldType *rawType
	data      unsafe.Pointer // various bits of information, packed in a byte array
}

// rawStructField is the same as StructField but with the Type member replaced
// with rawType. For internal use only. Avoiding this conversion to the Type
// interface improves code size in many cases.
type rawStructField struct {
	Name      string
	PkgPath   string
	Type      *rawType
	Tag       StructTag
	Offset    uintptr
	Anonymous bool
}

// Flags stored in the first byte of the struct field byte array. Must be kept
// up to date with compiler/interface.go.
const (
	structFieldFlagAnonymous = 1 << iota
	structFieldFlagHasTag
	structFieldFlagIsExported
	structFieldFlagIsEmbedded
)

func TypeOf(i interface{}) Type {
	if i == nil {
		return nil
	}
	typecode, _ := decomposeInterface(i)
	return (*rawType)(typecode)
}

func (t *rawType) ptrtag() uintptr {
	return uintptr(unsafe.Pointer(t)) & 0b11
}

func (t *rawType) isNamed() bool {
	if tag := t.ptrtag(); tag != 0 {
		return false
	}

	return t.meta&flagNamed != 0
}

func (t *rawType) underlying() *rawType {
	if t.isNamed() {
		return (*elemType)(unsafe.Pointer(t)).elem
	}
	return t
}

func (t *rawType) arrayLen() uintptr {
	return (*arrayType)(unsafe.Pointer(t.underlying())).arrayLen
}

func (t *rawType) numField() int {
	return int((*structType)(unsafe.Pointer(t.underlying())).numField)
}

func (t *rawType) name() string {
	ntype := (*namedType)(unsafe.Pointer(t))
	return readStringZ(unsafe.Pointer(&ntype.name[0]))
}

// Return the size (in bytes) of the given type.
func typeSize(t *rawType) uintptr {
	switch t.Kind() {
	case Bool, Int8, Uint8:
		return 1
	case Int16, Uint16:
		return 2
	case Int32, Uint32:
		return 4
	case Int64, Uint64:
		return 8
	case Int, Uint:
		return unsafe.Sizeof(int(0))
	case Uintptr:
		return unsafe.Sizeof(uintptr(0))
	case Float32:
		return 4
	case Float64:
		return 8
	case Complex64:
		return 8
	case Complex128:
		return 16
	case String:
		return unsafe.Sizeof("")
	case UnsafePointer, Chan, Map, Pointer:
		return unsafe.Sizeof(uintptr(0))
	case Slice:
		return unsafe.Sizeof([]int{})
	case Interface:
		return unsafe.Sizeof(interface{}(nil))
	case Func:
		var f func()
		return unsafe.Sizeof(f)
	case Array:
		return typeElem(t).Size() * t.arrayLen()
	case Struct:
		u := t.underlying()
		return uintptr((*structType)(unsafe.Pointer(u)).size)
	default:
		panic("unimplemented: size of type")
	}
}

// Return the type kind of the given type.
func typeKind(t *rawType) Kind {
	if t == nil {
		return Invalid
	}

	if tag := t.ptrtag(); tag != 0 {
		return Pointer
	}

	return Kind(t.meta & kindMask)
}

func typeString(t *rawType) string {
	if t.isNamed() {
		s := t.name()
		if s[0] == '.' {
			return s[1:]
		}
		return s
	}

	switch t.Kind() {
	case Chan:
		elem := typeElem(t).String()
		switch typeChanDir(t) {
		case SendDir:
			return "chan<- " + elem
		case RecvDir:
			return "<-chan " + elem
		case BothDir:
			if elem[0] == '<' {
				// typ is recv chan, need parentheses as "<-" associates with leftmost
				// chan possible, see:
				// * https://golang.org/ref/spec#Channel_types
				// * https://github.com/golang/go/issues/39897
				return "chan (" + elem + ")"
			}
			return "chan " + elem
		}

	case Pointer:
		return "*" + typeElem(t).String()
	case Slice:
		return "[]" + typeElem(t).String()
	case Array:
		return "[" + itoa.Itoa(int(t.arrayLen())) + "]" + typeElem(t).String()
	case Map:
		return "map[" + typeKey(t).String() + "]" + typeElem(t).String()
	case Struct:
		numField := t.numField()
		if numField == 0 {
			return "struct {}"
		}
		s := "struct {"
		for i := 0; i < numField; i++ {
			f := typeRawField(t, i)
			s += " " + f.Name + " " + f.Type.String()
			if f.Tag != "" {
				s += " " + quote(string(f.Tag))
			}
			// every field except the last needs a semicolon
			if i < numField-1 {
				s += ";"
			}
		}
		s += " }"
		return s
	case Interface:
		// TODO(dgryski): Needs actual method set info
		return "interface {}"
	default:
		return t.Kind().String()
	}

	return t.Kind().String()
}

// Return the element type given a type. Panics if this type doesn't have an
// element type.
func typeElem(t *rawType) *rawType {
	if tag := t.ptrtag(); tag != 0 {
		return (*rawType)(unsafe.Add(unsafe.Pointer(t), -1))
	}

	underlying := t.underlying()
	switch underlying.Kind() {
	case Pointer:
		return (*ptrType)(unsafe.Pointer(underlying)).elem
	case Chan, Slice, Array, Map:
		return (*elemType)(unsafe.Pointer(underlying)).elem
	default:
		panic(errTypeElem)
	}
}

func typeKey(t *rawType) *rawType {
	underlying := t.underlying()
	if underlying.Kind() != Map {
		panic(errTypeKey)
	}
	return (*mapType)(unsafe.Pointer(underlying)).key
}

func typeChanDir(t *rawType) ChanDir {
	if t.Kind() != Chan {
		panic(errTypeChanDir)
	}

	dir := int((*elemType)(unsafe.Pointer(t)).numMethod)

	// nummethod is overloaded for channel to store channel direction
	return ChanDir(dir)
}

func typeNumMethod(t *rawType) int {
	if t.isNamed() {
		return int((*namedType)(unsafe.Pointer(t)).numMethod)
	}

	switch t.Kind() {
	case Pointer:
		return int((*ptrType)(unsafe.Pointer(t)).numMethod)
	case Struct:
		return int((*structType)(unsafe.Pointer(t)).numMethod)
	case Interface:
		//FIXME: Use len(methods)
		return typeNumMethod((*interfaceType)(unsafe.Pointer(t)).ptrTo)
	}

	// Other types have no methods attached.  Note we don't panic here.
	return 0
}

func typeAssignableTo(t, u *rawType) bool {
	if t == u {
		return true
	}

	if t.underlying() == u.underlying() && (!t.isNamed() || !u.isNamed()) {
		return true
	}

	if u.Kind() == Interface && typeNumMethod(u) == 0 {
		return true
	}

	if u.Kind() == Interface {
		panic("reflect: unimplemented: AssignableTo with interface")
	}
	return false
}

func rawStructFieldFromPointer(descriptor *structType, fieldType *rawType, data unsafe.Pointer, flagsByte uint8, name string, offset uint32) rawStructField {
	// Read the field tag, if there is one.
	var tag string
	if flagsByte&structFieldFlagHasTag != 0 {
		data = unsafe.Add(data, 1) // C: data+1
		tagLen := uintptr(*(*byte)(data))
		data = unsafe.Add(data, 1) // C: data+1
		tag = *(*string)(unsafe.Pointer(&stringHeader{
			data: data,
			len:  tagLen,
		}))
	}

	// Set the PkgPath to some (arbitrary) value if the package path is not
	// exported.
	pkgPath := ""
	if flagsByte&structFieldFlagIsExported == 0 {
		// This field is unexported.
		pkgPath = readStringZ(unsafe.Pointer(descriptor.pkgpath))
	}

	return rawStructField{
		Name:      name,
		PkgPath:   pkgPath,
		Type:      fieldType,
		Tag:       StructTag(tag),
		Anonymous: flagsByte&structFieldFlagAnonymous != 0,
		Offset:    uintptr(offset),
	}
}

func typeRawField(t *rawType, n int) rawStructField {
	if t.Kind() != Struct {
		panic(errTypeField)
	}
	descriptor := (*structType)(unsafe.Pointer(t.underlying()))
	if uint(n) >= uint(descriptor.numField) {
		panic("reflect: field index out of range")
	}

	// Iterate over all the fields to calculate the offset.
	// This offset could have been stored directly in the array (to make the
	// lookup faster), but by calculating it on-the-fly a bit of storage can be
	// saved.
	field := (*structField)(unsafe.Add(unsafe.Pointer(&descriptor.fields[0]), uintptr(n)*unsafe.Sizeof(structField{})))
	data := field.data

	// Read some flags of this field, like whether the field is an embedded
	// field. See structFieldFlagAnonymous and similar flags.
	flagsByte := *(*byte)(data)
	data = unsafe.Add(data, 1)
	offset, lenOffs := uvarint32(unsafe.Slice((*byte)(data), maxVarintLen32))
	data = unsafe.Add(data, lenOffs)

	name := readStringZ(data)
	data = unsafe.Add(data, len(name))

	return rawStructFieldFromPointer(descriptor, field.fieldType, data, flagsByte, name, offset)
}

func (t *rawType) Name() string    { panic("todo: internal/reflectlite.Type.Name") }
func (t *rawType) PkgPath() string { panic("todo: internal/reflectlite.Type.PkgPath") }

func (t *rawType) Size() uintptr {
	return typeSize(t)
}

func (t *rawType) Kind() Kind {
	return typeKind(t)
}

func (t *rawType) Implements(u Type) bool {
	uraw := u.(*rawType)
	if uraw.Kind() != Interface {
		panic("reflect: non-interface type passed to Type.Implements")
	}
	return typeAssignableTo(t, uraw)
}

func (t *rawType) AssignableTo(u Type) bool {
	return typeAssignableTo(t, u.(*rawType))
}

func (t *rawType) Comparable() bool {
	return (t.meta & flagComparable) == flagComparable
}

func (t *rawType) String() string {
	return typeString(t)
}

func (t *rawType) Elem() Type {
	return typeElem(t)
}

// A StructTag is the tag string in a struct field.
type StructTag string

// TODO: it would be feasible to do the key/value splitting at compile time,
// avoiding the code size cost of doing it at runtime

// Get returns the value associated with key in the tag string.
func (tag StructTag) Get(key string) string {
	v, _ := tag.Lookup(key)
	return v
}

// Lookup returns the value associated with key in the tag string.
func (tag StructTag) Lookup(key string) (value string, ok bool) {
	for tag != "" {
		// Skip leading space.
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// Scan to colon. A space, a quote or a control character is a syntax error.
		// Strictly speaking, control chars include the range [0x7f, 0x9f], not just
		// [0x00, 0x1f], but in practice, we ignore the multi-byte control characters
		// as it is simpler to inspect the tag's bytes than the tag's runes.
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}
		name := string(tag[:i])
		tag = tag[i+1:]

		// Scan quoted string to find value.
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			break
		}
		qvalue := string(tag[:i+1])
		tag = tag[i+1:]

		if key == name {
			value, err := unquote(qvalue)
			if err != nil {
				break
			}
			return value, true
		}
	}
	return "", false
}

// Read and return a null terminated string starting from data.
func readStringZ(data unsafe.Pointer) string {
	start := data
	var len uintptr
	for *(*byte)(data) != 0 {
		len++
		data = unsafe.Add(data, 1) // C: data++
	}

	return *(*string)(unsafe.Pointer(&stringHeader{
		data: start,
		len:  len,
	}))
}

const maxVarintLen32 = 5

// encoding/binary.Uvarint, specialized for uint32
func uvarint32(buf []byte) (uint32, int) {
	var x uint32
	var s uint
	for i, b := range buf {
		if b < 0x80 {
			return x | uint32(b)<<s, i + 1
		}
		x |= uint32(b&0x7f) << s
		s += 7
	}
	return 0, 0
}
