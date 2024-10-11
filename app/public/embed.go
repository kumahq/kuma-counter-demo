package public

import "embed"

//go:embed *.html *.svg *.png
var Files embed.FS
