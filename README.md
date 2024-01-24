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

```Caddyfile
{
   default_bind stub://0.0.0.0
   point-c {
      <submodule name> <submodule config>
      system sys 0.0.0.0
   }
   point-c netops {
      <submodule name> <submodule config>
      forward sys:sys {
         <submodule name> <submodule config>
         tcp 80:8080 1024
      }
   }
   servers :80 {
      listener_wrappers {
         merge {
            <submodule name> <submodule config>
            point-c sys 8080
         }
      }
   }
}

:80 {
   log
}
```

### JSON Configuration

```json5
{
  "apps": {
    "http": {
      "servers": {
        "srv0": {
          "listen": ["stub://0.0.0.0:80"]
        }
      }
    },
    "point-c": {
      "networks": [
        {
          "addr": "0.0.0.0",
          "hostname": "sys",
          "type": "system"
        }
      ],
      "net-ops": [
        {
          "forwards": [
            {
              "buf": null,
              "forward": "tcp",
              "ports": "80:8080"
            }
          ],
          "hosts": "sys:sys",
          "op": "forward"
        }
      ]
    }
  }
}
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