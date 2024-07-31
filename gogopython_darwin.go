//go:build darwin && (amd64 || arm64)

package gogopython

import (
	"github.com/ebitengine/purego"
)

var (
	Py_PreInitialize            func(*PyPreConfig) PyStatus
	PyConfig_SetBytesString     func(*PyConfig_3_12, *WCharPtr, string) PyStatus
	Py_InitializeFromConfig     func(*PyConfig_3_12) PyStatus
	Py_NewInterpreterFromConfig func(state *PyThreadStatePtr, c *PyInterpreterConfig) PyStatus
)

func registerFuncsPlatDependent(lib PythonLibraryPtr) {
	purego.RegisterLibFunc(&Py_PreInitialize, lib, "Py_PreInitialize")
	purego.RegisterLibFunc(&PyConfig_SetBytesString, lib, "PyConfig_SetBytesString")
	purego.RegisterLibFunc(&Py_InitializeFromConfig, lib, "Py_InitializeFromConfig")
	purego.RegisterLibFunc(&Py_NewInterpreterFromConfig, lib, "Py_NewInterpreterFromConfig")
}
