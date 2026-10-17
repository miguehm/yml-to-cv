package render

import (
	"regexp"
	"strings"
)

// segment is a piece of text, optionally linked to a URL.
type segment struct {
	Text string
	URL  string
}

// linkPattern matches markdown-style links: [text](url).
var linkPattern = regexp.MustCompile(`\[([^\]]+)\]\(([^)\s]+)\)`)

// splitSegments parses markdown-style links out of s. Plain text between links
// becomes unlinked segments. The returned slice always has at least one element
// when s is non-empty.
func splitSegments(s string) []segment {
	var segs []segment
	last := 0
	for _, m := range linkPattern.FindAllStringSubmatchIndex(s, -1) {
		if m[0] > last {
			segs = append(segs, segment{Text: s[last:m[0]]})
		}
		segs = append(segs, segment{Text: s[m[2]:m[3]], URL: s[m[4]:m[5]]})
		last = m[1]
	}
	if last < len(s) {
		segs = append(segs, segment{Text: s[last:]})
	}
	if len(segs) == 0 {
		segs = append(segs, segment{Text: s})
	}
	return segs
}

// shortURL returns the print-friendly form of a URL: the host and path without
// the scheme or a leading "www.". The full URL is still used as the link target
// so the document stays clickable on screen while remaining legible on paper.
func shortURL(u string) string {
	s := strings.TrimSpace(u)
	if s == "" {
		return ""
	}
	if scheme := strings.Index(s, "://"); scheme >= 0 {
		s = s[scheme+len("://"):]
	}
	s = strings.TrimPrefix(s, "mailto:")
	s = strings.TrimPrefix(s, "www.")
	return strings.TrimRight(s, "/")
}

// hasURL reports whether any segment carries a link.
func hasURL(segs []segment) bool {
	for _, s := range segs {
		if s.URL != "" {
			return true
		}
	}
	return false
}
