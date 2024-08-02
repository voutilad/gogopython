package gogopython

// Opaque pointer to an underlying PyObject instance.
type PyObjectPtr uintptr

// Opaque pointer to an underlying PyTypeObject instance.
type PyTypeObjectPtr uintptr

// Opaque pointer to a Python wchar_t string.
type WCharPtr *byte

// Opaque state of the GIL, used sort of as a cookie in the ensure/release
// function calls.
type PyGILState int32

// Represents a NULL pointer to a Python PyObject
const NullPyObjectPtr PyObjectPtr = 0

// Represents a NULL pointer to a Python PyTypeObject
const NullPyTypeObjectPtr PyTypeObjectPtr = 0

// Confusingly named, but used to dictate to the Python interpretser & compiler
// how to interpret the provided Python script in string form.
type StartToken = int32

const (
	PySingleInput   StartToken = 256 // Used for single statements.
	PyFileInput     StartToken = 257 // Used for modules (many statements).
	PyEvalInput     StartToken = 258 // Used for expressions(?).
	PyFuncTypeInput StartToken = 345 // No idea what this is!
)

// PyObject types based on inspecting the tpflags of a PyTypeObject
type Type uint64

const (
	Long    Type = (1 << 24) // Python long.
	List    Type = (1 << 25) // Python list.
	Tuple   Type = (1 << 26) // Python tuple.
	Bytes   Type = (1 << 27) // Python bytes (not bytearray).
	String  Type = (1 << 28) // Python unicode string.
	Dict    Type = (1 << 29) // Python dictionary.
	None    Type = 0         // The Python "None" type.
	Unknown Type = 1         // We have no idea what the type is...
)

const (
	typeMask              = (0x3f << 24) // flags mask to get type bits
	immutableFlag         = (1 << 8)     // bit that describes an immutable object
	allowsSubclassingFlag = (1 << 10)    // bit that describes if a type can be subclassed
	// Our heuristic for detecting a Python None: it cannot be mutated and
	// it cannot be subclassed.
	noneMask = (immutableFlag | allowsSubclassingFlag)
)

// Return status of some specific Python C API calls.
//
// This is the biggest headache of this whole thing. A few functions return
// this struct directly instead of either via a pointer or by reference in
// the function args. It creates a nightware to deal with the various ABI
// logic for how structs get returned that don't fit into a cpu register
// width.
//
// If someone has a time machine, please go back and tell Guido not to do
// this. Please.
type PyStatus struct {
	Type     int32
	Func     WCharPtr
	ErrMsg   WCharPtr
	ExitCode int32
}

type PyMemAllocator = int32

const (
	PyMemAllocator_NotSet        = iota // Don't change allocator (use defaults).
	PyMemAllocator_Default              // Use defaults allocators.
	PyMemAllocator_Debug                // Default with debug hooks.
	PyMemAllocator_Malloc               // Use malloc(3).
	PyMemAllocator_MallocDebug          // Use malloc(3) with debug hooks.
	PyMemAllocator_PyMalloc             // Use Python's pymalloc.
	PyMemAllocator_PyMallocDebug        // Use Python's pymalloc with debug hooks.
)

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
	Allocator PyMemAllocator
}

type pyWideStringList struct {
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

	OrigArgv    pyWideStringList
	Argv        pyWideStringList
	XOptions    pyWideStringList
	WarnOptions pyWideStringList

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
	ModuleSearchPaths    pyWideStringList
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
	DefaultGil GilType = 0 // On Python 3.12, defaults to SharedGil
	SharedGil  GilType = 1 // On Python 3.12 and newer, uses unified GIL.
	OwnGil     GilType = 2 // On Python 3.12 and newer, creates a unique GIL.
)

// Configuration for a sub-interpreter. All int32 values are really booleans,
// so 0 = false, 1 = true. (Non-zero may also = true. Not sure!)
type PyInterpreterConfig struct {
	// Whether to share the main interpreters object allocator state.
	//
	// If this is 0, you must set CheckMultiInterpExtensions to 1.
	// If this is 1, you must set Gil to OwnGil.
	UseMainObMalloc int32

	// Whether to allow using Python's os.fork funcion.
	// Note: this doesn't block exec syscalls and subprocess module will still work.
	AllowFork int32

	// Whether to allow using Python's os.exec* functions.
	// Note: this doesn't block exec syscalls and subprocess module will still work.
	AllowExec int32

	// Whether to allow creating Python threads using the threading module.
	AllowThreads int32

	// Whether to allow creating Python daemon threads.
	AllowDaemonThreads int32

	// If 1, require multi-phase (non-legacy) extension modules. Must be 1 if you
	// enable UseMainObMalloc.
	CheckMultiInterpExtensions int32

	// The GIL mode for this sub-interpreter.
	Gil GilType
}

// Opaque pointer to a Python ThreadState.
type PyThreadStatePtr uintptr

// Opaque pointer to a Python InterpreterState.
type PyInterpreterStatePtr uintptr

// NULL version of a Python ThreadState.
const NullThreadState PyThreadStatePtr = 0

// NULL version of a Python InterpreterState.
const NullInterpreterState PyInterpreterStatePtr = 0

// Opaque pointer to a Python dynamic library state in purego.
type PythonLibraryPtr = uintptr
