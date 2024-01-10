# Caddy Stub Listener Plugin

## Overview
The Caddy Stub Listener Plugin provides a stub network listener designed to be used as a plugin for the Caddy server. This stub listener doesn't accept actual network connections but blocks on Accept calls until Close is called. It's primarily used when tunnel listeners are required or for situations where the network handling is managed elsewhere.

## Features
- **Caddy Plugin**: Seamlessly integrates with the Caddy server as a plugin.
- **Stub Network Listener**: Creates a network listener that simulates network behavior without establishing actual connections.

## Getting Started

### Prerequisites
Ensure you have Go installed on your system.

### Installation
To install the Stub Listener as a Caddy module using the `xcaddy` utility:

1. **Install xcaddy**:
   If you don't have `xcaddy` installed, you can install it by running:
    ```sh
    go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest
    ```

2. **Build Caddy with the Stub Listener Plugin**:
   Use `xcaddy` to build Caddy with the Stub Listener plugin included:
    ```sh
    xcaddy build --with github.com/point-c/stub-listener
    ```

3. **Run Your Custom Caddy Build**:
   After the build is complete, you'll have a custom Caddy binary with the Stub Listener plugin. Run it using:
    ```sh
    ./caddy run
    ```

### Usage
After installing the Stub Listener plugin, you need to configure it in your Caddy setup. Here's how you can do it with both JSON and Caddyfile configurations:

#### JSON Configuration
To use the stub listener in a Caddy JSON configuration:

1. **Edit Your Caddy JSON File**: Add a listener with the network type set to "stub". Here's an example snippet:

    ```json
    {
        "apps": {
            "http": {
                "servers": {
                    "example": {
                        "listen": ["stub://localhost:443"]
                        // Other server configurations...
                    }
                }
            }
        }
    }
    ```

2. **Run Caddy with the JSON Config**: Use the command `./caddy run --config /path/to/your/config.json` to start Caddy with the specified configuration.

#### Caddyfile Configuration
To use the stub listener in a Caddyfile:

1. **Edit Your Caddyfile**: Define a site and specify the use of the stub listener. Here's an example:

    ```
    {
        default_bind stub://0.0.0.0
    }
    
    foo.bar.com {
    }
    ```

2. **Run Caddy**: Start your Caddy server normally using the command `./caddy run` or by specifying the Caddyfile directly if not using the default location.