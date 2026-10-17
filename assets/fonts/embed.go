// Package fonts embeds the Tinos typeface (metric-compatible with Times New
// Roman, SIL Open Font License) used to render the CV in a classic serif
// Harvard style while supporting full UTF-8 text.
package fonts

import _ "embed"

// TinosRegular is the regular weight.
//
//go:embed Tinos-Regular.ttf
var TinosRegular []byte

// TinosBold is the bold weight.
//
//go:embed Tinos-Bold.ttf
var TinosBold []byte

// TinosItalic is the italic style.
//
//go:embed Tinos-Italic.ttf
var TinosItalic []byte

// TinosBoldItalic is the bold italic style.
//
//go:embed Tinos-BoldItalic.ttf
var TinosBoldItalic []byte
