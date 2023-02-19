package static

import (
	"embed"
	"io/fs"
)

//go:embed index.html status
var content embed.FS

func FS() fs.FS { return content }
