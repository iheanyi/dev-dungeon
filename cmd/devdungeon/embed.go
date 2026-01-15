package main

import (
	"embed"
	"io/fs"
)

//go:embed all:static
var embeddedStatic embed.FS

// getEmbeddedStaticFS returns the embedded static files.
// Returns nil if no embedded files are available or if just a placeholder exists.
func getEmbeddedStaticFS() fs.FS {
	sub, err := fs.Sub(embeddedStatic, "static")
	if err != nil {
		return nil
	}

	// Check if we have actual content (index.html), not just a placeholder
	if _, err := fs.Stat(sub, "index.html"); err != nil {
		return nil
	}

	return sub
}
