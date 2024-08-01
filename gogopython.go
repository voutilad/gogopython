package gogopython

import (
	"bufio"
	"errors"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"unsafe"

	"github.com/ebitengine/purego"
)

type PyObjectPtr uintptr
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
	Items  uintptr
}

type PyConfig_3_12 struct {
	ConfigInit int32

	Isolated              int32
	UseEnvironment        int32
	DevMode               int32
	InstallSignalHandlers int32
	UseHashSeed           int32
	HashSeed              uint
	FaultHandler          int32
	TraceMalloc           int32
	PerfProfiling         int32
	ImportTime            int32
	CodeDebugRanges       int32
	ShowRefCount          int32
	DumpRefs              int32
	DumpRefsFile          WCharPtr
	MallocStats           int32
	FilesystemEncoding    WCharPtr
	FilesystemErrors      WCharPtr
	PycachePrefix         WCharPtr
	ParseArgv             int32

	OrigArgv    PyWideStringList
	Argv        PyWideStringList
	XOptions    PyWideStringList
	WarnOptions PyWideStringList

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

type PyThreadStatePtr uintptr
type PyInterpreterStatePtr uintptr

const NullThreadState PyThreadStatePtr = 0

var (
	Py_DecodeLocale func(string, uint64) WCharPtr
	Py_EncodeLocale func(WCharPtr, uint64) string

	PyPreConfig_InitIsolatedConfig func(*PyPreConfig)

	PyConfig_InitPythonConfig         func(*PyConfig_3_12)
	PyConfig_InitIsolatedPythonConfig func(*PyConfig_3_12)
	PyConfig_Clear                    func(*PyConfig_3_12)

	Py_FinalizeEx func() int32

	Py_EndInterpreter func(PyThreadStatePtr)

	// Check if we have the GIL. 1 if true, 0 if false.
	PyGILState_Check func() int32
	// Take a reference to the GIL. Caution: this is recursive.
	PyGILState_Ensure func() PyGILState
	// Release a reference to the GIL.
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

type PythonLibraryPtr = uintptr

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

	purego.RegisterLibFunc(&PyBool_FromLong, lib, "PyBool_FromLong")

	purego.RegisterLibFunc(&PyLong_FromLong, lib, "PyLong_FromLong")
	purego.RegisterLibFunc(&PyLong_FromUnsignedLong, lib, "PyLong_FromUnsignedLong")
	purego.RegisterLibFunc(&PyLong_FromLongLong, lib, "PyLong_FromLongLong")
	purego.RegisterLibFunc(&PyLong_FromUnsignedLongLong, lib, "PyLong_FromUnsignedLongLong")

	purego.RegisterLibFunc(&PyTuple_New, lib, "PyTuple_New")
	purego.RegisterLibFunc(&PyTuple_SetItem, lib, "PyTuple_SetItem")

	purego.RegisterLibFunc(&PyDict_New, lib, "PyDict_New")
	purego.RegisterLibFunc(&PyDictProxy_New, lib, "PyDictProxy_New")
	purego.RegisterLibFunc(&PyDict_Clear, lib, "PyDict_Clear")
	purego.RegisterLibFunc(&PyDict_SetItem, lib, "PyDict_SetItem")
	purego.RegisterLibFunc(&PyDict_SetItemString, lib, "PyDict_SetItemString")

	purego.RegisterLibFunc(&Py_DecRef, lib, "Py_DecRef")
	purego.RegisterLibFunc(&Py_IncRef, lib, "Py_IncRef")

	purego.RegisterLibFunc(&PyErr_Clear, lib, "PyErr_Clear")
	purego.RegisterLibFunc(&PyErr_Print, lib, "PyErr_Print")

	registerFuncsPlatDependent(lib)
}

func Load_library(exe string) error {
	var library string
	var err error

	switch os := runtime.GOOS; os {
	case "darwin":
		// On macOS, let's assume Python3 was installed not via XCode
		// (which ships a fat binary with amd64 & arm64).
		// We can use otool, if available, to find the python framework.
		library, err = findLibraryOnMacOS(exe)
		if err != nil {
			return err
		}
	case "linux":
		library = "libpython3.so"
	default:
		log.Fatalln("unsupported runtime:", os)
	}

	lib, err := purego.Dlopen(library, purego.RTLD_NOW)
	if err != nil {
		return err
	}

	registerFuncs(lib)

	return nil
}

// Try using otool to find the Python library.
// XXX maybe move this to gogopython_darwin.go?
func findLibraryOnMacOS(exe string) (string, error) {
	lib := ""

	// First resolve the location if we're given just "python3"
	cmd := exec.Command("command", "-v", exe)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return lib, err
	}
	if err = cmd.Start(); err != nil {
		return lib, err
	}
	path, err := bufio.NewReader(stdout).ReadString(byte('\n'))
	if err != nil {
		return lib, err
	}
	if err = cmd.Wait(); err != nil {
		return lib, err
	}

	cmd = exec.Command("otool", "-L", strings.TrimRight(path, "\n"))
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		return lib, err
	}

	if err = cmd.Start(); err != nil {
		return lib, err
	}

	scanner := bufio.NewScanner(bufio.NewReader(stdout))
	for scanner.Scan() {
		// We should have a line pointing to a Python.framework location.
		text := scanner.Text()
		log.Printf("text: %s\n", text)
		if len(lib) == 0 && strings.Contains(text, "Python.framework") {
			// Should look something like:
			//    /something/Python.framework/Versions/3.12/Python (compatibility ...)
			parts := strings.SplitAfterN(strings.TrimLeft(text, " \t"), " ", 2)
			if len(parts) < 2 {
				return "", errors.New("could not parse otool output")
			}
			lib = strings.TrimRight(parts[0], " ")
		}
	}
	err = cmd.Wait()
	return lib, err
}
