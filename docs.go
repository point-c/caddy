//go:build docs

package caddy

//go:generate rm -rf docs
//go:generate go run "github.com/johnstarich/go/gopages" -base /$GOPACKAGE -out docs

import _ "github.com/johnstarich/go/gopages"
