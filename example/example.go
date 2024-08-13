package main

import (
	_ "embed"
	py "github.com/voutilad/gogopython"
	"log"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
)

//go:embed script.py
var script string

var helperCode string = `
global content
def content():
	global __content__
	return __content__

go_func(0, 1, 2)
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

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Pre-initialize Python.
	preConfig := py.PyPreConfig{}
	py.PyPreConfig_InitIsolatedConfig(&preConfig)
	preConfig.Allocator = py.PyMemAllocator_Malloc
	status := py.Py_PreInitialize(&preConfig)
	if status.Type != 0 {
		msg, _ := py.WCharToString(status.ErrMsg)
		log.Fatalln("Failed to preinitialize python:", msg)
	}
	log.Println("Pre-initialization complete.")

	// Configure the main interpreter.
	config := py.PyConfig_3_12{}
	py.PyConfig_InitPythonConfig(&config)
	defer py.PyConfig_Clear(&config)

	status = py.PyConfig_SetBytesString(&config, &config.Home, home)
	if status.Type != 0 {
		msg, _ := py.WCharToString(status.ErrMsg)
		log.Fatalln("Failed to set home:", msg)
	}
	log.Println("Set python home:", home)

	status = py.PyConfig_SetBytesString(&config, &config.PythonPathEnv, path)
	if status.Type != 0 {
		msg, _ := py.WCharToString(status.ErrMsg)
		log.Fatalln("failed to set path:", msg)
	}
	log.Println("Set python path:", path)

	// Initialize our main interpreter in our main Go routine.
	status = py.Py_InitializeFromConfig(&config)
	if status.Type != 0 {
		msg, _ := py.WCharToString(status.ErrMsg)
		log.Fatalln("Failed to initialize Python:", msg)
	}
	mainTs := py.PyThreadState_Get()

	// Unload our main interpreter state from this thread.
	py.PyThreadState_Swap(py.NullThreadState)

	// Create a sub-interpreter in our main go routine so it's tied to the current thread.
	var subThreadPtr py.PyThreadStatePtr
	interpreterConfig := py.PyInterpreterConfig{}
	interpreterConfig.Gil = py.OwnGil
	interpreterConfig.CheckMultiInterpExtensions = 1
	status = py.Py_NewInterpreterFromConfig(&subThreadPtr, &interpreterConfig)
	if status.Type != 0 {
		msg, _ := py.WCharToString(status.ErrMsg)
		log.Fatalln("Failed to create sub-interpreter:", msg)
	}

	// Get a pointer to our interpreter state, which should not be thread local afaik.
	subIntState := py.PyInterpreterState_Get()

	// Remove our thread state and release the GIL
	ts := py.PyEval_SaveThread()

	// Launch a go routine and busy wait for it to finish. Use an atomic and
	// not a channel so we can do the busy wait in a for-loop.
	var signal atomic.Bool
	signal.Store(true)

	go func() {
		runtime.LockOSThread()
		newTs := py.PyThreadState_New(subIntState)
		py.PyEval_RestoreThread(newTs)

		// Demonstrate running a simple script without global/local state.
		if py.PyRun_SimpleString(script) != 0 {
			py.PyErr_Print()
			log.Fatalln("failed to run script")
		}

		// Create mappings (dicts) for global and local state.
		globals := py.PyDict_New()
		locals := py.PyDict_New()

		// Create a callback into Go.
		pyFn := py.NewFunction("go_func", py.NullPyObjectPtr,
			func(self, args py.PyObjectPtr) py.PyObjectPtr {
				log.Printf("Go func called: self=0x%x, args=0x%x\n", self, args)

				argsType := py.BaseType(args)
				if argsType != py.Tuple {
					log.Fatalln("Expected a Python Tuple, got", argsType.String())
				}
				sz := py.PyTuple_Size(args)
				log.Println("positional args of", sz, "items")
				for i := int64(0); i < sz; i++ {
					obj := py.PyTuple_GetItem(args, i)
					t := py.BaseType(obj)
					if t == py.Long {
						val := py.PyLong_AsLong(obj)
						log.Printf(" item[%d] = %d\n", i, val)
					} else {
						log.Fatalln("Expected a Long in the Tuple.")
					}
				}
				return py.PyLong_FromLong(0) // need something non-null
			})
		py.PyDict_SetItemString(globals, "go_func", pyFn)

		// Install some helper code in our global state.
		result := py.PyRun_String(helperCode, py.PyFileInput, globals, locals)
		if result == py.NullPyObjectPtr {
			py.PyErr_Print()
			log.Fatalln("PyRun_String failed.")
		}
		py.Py_DecRef(result)

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
			py.Py_DecRef(output)

			// Extract our "root" local defined in the code.
			root := py.PyDict_GetItemString(locals, "root")
			if root == py.NullPyObjectPtr {
				log.Fatalln("no root object found")
			}

			// Test out type detection.
			if py.BaseType(root) != py.Dict {
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
				if py.BaseType(obj) != v {
					log.Fatalf("Value for key %s is not %s\n", k, v.String())
				}
				log.Printf("Detected root['%s'] as a %s\n", k, v.String())
			}

			// Test string copy-out.
			pyString := py.PyDict_GetItemString(root, "string")
			s, err := py.UnicodeToString(pyString)
			if err != nil {
				log.Fatalln("Failed to extract string:", err)
			}
			log.Println("Extracted string:", s)
		}

		// Drop ref counts.
		py.Py_DecRef(globals)
		py.Py_DecRef(locals)

		// Clean up our thread.
		py.PyThreadState_Clear(newTs)
		py.PyThreadState_DeleteCurrent()

		runtime.UnlockOSThread()
		signal.Store(false)
	}()

	working := true
	for working {
		// busy busy busy
		working = signal.Load()
	}
	log.Println("Go routine looks finished.")

	py.PyEval_RestoreThread(ts)
	log.Println("Reloaded original sub-interpreter thread state on main go routine.")

	py.PyThreadState_Clear(ts)
	log.Println("Cleared original sub-interpreter state on main go routine.")

	py.PyInterpreterState_Clear(subIntState)
	log.Println("Reset interpreter on main go routine.")

	py.PyInterpreterState_Delete(subIntState)
	log.Println("Deleted sub interpreter state on main go routine.")

	py.PyEval_RestoreThread(mainTs)
	log.Println("Restored main thread on main go routine.")

	py.Py_FinalizeEx()
}
