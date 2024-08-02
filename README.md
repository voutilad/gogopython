# gogopython

Python, but in Go.

```bash
CGO_ENABLED=0 go build
```

> Heads up: this currently requires Python 3.12. No if's, and's or but's.

## Using

`gogopython` merely exposes the Python C-Api, so you'll need to use it just 
like you would without Go. For an example of spinning up a sub-interpreter,
see the example program in `cmd/main.go`.

## Library Detection

The biggest pain is finding the Python dynamic library. On some Linux systems,
it must be installed separately (Debian-based distros for sure).

```bash
sudo apt install libpython3.12
```

`gogopython` will try to find the library using `distutils` via the given
Python binary. This may require installing `setuptools` via `pip`.

On macOS, it will use `otool` to 

## Quick command line test

Simply point the test program at your Python3 binary.

```
# Create and activate virtual environment.
python3 -m venv venv
. venv/bin/activate

# Install setuptools. This is used for library discovery.
pip install setuptools

# We no longer need the virtual environment enabled.
deactivate

# Run the test app.
go run example/example.go ./venv/bin/python3
```

> Note: if on Linux, make sure you have `setuptools` installed.

## Known Issues

- Requires Python 3.12 as it uses sub-interpreters. Sorry, not sorry.

- Linux requires a shim using the `ffi` Go module that uses `purego` 
  to leverage `libffi`, so on Linux `libffi` must be available. This
  is all because some super old Python C API functions decide to
  return a struct on the stack and `purego` only supports that on
  macOS currently.

- The Python api is super thread local storage oriented. Using it with
  Go is a small nightmare. Gratuitous use of `runtime.LockOSThread()`
  is required.

- The helper function for finding the Python dynamic library won't
  work with Python installed via XCode as it's a funky dual-arch
  binary with some dynamic library funkiness and `otool` can't
  resolve the actual Python dylib location (if there even is one!)
