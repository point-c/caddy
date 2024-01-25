# point-c

[![Go Reference](https://img.shields.io/badge/godoc-reference-%23007d9c.svg)](https://point-c.github.io/caddy)

point-c is a collection of Caddy modules for handling traffic between host systems and WireGuard tunnels. 

## Modules

- `merge-listener-wrapper`: `caddy.ListenerWrapper` that wraps multiple TCP listeners.
- `point-c`: Handles registered networks and network operations.
- `sysnet`: A `point-c` network for the host system.
- `wg`: A `point-c` network for WireGuard tunnels.
- `listener`: Allows registering a `point-c` network with `merge-listener-wrapper`.
- `forward`: Network operation that manages modules that move traffic from a source to a destination.
- `forward-tcp`: A forward submodule that forwards TCP traffic.
- `stub-listener`: Prevent caddy from listening on the host system.
- `rand`: A `caddyhttp.MiddlewareHandler` that returns random data.

### Installation
To install a module from `point-c`, you will need to build a custom Caddy binary that includes that module. This can be achieved using the `xcaddy` utility:

1. **Install xcaddy**:
   ```sh
   go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest
   ```

2. **Build Caddy with all modules**:
   ```sh
   xcaddy build \
      --with github.com/point-c/caddy/module/forward@latest \
      --with github.com/point-c/caddy/module/forward-tcp@latest \
      --with github.com/point-c/caddy/module/listener@latest \
      --with github.com/point-c/caddy/module/merge-listener-wrapper@latest \
      --with github.com/point-c/caddy/module/point-c@latest \
      --with github.com/point-c/caddy/module/rand@latest \
      --with github.com/point-c/caddy/module/stub-listener@latest \
      --with github.com/point-c/caddy/module/sysnet@latest \
      --with github.com/point-c/caddy/module/wg@latest
   ```

3. **Run Your Custom Caddy Build**:
   ```sh
   ./caddy run
   ```

## Configuration

### Quickstart

#### Server

```Caddyfile
{
    # Don't bind to the host system
    default_bind stub://0.0.0.0
    point-c {
        # Default network
        system sys 0.0.0.0
        # WireGuard server config
        wgserver server {
            ip 192.168.45.1
            port 51820
            private {{ txt .Private }}
            peer client {
                ip 192.168.45.2
                public {{ txt .Public }}
                shared {{ txt .Shared }}
            }
        }
    }
    # Forward traffic
    point-c netops {
        forward sys:client{
            tcp 80:80
        }
    }
}

:80 {
   # Run HTTP server to prevent HTTPS from being used
}
```

#### Client

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