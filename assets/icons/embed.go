// Package icons embeds the Bootstrap Icons typeface (MIT License) used to draw
// the small monochrome glyphs next to the contact details in the CV header.
//
// The bundled TrueType file was converted from the official WOFF distribution
// (bootstrap-icons@1.13.2); see LICENSE in this directory.
package icons

import _ "embed"

// BootstrapIcons is the Bootstrap Icons TrueType font.
//
//go:embed bootstrap-icons.ttf
var BootstrapIcons []byte
