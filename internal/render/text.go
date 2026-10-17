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

// segmentText returns the concatenated display text of the segments.
func segmentText(segs []segment) string {
	var b strings.Builder
	for _, s := range segs {
		b.WriteString(s.Text)
	}
	return b.String()
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
