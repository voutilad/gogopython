package main

import (
	"log"
	"runtime"
	"strings"
	"unsafe"

	"github.com/ebitengine/purego"
)

type PyStatus struct {
	Type     int32
	Func     *byte
	ErrMsg   *byte
	ExitCode int32
}

func PyBytesToString(b *byte) string {
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
	StdioEncodings      *byte
	StdioErrors         *byte
	// LegacyWindowsStdio  int32 // if windows
	CheckHashPycsMode *byte
	UseFrozenModules  int32
	SafePath          int32
	IntMaxStrDigits   int32
	// CpuCount          int32
	// EnableGil         int32 // if gil disabled

	/* Path configuration inputs */
	PathConfigWarnings int32
	ProgramName        *byte
	PythonPathEnv      *byte
	Home               *byte
	PlatLibDir         *byte

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
	installImportLib int32
	initMain         int32
	isPythonBuild    int32
	// pystats          int32 // if Py_Stats
	// runPresite       *byte // if Py_DEBUG
}

var (
	py_GetPath func() *byte
	py_SetPath func(*byte)

	py_DecodeLocale         func(string, uint64) *byte
	py_EncodeLocale         func(*byte, uint64) string
	pyWideStringList_Append func(*PyWideStringList, *byte) PyStatus

	pyPreConfig_InitIsolatedConfig func(*PyPreConfig)
	py_PreInitialize               func(*PyPreConfig) PyStatus

	pyConfig_InitPythonConfig         func(*PyConfig)
	pyConfig_InitIsolatedPythonConfig func(*PyConfig)
	pyConfig_SetBytesString           func(*PyConfig, **byte, string) PyStatus
	pyConfig_Clear                    func(*PyConfig)
	py_InitializeFromConfig           func(*PyConfig) PyStatus
	pyConfig_Read                     func(*PyConfig) PyStatus

	pyrun_SimpleString func(string) int32
)

func main() {
	var library string
	switch runtime.GOOS {
	case "darwin":
		library = "/opt/homebrew/opt/python3/Frameworks/Python.framework/Versions/3.12/lib/libpython3.12.dylib"
		break
	case "linux":
		library = "libpython3.so"
		break
	default:
		log.Fatalln("unsupported runtime: ", runtime.GOOS)
	}

	python, err := purego.Dlopen(library, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		log.Fatalln("dlopen: ", err)
	}

	// ============================

	purego.RegisterLibFunc(&py_GetPath, python, "Py_GetPath")
	purego.RegisterLibFunc(&py_SetPath, python, "Py_SetPath")

	purego.RegisterLibFunc(&py_DecodeLocale, python, "Py_DecodeLocale")
	purego.RegisterLibFunc(&py_EncodeLocale, python, "Py_EncodeLocale")
	purego.RegisterLibFunc(&pyWideStringList_Append, python, "PyWideStringList_Append")

	purego.RegisterLibFunc(&pyPreConfig_InitIsolatedConfig, python, "PyPreConfig_InitIsolatedConfig")
	purego.RegisterLibFunc(&py_PreInitialize, python, "Py_PreInitialize")

	purego.RegisterLibFunc(&pyConfig_InitPythonConfig, python, "PyConfig_InitPythonConfig")
	purego.RegisterLibFunc(&pyConfig_InitIsolatedPythonConfig, python, "PyConfig_InitIsolatedConfig")
	purego.RegisterLibFunc(&pyConfig_SetBytesString, python, "PyConfig_SetBytesString")
	purego.RegisterLibFunc(&pyConfig_Clear, python, "PyConfig_Clear")
	purego.RegisterLibFunc(&pyConfig_Read, python, "PyConfig_Read")

	purego.RegisterLibFunc(&py_InitializeFromConfig, python, "Py_InitializeFromConfig")

	purego.RegisterLibFunc(&pyrun_SimpleString, python, "PyRun_SimpleString")

	// ============================

	preConfig := PyPreConfig{}
	pyPreConfig_InitIsolatedConfig(&preConfig)
	status := py_PreInitialize(&preConfig)
	if status.Type != 0 {
		log.Fatalln("failed to preinitialize python:", PyBytesToString(status.ErrMsg))
	}
	log.Println("preinitialization complete")

	/* Configure our Paths */
	config := PyConfig{}
	pyConfig_InitPythonConfig(&config)
	defer pyConfig_Clear(&config)
	config.PathConfigWarnings = 0
	config.TraceMalloc = 0
	config.ParseArgv = 0
	config.SafePath = 1
	config.UserSiteDirectory = 0

	home := "/Users/dv/src/gogopython/venv"
	status = pyConfig_SetBytesString(&config, &config.Home, home)
	if status.Type != 0 {
		log.Fatalln("failed to set home:", PyBytesToString(status.ErrMsg))
	}
	log.Println("set home:", home)

	path := strings.Join([]string{
		"/opt/homebrew/Cellar/python@3.12/3.12.4/Frameworks/Python.framework/Versions/3.12/lib/python3.12",
		"/opt/homebrew/Cellar/python@3.12/3.12.4/Frameworks/Python.framework/Versions/3.12/lib/python3.12/lib-dynload",
		"/Users/dv/src/gogopython/venv/lib/python3.12/site-packages",
	}, ":")
	status = pyConfig_SetBytesString(&config, &config.PythonPathEnv, path)
	if status.Type != 0 {
		log.Fatalln("failed to set path:", PyBytesToString(status.ErrMsg))
	}
	log.Println("set path:", path)

	status = py_InitializeFromConfig(&config)
	if status.Type != 0 {
		log.Fatalln("failed to initialize Python:", PyBytesToString(status.ErrMsg))
	}
	log.Println("initialized")

	program := `
import sys
print(sys.path)

import httpx
`
	if pyrun_SimpleString(program) != 0 {
		log.Fatalln("failed to run program")
	}
}
