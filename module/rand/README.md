# Rand

[![Go Reference](https://img.shields.io/badge/godoc-reference-%23007d9c.svg)](https://point-c.github.io/caddy-randhandler)

The `rand` module is a custom HTTP handler for Caddy, designed to generate and serve random data in response to HTTP requests. This module can be particularly useful for testing, simulations, or any scenario where random data generation is required.

## Features

- Generates random data based on the provided seed.
- Allows specifying the size of the random data to be generated.
- Integrates seamlessly with Caddy's modular design.

### Installation
To install `rand`, you will need to build a custom Caddy binary that includes this module. This can be achieved using the `xcaddy` utility:

1. **Install xcaddy**:
   ```sh
   go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest
   ```

2. **Build Caddy with the random data generator**:
   ```sh
   xcaddy build --with github.com/point-c/caddy-randhandler
   ```

3. **Run Your Custom Caddy Build**:
   ```sh
   ./caddy run
   ```

## Configuration

### Caddyfile

You can configure the `rand` module in your Caddyfile using the `rand` directive. Here's an example configuration:

```Caddyfile
:80 {
    route {
        rand
    }
}
```

This configuration sets up a Caddy server listening on port 80, where the `rand` module handles all incoming requests.

### JSON Configuration

Alternatively, you can configure the module using Caddy's JSON configuration. Here's an example:

```json
{
   "apps": {
      "http": {
         "servers": {
            "srv0": {
               "listen": [":80"],
               "routes": [
                  {
                     "handle": [
                        {
                           "handler": "subroute",
                           "routes": [{"handle": [{"handler": "rand"}]}]
                        }
                     ]
                  }
               ]
            }
         }
      }
   }
}
```

This JSON configuration achieves the same setup as the Caddyfile example.

## Usage

Once configured, the module will respond to HTTP requests with random data. You can control the seed and size of the random data using HTTP headers:

- `Rand-Seed`: The seed for the random data generator. If not specified, the current Unix microsecond timestamp is used.
- `Rand-Size`: The size of the random data in bytes. If not specified or set to a negative value, the module will stream random data indefinitely.

### Example Request

Here's how you might use `curl` to make a request to a server using the `rand` module:

```bash
curl -H "Rand-Seed: 12345" -H "Rand-Size: 1024" http://localhost:80
```

This request will return 1024 bytes of random data generated using the seed 12345.

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