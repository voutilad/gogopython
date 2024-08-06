package main

import (
	"log"
	"os"
	"runtime"
	"strings"
	"sync/atomic"

	py "github.com/voutilad/gogopython"
)

var script string = `
import time
print('python is sleeping')
time.sleep(0.2)
print('python is awake!')
`

var helperCode string = `
global content
def content():
	global __content__
	return __content__
`

var program string = `
root = {
  "long": 123,
  "list": [],
  "tuple": (),
  "bytes": content(),
  "string": "hey",
  "float": 3.14,
  "set": set(),
  "none": None,
}
`

func main() {
	exe := "python3"
	if len(os.Args) > 1 {
		exe = os.Args[1]
	}
	log.Println("Using python exe:", exe)

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
		log.Fatalln("Failed to preinitialize python:", py.PyBytesToString(status.ErrMsg))
	}
	log.Println("Pre-initialization complete.")

	// Configure the main interpreter.
	config := py.PyConfig_3_12{}
	py.PyConfig_InitPythonConfig(&config)
	defer py.PyConfig_Clear(&config)

	status = py.PyConfig_SetBytesString(&config, &config.Home, home)
	if status.Type != 0 {
		log.Fatalln("Failed to set home:", py.PyBytesToString(status.ErrMsg))
	}
	log.Println("Set python home:", home)

	status = py.PyConfig_SetBytesString(&config, &config.PythonPathEnv, path)
	if status.Type != 0 {
		log.Fatalln("failed to set path:", py.PyBytesToString(status.ErrMsg))
	}
	log.Println("Set python path:", path)

	// Initialize our main interpreter in our main Go routine.
	status = py.Py_InitializeFromConfig(&config)
	if status.Type != 0 {
		log.Fatalln("Failed to initialize Python:", py.PyBytesToString(status.ErrMsg))
	}
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
		log.Fatalln("Failed to create sub-interpreter:", py.PyBytesToString(status.ErrMsg))
	}
	print_current_interpreter()

	// Get a pointer to our interpreter state, which should not be thread local afaik.
	subint := py.PyInterpreterState_Get()

	// Remove our threadstate and release the GIL
	ts := py.PyEval_SaveThread()

	// Launch a go routine and busy wait for it to finish. Use an atomic and
	// not a channel so we can do the busy wait in a for-loop.
	var signal atomic.Bool
	signal.Store(true)

	go func() {
		log.Println("Started go routine.")

		runtime.LockOSThread()
		ts := py.PyThreadState_New(subint)
		py.PyEval_RestoreThread(ts)

		print_current_interpreter()

		// Demonstrate running a simple script without global/local state.
		if py.PyRun_SimpleString(script) != 0 {
			py.PyErr_Print()
			log.Fatalln("failed to run script")
		}

		// Create mappings (dicts) for global and local state.
		globals := py.PyDict_New()
		locals := py.PyDict_New()

		// Install some helper code in our global state.
		py.PyRun_String(helperCode, py.PyFileInput, globals, locals)

		// Populate some global state.
		bytes := py.PyBytes_FromString("hello world")
		py.PyDict_SetItemString(globals, "__content__", bytes)
		py.Py_DecRef(bytes)

		// Compile our program.
		code := py.Py_CompileString(program, "program.py", py.PyFileInput)
		if code == py.NullPyCodeObjectPtr {
			py.PyErr_Print()
			log.Fatalln("failed to compile python program")
		}

		// Run our program.
		output := py.PyEval_EvalCode(code, globals, locals)
		if output == py.NullPyObjectPtr {
			py.PyErr_Print()
			log.Fatalln("exception in python script")
		} else {
			// Extract our "root" local defined in the code.
			root := py.PyDict_GetItemString(locals, "root")
			if root == py.NullPyObjectPtr {
				log.Fatalln("no root object found")
			}

			// Test out type detection.
			if py.Py_BaseType(root) != py.Dict {
				log.Fatalln("root should be a dict")
			}
			m := map[string]py.Type{
				"long":   py.Long,
				"list":   py.List,
				"tuple":  py.Tuple,
				"bytes":  py.Bytes,
				"string": py.String,
				"float":  py.Float,
				"set":    py.Set,
				"none":   py.None,
			}
			for k, v := range m {
				obj := py.PyDict_GetItemString(root, k)
				if obj == py.NullPyObjectPtr {
					log.Fatalf("Failed to find key %s in root dict\n", k)
				}
				if py.Py_BaseType(obj) != v {
					log.Fatalf("Value for key %s is not %s\n", k, v.String())
				}
				log.Printf("Detected root['%s'] as a %s\n", k, v.String())
			}
		}

		// Drop ref counts.
		py.Py_DecRef(globals)
		py.Py_DecRef(locals)

		// Clean up our thread.
		py.PyThreadState_Clear(ts)
		py.PyThreadState_DeleteCurrent()

		runtime.UnlockOSThread()
		signal.Store(false)
	}()

	working := true
	for working {
		// busy busy busy
		working = signal.Load()
	}
	log.Println("Go routine looks fininshed.")

	py.PyEval_RestoreThread(ts)
	log.Println("Reloaded subthread state on main go routine.")
	print_current_interpreter()

	py.PyThreadState_Clear(ts)
	log.Println("Cleared subthread state on main go routine.")

	py.PyInterpreterState_Clear(subint)
	log.Println("Reset interpreter on main go routine.")

	py.PyInterpreterState_Delete(subint)
	log.Println("Deleted sub interpreter state on main go routine.")

	py.PyEval_RestoreThread(mainTs)
	log.Println("Restored main thread on main go routine.")
	print_current_interpreter()

	py.Py_FinalizeEx()
}

func print_current_interpreter() {
	/// See https://github.com/python/cpython/blob/2b163aa9e796b312bb0549d49145d26e4904768e/Programs/_testembed.c#L100-L115
	ts := py.PyThreadState_Get()
	me := py.PyThreadState_GetInterpreter(ts)
	id := py.PyInterpreterState_GetID(me)
	log.Printf("Active interpreter: 0x%x, thread state 0x%x, id %d\n", me, ts, id)
	// py.PyRun_SimpleString("import sys; print('id(modules) =', id(sys.modules)); sys.stdout.flush()")
}
