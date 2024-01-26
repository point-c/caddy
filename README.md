# point-c

[![Go Reference](https://img.shields.io/badge/godoc-reference-%23007d9c.svg)](https://point-c.github.io/caddy)

point-c is a collection of Caddy modules for handling traffic between host systems and WireGuard tunnels.

`point-c` is a collection of Caddy modules designed to handle network traffic between host systems and WireGuard tunnels. This library is for users looking to integrate robust network handling into their Caddy server configurations, particularly in VPN scenarios.

## Features

- Easy integration with Caddy and WireGuard.
- Customizable modules for different network operations.
- Streamlined handling of TCP traffic and listener wrapping.

## Installation

To install the modules from `point-c`, you will need to build a custom Caddy binary that includes that module. This can be achieved using the `xcaddy` utility:

1. **Install xcaddy**:
   ```sh
   go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest
   ```

2. **Build Caddy with all modules**:
   ```sh
   xcaddy build --with github.com/point-c/caddy/module@latest
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

The merge-listener-wrapper module acts as a wrapper for TCP listeners within the Caddy server. This module enables the bundling of multiple listeners into a single `net.Listener`. It's useful for allowing different networks to combine traffic.

#### Config

##### Caddyfile

```Caddyfile
{
    servers <listen address> {
        listener_wrappers {
            merge {
               # Listener definitions go here 
            }
        }
    }
}
```

##### JSON

```json
{
  "apps": {
    "http": {
      "servers": {
        "srv0": {
          "listen": [
            "<listen address>"
          ],
          "listener_wrappers": [
            {
              "listeners": [],
              "wrapper": "merge"
            },
            {
              "wrapper": "tls"
            }
          ]
        }
      }
    }
  }
}
```

### `point-c`

The centerpiece of this library, `point-c`, is designed to handle and manage registered networks and their associated operations within Caddy. This module serves as the backbone for integrating Caddy with various network setups, including both system networks and WireGuard configurations. It's essential for users looking to leverage Caddy in network-intensive applications or in scenarios requiring network control.

#### Config

##### Caddyfile

```Caddyfile
{
	point-c {
	}
	point-c netops {
	}
}
```

##### JSON

```json
{
   "apps": {
      "point-c": {
         "networks": [],
         "net-ops": []
      }
   }
}
```

### `sysnet`

`sysnet` is a dedicated point-c network module for the host system. It provides a streamlined way to integrate the host's network settings into the point-c ecosystem. This module is for scenarios where the host system's network needs to be used in a `point-c` network operation.

#### Config

```Caddyfile
{
	point-c {
		system <hostname> <ip address>
	}
}
```

##### JSON

```json
{
   "apps": {
      "point-c": {
         "networks": [
            {
               "addr": "<ip address>",
               "hostname": "<hostname>",
               "type": "system"
            }
         ]
      }
   }
}
```

### `wg`

The `wg` module enables registration of WireGuard tunnels, allowing integration of WireGuard-based VPN configurations with Caddy. The module is for users looking to combine WireGuard Caddy's web-serving functionality.

#### Config

```Caddyfile
{
	point-c {
		wgclient <hostname> {
			ip <address on the virtual network>
			endpoint <server address>
			private <client private key>
			public <server public key>
			shared <shared key>
		}
		wgserver <hostname> {
			ip <address on the virtual network>
			port <server listen port>
			private <server private key>
			peer <hostname> {
				ip <address on the virtual network>
				public <client public key>
				shared <shared key>
			}
		}
	}
}
```

##### JSON

```json
{
   "apps": {
      "point-c": {
         "networks": [
            {
               "endpoint": "<server address>",
               "ip": "<address on the virtual network>",
               "name": "<hostname>",
               "preshared": "<shared key>",
               "private": "<client private key>",
               "public": "<server public key>",
               "type": "wgclient"
            },
            {
               "hostname": "<hostname>",
               "ip": "<address on the virtual network>",
               "listen-port": <server listen port>,
               "peers": [
                  {
                     "hostname": "<hostname>",
                     "ip": "<address on the virtual network>",
                     "preshared": "<shared key>",
                     "public": "<client public key>"
                  }
               ],
               "private": "<server private key>",
               "type": "wgserver"
            }
         ]
      }
   }
}
```

### `listener`

The `listener` module for `point-c` allows for the use of registered networks with the `merge-listener-wrapper`. This module allows using networks as Caddy listeners.

#### Config

##### Caddyfile

```Caddyfile
{
    servers <listen address> {
        listener_wrappers {
            merge {
               point-c <hostname> <listen port>
            }
            tls # should always be last
        }
    }
}
```

##### JSON

```json
{
  "apps": {
    "http": {
      "servers": {
        "srv0": {
          "listen": [
            "<listen address>"
          ],
          "listener_wrappers": [
            {
              "listeners": [
                {
                  "listener": "<hostname>",
                  "name": "server",
                  "port": <listen port>
                }
              ],
              "wrapper": "merge"
            },
            {
              "wrapper": "tls"
            }
          ]
        }
      }
    }
  }
}
```

### `forward`

The `forward` module specializes in directing traffic from a source to a destination. It allows multiple port forwards to be made for a `src host:dst host` pair.

#### Config

##### Caddyfile

```Caddyfile
{
	point-c netops {
		forward <src host>:<dst host> {
		}
	}
}
```

##### JSON

```json
{
   "apps": {
      "point-c": {
         "net-ops": [
            {
               "forwards": [],
               "hosts": "<src host>:<dst host>",
               "op": "forward"
            }
         ]
      }
   }
}
```

### `forward-tcp`

As a submodule of `forward`, `forward-tcp` focuses on forwarding TCP traffic. It allows specifying a specific buffer size in bytes. The default is `4096`.

#### Config

##### Caddyfile

```Caddyfile
{
	point-c netops {
		forward <src host>:<dst host> {
			tcp <src port>:<dst port> [buffer size in bytes]
		}
	}
}
```

##### JSON

```json
{
   "apps": {
      "point-c": {
         "net-ops": [
            {
               "forwards": [
                  {
                     "buf": <null | buffer size in bytes>,
                     "forward": "tcp",
                     "ports": "<src port>:<dst port>"
                  }
               ],
               "hosts": "<src host>:<dst host>",
               "op": "forward"
            }
         ]
      }
   }
}
```

### `stub-listener`

`stub-listener` allows binding to nothing, preventing Caddy from listening on the host system. This is useful for isolating Caddy from the host environment, such as when it is unable to open ports on the host system.

#### Config

##### Caddyfile

```Caddyfile
{
	default_bind stub://<ip address>
}
```

##### JSON

```json

{
  "apps": {
    "http": {
      "servers": {
        "srv0": {
          "listen": [
            "stub://<ip address>:<port>"
          ]
        }
      }
    }
  }
}
```

### `rand`

`rand` is a `caddyhttp.MiddlewareHandler` that returns random data, useful for testing.

Accepts the following headers:
- `Rand-Seed`: The seed for the random data generator. If not specified, the current Unix microsecond timestamp is used.
- `Rand-Size`: The size of the random data in bytes. If not specified or set to a negative value, the module will stream random data indefinitely.

#### Config

##### Caddyfile

```Caddyfile
<site block> {
    route {
        rand
    }
}
```

##### JSON Configuration

```json
{
   "apps": {
      "http": {
         "servers": {
            "srv0": {
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

## Full Configuration

For advanced users, point-c offers extended configuration options to tailor the modules to specific needs.

### Caddyfile

```Caddyfile
{
	default_bind stub://0.0.0.0
	point-c {
		system sys 0.0.0.0
		wgclient client {
			ip 192.168.45.2
			endpoint 127.0.0.1:51820
			private UCoEdsc8Mw7ZY81jSAHOGIw23QxqxfN8SQ8YktOrw0I=
			public 5GIGlLmvYnTyoQ59QIUYEo2FFUgubTibAO2qFI859hY=
			shared Z9Ad3ZhTQbIUCLEKATYXS1m380vYrYFhGA75tspxsOU=
		}
		wgserver server {
			ip 192.168.45.1
			port 51820
			private 2Jgm2q3tFu21cO1IMyhjENqp7t5qep0++novkdKHe0k=
			peer client-1 {
				ip 192.168.45.3
				public Tdbxgh9AHWXodT60AiwCUPDTEITyVD+ecMhp2TDY1xw=
				shared Z9Ad3ZhTQbIUCLEKATYXS1m380vYrYFhGA75tspxsOU=
			}
		}
	}
	point-c netops {
		forward server:client-1 {
			tcp 443:443
		}
		forward sys:server {
			tcp 443:443 8192
		}
	}
	servers :443 {
		listener_wrappers {
			merge {
				point-c server 443
				point-c client 443
			}
			tls
		}
	}
}

:443 {
	route {
		rand
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
               "listen": [
                  "stub://0.0.0.0:443"
               ],
               "listener_wrappers": [
                  {
                     "listeners": [
                        {
                           "listener": "point-c",
                           "name": "server",
                           "port": 443
                        },
                        {
                           "listener": "point-c",
                           "name": "client",
                           "port": 443
                        }
                     ],
                     "wrapper": "merge"
                  },
                  {
                     "wrapper": "tls"
                  }
               ],
               "routes": [
                  {
                     "handle": [
                        {
                           "handler": "subroute",
                           "routes": [
                              {
                                 "handle": [
                                    {
                                       "handler": "rand"
                                    }
                                 ]
                              }
                           ]
                        }
                     ]
                  }
               ]
            }
         }
      },
      "point-c": {
         "networks": [
            {
               "addr": "0.0.0.0",
               "hostname": "sys",
               "type": "system"
            },
            {
               "endpoint": "127.0.0.1:51820",
               "ip": "192.168.45.2",
               "name": "client",
               "preshared": "Z9Ad3ZhTQbIUCLEKATYXS1m380vYrYFhGA75tspxsOU=",
               "private": "UCoEdsc8Mw7ZY81jSAHOGIw23QxqxfN8SQ8YktOrw0I=",
               "public": "5GIGlLmvYnTyoQ59QIUYEo2FFUgubTibAO2qFI859hY=",
               "type": "wgclient"
            },
            {
               "hostname": "server",
               "ip": "192.168.45.1",
               "listen-port": 51820,
               "peers": [
                  {
                     "hostname": "client-1",
                     "ip": "192.168.45.3",
                     "preshared": "Z9Ad3ZhTQbIUCLEKATYXS1m380vYrYFhGA75tspxsOU=",
                     "public": "Tdbxgh9AHWXodT60AiwCUPDTEITyVD+ecMhp2TDY1xw="
                  }
               ],
               "private": "2Jgm2q3tFu21cO1IMyhjENqp7t5qep0++novkdKHe0k=",
               "type": "wgserver"
            }
         ],
         "net-ops": [
            {
               "forwards": [
                  {
                     "buf": null,
                     "forward": "tcp",
                     "ports": "443:443"
                  }
               ],
               "hosts": "server:client-1",
               "op": "forward"
            },
            {
               "forwards": [
                  {
                     "buf": 8192,
                     "forward": "tcp",
                     "ports": "443:443"
                  }
               ],
               "hosts": "sys:server",
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
go test ./...
```