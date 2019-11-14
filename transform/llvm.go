package transform

import (
	"reflect"

	"tinygo.org/x/go-llvm"
)

// Return a list of values (actually, instructions) where this value is used as
// an operand.
func getUses(value llvm.Value) []llvm.Value {
	if value.IsNil() {
		return nil
	}
	var uses []llvm.Value
	use := value.FirstUse()
	for !use.IsNil() {
		uses = append(uses, use.User())
		use = use.NextUse()
	}
	return uses
}

// makeGlobalArray creates a new LLVM global with the given name and integers as
// contents, and returns the global.
// Note that it is left with the default linkage etc., you should set
// linkage/constant/etc properties yourself.
func makeGlobalArray(mod llvm.Module, bufItf interface{}, name string, elementType llvm.Type) llvm.Value {
	buf := reflect.ValueOf(bufItf)
	globalType := llvm.ArrayType(elementType, buf.Len())
	global := llvm.AddGlobal(mod, globalType, name)
	value := llvm.Undef(globalType)
	for i := 0; i < buf.Len(); i++ {
		ch := buf.Index(i).Uint()
		value = llvm.ConstInsertValue(value, llvm.ConstInt(elementType, ch, false), []uint32{uint32(i)})
	}
	global.SetInitializer(value)
	return global
}

// getGlobalBytes returns the slice contained in the array of the provided
// global. It can recover the bytes originally created using makeGlobalArray, if
// makeGlobalArray was given a byte slice.
func getGlobalBytes(global llvm.Value) []byte {
	value := global.Initializer()
	buf := make([]byte, value.Type().ArrayLength())
	for i := range buf {
		buf[i] = byte(llvm.ConstExtractValue(value, []uint32{uint32(i)}).ZExtValue())
	}
	return buf
}

// replaceGlobalByteWithArray replaces a global integer type in the module with
// an integer array, using a GEP to make the types match. It is a convenience
// function used for creating reflection sidetables, for example.
func replaceGlobalIntWithArray(mod llvm.Module, name string, buf interface{}) llvm.Value {
	oldGlobal := mod.NamedGlobal(name)
	global := makeGlobalArray(mod, buf, name+".tmp", oldGlobal.Type().ElementType())
	gep := llvm.ConstGEP(global, []llvm.Value{
		llvm.ConstInt(mod.Context().Int32Type(), 0, false),
		llvm.ConstInt(mod.Context().Int32Type(), 0, false),
	})
	oldGlobal.ReplaceAllUsesWith(gep)
	oldGlobal.EraseFromParentAsGlobal()
	global.SetName(name)
	return global
}
