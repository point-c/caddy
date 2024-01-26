package module

import (
	_ "github.com/point-c/caddy/module/forward"
	_ "github.com/point-c/caddy/module/forward-tcp"
	_ "github.com/point-c/caddy/module/listener"
	_ "github.com/point-c/caddy/module/merge-listener-wrapper"
	_ "github.com/point-c/caddy/module/point-c"
	_ "github.com/point-c/caddy/module/rand"
	_ "github.com/point-c/caddy/module/stub-listener"
	_ "github.com/point-c/caddy/module/sysnet"
	_ "github.com/point-c/caddy/module/wg"
)

func init() { _ = 1 }
