package main

import (
	"unsafe"

	"github.com/ebitengine/purego"
)

type PyObjectPtr *byte
type WCharPtr *byte
type PyGILState int32

type StartToken = int32

const (
	PySingleInput   StartToken = 256 // Used for single statements
	PyFileInput     StartToken = 257 // Used for modules (i.e. many statements)
	PyEvalInput     StartToken = 258 // Used for expressions(?)
	PyFuncTypeInput StartToken = 345 // ??? no idea
)

type PyStatus struct {
	Type     int32
	Func     WCharPtr
	ErrMsg   WCharPtr
	ExitCode int32
}

func PyBytesToString(b WCharPtr) string {
	ptr := unsafe.Pointer(b)

	for len := 0; len < 1024; len++ {
		if *(*uint8)(ptr) == 0 {
			return unsafe.String(b, len)
		}
		ptr = unsafe.Add(ptr, 1)
	}

	return ""
}

type PyPreConfig struct {
	ConfigInit        int32
	ParseArgv         int32
	Isolated          int32
	UseEnvironment    int32
	ConfigureLocale   int32
	CoerceCLocale     int32
	CoerceCLocaleWarn int32
	// LegacyWindowsFSEncoding // only on Windows
	Utf8Mode  int32
	DevMode   int32
	Allocator int32
}

type PyWideStringList struct {
	Length int64
	Items  **byte
}

type PyConfig struct {
	ConfigInit int32

	Isolated              int32
	UseEnvironment        int32
	DevMode               int32
	InstallSignalHandlers int32
	UseHashSeed           int32
	HashSeed              uint64
	FaultHandler          int32
	TraceMalloc           int32
	PerfProfiling         int32
	ImportTime            int32
	CodeDebugRanges       int32
	ShowRefCount          int32
	DumpRefs              int32
	DumpRefsFile          *byte
	MallocStats           int32
	FilesystemEncoding    *byte
	FilesystemErrors      *byte
	PycachePrefix         *byte
	ParseArgv             int32

	OrigArgv    PyWideStringList
	Argv        PyWideStringList
	XOptions    PyWideStringList
	WarnOptions PyWideStringList
	Padding     [4]int8 // xxx alignment issues?

	SiteImport          int32
	BytesWarning        int32
	WarnDefaultEncoding int32
	Inspect             int32
	Interactive         int32
	OptimizationLevel   int32
	ParserDebug         int32
	WriteBytecode       int32
	Verbose             int32
	Quiet               int32
	UserSiteDirectory   int32
	ConfigureCStdio     int32
	BufferedStdio       int32
	StdioEncodings      WCharPtr
	StdioErrors         WCharPtr
	// LegacyWindowsStdio  int32 // if windows
	CheckHashPycsMode WCharPtr
	UseFrozenModules  int32
	SafePath          int32
	IntMaxStrDigits   int32
	// CpuCount          int32
	// EnableGil         int32 // if gil disabled

	/* Path configuration inputs */
	PathConfigWarnings int32
	ProgramName        WCharPtr
	PythonPathEnv      WCharPtr
	Home               WCharPtr
	PlatLibDir         WCharPtr

	/* Path configuration outputs */
	ModuleSearchPathsSet int32
	ModuleSearchPaths    PyWideStringList
	StdlibDir            *byte
	Executable           *byte
	BaseExecutable       *byte
	Prefix               *byte
	BasePrefix           *byte
	ExecPrefix           *byte
	BaseExecPrefix       *byte

	/* Parameter only used by Py_Main */
	SkipSourceFirstLine int32
	RunCommand          *byte
	RunModule           *byte
	RunFilename         *byte

	/* Set by Py_Main */
	SysPath0 *byte

	/* Private Fields */
	InstallImportLib int32
	InitMain         int32
	IsPythonBuild    int32
	// pystats          int32 // if Py_Stats
	// runPresite       *byte // if Py_DEBUG
}

type GilType int32

const (
	DefaultGil GilType = 0
	SharedGil  GilType = 1
	OwnGil     GilType = 2
)

type PyInterpreterConfig struct {
	UseMainObMalloc            int32
	AllowFork                  int32
	AllowExec                  int32
	AllowThreads               int32
	AllowDaemonThreads         int32
	CheckMultiInterpExtensions int32
	Gil                        GilType
}

type pyTrashCan struct {
	DeleteNesting int32
	DeleteLater   PyObjectPtr
}

type pyCFrame struct {
	CurrentFrame uintptr
	Previous     uintptr
}

type pyStackChunk struct {
	Previous uintptr
	Size     int //xxx size_t?
	Top      int //xxx size_t?
	Data     uintptr
}

type pyErrStackItem struct {
	ExcValue     PyObjectPtr
	PreviousItem uintptr
}

type PyThreadState_3_12 struct {
	Prev                         uintptr
	Next                         uintptr
	Interp                       uintptr
	Status                       uint32
	PyRecursionRemaining         int32
	PyRecursionLimit             int32
	CRecursionRemaining          int32
	RecursionHeadroom            int32
	Tracing                      int32
	WhatEvent                    int32
	CFrame                       uintptr
	CProfileFunc                 uintptr
	CTraceFunc                   uintptr
	CProfileObj                  PyObjectPtr
	CTraceObj                    PyObjectPtr
	CurrentException             PyObjectPtr
	ExcInfo                      uintptr
	Dict                         PyObjectPtr
	GilstateCounter              int32
	AsyncExc                     PyObjectPtr
	ThreadId                     uint
	NativeThreadId               uint
	Trash                        pyTrashCan
	OnDelete                     uintptr
	OnDeleteData                 uintptr
	CoRoutineOriginTrackingDepth int32
	AsyncGenFirstIter            PyObjectPtr
	AsyncGenFinalizer            PyObjectPtr
	Context                      PyObjectPtr
	ContextVer                   uint64
	Id                           uint64
	DataStackChunk               pyStackChunk
	DataStackTop                 uintptr
	DataStackLimit               uintptr
	ExcState                     pyErrStackItem
	RootCFrame                   pyCFrame
}

var (
	Py_DecodeLocale         func(string, uint64) WCharPtr
	Py_EncodeLocale         func(WCharPtr, uint64) string
	PyWideStringList_Append func(*PyWideStringList, WCharPtr) PyStatus

	PyPreConfig_InitIsolatedConfig func(*PyPreConfig)
	Py_PreInitialize               func(*PyPreConfig) PyStatus

	PyConfig_InitPythonConfig         func(*PyConfig)
	PyConfig_InitIsolatedPythonConfig func(*PyConfig)
	PyConfig_SetBytesString           func(*PyConfig, *WCharPtr, string) PyStatus
	PyConfig_Clear                    func(*PyConfig)
	Py_InitializeFromConfig           func(*PyConfig) PyStatus
	PyConfig_Read                     func(*PyConfig) PyStatus

	Py_FinalizeEx func() int32

	Py_NewInterpreterFromConfig func(**PyThreadState_3_12, *PyInterpreterConfig) PyStatus
	Py_EndInterpreter           func(*PyThreadState_3_12)

	PyGILState_Check   func() int32
	PyGILState_Ensure  func() PyGILState
	PyGILState_Release func(PyGILState)

	PyEval_SaveThread    func() *PyThreadState_3_12
	PyEval_RestoreThread func(*PyThreadState_3_12)

	PyRun_SimpleString func(string) int32
	PyRun_String       func(s string, token StartToken, globals, locals PyObjectPtr) PyObjectPtr

	PyModule_New          func(string) PyObjectPtr
	PyModule_AddObjectRef func(module PyObjectPtr, name string, item PyObjectPtr) int32

	PyBool_FromLong func(int) PyObjectPtr

	PyLong_FromLong             func(int) PyObjectPtr
	PyLong_FromUnsignedLong     func(uint) PyObjectPtr
	PyLong_FromLongLong         func(int64) PyObjectPtr
	PyLong_FromUnsignedLongLong func(uint64) PyObjectPtr

	PyTuple_New     func(int64) PyObjectPtr
	PyTuple_SetItem func(tuple PyObjectPtr, pos int64, item PyObjectPtr) int32

	PyDict_New           func() PyObjectPtr
	PyDictProxy_New      func(mapping PyObjectPtr) PyObjectPtr
	PyDict_Clear         func(PyObjectPtr)
	PyDict_SetItem       func(dict, key, val PyObjectPtr) int32
	PyDict_SetItemString func(dict PyObjectPtr, key string, val PyObjectPtr) int

	Py_DecRef func(PyObjectPtr)
	Py_IncRef func(PyObjectPtr)

	PyErr_Clear func()
	PyErr_Print func()
)

func registerFuncs(pythonLib uintptr) {
	purego.RegisterLibFunc(&Py_DecodeLocale, pythonLib, "Py_DecodeLocale")
	purego.RegisterLibFunc(&Py_EncodeLocale, pythonLib, "Py_EncodeLocale")
	purego.RegisterLibFunc(&PyWideStringList_Append, pythonLib, "PyWideStringList_Append")

	purego.RegisterLibFunc(&PyPreConfig_InitIsolatedConfig, pythonLib, "PyPreConfig_InitIsolatedConfig")
	purego.RegisterLibFunc(&Py_PreInitialize, pythonLib, "Py_PreInitialize")

	purego.RegisterLibFunc(&PyConfig_InitPythonConfig, pythonLib, "PyConfig_InitPythonConfig")
	purego.RegisterLibFunc(&PyConfig_InitIsolatedPythonConfig, pythonLib, "PyConfig_InitIsolatedConfig")
	purego.RegisterLibFunc(&PyConfig_SetBytesString, pythonLib, "PyConfig_SetBytesString")
	purego.RegisterLibFunc(&PyConfig_Clear, pythonLib, "PyConfig_Clear")
	purego.RegisterLibFunc(&PyConfig_Read, pythonLib, "PyConfig_Read")

	purego.RegisterLibFunc(&Py_FinalizeEx, pythonLib, "Py_FinalizeEx")

	purego.RegisterLibFunc(&Py_NewInterpreterFromConfig, pythonLib, "Py_NewInterpreterFromConfig")

	purego.RegisterLibFunc(&PyGILState_Check, pythonLib, "PyGILState_Check")
	purego.RegisterLibFunc(&PyGILState_Ensure, pythonLib, "PyGILState_Ensure")
	purego.RegisterLibFunc(&PyGILState_Release, pythonLib, "PyGILState_Release")

	purego.RegisterLibFunc(&PyEval_SaveThread, pythonLib, "PyEval_SaveThread")
	purego.RegisterLibFunc(&PyEval_RestoreThread, pythonLib, "PyEval_RestoreThread")

	purego.RegisterLibFunc(&Py_InitializeFromConfig, pythonLib, "Py_InitializeFromConfig")

	purego.RegisterLibFunc(&PyRun_SimpleString, pythonLib, "PyRun_SimpleString")
	purego.RegisterLibFunc(&PyRun_String, pythonLib, "PyRun_String")

	purego.RegisterLibFunc(&PyModule_New, pythonLib, "PyModule_New")
	purego.RegisterLibFunc(&PyModule_AddObjectRef, pythonLib, "PyModule_AddObjectRef")

	purego.RegisterLibFunc(&PyBool_FromLong, pythonLib, "PyBool_FromLong")

	purego.RegisterLibFunc(&PyLong_FromLong, pythonLib, "PyLong_FromLong")
	purego.RegisterLibFunc(&PyLong_FromUnsignedLong, pythonLib, "PyLong_FromUnsignedLong")
	purego.RegisterLibFunc(&PyLong_FromLongLong, pythonLib, "PyLong_FromLongLong")
	purego.RegisterLibFunc(&PyLong_FromUnsignedLongLong, pythonLib, "PyLong_FromUnsignedLongLong")

	purego.RegisterLibFunc(&PyTuple_New, pythonLib, "PyTuple_New")
	purego.RegisterLibFunc(&PyTuple_SetItem, pythonLib, "PyTuple_SetItem")

	purego.RegisterLibFunc(&PyDict_New, pythonLib, "PyDict_New")
	purego.RegisterLibFunc(&PyDictProxy_New, pythonLib, "PyDictProxy_New")
	purego.RegisterLibFunc(&PyDict_Clear, pythonLib, "PyDict_Clear")
	purego.RegisterLibFunc(&PyDict_SetItem, pythonLib, "PyDict_SetItem")
	purego.RegisterLibFunc(&PyDict_SetItemString, pythonLib, "PyDict_SetItemString")

	purego.RegisterLibFunc(&Py_DecRef, pythonLib, "Py_DecRef")
	purego.RegisterLibFunc(&Py_IncRef, pythonLib, "Py_IncRef")

	purego.RegisterLibFunc(&PyErr_Clear, pythonLib, "PyErr_Clear")
	purego.RegisterLibFunc(&PyErr_Print, pythonLib, "PyErr_Print")
}
