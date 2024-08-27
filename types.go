package gogopython

// PyObjectPtr is a pointer to an underlying PyObject instance.
type PyObjectPtr uintptr

// PyTypeObjectPtr is a pointer to an underlying PyTypeObject instance.
type PyTypeObjectPtr PyObjectPtr

// PyCodeObjectPtr is a pointer to an underlying Python code object.
type PyCodeObjectPtr PyObjectPtr

// WCharPtr is a pointer to a Python wchar_t string.
type WCharPtr *byte

// PyGILState is used sort as a cookie in the ensure/release function calls.
type PyGILState int32

// NullPyObjectPtr represents a NULL pointer to a Python PyObject
const NullPyObjectPtr PyObjectPtr = 0

// NullPyTypeObjectPtr represents a NULL pointer to a Python PyTypeObject
const NullPyTypeObjectPtr PyTypeObjectPtr = 0

// NullPyCodeObjectPtr represents a NULL pointer to a Python code object.
const NullPyCodeObjectPtr PyCodeObjectPtr = 0

type EncodingErrors = string

const (
	Strict            EncodingErrors = "strict"
	Ignore            EncodingErrors = "ignore"
	Replace           EncodingErrors = "replace"
	XmlCharRefReplace EncodingErrors = "xmlcharrefreplace"
	BackslashReplace  EncodingErrors = "backslashreplace"
)

// StartToken (confusingly named) is used to dictate to the Python
// interpreter and compiler how to evaluate a Python script.
type StartToken = int32

const (
	PySingleInput   StartToken = 256 // Used for single statements.
	PyFileInput     StartToken = 257 // Used for modules (many statements).
	PyEvalInput     StartToken = 258 // Used for expressions(?).
	PyFuncTypeInput StartToken = 345 // No idea what this is!
)

// Type describes the native Python type of PyObject.
type Type uint64

const (
	Long      Type = (1 << 24) // Python long.
	List      Type = (1 << 25) // Python list.
	Tuple     Type = (1 << 26) // Python tuple.
	Bytes     Type = (1 << 27) // Python bytes (not bytearray).
	String    Type = (1 << 28) // Python unicode string.
	Dict      Type = (1 << 29) // Python dictionary.
	None      Type = 0         // The Python "None" type.
	Float     Type = 1         // Python float.
	Set       Type = 2         // Python set.
	Function  Type = 3         // Python function.
	Generator Type = 4         // Python generator.
	Module    Type = 5         // Python module.

	Unknown Type = 0xffffffff // We have no idea what the type is...
)

// String converts a Type to a human-readable string representation.
func (t Type) String() string {
	switch t {
	case Long:
		return "Long"
	case List:
		return "List"
	case Tuple:
		return "Tuple"
	case Bytes:
		return "Bytes"
	case String:
		return "String"
	case Dict:
		return "Dict"
	case None:
		return "None"
	case Float:
		return "Float"
	case Set:
		return "Set"
	case Function:
		return "Function"
	case Generator:
		return "Generator"
	case Module:
		return "Module"
	}
	return "Unknown"
}

const (
	typeMask              = (0x3f << 24) // flags mask to get type bits
	builtIn               = (1 << 1)     // flags for built-in types
	disallowInstantiation = (1 << 7)     // disallow creating new instances
	immutableFlag         = (1 << 8)     // bit that describes an immutable object
	allowsSubclassingFlag = (1 << 10)    // bit that describes if a type can be subclassed
	hasVectorCall         = (1 << 11)    // set if the type implements the vectorcall protocol (PEP 590)
	ready                 = (1 << 12)    // type is fully initialized
	supportsGc            = (1 << 14)    // Type supports garbage collection
	methodDescriptor      = (1 << 17)    // Object behaves like an unbound method
	validVersion          = (1 << 19)    // unused legacy flag
	matchSelf             = (1 << 22)    // "undocumented" flag for some built-ins regarding pattern matching

	// Heuristic for detecting a Python None. Only these bits should be set.
	noneMask = (builtIn | immutableFlag | ready | validVersion)

	// Our heuristic for detecting a Python float since it's not a type in
	// the set of types that have special bits.
	floatMask = (noneMask | matchSelf)

	// Our heuristic for detecting a Python set. It looks like a float, but
	// supports garbage collection as it's a container.
	setMask = (noneMask | matchSelf | supportsGc)

	// Our heuristic for detecting a Python function.
	fnMask = (noneMask | hasVectorCall | supportsGc | methodDescriptor)

	// Our heuristic for detecting a Python generator or iterator.
	genMask = (noneMask | disallowInstantiation | supportsGc)

	// Our heuristic for detecting a Python module.
	moduleMask = (noneMask | allowsSubclassingFlag | supportsGc)
)

// PyStatus is returned by some Python C API calls.
//
// This is the biggest headache of this whole thing. A few functions return
// this struct directly instead of either via a pointer or by reference in
// the function args. It creates a nightmare to deal with the various ABI
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

type PyCompilerFlags struct {
	Flags          int32
	FeatureVersion int32
}

type OptimizeLevel = int32

const (
	UseInterpreterLevel              OptimizeLevel = -1 // Uses whatever the interpreter was built with.
	NoOptimization                   OptimizeLevel = 0  // No optimization, __debug__ is True.
	RemoveDebugsAndAsserts           OptimizeLevel = 1  // __debug__ is False, no asserts.
	RemoveDebugsAssertsAndDocstrings OptimizeLevel = 2  // __debug__ is False, no asserts, no docstrings.
)

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

// PyConfig_3_12 configures a Python 3.12 interpreter.
//
// This is a version-dependent structure, unfortunately. We need this because
// it's the stable way of configuring the Home and Path (PythonPathEnv).
//
// Sadly this is also dependent on platform (Windows vs. not-Windows) and some
// compile time decisions for the Python implementation (e.g. debug, stats).
//
// Ultimately, this should be made private and the configuration complexity
// hidden from the programmer.
type PyConfig_3_12 struct {
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

// PyInterpreterConfig defines settings for a sub-interpreter. All int32
// values are really booleans, so 0 = false, 1 = true. (Non-zero may
// also be equivalent to true, but I'm not sure!)
type PyInterpreterConfig struct {
	// Whether to share the main interpreters object allocator state.
	//
	// If this is 0, you must set CheckMultiInterpExtensions to 1.
	// If this is 1, you must set Gil to OwnGil.
	UseMainObMalloc int32

	// Whether to allow using Python's os.fork function.
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

// PyThreadStatePtr is a pointer to a Python ThreadState.
type PyThreadStatePtr uintptr

// PyInterpreterStatePtr is a pointer to a Python InterpreterState.
type PyInterpreterStatePtr uintptr

// NullThreadState is a NULL version of a Python ThreadState.
const NullThreadState PyThreadStatePtr = 0

// NullInterpreterState is a NULL version of a Python InterpreterState.
const NullInterpreterState PyInterpreterStatePtr = 0

// PythonLibraryPtr is a pointer to a Python dynamic library state.
type PythonLibraryPtr = uintptr

// PyCFunction points to a C function implementation.
type PyCFunction = uintptr

// MethodFlags bits indicate how the method call is constructed.
type MethodFlags int32

const (
	MethodVarArgs  MethodFlags = 0x01
	MethodKeywords MethodFlags = 0x02
	MethodNoArgs   MethodFlags = 0x04

	// More exist...left out for brevity.

	MethodFastCall MethodFlags = 0x80
)

const (
	// PyCFunctionDefault has the C function signature:
	//   PyObject *PyCFunction(PyObject *self, PyObject *args)
	PyCFunctionDefault MethodFlags = MethodVarArgs

	// PyCFunctionWithKeywords has the C function signature:
	//   PyObject *fn(PyObject *self, PyObject *args, PyObject *kwargs)
	PyCFunctionWithKeywords MethodFlags = MethodVarArgs | MethodKeywords

	// PyCFunctionFast has the C function signature:
	//   PyObject *fn(PyObject *self, PyObject *const *args, Py_ssize_t nargs)
	PyCFunctionFast MethodFlags = MethodFastCall
)

// PyMethodDef describes a C function made callable from Python.
type PyMethodDef struct {
	Name      *byte       // Name is a C string identifying the method.
	Method    PyCFunction // Method points to the C function to be called.
	Flags     MethodFlags // Flags is the bit-wise configuration of how the method is invoked.
	Docstring *byte       // Docstring is a C string describing documentation for the method.
}

type PySendResult int32

const (
	PyGen_Return PySendResult = 0
	PyGen_Error  PySendResult = -1
	PyGen_Next   PySendResult = 1
)
