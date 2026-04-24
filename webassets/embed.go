package webassets

import "embed"

//go:embed index.html
var IndexHTML embed.FS

//go:embed all:static
var Static embed.FS
