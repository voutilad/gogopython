package gogopython

import "github.com/ebitengine/purego"

var (
	// Converts a Go string into a Python *wchar_t, optionally storing some
	// error information in the provided index (if non-nil).
	Py_DecodeLocale func(s string, index *int) WCharPtr
	// Converts a Python *wchar_t to a Go string, optionally storing some
	// error information in the provided index (if non-nil).
	Py_EncodeLocale func(p WCharPtr, index *int) string

	// Pre-initialize the provided Python interpreter config using "isolated"
	// defaults.
	PyPreConfig_InitIsolatedConfig func(*PyPreConfig)

	// Initialize the provided Python interpreter config using defaults.
	PyConfig_InitPythonConfig func(*PyConfig_3_12)
	// Initialize the provided Python interpreter config using "isolated"
	// defaults.
	PyConfig_InitIsolatedPythonConfig func(*PyConfig_3_12)
	PyConfig_Clear                    func(*PyConfig_3_12)

	// Tear down the global Python interpreter state.
	//
	// This can deadlock depending on the GIL state. It can also panic. Be
	// careful!
	Py_FinalizeEx func() int32

	// Tear down a sub-interpreter using the provided thread state.
	//
	// Can panic or deadlock. Be careful!.
	Py_EndInterpreter func(PyThreadStatePtr)

	// Reports whether we have the GIL. 1 if true, 0 if false.
	PyGILState_Check func() int32

	// Take a reference to the GIL. Caution: this is recursive.
	PyGILState_Ensure func() PyGILState

	// Release a reference to the GIL. Caution: this is recursive.
	PyGILState_Release func(PyGILState)

	PyEval_AcquireThread func(PyThreadStatePtr)
	PyEval_ReleaseThread func(PyThreadStatePtr)
	PyEval_SaveThread    func() PyThreadStatePtr
	PyEval_RestoreThread func(PyThreadStatePtr)

	PyThreadState_Get            func() PyThreadStatePtr
	PyThreadState_New            func(PyInterpreterStatePtr) PyThreadStatePtr
	PyThreadState_Swap           func(PyThreadStatePtr) PyThreadStatePtr
	PyThreadState_Clear          func(PyThreadStatePtr)
	PyThreadState_Delete         func(PyThreadStatePtr)
	PyThreadState_DeleteCurrent  func()
	PyThreadState_GetInterpreter func(PyThreadStatePtr) PyInterpreterStatePtr

	PyInterpreterState_Get    func() PyInterpreterStatePtr
	PyInterpreterState_GetID  func(PyInterpreterStatePtr) int64
	PyInterpreterState_Clear  func(PyInterpreterStatePtr)
	PyInterpreterState_Delete func(PyInterpreterStatePtr)

	// Run a given Python script in the current interpreter, returning an exit
	// code based on if there was a Python exception raised.
	PyRun_SimpleString func(script string) int32
	// Run a given Python script in the current interpreter using the given
	// StartToken mode and globals/locals dicts.
	//
	// Globals will be accessible like any global and the script can mutate the
	// globals mapping using the "globals" keyword in the script.
	//
	// Locals will contain any declared local values from the script and is a
	// simple way to "return" Python data.
	PyRun_String func(script string, token StartToken, globals, locals PyObjectPtr) PyObjectPtr

	PyModule_New          func(string) PyObjectPtr
	PyModule_AddObjectRef func(module PyObjectPtr, name string, item PyObjectPtr) int32

	PyBool_FromLong func(int) PyObjectPtr

	PyLong_FromLong             func(int) PyObjectPtr
	PyLong_FromUnsignedLong     func(uint) PyObjectPtr
	PyLong_FromLongLong         func(int64) PyObjectPtr
	PyLong_FromUnsignedLongLong func(uint64) PyObjectPtr

	PyTuple_New     func(int64) PyObjectPtr
	PyTuple_SetItem func(tuple PyObjectPtr, pos int64, item PyObjectPtr) int32

	PyList_New     func(PyObjectPtr) int32
	PyList_Size    func(PyObjectPtr) int
	PyList_GetItem func(PyObjectPtr, int) PyObjectPtr
	PyList_SetItem func(list PyObjectPtr, index int, item PyObjectPtr) int32
	PyList_Append  func(list, item PyObjectPtr) int32
	PyList_Insert  func(list PyObjectPtr, index int, item PyObjectPtr) int32

	PyDict_New           func() PyObjectPtr
	PyDictProxy_New      func(mapping PyObjectPtr) PyObjectPtr
	PyDict_Clear         func(PyObjectPtr)
	PyDict_SetItem       func(dict, key, val PyObjectPtr) int32
	PyDict_SetItemString func(dict PyObjectPtr, key string, val PyObjectPtr) int
	PyDict_GetItem       func(dict, key, val PyObjectPtr) PyObjectPtr
	PyDict_GetItemString func(dict PyObjectPtr, key string) PyObjectPtr

	PyBytes_FromString            func(string) PyObjectPtr
	PyBytes_FromStringAndSize     func(*byte, int) PyObjectPtr
	PyByteArray_FromStringAndSize func(*byte, int) PyObjectPtr
	PyBytes_AsString              func(PyObjectPtr) *byte
	PyBytes_Size                  func(PyObjectPtr) int

	PyUnicode_FromString       func(string) PyObjectPtr
	PyUnicode_AsWideCharString func(PyObjectPtr, *int) WCharPtr

	Py_DecRef func(PyObjectPtr)
	Py_IncRef func(PyObjectPtr)

	PyErr_Clear func()
	PyErr_Print func()

	PyMem_Free func(*byte)

	PyObject_Type   func(PyObjectPtr) PyTypeObjectPtr
	PyType_GetFlags func(PyTypeObjectPtr) uint64
)

// Our problem children. These all return PyStatus, a struct. These need
// special handling to work on certain platforms like Linux due to how
// purego is currently written.
var (
	Py_PreInitialize            func(*PyPreConfig) PyStatus
	PyConfig_SetBytesString     func(*PyConfig_3_12, *WCharPtr, string) PyStatus
	Py_InitializeFromConfig     func(*PyConfig_3_12) PyStatus
	Py_NewInterpreterFromConfig func(state *PyThreadStatePtr, c *PyInterpreterConfig) PyStatus
)

func registerFuncs(lib PythonLibraryPtr) {
	purego.RegisterLibFunc(&Py_DecodeLocale, lib, "Py_DecodeLocale")
	purego.RegisterLibFunc(&Py_EncodeLocale, lib, "Py_EncodeLocale")

	purego.RegisterLibFunc(&PyPreConfig_InitIsolatedConfig, lib, "PyPreConfig_InitIsolatedConfig")

	purego.RegisterLibFunc(&PyConfig_InitPythonConfig, lib, "PyConfig_InitPythonConfig")
	purego.RegisterLibFunc(&PyConfig_InitIsolatedPythonConfig, lib, "PyConfig_InitIsolatedConfig")
	purego.RegisterLibFunc(&PyConfig_Clear, lib, "PyConfig_Clear")

	purego.RegisterLibFunc(&Py_FinalizeEx, lib, "Py_FinalizeEx")

	purego.RegisterLibFunc(&Py_EndInterpreter, lib, "Py_EndInterpreter")

	purego.RegisterLibFunc(&PyGILState_Check, lib, "PyGILState_Check")
	purego.RegisterLibFunc(&PyGILState_Ensure, lib, "PyGILState_Ensure")
	purego.RegisterLibFunc(&PyGILState_Release, lib, "PyGILState_Release")

	purego.RegisterLibFunc(&PyEval_AcquireThread, lib, "PyEval_AcquireThread")
	purego.RegisterLibFunc(&PyEval_ReleaseThread, lib, "PyEval_ReleaseThread")
	purego.RegisterLibFunc(&PyEval_SaveThread, lib, "PyEval_SaveThread")
	purego.RegisterLibFunc(&PyEval_RestoreThread, lib, "PyEval_RestoreThread")

	purego.RegisterLibFunc(&PyThreadState_Get, lib, "PyThreadState_Get")
	purego.RegisterLibFunc(&PyThreadState_New, lib, "PyThreadState_New")
	purego.RegisterLibFunc(&PyThreadState_Swap, lib, "PyThreadState_Swap")
	purego.RegisterLibFunc(&PyThreadState_Clear, lib, "PyThreadState_Clear")
	purego.RegisterLibFunc(&PyThreadState_Delete, lib, "PyThreadState_Delete")
	purego.RegisterLibFunc(&PyThreadState_DeleteCurrent, lib, "PyThreadState_DeleteCurrent")
	purego.RegisterLibFunc(&PyThreadState_GetInterpreter, lib, "PyThreadState_GetInterpreter")

	purego.RegisterLibFunc(&PyInterpreterState_Get, lib, "PyInterpreterState_Get")
	purego.RegisterLibFunc(&PyInterpreterState_GetID, lib, "PyInterpreterState_GetID")
	purego.RegisterLibFunc(&PyInterpreterState_Clear, lib, "PyInterpreterState_Clear")
	purego.RegisterLibFunc(&PyInterpreterState_Delete, lib, "PyInterpreterState_Delete")

	purego.RegisterLibFunc(&PyRun_SimpleString, lib, "PyRun_SimpleString")
	purego.RegisterLibFunc(&PyRun_String, lib, "PyRun_String")

	purego.RegisterLibFunc(&PyModule_New, lib, "PyModule_New")
	purego.RegisterLibFunc(&PyModule_AddObjectRef, lib, "PyModule_AddObjectRef")

	// ==== Data types
	purego.RegisterLibFunc(&PyBool_FromLong, lib, "PyBool_FromLong")

	purego.RegisterLibFunc(&PyLong_FromLong, lib, "PyLong_FromLong")
	purego.RegisterLibFunc(&PyLong_FromUnsignedLong, lib, "PyLong_FromUnsignedLong")
	purego.RegisterLibFunc(&PyLong_FromLongLong, lib, "PyLong_FromLongLong")
	purego.RegisterLibFunc(&PyLong_FromUnsignedLongLong, lib, "PyLong_FromUnsignedLongLong")

	purego.RegisterLibFunc(&PyTuple_New, lib, "PyTuple_New")
	purego.RegisterLibFunc(&PyTuple_SetItem, lib, "PyTuple_SetItem")

	purego.RegisterLibFunc(&PyList_New, lib, "PyList_New")
	purego.RegisterLibFunc(&PyList_Size, lib, "PyList_Size")
	purego.RegisterLibFunc(&PyList_GetItem, lib, "PyList_GetItem")
	purego.RegisterLibFunc(&PyList_SetItem, lib, "PyList_SetItem")
	purego.RegisterLibFunc(&PyList_Append, lib, "PyList_Append")
	purego.RegisterLibFunc(&PyList_Insert, lib, "PyList_Insert")

	purego.RegisterLibFunc(&PyDict_New, lib, "PyDict_New")
	purego.RegisterLibFunc(&PyDictProxy_New, lib, "PyDictProxy_New")
	purego.RegisterLibFunc(&PyDict_Clear, lib, "PyDict_Clear")
	purego.RegisterLibFunc(&PyDict_SetItem, lib, "PyDict_SetItem")
	purego.RegisterLibFunc(&PyDict_SetItemString, lib, "PyDict_SetItemString")
	purego.RegisterLibFunc(&PyDict_GetItem, lib, "PyDict_GetItem")
	purego.RegisterLibFunc(&PyDict_GetItemString, lib, "PyDict_GetItemString")

	purego.RegisterLibFunc(&PyBytes_FromString, lib, "PyBytes_FromString")
	purego.RegisterLibFunc(&PyBytes_FromStringAndSize, lib, "PyBytes_FromStringAndSize")
	purego.RegisterLibFunc(&PyByteArray_FromStringAndSize, lib, "PyByteArray_FromStringAndSize")
	purego.RegisterLibFunc(&PyBytes_AsString, lib, "PyBytes_AsString")
	purego.RegisterLibFunc(&PyBytes_Size, lib, "PyBytes_Size")

	purego.RegisterLibFunc(&PyUnicode_FromString, lib, "PyUnicode_FromString")
	purego.RegisterLibFunc(&PyUnicode_AsWideCharString, lib, "PyUnicode_AsWideCharString")

	purego.RegisterLibFunc(&Py_DecRef, lib, "Py_DecRef")
	purego.RegisterLibFunc(&Py_IncRef, lib, "Py_IncRef")

	purego.RegisterLibFunc(&PyErr_Clear, lib, "PyErr_Clear")
	purego.RegisterLibFunc(&PyErr_Print, lib, "PyErr_Print")

	purego.RegisterLibFunc(&PyMem_Free, lib, "PyMem_Free")

	purego.RegisterLibFunc(&PyObject_Type, lib, "PyObject_Type")
	purego.RegisterLibFunc(&PyType_GetFlags, lib, "PyType_GetFlags")

	// For the functions that return structs, we need to use some platform
	// dependent approaches.
	registerFuncsPlatDependent(lib)
}
