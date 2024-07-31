# gogopython

Python, but in Go.

```bash
CGO_ENABLED=0 go build
```

To run the test program, figure out your python home and paths:

```bash
python3 -c 'import sys; print(f"home={sys.prefix}\npaths={sys.path}")'
```

Run the test program:

```
go run cmd/main.go [your home] [each path entry]
```

