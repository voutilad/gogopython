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
	preConfig.Allocator = 3
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
	config := PyConfig_3_12{}
	PyConfig_InitPythonConfig(&config)
	config.ParseArgv = 0
	config.SafePath = 1
	config.UserSiteDirectory = 0
	config.InstallSignalHandlers = 0
	log.Printf("config: %p", &config)

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

	/////////

	status = Py_InitializeFromConfig(&config)
	if status.Type != 0 {
		log.Fatalln("failed to initialize Python:", PyBytesToString(status.ErrMsg))
	}
	log.Println("initialized")

	// Create a sub-interpreter to partially isolate the script.
	interpreterConfig := PyInterpreterConfig{}
	interpreterConfig.Gil = SharedGil
	interpreterConfig.AllowThreads = 1
	interpreterConfig.CheckMultiInterpExtensions = 1
	log.Printf("interpreter config: %p", &interpreterConfig)

	mainStatePtr := PyThreadState_Get()

	PyEval_ReleaseThread(mainStatePtr)

	gil := PyGILState_Ensure()
	print_current_interpreter()
	PyThreadState_Swap(NullThreadState)

	var subThreadPtr PyThreadStatePtr
	status = Py_NewInterpreterFromConfig(&subThreadPtr, &interpreterConfig)
	if status.Type != 0 {
		log.Fatalln("failed to create sub-interpreter:", PyBytesToString(status.ErrMsg))
	}
	print_current_interpreter()
	Py_EndInterpreter(subThreadPtr)

	PyThreadState_Swap(mainStatePtr)
	print_current_interpreter()
	PyGILState_Release(gil)

	PyEval_RestoreThread(mainStatePtr)
	Py_FinalizeEx()
}

func print_current_interpreter() {
	ts := PyThreadState_Get()
	me := PyThreadState_GetInterpreter(ts)
	id := PyInterpreterState_GetID(me)
	log.Printf("interp 0x%x, ts 0x%x, id %d\n", me, ts, id)
	PyRun_SimpleString("import sys; print('id(modules) =', id(sys.modules)); sys.stdout.flush()")
}
