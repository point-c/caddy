# Point-c

## Overview

These plugins allow caddy to manage networks more efficiently. 
Networking applications can be run under this module allowing fine grained networking.

## Features

- **Lifecycle Manegement**: Multiple submodules can be run from this module.
- **Networking Apps**: Networking applications can be created and used via this module.
- **Network Listener**: Listen on custom networks.

## Getting Started

### Prerequisites
Ensure you have Go installed on your system.

### Installation
To install the Caddy Merge Listener Wrapper, you will need to build a custom Caddy binary that includes this module. This can be achieved using the `xcaddy` utility:

1. **Install xcaddy**:
   ```sh
   go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest
   ```

2. **Build Caddy with the Merge Listener Wrapper**:
   ```sh
   xcaddy build --with github.com/point-c/caddy
   ```

3. **Run Your Custom Caddy Build**:
   ```sh
   ./caddy run
   ```

### Configuration
#### JSON Configuration
Edit your Caddy JSON configuration to include the Merge Listener Wrapper. Here's a snippet to get you started:

```json
{
    // Other Caddy configurations...
    "apps": {
        "http": {
            "servers": {
                "example": {
                    "listener_wrappers": [
                        {
                            "wrapper": "multi",
                            "listeners": [
                            	{
                            		"listener": "<listener>",
                            		...
                            	},
                            	...
                            ]
                        }
                    ],
                    // Other server configurations...
                }
            }
        }
    }
}
```

#### Caddyfile Configuration
In your Caddyfile, you can use the Merge Listener Wrapper as follows:

```
{
   # Global config section
    point-c {
        <submodule name> <submodule options...>
    }
    netop {
        <submodule name> <submodule options...>
    }
}
```