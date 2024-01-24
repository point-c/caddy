//go:build docs

package point_c

//go:generate rm -rf docs
//go:generate go run "github.com/johnstarich/go/gopages" -base /caddy -internal -out docs -source-link "https://github.com/point-c/caddy/blob/main/{{.Path}}{{if .Line}}#L{{.Line}}{{end}}"

import _ "github.com/johnstarich/go/gopages"
