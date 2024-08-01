package main

import (
	"log"
	"os"
	"runtime"
	"strings"
	"sync/atomic"

	py "github.com/voutilad/gogopython"
)

func main() {
	home := ""
	path := ""
	if len(os.Args) > 1 {
		home = os.Args[1]
	}
	if len(os.Args) > 2 {
		path = strings.Join(os.Args[2:], ":")
	}

	err := py.Load_library("python3")
	if err != nil {
		log.Fatalln(err)
	}

	// ==============

	preConfig := py.PyPreConfig{}
	py.PyPreConfig_InitIsolatedConfig(&preConfig)
	preConfig.Allocator = 3
	var status py.PyStatus
	status = py.Py_PreInitialize(&preConfig)
	log.Println("HERE")
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

	log.Println("using home:", home)
	status = py.PyConfig_SetBytesString(&config, &config.Home, home)
	if status.Type != 0 {
		log.Fatalln("failed to set home:", py.PyBytesToString(status.ErrMsg))
	}
	log.Println("set home:", home)

	log.Println("using path:", path)
	status = py.PyConfig_SetBytesString(&config, &config.PythonPathEnv, path)
	if status.Type != 0 {
		log.Fatalln("failed to set path:", py.PyBytesToString(status.ErrMsg))
	}
	log.Println("set path:", path)

	// Initialize our main interpreter in our main go routine.
	status = py.Py_InitializeFromConfig(&config)
	if status.Type != 0 {
		log.Fatalln("failed to initialize Python:", py.PyBytesToString(status.ErrMsg))
	}
	log.Println("initialized")
	mainTs := py.PyThreadState_Get()
	//mainInt := py.PyInterpreterState_Get()

	// Add a GIL reference, print, and drop our thread state.
	//gil := py.PyGILState_Ensure()
	print_current_interpreter()
	py.PyThreadState_Swap(py.NullThreadState)

	// Create a Subinterpreter in our main go routine so it's tied to the current thread.
	var subThreadPtr py.PyThreadStatePtr
	interpreterConfig := py.PyInterpreterConfig{}
	interpreterConfig.Gil = py.DefaultGil // OwnGil works in 3.12, but is hard to use.
	interpreterConfig.CheckMultiInterpExtensions = 1
	status = py.Py_NewInterpreterFromConfig(&subThreadPtr, &interpreterConfig)
	if status.Type != 0 {
		log.Fatalln("failed to create sub-interpreter:", py.PyBytesToString(status.ErrMsg))
	}
	print_current_interpreter()

	// Get a pointer to our interpreter state, which should not be thread local afaik.
	subint := py.PyInterpreterState_Get()

	// Remove our threadstate and release the GIL
	ts := py.PyEval_SaveThread()

	// Launch a go routine and busy wait for it to finish (trying to force another os thread to pick it up)
	var signal atomic.Bool
	signal.Store(true)
	log.Println("launching go routine")
	go func() {
		log.Println("go routine starting")
		runtime.LockOSThread()
		ts := py.PyThreadState_New(subint)

		log.Println("created new thread state")

		log.Printf("gil? %d\n", py.PyGILState_Check())
		py.PyEval_RestoreThread(ts)

		print_current_interpreter()

		log.Println("go routine running python script")
		py.PyRun_SimpleString("import time; print('python is sleeping'); time.sleep(0.2); print('python is awake!')")

		globals := py.PyDict_New()
		locals := py.PyDict_New()
		program := "x = {'name': 'dave'}"
		output := py.PyRun_String(program, py.PyFileInput, globals, locals)
		if output == py.NullPyObjectPtr {
			log.Println("null result? Huh.")
		} else {
			x := py.PyDict_GetItemString(locals, "x")
			outputType := py.Py_BaseType(x)
			log.Println("base type is dict?:", outputType == py.Dict)
		}

		log.Println("clearing thread state")
		py.PyThreadState_Clear(ts)

		py.PyThreadState_DeleteCurrent()
		log.Println("go routine ending")
		runtime.UnlockOSThread()
		signal.Store(false)
	}()

	working := true
	for working {
		// busy busy busy
		working = signal.Load()
	}
	log.Println("go routine looks fininshed")

	py.PyEval_RestoreThread(ts)
	log.Println("reloaded subthread state on main go routine")
	print_current_interpreter()

	py.PyThreadState_Clear(ts)
	log.Println("cleared subthread state on main go routine")

	py.PyInterpreterState_Clear(subint)
	log.Println("reset interpreter on main go routine")

	py.PyInterpreterState_Delete(subint)
	log.Println("deleted sub interpreter state on main go routine")

	py.PyEval_RestoreThread(mainTs)
	log.Println("restored main thread on main go routine")
	print_current_interpreter()

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
