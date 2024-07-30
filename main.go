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
		library = "libpython3.so"
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
	pyPreConfig_InitIsolatedConfig(&preConfig)
	status := py_PreInitialize(&preConfig)
	if status.Type != 0 {
		log.Fatalln("failed to preinitialize python:", PyBytesToString(status.ErrMsg))
	}
	log.Println("preinitialization complete")

	/* Configure our Paths */
	config := PyConfig{}
	pyConfig_InitPythonConfig(&config)
	defer pyConfig_Clear(&config)
	config.PathConfigWarnings = 0
	config.TraceMalloc = 0
	config.ParseArgv = 0
	config.SafePath = 1
	config.UserSiteDirectory = 0
	config.InstallSignalHandlers = 0

	home := "/Users/dv/src/gogopython/venv"
	status = pyConfig_SetBytesString(&config, &config.Home, home)
	if status.Type != 0 {
		log.Fatalln("failed to set home:", PyBytesToString(status.ErrMsg))
	}
	log.Println("set home:", home)

	path := strings.Join([]string{
		"/opt/homebrew/Cellar/python@3.12/3.12.4/Frameworks/Python.framework/Versions/3.12/lib/python3.12",
		"/opt/homebrew/Cellar/python@3.12/3.12.4/Frameworks/Python.framework/Versions/3.12/lib/python3.12/lib-dynload",
		"/Users/dv/src/gogopython/venv/lib/python3.12/site-packages",
	}, ":")
	status = pyConfig_SetBytesString(&config, &config.PythonPathEnv, path)
	if status.Type != 0 {
		log.Fatalln("failed to set path:", PyBytesToString(status.ErrMsg))
	}
	log.Println("set path:", path)

	status = py_InitializeFromConfig(&config)
	if status.Type != 0 {
		log.Fatalln("failed to initialize Python:", PyBytesToString(status.ErrMsg))
	}
	log.Println("initialized")
	defer py_FinalizeEx()

	globals := PyDict_New()
	defer Py_DecRef(globals)
	locals := PyDict_New()
	defer Py_DecRef(locals)

	tuple := PyTuple_New(2)
	defer Py_DecRef(tuple)
	PyTuple_SetItem(tuple, 0, PyLong_FromLong(69))
	PyTuple_SetItem(tuple, 1, PyLong_FromLong(420))

	if PyDict_SetItemString(globals, "junk", tuple) != 0 {
		log.Fatalln("failed to set globals")
	}

	program := `
import numpy as np
import torch
data = [[1, 2], [3, 4]]
print(torch.tensor(data))
`
	if PyRun_String(program, PyFileInput, globals, locals) == nil {
		PyErr_Print()
		PyErr_Clear()
		log.Fatalln("crap1")
	}

	if PyRun_String("print(junk)", PySingleInput, globals, locals) == nil {
		PyErr_Print()
		PyErr_Clear()
		log.Fatalln("crap2")
	}
}
