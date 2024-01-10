# Caddy Merge Listener Wrapper

## Overview
The Caddy Merge Listener Wrapper is a module for the Caddy server that allows merging multiple network listeners. It provides the functionality to aggregate connections from multiple sources into a single listener.

## Features
- **Multiple Listener Support**: Merge multiple `net.Listener` instances into a single listener.
- **Flexible Configuration**: Supports dynamic configuration of listeners via JSON and Caddyfile.

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
   xcaddy build --with github.com/point-c/caddy-merge-listener-wrapper
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
# Caddyfile configuration...
{
    servers :443 {
    	listener_wrappers {
    		multi {
			<submodule name> <submodule config>
    		}
    		# Important! `tls` must be defined and the last wrapper in the list
    		tls
    	}
    }
}

# Server configuration...
```