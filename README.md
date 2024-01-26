# point-c

[![Go Reference](https://img.shields.io/badge/godoc-reference-%23007d9c.svg)](https://point-c.github.io/caddy)

point-c is a collection of Caddy modules for handling traffic between host systems and WireGuard tunnels.

### Installation
To install the modules from `point-c`, you will need to build a custom Caddy binary that includes that module. This can be achieved using the `xcaddy` utility:

1. **Install xcaddy**:
   ```sh
   go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest
   ```

2. **Build Caddy with all modules**:
   ```sh
   xcaddy build \
      --with github.com/point-c/caddy/module@latest
   ```

3. **Run Your Custom Caddy Build**:
   ```sh
   ./caddy run
   ```

## Quickstart

### Caddy as Server Configuration

```Caddyfile
{
    # (Optional) Don't bind to the host system
    default_bind stub://0.0.0.0
    point-c {
        # WireGuard server config
        wgserver server {
            ip 192.168.45.1
            port 51820
            private 2Jgm2q3tFu21cO1IMyhjENqp7t5qep0++novkdKHe0k=
            # Add peer blocks for each client
            peer client-1 {
                ip 192.168.45.2
                public Tdbxgh9AHWXodT60AiwCUPDTEITyVD+ecMhp2TDY1xw=
                shared Z9Ad3ZhTQbIUCLEKATYXS1m380vYrYFhGA75tspxsOU=
            }
        }
    }
    servers :443 {
        listener_wrappers {
            merge {
                point-c client-1 443
                # Add `point-c` for each client
            }
        }
    }
}

# Rest of Caddy config
```

### Remote Listen Configuration

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
            private 2Jgm2q3tFu21cO1IMyhjENqp7t5qep0++novkdKHe0k=
            peer client {
                ip 192.168.45.2
                public Tdbxgh9AHWXodT60AiwCUPDTEITyVD+ecMhp2TDY1xw=
                shared Z9Ad3ZhTQbIUCLEKATYXS1m380vYrYFhGA75tspxsOU=
            }
        }
    }
    # Forward traffic
    point-c netops {
        forward sys:client{
            tcp 443:443
        }
    }
}

:80 {
   # Run HTTP server to prevent HTTPS from being used
}
```

#### Client

```Caddyfile
{
    # (Optional) Don't bind to the host system
    default_bind stub://0.0.0.0
    point-c {
        wgclient client {
            ip 192.168.45.2
            endpoint 127.0.0.1:51820
            private UCoEdsc8Mw7ZY81jSAHOGIw23QxqxfN8SQ8YktOrw0I=
            public 5GIGlLmvYnTyoQ59QIUYEo2FFUgubTibAO2qFI859hY=
            shared Z9Ad3ZhTQbIUCLEKATYXS1m380vYrYFhGA75tspxsOU=
        }
    }
    servers :443 {
        listener_wrappers {
            merge {
                point-c client 443
            }
        }
    }
}

# Rest of Caddy config
```

## Modules

### `merge-listener-wrapper`

`caddy.ListenerWrapper` that wraps multiple TCP listeners.

```Caddyfile
{
    servers :443 {
        listener_wrappers {
            merge {
               # Listener definitions go here 
            }
        }
    }
}
```

#### Config

### `point-c`

Handles registered networks and network operations.

#### Config

### `sysnet`

A `point-c` network for the host system.

#### Config

### `wg`

A `point-c` network for WireGuard tunnels.

#### Config

### `listener`

Allows registering a `point-c` network with `merge-listener-wrapper`.

#### Config

### `forward`

Network operation that manages modules that move traffic from a source to a destination.

#### Config

### `forward-tcp`

A forward submodule that forwards TCP traffic.

#### Config

### `stub-listener`

Prevent caddy from listening on the host system.

#### Config

### `rand`

A `caddyhttp.MiddlewareHandler` that returns random data.

#### Config

## Full Configuration

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