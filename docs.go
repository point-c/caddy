//go:build docs

package caddy

//go:generate rm -rf docs
//go:generate go run "github.com/therve/go/gopages" -base /$GOPACKAGE -internal -out docs -source-link "https://github.com/point-c/$GOPACKAGE/blob/main/{{.Path}}{{if .Line}}#L{{.Line}}{{end}}"

import _ "github.com/therve/go/gopages"
