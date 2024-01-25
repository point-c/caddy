# channel-listener

[![Go Reference](https://img.shields.io/badge/godoc-reference-%23007d9c.svg)](https://point-c.github.io/channel-listener)

channel-listener provides a simple way to create a manually controlled `net.Listener`.

## Installation

To use channel-listener in your Go project, install it using `go get`:

```bash
go get github.com/point-c/channel-listener
```

## Usage

```go
c := make(chan net.Conn)
// The provided address will be passed through to `ln.Addr()`
ln := channel_listener.New(c, &net.TCPAddr{})
defer ln.Close() // Closing `c` will also close the listener
for {
    // Any `net.Conn` send through `c` will be accepted here
	ln, err := ln.Accept()
	if err != nil {
		log.Fatal(err)
    }
	// Use ln...
}
```

### Closing With An Error

```go
ln.CloseWithErr(errors.New("error"))
// Will only return the above error
_, err := ln.Accept()
// Will only return the above error
err := ln.Close()
// This error will be ignored and the first error will be returned
err := ln.CloseWithErr(errors.New("another error"))
```

## Testing

The package includes tests that demonstrate its functionality. Use Go's testing tools to run the tests:

```bash
go test
```

## Godocs

To regenerate godocs:

```bash
go generate -tags docs ./...
```