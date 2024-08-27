// Package gogopython wraps a Python dynamic library with Go native functions, making
// embedding of Python in a native Go app relatively easy. (For some definition
// of easy)
//
// It wraps common Python C API functions needed to manage interpreters and
// create/modify Python objects. Not all functions are wrapped. Not all
// features are wrappable in pure Go as in some cases they're C macros.
//
// Since the #1 headache in using an embedding Python interpreter is finding
// the necessary library, Python home, and paths for packages, gogopython
// provides a few helper functions to try figuring this out for the user.
//
// Note: Currently only Python 3.12 is supported.
package gogopython

import (
	"bufio"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"unsafe"

	"github.com/ebitengine/purego"
)

// LoadLibrary attempts to load and wrap the appropriate dynamic library for
// embedding Python, given a particular Python binary.
//
// Currently, assumes the provided binary is for Python 3.12.
func LoadLibrary(exe string) error {
	var dll string
	var err error
	os := runtime.GOOS

	// TODO: detect Python version.

	switch os {
	case "darwin":
		dll = "libpython3.12.dylib"
	case "linux":
		dll = "libpython3.12.so.1.0" // todo: maybe find this dynamically?
	default:
		return fmt.Errorf("unsupported os: %s", os)
	}

	base, err := findLibraryBaseUsingDistutils(exe)
	if err != nil {
		// Use a fallback method that's OS dependent.
		switch os {
		case "darwin":
			base, err = findLibraryBaseFallbackToOtool(exe)
		case "linux":
			// todo: figure out a heuristic for guessing on linux
		default:
			// nothing
		}
		if err != nil {
			return errors.New("failed to find library base path")
		}
	}
	library := base + "/" + dll

	lib, err := purego.Dlopen(library, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return err
	}

	registerFuncs(lib)

	return nil
}

// Find the location of the Python dynamic library using Python's setuptools
// package.
//
// Returns a pointer to the base directory as a string or an error on failure.
func findLibraryBaseUsingDistutils(exe string) (string, error) {
	// TODO: context with deadline
	// One approach is, assuming setuptools is available, is to use distutils.
	cmd := exec.Command(exe, "-c", "from distutils import sysconfig; print(sysconfig.get_config_var('LIBDIR'))")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	if err = cmd.Start(); err != nil {
		return "", err
	}
	base, err := bufio.NewReader(stdout).ReadString(byte('\n'))
	if err != nil {
		return "", err
	}
	if err = cmd.Wait(); err != nil {
		return "", err
	}

	if base != "" {
		lib := strings.TrimRight(base, " \n")
		return lib, nil
	}
	return "", errors.New("failed to find library base")
}

// Try using otool (on macOS) and see if we can find the dynamic library path.
// This is "best effort"...and "best" is a bit of a stretch.
//
// Returns the base path as a pointer to a string or an error on failure.
func findLibraryBaseFallbackToOtool(exe string) (string, error) {
	lib := ""

	// First resolve the location if we're given just "python3"
	cmd := exec.Command("command", "-v", exe)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	if err = cmd.Start(); err != nil {
		return "", err
	}
	path, err := bufio.NewReader(stdout).ReadString(byte('\n'))
	if err != nil {
		return "", err
	}
	if err = cmd.Wait(); err != nil {
		return "", err
	}

	cmd = exec.Command("otool", "-L", strings.TrimRight(path, "\n"))
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	if err = cmd.Start(); err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(bufio.NewReader(stdout))
	for scanner.Scan() {
		// We should have a line pointing to a Python.framework location.
		text := scanner.Text()
		if strings.Contains(text, "Python.framework") {
			// Should look something like:
			//    /something/Python.framework/Versions/3.12/Python (compatibility ...)
			parts := strings.SplitAfterN(strings.TrimLeft(text, " \t"), " ", 2)
			if len(parts) < 2 {
				return "", errors.New("could not parse otool output")
			}
			lib = strings.TrimRight(parts[0], " ")
			lib = strings.TrimSuffix(lib, "Python")

			// At this point, we should have the base directory for the lib dir.
			lib = lib + "/lib"
		}
	}
	err = cmd.Wait()
	if err != nil {
		return "", err
	}
	if lib != "" {
		return lib, nil
	}
	return "", errors.New("failed to find library base")
}

// Python snippet for discovering home and path.
const helper string = "import sys; print(sys.prefix); [print(p) for p in sys.path if len(p) > 0]"

// FindPythonHomeAndPaths uses the provided Python executable to discovery the
// Python Home and path settings.
func FindPythonHomeAndPaths(exe string) (string, []string, error) {
	home := ""

	// Start with empty string, which is for the current directory.
	// Without this, we can't load adjacent py files.
	paths := []string{""}

	cmd := exec.Command(exe, "-c", helper)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return home, paths, err
	}
	if err = cmd.Start(); err != nil {
		return home, paths, err
	}
	scanner := bufio.NewScanner(bufio.NewReader(stdout))

	// First line is our home, subsequent are the path
	first := true
	for scanner.Scan() {
		text := scanner.Text()
		if first {
			home = text
			first = false
		} else {
			paths = append(paths, text)
		}
	}
	if err = cmd.Wait(); err != nil {
		return home, paths, err
	}
	return home, paths, nil
}

// WCharToString copies out a Python *wchar_t to a Go string.
func WCharToString(text WCharPtr) (string, error) {
	p := Py_EncodeLocale(text, nil)
	if p == nil {
		return "", errors.New("failed to encode text")
	}

	// We don't own p.
	defer PyMem_Free(p)

	// We need to find the length of the string. It should be NULL-terminated,
	// but cap the possible string length to something arbitrary: 10 MiB.
	ptr := unsafe.Pointer(p)
	for _len := 0; _len < (10 << 20); _len++ {
		if *(*uint8)(ptr) == 0 {
			// Found our NULL.
			s := unsafe.String(p, _len)
			// We don't own the backing bytes, the interpreter does.
			// As such, we need to make a copy.
			return strings.Clone(s), nil
		}
		ptr = unsafe.Add(ptr, 1)
	}

	return "", errors.New("text too long")
}

// UnicodeToString converts a Python Unicode object (i.e. a Python string)
// to a Go string.
//
// Note: this currently involves a lot of data copying out from the Python
// interpreter. It's far from optimized.
func UnicodeToString(unicode PyObjectPtr) (string, error) {
	buf := PyUnicode_AsEncodedString(unicode, "utf-8", Strict)
	if buf == NullPyObjectPtr {
		PyErr_Print()
		return "", errors.New("failed to encode python object")
	}
	defer Py_DecRef(buf)

	sz := PyBytes_Size(buf)
	p := PyBytes_AsString(buf)
	str := unsafe.String(p, sz)

	// We need to make a copy as the backing bytes are not owned by str.
	return strings.Clone(str), nil
}

// BaseType identifies the Python base type from a Python *PyObject.
//
// This uses a heuristic based on inspecting some internal object flags as
// most of the Python C API for type inspection is written in macros.
//
// See https://docs.python.org/3/c-api/type.html#c.PyType_GetFlags if
// curious about the flags.
func BaseType(obj PyObjectPtr) Type {
	if obj == NullPyObjectPtr {
		return Unknown
	}

	tp := PyObject_Type(obj)
	if tp == NullPyTypeObjectPtr {
		return Unknown
	}

	flags := PyType_GetFlags(tp)
	if (flags & typeMask) != 0 {
		if (flags & (uint64)(Long)) != 0 {
			return Long
		} else if (flags & (uint64)(List)) != 0 {
			return List
		} else if (flags & (uint64)(Tuple)) != 0 {
			return Tuple
		} else if (flags & (uint64)(Bytes)) != 0 {
			return Bytes
		} else if (flags & (uint64)(String)) != 0 {
			return String
		} else if (flags & (uint64)(Dict)) != 0 {
			return Dict
		}
	} else {
		if (flags & fnMask) == fnMask {
			return Function
		}
		if (flags & genMask) == genMask {
			return Generator
		}
		if (flags & setMask) == setMask {
			return Set
		}
		if (flags & moduleMask) == moduleMask {
			return Module
		}
		if (flags & floatMask) == floatMask {
			return Float
		}
		if (flags & noneMask) == noneMask {
			return None
		}
	}
	return Unknown
}

// NewFunction creates a new Python function object, with the given name, that
// calls the provided Go func.
func NewFunction(name string, self PyObjectPtr, fn func(self, tuple PyObjectPtr) PyObjectPtr) PyObjectPtr {
	def := PyMethodDef{
		Name:   unsafe.StringData(name),
		Flags:  MethodVarArgs,
		Method: purego.NewCallback(fn),
	}
	return PyCFunction_NewEx(&def, self, NullPyObjectPtr)
}
