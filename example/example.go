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
	exe := "python3"
	if len(os.Args) > 1 {
		exe = os.Args[1]
	}
	log.Println("using python exe:", exe)

	// Discovery our Python Home and Path settings.
	home, paths, err := py.FindPythonHomeAndPaths(exe)
	if err != nil {
		log.Fatalln(err)
	}
	path := strings.Join(paths, ":")

	// Initialize the bindings.
	err = py.Load_library(exe)
	if err != nil {
		log.Fatalln(err)
	}

	// Pre-initialize Python.
	preConfig := py.PyPreConfig{}
	py.PyPreConfig_InitIsolatedConfig(&preConfig)
	preConfig.Allocator = py.PyMemAllocator_Malloc
	status := py.Py_PreInitialize(&preConfig)
	if status.Type != 0 {
		log.Fatalln("failed to preinitialize python:", py.PyBytesToString(status.ErrMsg))
	}
	log.Println("preinitialization complete")

	// Configure the main interpreter.
	config := py.PyConfig_3_12{}
	py.PyConfig_InitPythonConfig(&config)
	defer py.PyConfig_Clear(&config)

	status = py.PyConfig_SetBytesString(&config, &config.Home, home)
	if status.Type != 0 {
		log.Fatalln("failed to set home:", py.PyBytesToString(status.ErrMsg))
	}
	log.Println("set python home:", home)

	status = py.PyConfig_SetBytesString(&config, &config.PythonPathEnv, path)
	if status.Type != 0 {
		log.Fatalln("failed to set path:", py.PyBytesToString(status.ErrMsg))
	}
	log.Println("set python path:", path)

	// Initialize our main interpreter in our main Go routine.
	status = py.Py_InitializeFromConfig(&config)
	if status.Type != 0 {
		log.Fatalln("failed to initialize Python:", py.PyBytesToString(status.ErrMsg))
	}
	log.Println("initialized")
	mainTs := py.PyThreadState_Get()

	// Add a GIL reference, print, and drop our thread state.
	print_current_interpreter()
	py.PyThreadState_Swap(py.NullThreadState)

	// Create a Subinterpreter in our main go routine so it's tied to the current thread.
	var subThreadPtr py.PyThreadStatePtr
	interpreterConfig := py.PyInterpreterConfig{}
	interpreterConfig.Gil = py.OwnGil
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

		py.PyRun_String(`
global content
def content():
	global __content__
	return __content__
		`, py.PyFileInput, globals, locals)

		bytes := py.PyBytes_FromString("hello world")
		py.PyDict_SetItemString(globals, "__content__", bytes)
		py.Py_DecRef(bytes)

		program := `
x = {"name": "Dave"}
print(content())
root = "hello world"
`
		output := py.PyRun_String(program, py.PyFileInput, globals, locals)
		if output == py.NullPyObjectPtr {
			py.PyErr_Print()
			log.Fatalln("exception in python script")
		} else {
			// We should be able to see our mutated local variable state.
			x := py.PyDict_GetItemString(locals, "x")
			outputType := py.Py_BaseType(x)
			log.Println("Is 'x' a dict?", outputType == py.Dict)

			// Extract our message string and convert to Go string.
			root := py.PyDict_GetItemString(locals, "root")
			if root == py.NullPyObjectPtr {
				log.Fatalln("no root object found")
			}
			msg := py.PyUnicode_AsWideCharString(root, nil)
			defer py.PyMem_Free(msg)
			s := py.Py_EncodeLocale(msg, nil)
			log.Println("root:", s)
		}

		py.Py_DecRef(globals)
		py.Py_DecRef(locals)

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
