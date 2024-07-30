package main

import (
	"log"
	"runtime"
	"strings"

	"github.com/ebitengine/purego"
)

func main() {
	var library string
	switch os := runtime.GOOS; os {
	case "darwin":
		library = "/opt/homebrew/opt/python3/Frameworks/Python.framework/Versions/3.12/lib/libpython3.12.dylib"
	case "linux":
		// Need to update purego to handle unmarshaling structs returned on the stack.
		// Python functions sometimes return a PyStatus object. So, so annoying. Who
		// does this?! I guess it's a consequence of being written for i386...sigh.
		log.Fatalln("ugh, python does dumb things like returning a struct on the stack...no worky yet on Linux")
		// library = "libpython3.so"
	default:
		log.Fatalln("unsupported runtime:", os)
	}

	python, err := purego.Dlopen(library, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		log.Fatalln("dlopen: ", err)
	}
	registerFuncs(python)

	// ==============

	preConfig := PyPreConfig{}
	PyPreConfig_InitIsolatedConfig(&preConfig)
	status := Py_PreInitialize(&preConfig)
	if status.Type != 0 {
		log.Fatalln("failed to preinitialize python:", PyBytesToString(status.ErrMsg))
	}
	log.Println("preinitialization complete")

	/*
	 * Configure our Paths. We need to approximate an isolated config from regular
	 * because Python will ignore our modifying some values if we initialize an
	 * isolated config. Annoying.
	 */
	config := PyConfig{}
	PyConfig_InitPythonConfig(&config)
	defer PyConfig_Clear(&config)
	config.ParseArgv = 0
	config.SafePath = 1
	config.UserSiteDirectory = 0
	config.InstallSignalHandlers = 0

	home := "/Users/dv/src/gogopython/venv"
	status = PyConfig_SetBytesString(&config, &config.Home, home)
	if status.Type != 0 {
		log.Fatalln("failed to set home:", PyBytesToString(status.ErrMsg))
	}
	log.Println("set home:", home)

	path := strings.Join([]string{
		"/opt/homebrew/Cellar/python@3.12/3.12.4/Frameworks/Python.framework/Versions/3.12/lib/python3.12",
		"/opt/homebrew/Cellar/python@3.12/3.12.4/Frameworks/Python.framework/Versions/3.12/lib/python3.12/lib-dynload",
		"/Users/dv/src/gogopython/venv/lib/python3.12/site-packages",
	}, ":")
	status = PyConfig_SetBytesString(&config, &config.PythonPathEnv, path)
	if status.Type != 0 {
		log.Fatalln("failed to set path:", PyBytesToString(status.ErrMsg))
	}
	log.Println("set path:", path)

	status = Py_InitializeFromConfig(&config)
	if status.Type != 0 {
		log.Fatalln("failed to initialize Python:", PyBytesToString(status.ErrMsg))
	}
	log.Println("initialized")
	defer Py_FinalizeEx()

	log.Println("GIL held?", PyGILState_Check())

	// Globals and Locals
	globals := PyDict_New()
	defer Py_DecRef(globals)
	locals := PyDict_New()
	defer Py_DecRef(locals)

	// This is the input.
	this := PyDict_New()
	defer Py_DecRef(this)
	proxy := PyDictProxy_New(this)
	defer Py_DecRef(proxy)

	PyDict_SetItemString(this, "junk", PyDict_New())

	// Root is the output.
	root := PyDict_New()
	defer Py_DecRef(root)

	state := PyEval_SaveThread()
	log.Println("GIL held?", PyGILState_Check())
	PyEval_RestoreThread(state)
	log.Println("GIL held?", PyGILState_Check())

	if PyDict_SetItemString(globals, "this", proxy) != 0 {
		log.Fatalln("failed to add 'this' proxy to globals")
	}
	if PyDict_SetItemString(globals, "root", root) != 0 {
		log.Fatalln("failed to add 'root' to globals")
	}

	program := `
print(this)
this["junk"].update({"name": "Dave"})
print(this)
`
	if PyRun_String(program, PyFileInput, globals, locals) == nil {
		PyErr_Print()
		PyErr_Clear()
	}
}
