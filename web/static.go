package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed all:static
var static embed.FS

// StaticFS returns the embedded front-end filesystem rooted at web/static.
func StaticFS() (http.FileSystem, error) {
	sub, err := fs.Sub(static, "static")
	if err != nil {
		return nil, err
	}
	return http.FS(sub), nil
}
