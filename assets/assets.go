// Package assets provides static assets for rich content
package assets

import "embed"

//go:embed src
// FS is the virtual static filesystem
var FS embed.FS
