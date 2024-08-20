package gogopython

import "github.com/ebitengine/purego"

var (
	// Py_DecodeLocale converts a Go string into a Python *wchar_t, optionally
	// storing some error information in the provided index (if non-nil).
	Py_DecodeLocale func(s string, index *int) WCharPtr

	// Py_EncodeLocale converts a Python *wchar_t to a C char*, optionally
	// storing some error information in the provided index (if non-nil).
	Py_EncodeLocale func(p WCharPtr, index *int) *byte

	// PyPreConfig_InitIsolatedConfig pre-initializes the provided Python
	// interpreter config using "isolated" defaults.
	PyPreConfig_InitIsolatedConfig func(*PyPreConfig)

	// PyConfig_InitPythonConfig initializes the provided Python interpreter
	// config using defaults.
	PyConfig_InitPythonConfig func(*PyConfig_3_12)

	// PyConfig_InitIsolatedPythonConfig initializes the provided Python
	// interpreter config using "isolated" defaults.
	PyConfig_InitIsolatedPythonConfig func(*PyConfig_3_12)

	// PyConfig_Clear clears set values in a given PyConfig_3_12.
	PyConfig_Clear func(*PyConfig_3_12)

	// Py_FinalizeEx tears down the global Python interpreter state.
	//
	// This can deadlock depending on the GIL state. It can also panic.
	Py_FinalizeEx func() int32

	// Py_EndInterpreter tears down a sub-interpreter using the provided
	// thread state.
	//
	// This can panic or deadlock. Be careful!
	Py_EndInterpreter func(PyThreadStatePtr)

	// PyGILState_Check reports whether the caller has the GIL.
	// It returns 1 if true, 0 if false.
	PyGILState_Check func() int32

	// PyGILState_Ensure takes a reference to the GIL.
	// Caution: this is recursive.
	PyGILState_Ensure func() PyGILState

	// PyGILState_Release releases a reference to the GIL.
	// Caution: this is recursive.
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

	// PyRun_SimpleString evaluates the given Python script in the current
	// interpreter, returning an exit code based on if there was a Python
	// exception raised.
	PyRun_SimpleString func(script string) int32

	// PyRun_String evaluates a given Python script in the current interpreter
	// using the given StartToken mode and globals/locals dicts.
	//
	// Globals will be accessible like any global and the script can mutate the
	// globals mapping using the "globals" keyword in the script.
	//
	// Locals will contain any declared local values from the script and is a
	// simple way to "return" Python data.
	PyRun_String func(str string, start StartToken, globals, locals PyObjectPtr) PyObjectPtr

	// Py_CompileString is a simplified form of Py_CompileStringFlags using
	// default compiler flags.
	Py_CompileString func(str, filename string, start StartToken) PyCodeObjectPtr
	// Py_CompileStringFlags is a simplified form of Py_CompileStringExFlags
	// with optimizations set to UseInterpreterLevel.
	Py_CompileStringFlags func(str, filename string, start StartToken, flags *PyCompilerFlags) PyCodeObjectPtr

	// Py_CompileStringExFlags compiles the Python script in str and returns
	// the compiled Python code object. The filename is used to populate the
	// __file__ information for tracebacks and exception messages.
	//
	// Returns NullPyCodeObjectPtr on error.
	Py_CompileStringExFlags func(str, filename string, start StartToken,
		flags *PyCompilerFlags, optimize OptimizeLevel) PyCodeObjectPtr

	PyEval_EvalCode func(co PyCodeObjectPtr, globals, locals PyObjectPtr) PyObjectPtr

	PyModule_New          func(string) PyObjectPtr
	PyModule_AddObjectRef func(module PyObjectPtr, name string, item PyObjectPtr) int32

	PyCFunction_NewEx func(def *PyMethodDef, self, module PyObjectPtr) PyObjectPtr

	PyBool_FromLong func(int64) PyObjectPtr

	PyLong_AsLong               func(PyObjectPtr) int64
	PyLong_AsLongAndOverflow    func(PyObjectPtr, *int64) int64
	PyLong_AsUnsignedLong       func(PyObjectPtr) uint64
	PyLong_FromLong             func(int64) PyObjectPtr
	PyLong_FromUnsignedLong     func(uint64) PyObjectPtr
	PyLong_FromLongLong         func(int64) PyObjectPtr
	PyLong_FromUnsignedLongLong func(uint64) PyObjectPtr

	PyFloat_AsDouble   func(PyObjectPtr) float64
	PyFloat_FromDouble func(float64) PyObjectPtr

	PyTuple_New     func(int64) PyObjectPtr
	PyTuple_GetItem func(tuple PyObjectPtr, pos int64) PyObjectPtr
	PyTuple_SetItem func(tuple PyObjectPtr, pos int64, item PyObjectPtr) int32
	PyTuple_Size    func(tuple PyObjectPtr) int64

	PyList_New     func(PyObjectPtr) int32
	PyList_Size    func(PyObjectPtr) int64
	PyList_GetItem func(PyObjectPtr, int64) PyObjectPtr
	PyList_SetItem func(list PyObjectPtr, index int, item PyObjectPtr) int32
	PyList_Append  func(list, item PyObjectPtr) int32
	PyList_Insert  func(list PyObjectPtr, index int, item PyObjectPtr) int32

	PyDict_New           func() PyObjectPtr
	PyDictProxy_New      func(mapping PyObjectPtr) PyObjectPtr
	PyDict_Clear         func(PyObjectPtr)
	PyDict_SetItem       func(dict, key, val PyObjectPtr) int32
	PyDict_SetItemString func(dict PyObjectPtr, key string, val PyObjectPtr) int64
	PyDict_GetItem       func(dict, key PyObjectPtr) PyObjectPtr
	PyDict_GetItemString func(dict PyObjectPtr, key string) PyObjectPtr
	PyDict_Keys          func(dict PyObjectPtr) PyObjectPtr
	PyDict_Values        func(dict PyObjectPtr) PyObjectPtr
	PyDict_Size          func(dict PyObjectPtr) int64

	PyIter_Check func(iter PyObjectPtr) int32
	PyIter_Next  func(iter PyObjectPtr) PyObjectPtr
	PyIter_Send  func(iter, arg PyObjectPtr, result *PyObjectPtr) PySendResult

	PyFunction_GetCode func(fn PyObjectPtr) PyCodeObjectPtr

	PyObject_Call       func(callable, args, kwargs PyObjectPtr) PyObjectPtr
	PyObject_CallNoArgs func(callable PyObjectPtr) PyObjectPtr
	PyObject_CallOneArg func(callable, args PyObjectPtr) PyObjectPtr
	PyObject_CallObject func(callable, args PyObjectPtr) PyObjectPtr

	PySet_New       func(iterable PyObjectPtr) PyObjectPtr
	PyFrozenSet_New func(iterable PyObjectPtr) PyObjectPtr
	PySet_Size      func(PyObjectPtr) int64
	PySet_Contains  func(set, key PyObjectPtr) int32
	PySet_Add       func(set, key PyObjectPtr) int32
	PySet_Discard   func(set, key PyObjectPtr) int32
	PySet_Pop       func(set, key PyObjectPtr) PyObjectPtr
	PySet_Clear     func(set PyObjectPtr) int32

	PyBytes_FromString            func(string) PyObjectPtr
	PyBytes_FromStringAndSize     func(*byte, int64) PyObjectPtr
	PyByteArray_FromStringAndSize func(*byte, int64) PyObjectPtr
	PyBytes_AsString              func(PyObjectPtr) *byte
	PyBytes_Size                  func(PyObjectPtr) int64

	PyUnicode_FromString       func(string) PyObjectPtr
	PyUnicode_AsWideCharString func(PyObjectPtr, *int) WCharPtr
	PyUnicode_DecodeFSDefault  func(string) PyObjectPtr
	PyUnicode_EncodeFSDefault  func(PyObjectPtr) PyObjectPtr

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

	purego.RegisterLibFunc(&Py_CompileString, lib, "Py_CompileString")
	purego.RegisterLibFunc(&Py_CompileStringFlags, lib, "Py_CompileStringFlags")
	purego.RegisterLibFunc(&Py_CompileStringExFlags, lib, "Py_CompileStringExFlags")

	purego.RegisterLibFunc(&PyEval_EvalCode, lib, "PyEval_EvalCode")

	purego.RegisterLibFunc(&PyModule_New, lib, "PyModule_New")
	purego.RegisterLibFunc(&PyModule_AddObjectRef, lib, "PyModule_AddObjectRef")

	purego.RegisterLibFunc(&PyCFunction_NewEx, lib, "PyCFunction_NewEx")

	// ==== Data types
	purego.RegisterLibFunc(&PyBool_FromLong, lib, "PyBool_FromLong")

	purego.RegisterLibFunc(&PyLong_AsLong, lib, "PyLong_AsLong")
	purego.RegisterLibFunc(&PyLong_AsLongAndOverflow, lib, "PyLong_AsLongAndOverflow")
	purego.RegisterLibFunc(&PyLong_AsUnsignedLong, lib, "PyLong_AsUnsignedLong")
	purego.RegisterLibFunc(&PyLong_FromLong, lib, "PyLong_FromLong")
	purego.RegisterLibFunc(&PyLong_FromUnsignedLong, lib, "PyLong_FromUnsignedLong")
	purego.RegisterLibFunc(&PyLong_FromLongLong, lib, "PyLong_FromLongLong")
	purego.RegisterLibFunc(&PyLong_FromUnsignedLongLong, lib, "PyLong_FromUnsignedLongLong")

	purego.RegisterLibFunc(&PyFloat_AsDouble, lib, "PyFloat_AsDouble")
	purego.RegisterLibFunc(&PyFloat_FromDouble, lib, "PyFloat_FromDouble")

	purego.RegisterLibFunc(&PyTuple_New, lib, "PyTuple_New")
	purego.RegisterLibFunc(&PyTuple_GetItem, lib, "PyTuple_GetItem")
	purego.RegisterLibFunc(&PyTuple_SetItem, lib, "PyTuple_SetItem")
	purego.RegisterLibFunc(&PyTuple_Size, lib, "PyTuple_Size")

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
	purego.RegisterLibFunc(&PyDict_Keys, lib, "PyDict_Keys")
	purego.RegisterLibFunc(&PyDict_Values, lib, "PyDict_Values")
	purego.RegisterLibFunc(&PyDict_Size, lib, "PyDict_Size")

	purego.RegisterLibFunc(&PyIter_Check, lib, "PyIter_Check")
	purego.RegisterLibFunc(&PyIter_Next, lib, "PyIter_Next")
	purego.RegisterLibFunc(&PyIter_Send, lib, "PyIter_Send")

	purego.RegisterLibFunc(&PyFunction_GetCode, lib, "PyFunction_GetCode")

	purego.RegisterLibFunc(&PyObject_Call, lib, "PyObject_Call")
	purego.RegisterLibFunc(&PyObject_CallOneArg, lib, "PyObject_CallOneArg")
	purego.RegisterLibFunc(&PyObject_CallNoArgs, lib, "PyObject_CallNoArgs")
	purego.RegisterLibFunc(&PyObject_CallObject, lib, "PyObject_CallObject")

	purego.RegisterLibFunc(&PySet_New, lib, "PySet_New")
	purego.RegisterLibFunc(&PyFrozenSet_New, lib, "PyFrozenSet_New")
	purego.RegisterLibFunc(&PySet_Size, lib, "PySet_Size")
	purego.RegisterLibFunc(&PySet_Contains, lib, "PySet_Contains")
	purego.RegisterLibFunc(&PySet_Add, lib, "PySet_Add")
	purego.RegisterLibFunc(&PySet_Discard, lib, "PySet_Discard")
	purego.RegisterLibFunc(&PySet_Clear, lib, "PySet_Clear")
	purego.RegisterLibFunc(&PySet_Pop, lib, "PySet_Pop")

	purego.RegisterLibFunc(&PyBytes_FromString, lib, "PyBytes_FromString")
	purego.RegisterLibFunc(&PyBytes_FromStringAndSize, lib, "PyBytes_FromStringAndSize")
	purego.RegisterLibFunc(&PyByteArray_FromStringAndSize, lib, "PyByteArray_FromStringAndSize")
	purego.RegisterLibFunc(&PyBytes_AsString, lib, "PyBytes_AsString")
	purego.RegisterLibFunc(&PyBytes_Size, lib, "PyBytes_Size")

	purego.RegisterLibFunc(&PyUnicode_FromString, lib, "PyUnicode_FromString")
	purego.RegisterLibFunc(&PyUnicode_AsWideCharString, lib, "PyUnicode_AsWideCharString")
	purego.RegisterLibFunc(&PyUnicode_DecodeFSDefault, lib, "PyUnicode_DecodeFSDefault")
	purego.RegisterLibFunc(&PyUnicode_EncodeFSDefault, lib, "PyUnicode_EncodeFSDefault")

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
