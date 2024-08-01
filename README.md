# gogopython

Python, but in Go.

```bash
CGO_ENABLED=0 go build
```

## Using

`gogopython` merely exposes the Python C-Api, so you'll need to use it just 
like you would without Go. For an example of spinning up a sub-interpreter,
see the example program in `cmd/main.go`.


## Quick command line test

To run the test program, figure out your python home and paths:

```bash
python3 -c 'import sys; print(f"home={sys.prefix}\npaths={sys.path}")'
```

Run the test program:

```
go run cmd/main.go [your home] [each path entry]
```

> note: home and path entries _are_ optional, but if you're trying to use a
> virtualenv you'll need to set them

## Known Issues

- Panics sometimes just happen during library loading, not sure why!

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
