package main

import (
	py "github.com/voutilad/gogopython"
	"log"
	"os"
	"runtime"
	"strings"
)

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

	script1 := `
import pandas as pd
import pickle

df = pd.DataFrame.from_dict({"names": ["maple", "moxie"], "age": [8, 3]})
print(f"script1: {df}")
buffer = pickle.dumps(df)
`
	script2 := `
import pandas as pd
import pickle

df = pickle.loads(buffer)
print(f"script2: {df}")
`
	code1 := py.Py_CompileString(script1, "script1", py.PyFileInput)
	code2 := py.Py_CompileString(script2, "script2", py.PyFileInput)

	m := py.PyImport_AddModule("__main__")
	log.Println("m:", py.BaseType(m).String())
	locals := py.PyModule_GetDict(m)

	result := py.PyEval_EvalCode(code1, locals, locals)
	if result == py.NullPyObjectPtr {
		py.PyErr_Print()
	}
	log.Println("result:", py.BaseType(result).String())

	keys := py.PyDict_Keys(locals)
	for i := int64(0); i < py.PyList_Size(keys); i++ {
		key := py.PyList_GetItem(keys, i)
		s, _ := py.UnicodeToString(key)
		log.Println(i, "key:", s)
	}

	result = py.PyEval_EvalCode(code2, locals, locals)
	if result == py.NullPyObjectPtr {
		py.PyErr_Print()
	}
	log.Println("result:", py.BaseType(result).String())

	py.Py_FinalizeEx()
}
