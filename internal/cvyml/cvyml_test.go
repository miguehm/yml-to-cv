package cvyml

import (
	"bytes"
	"strings"
	"testing"
)

const minimalYAML = `
basics:
  name: "Jane Doe"
education:
  - institution: "Example University"
    degree: "BS"
    area: "Computer Science"
    start: "2020-08"
    end: "2024-05"
`

func TestRenderRejectsUnknownLanguage(t *testing.T) {
	var buf bytes.Buffer
	err := Render(&buf, []byte(minimalYAML), Options{Lang: "fr"})
	if err == nil {
		t.Fatal("Render() with Lang \"fr\" expected error")
	}
	if !strings.Contains(err.Error(), "idioma no soportado") {
		t.Errorf("Render() error = %q, want mention of unsupported language", err)
	}
	if buf.Len() != 0 {
		t.Error("Render() wrote output despite the invalid language")
	}
}

func TestRenderAcceptsLanguageCaseInsensitive(t *testing.T) {
	for _, lang := range []string{"", "en", "EN", "es", "ES", " Es "} {
		var buf bytes.Buffer
		if err := Render(&buf, []byte(minimalYAML), Options{Lang: lang}); err != nil {
			t.Errorf("Render() with Lang %q unexpected error: %v", lang, err)
			continue
		}
		if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF-")) {
			t.Errorf("Render() with Lang %q did not produce a PDF", lang)
		}
	}
}
