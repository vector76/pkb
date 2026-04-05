// Package web embeds the PKB static assets and HTML templates.
package web

import "embed"

//go:embed templates static
var Assets embed.FS
