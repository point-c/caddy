# point-c

[![Go Reference](https://img.shields.io/badge/godoc-reference-%23007d9c.svg)](https://point-c.github.io/caddy)



## Features

### Installation
To install `point-c`, you will need to build a custom Caddy binary that includes this module. This can be achieved using the `xcaddy` utility:

1. **Install xcaddy**:
   ```sh
   go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest
   ```

2. **Build Caddy with the random data generator**:
   ```sh
   xcaddy build --with github.com/point-c/caddy
   ```

3. **Run Your Custom Caddy Build**:
   ```sh
   ./caddy run
   ```

## Configuration

### Caddyfile

### JSON Configuration

## Usage

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