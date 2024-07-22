package main

import (
	"log"

	"github.com/ebitengine/purego"
)

type PyStatus struct {
	Type     int32
	Func     *uint8
	ErrMsg   *uint8
	ExitCode int32
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
	Items  **uint16
}

type PyConfig struct {
	ConfigInit            int32
	Isolated              int32
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
	DumpRefsFile          *uint16
	MallocStats           int32
	FilesystemEncoding    *uint16
	FilesystemErrors      *uint16
	PycachePrefix         *uint16
	ParseArgv             int32
	OrigArgv              PyWideStringList
	Argv                  PyWideStringList
	XOptions              PyWideStringList
	WarnOptions           PyWideStringList
	SiteImport            int32
	BytesWarning          int32
	WarnDefaultEncoding   int32
	Inspect               int32
	Interactive           int32
	OptimizationLevel     int32
	ParserDebug           int32
	WriteBytecode         int32
	Verbose               int32
	Quiet                 int32
	UserSiteDirectory     int32
	ConfigureCStdio       int32
	BufferedStdio         int32
	StdioEncodings        *uint16
	StdioErrors           *uint16
	// LegacyWindowsStdio // if windows
	CheckHashPycsMode *uint16
	UseFrozenModules  int32
	SafePath          int32
	IntMaxStrDigits   int32
	CpuCount          int32
	// EnableGil int // if gil disabled

	/* Path configuration outputs */
	ModuleSearchPathsSet int32
	ModuleSearchPaths    PyWideStringList
	StdlibDir            *uint16
	Executable           *uint16
	BaseExecutable       *uint16
	Prefix               *uint16
	BasePrefix           *uint16
	ExecPrefix           *uint16
	BaseExecPrefix       *uint16

	/* Parameter only used by Py_Main */
	SkipSourceFirstLine int32
	RunCommand          *uint16
	RunModule           *uint16
	RunFilename         *uint16

	/* Set by Py_Main */
	SysPath0 *uint16

	/* Private Fields */
	installImportLib int32
	initMain         int32
	isPythonBuild    int32
	// pystats          int // if Py_Stats
	// runPresite       *uint16t // if Py_DEBUG
}

var (
	py_InitializeEx func(int)
	py_FinalizeEx   func()
	py_GetPath      func() string
	py_SetPath      func(*byte)

	pyPreConfig_InitIsolatedConfig func(c *PyPreConfig)
	py_PreInitialize               func(c *PyPreConfig) int32 // this is crap

	pyrun_SimpleString func(string) int
)

func main() {
	python, err := purego.Dlopen("libpython3.so", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		log.Fatalln("dlopen: ", err)
	}

	program := `
import sys
print(f"path: {sys.path}")

import httpx
r = httpx.get("https://api.ipify.org?format=json")
print(r)
`

	purego.RegisterLibFunc(&py_InitializeEx, python, "Py_InitializeEx")
	purego.RegisterLibFunc(&py_FinalizeEx, python, "Py_FinalizeEx")
	purego.RegisterLibFunc(&py_GetPath, python, "Py_GetPath")
	purego.RegisterLibFunc(&py_SetPath, python, "Py_SetPath")
	purego.RegisterLibFunc(&pyPreConfig_InitIsolatedConfig, python, "PyPreConfig_InitIsolatedConfig")
	purego.RegisterLibFunc(&py_PreInitialize, python, "Py_PreInitialize")
	purego.RegisterLibFunc(&pyrun_SimpleString, python, "PyRun_SimpleString")

	config := PyPreConfig{}
	pyPreConfig_InitIsolatedConfig(&config)
	config.Utf8Mode = 1

	log.Println("config is isolated? ", config.Isolated)
	log.Println("utf8 is enabled? ", config.Utf8Mode)

	//status := py_PreInitialize(&config)
	//log.Println("status? ", status)

	py_InitializeEx(0)

	result := pyrun_SimpleString(program)
	log.Println("returned ", result)
	py_FinalizeEx()
}
