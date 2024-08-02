//go:build darwin && (amd64 || arm64)

package gogopython

import (
	"github.com/ebitengine/purego"
)

func registerFuncsPlatDependent(lib PythonLibraryPtr) {
	// On macOS, purego supports returning structs natively. Easy!

	purego.RegisterLibFunc(&Py_PreInitialize, lib, "Py_PreInitialize")
	purego.RegisterLibFunc(&PyConfig_SetBytesString, lib, "PyConfig_SetBytesString")
	purego.RegisterLibFunc(&Py_InitializeFromConfig, lib, "Py_InitializeFromConfig")
	purego.RegisterLibFunc(&Py_NewInterpreterFromConfig, lib, "Py_NewInterpreterFromConfig")
}
