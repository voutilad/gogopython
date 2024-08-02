//go:build linux && (amd64 || arm64)

package gogopython

import (
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/jupiterrider/ffi"
	"golang.org/x/sys/unix"
)

// Python's PyStatus struct definition for use with libffi.
var typePyStatus = ffi.Type{
	Type: ffi.Struct,
	Elements: &[]*ffi.Type{
		&ffi.TypeSint32,  // status type
		&ffi.TypePointer, // *wchar_t function name
		&ffi.TypePointer, // *wchar_t error message
		&ffi.TypeSint32,  // exit code
		nil,
	}[0],
}

// Register our problem child functions using [githubcom/jupiterrider/ffi],
// which gives us the ability to handle returning structs on the stack.
func registerFuncsPlatDependent(lib PythonLibraryPtr) {
	// Py_PreInitialize
	var cifPy_PreInitialize ffi.Cif
	status := ffi.PrepCif(&cifPy_PreInitialize, ffi.DefaultAbi, 1, &typePyStatus, &ffi.TypePointer)
	if status != ffi.OK {
		panic(status)
	}
	symPy_PreInitialize, err := purego.Dlsym(lib, "Py_PreInitialize")
	if err != nil {
		panic(err)
	}
	Py_PreInitialize = func(cfg *PyPreConfig) PyStatus {
		var status PyStatus
		ffi.Call(&cifPy_PreInitialize, symPy_PreInitialize, unsafe.Pointer(&status), unsafe.Pointer(&cfg))
		return status
	}

	// Py_InitializeFromConfig
	var cifPy_InitializeFromConfig ffi.Cif
	status = ffi.PrepCif(&cifPy_InitializeFromConfig, ffi.DefaultAbi, 1, &typePyStatus, &ffi.TypePointer)
	if status != ffi.OK {
		panic(status)
	}
	symPy_InitializeFromConfig, err := purego.Dlsym(lib, "Py_InitializeFromConfig")
	if err != nil {
		panic(err)
	}
	Py_InitializeFromConfig = func(cfg *PyConfig_3_12) PyStatus {
		var status PyStatus
		ffi.Call(&cifPy_InitializeFromConfig, symPy_InitializeFromConfig, unsafe.Pointer(&status), unsafe.Pointer(&cfg))
		return status
	}

	// PyConfig_SetBytesString
	var cifPyConfig_SetBytesString ffi.Cif
	status = ffi.PrepCif(&cifPyConfig_SetBytesString, ffi.DefaultAbi, 3, &typePyStatus, &ffi.TypePointer, &ffi.TypePointer, &ffi.TypePointer)
	if status != ffi.OK {
		panic(status)
	}
	symPyConfig_SetBytesString, err := purego.Dlsym(lib, "PyConfig_SetBytesString")
	if err != nil {
		panic(err)
	}
	PyConfig_SetBytesString = func(cfg *PyConfig_3_12, wchar *WCharPtr, s string) PyStatus {
		var status PyStatus
		text, _ := unix.BytePtrFromString(s)
		ffi.Call(&cifPyConfig_SetBytesString, symPyConfig_SetBytesString, unsafe.Pointer(&status), unsafe.Pointer(cfg), unsafe.Pointer(&wchar), unsafe.Pointer(&text))
		return status
	}

	// Py_NewInterpreterFromConfig
	var cifPy_NewInterpreterFromConfig ffi.Cif
	status = ffi.PrepCif(&cifPy_NewInterpreterFromConfig, ffi.DefaultAbi, 2, &typePyStatus, &ffi.TypePointer, &ffi.TypePointer)
	if status != ffi.OK {
		panic(status)
	}
	symPy_NewInterpreterFromConfig, err := purego.Dlsym(lib, "Py_NewInterpreterFromConfig")
	if err != nil {
		panic(err)
	}
	Py_NewInterpreterFromConfig = func(state *PyThreadStatePtr, c *PyInterpreterConfig) PyStatus {
		var status PyStatus
		ffi.Call(&cifPy_NewInterpreterFromConfig, symPy_NewInterpreterFromConfig, unsafe.Pointer(&status), unsafe.Pointer(&state), unsafe.Pointer(&c))
		return status
	}
}
