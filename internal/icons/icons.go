// Package icons defines the small monochrome glyphs drawn next to the contact
// details in the CV header and maps them to code points in the embedded
// Bootstrap Icons font.
package icons

import "strings"

// Supported icon names. They are the values accepted by the optional "icon"
// field of a profile.
const (
	GitHub    = "github"
	Globe     = "globe"
	Envelope  = "envelope"
	Telephone = "telephone"
	Location  = "geo-alt"
	LinkedIn  = "linkedin"
	Link      = "link"
)

// runes maps an icon name to its Bootstrap Icons code point.
var runes = map[string]rune{
	GitHub:    0xF3ED,
	Globe:     0xF3EE,
	Envelope:  0xF32F,
	Telephone: 0xF5C1,
	Location:  0xF3E8,
	LinkedIn:  0xF472,
	Link:      0xF471,
}

// IsName reports whether name is a supported icon.
func IsName(name string) bool {
	_, ok := runes[strings.TrimSpace(name)]
	return ok
}

// Rune returns the font code point for name, or 0 when the name is unknown.
func Rune(name string) rune {
	return runes[strings.TrimSpace(name)]
}

// ForNetwork returns the icon that best represents a social or personal
// profile network. Unknown networks fall back to a globe.
func ForNetwork(network string) string {
	switch strings.ToLower(strings.TrimSpace(network)) {
	case "github":
		return GitHub
	case "linkedin":
		return LinkedIn
	default:
		return Globe
	}
}
