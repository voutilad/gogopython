package main

import (
	"log"
	"strings"
)

func main() {
	python, err := load_library()
	if err != nil {
		log.Fatalln(err)
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
	defer PyConfig_Clear(&config)
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
	interpreterConfig.Gil = DefaultGil // OwnGil works in 3.12, but is hard to use.
	interpreterConfig.CheckMultiInterpExtensions = 1
	log.Printf("interpreter config: %p", &interpreterConfig)

	mainStatePtr := PyThreadState_Get()

	//PyEval_ReleaseThread(mainStatePtr)

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

	//PyEval_RestoreThread(mainStatePtr)
	Py_FinalizeEx()
}

func print_current_interpreter() {
	/// See https://github.com/python/cpython/blob/2b163aa9e796b312bb0549d49145d26e4904768e/Programs/_testembed.c#L100-L115
	ts := PyThreadState_Get()
	me := PyThreadState_GetInterpreter(ts)
	id := PyInterpreterState_GetID(me)
	log.Printf("interp 0x%x, ts 0x%x, id %d\n", me, ts, id)
	PyRun_SimpleString("import sys; print('id(modules) =', id(sys.modules)); sys.stdout.flush()")
}
