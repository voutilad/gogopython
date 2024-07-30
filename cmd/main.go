package main

import (
	"log"
	"strings"

	py "github.com/voutilad/gogopython"
)

func main() {
	err := py.Load_library()
	if err != nil {
		log.Fatalln(err)
	}

	// ==============

	preConfig := py.PyPreConfig{}
	py.PyPreConfig_InitIsolatedConfig(&preConfig)
	preConfig.Allocator = 3
	status := py.Py_PreInitialize(&preConfig)
	if status.Type != 0 {
		log.Fatalln("failed to preinitialize python:", py.PyBytesToString(status.ErrMsg))
	}
	log.Println("preinitialization complete")

	/*
	 * Configure our Paths. We need to approximate an isolated config from regular
	 * because Python will ignore our modifying some values if we initialize an
	 * isolated config. Annoying.
	 */
	config := py.PyConfig_3_12{}
	py.PyConfig_InitPythonConfig(&config)
	defer py.PyConfig_Clear(&config)
	config.ParseArgv = 0
	config.SafePath = 1
	config.UserSiteDirectory = 0
	config.InstallSignalHandlers = 0
	log.Printf("config: %p", &config)

	home := "/Users/dv/src/gogopython/venv"
	status = py.PyConfig_SetBytesString(&config, &config.Home, home)
	if status.Type != 0 {
		log.Fatalln("failed to set home:", py.PyBytesToString(status.ErrMsg))
	}
	log.Println("set home:", home)

	path := strings.Join([]string{
		"/opt/homebrew/Cellar/python@3.12/3.12.4/Frameworks/Python.framework/Versions/3.12/lib/python3.12",
		"/opt/homebrew/Cellar/python@3.12/3.12.4/Frameworks/Python.framework/Versions/3.12/lib/python3.12/lib-dynload",
		"/Users/dv/src/gogopython/venv/lib/python3.12/site-packages",
	}, ":")
	status = py.PyConfig_SetBytesString(&config, &config.PythonPathEnv, path)
	if status.Type != 0 {
		log.Fatalln("failed to set path:", py.PyBytesToString(status.ErrMsg))
	}
	log.Println("set path:", path)

	/////////

	status = py.Py_InitializeFromConfig(&config)
	if status.Type != 0 {
		log.Fatalln("failed to initialize Python:", py.PyBytesToString(status.ErrMsg))
	}
	log.Println("initialized")

	// Create a sub-interpreter to partially isolate the script.
	interpreterConfig := py.PyInterpreterConfig{}
	interpreterConfig.Gil = py.DefaultGil // OwnGil works in 3.12, but is hard to use.
	interpreterConfig.CheckMultiInterpExtensions = 1
	log.Printf("interpreter config: %p", &interpreterConfig)

	mainStatePtr := py.PyThreadState_Get()

	//PyEval_ReleaseThread(mainStatePtr)

	gil := py.PyGILState_Ensure()
	print_current_interpreter()
	py.PyThreadState_Swap(py.NullThreadState)

	var subThreadPtr py.PyThreadStatePtr
	status = py.Py_NewInterpreterFromConfig(&subThreadPtr, &interpreterConfig)
	if status.Type != 0 {
		log.Fatalln("failed to create sub-interpreter:", py.PyBytesToString(status.ErrMsg))
	}
	print_current_interpreter()

	py.Py_EndInterpreter(subThreadPtr)

	py.PyThreadState_Swap(mainStatePtr)
	print_current_interpreter()
	py.PyGILState_Release(gil)

	//PyEval_RestoreThread(mainStatePtr)
	py.Py_FinalizeEx()
}

func print_current_interpreter() {
	/// See https://github.com/python/cpython/blob/2b163aa9e796b312bb0549d49145d26e4904768e/Programs/_testembed.c#L100-L115
	ts := py.PyThreadState_Get()
	me := py.PyThreadState_GetInterpreter(ts)
	id := py.PyInterpreterState_GetID(me)
	log.Printf("interp 0x%x, ts 0x%x, id %d\n", me, ts, id)
	py.PyRun_SimpleString("import sys; print('id(modules) =', id(sys.modules)); sys.stdout.flush()")
}
