package render

import (
	"bytes"
	"compress/zlib"
	"io"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/miguehm/yml-to-cv/internal/i18n"
	"github.com/miguehm/yml-to-cv/internal/model"
)

const sampleYAML = `
basics:
  name: "Jane Doe"
  email: "jane@example.com"
  location: "Puebla, Mexico"
  summary: >-
    A concise professional summary that is long enough to wrap onto more than a
    single line when rendered in the document body.
  profiles:
    - network: GitHub
      url: "https://github.com/jane"
education:
  - institution: "Example University"
    degree: "BS"
    area: "Computer Science"
    location: "Puebla, Mexico"
    start: "2020-08"
    end: "2024-05"
    highlights:
      - "Relevant coursework in algorithms and data structures."
experience:
  - company: "Acme Corp"
    position: "Backend Developer"
    location: "Remote"
    start: "2024-06"
    end: "2025-01"
    highlights:
      - "Built REST APIs with Go and PostgreSQL."
projects:
  - name: "ContraScripts"
    url: "https://contrascripts.miguehm.com"
    date: "2026-06"
    summary: "A web application for writing screenplays."
    highlights:
      - "Uses the Fountain plain-text screenplay format."
skills:
  - label: "Languages"
    details: "Go, Python, JavaScript"
achievements:
  - name: "Second place - Science Fiction Contest"
    date: "2022-08-11"
    highlights:
      - "Awarded for a short story submitted to a university contest."
`

const spanishSampleYAML = `
basics:
  name: "Jane Doe"
  email: "jane@example.com"
  summary: "Resumen profesional conciso."
education:
  - institution: "Example University"
    degree: "BS"
    area: "Computer Science"
    start: "2020-08"
    end: ""
experience:
  - company: "Acme Corp"
    position: "Backend Developer"
    start: "2024-06"
    end: ""
projects:
  - name: "ContraScripts"
    date: "2026-06"
    summary: "Aplicación web."
skills:
  - label: "Languages"
    details: "Go, Python"
achievements:
  - name: "Second place"
    date: "2022-08-11"
`

func renderYAML(t *testing.T, opts Options, doc string) []byte {
	t.Helper()
	cv, err := model.Parse([]byte(doc))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	var buf bytes.Buffer
	if err := NewHarvard(opts).Render(&buf, cv); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	return buf.Bytes()
}

func renderSample(t *testing.T, opts Options) []byte {
	t.Helper()
	return renderYAML(t, opts, sampleYAML)
}

func TestRenderProducesPDF(t *testing.T) {
	out := renderSample(t, Options{})
	if !bytes.HasPrefix(out, []byte("%PDF-")) {
		t.Fatalf("output is not a PDF: %q", out[:min(8, len(out))])
	}
	if !bytes.Contains(out, []byte("/Title")) {
		t.Error("missing /Title metadata")
	}
	if !bytes.Contains(out, []byte("/URI")) {
		t.Error("expected clickable links (/URI) in output")
	}
	if n := bytes.Count(out, []byte("/Type /Page")) - bytes.Count(out, []byte("/Type /Pages")); n < 1 {
		t.Errorf("page count = %d, want >= 1", n)
	}
}

// TestSummaryRendered guards against basics.summary being parsed but never
// drawn into the document.
func TestSummaryRendered(t *testing.T) {
	const summary = "A concise professional summary."
	const doc = `
basics:
  name: "Jane Doe"
  summary: "` + summary + `"
`
	content := pdfContentStreams(t, renderYAML(t, Options{}, doc))
	if !strings.Contains(content, utf16BE(summary)) {
		t.Error("basics.summary was not rendered")
	}
}

func TestRenderDeterministic(t *testing.T) {
	a := renderSample(t, Options{})
	b := renderSample(t, Options{})
	if !bytes.Equal(a, b) {
		t.Fatal("rendering is not deterministic with the default options")
	}
}

func TestRenderA4(t *testing.T) {
	out := renderSample(t, Options{Paper: "a4"})
	if !strings.HasPrefix(string(out), "%PDF-") {
		t.Fatal("A4 render did not produce a PDF")
	}
}

// TestRenderSpanish verifies that the PDF chrome (section headings, date
// markers and the degree connector) follows Options.Language, while the YAML
// content is left untouched.
func TestRenderSpanish(t *testing.T) {
	content := pdfContentStreams(t, renderYAML(t, Options{Language: i18n.Spanish}, spanishSampleYAML))

	wants := []string{
		"EDUCACIÓN",
		"EXPERIENCIA",
		"PROYECTOS",
		"HABILIDADES",
		"LOGROS",
		"ago 2020 \u2013 Actualidad",
		"jun 2024 \u2013 Actualidad",
		"BS en Computer Science",
		"11 ago 2022",
		"Resumen profesional conciso.",
	}
	for _, w := range wants {
		if !strings.Contains(content, utf16BE(w)) {
			t.Errorf("Spanish render is missing %q", w)
		}
	}

	// English chrome must be absent.
	for _, unwanted := range []string{"EDUCATION", "PRESENT"} {
		if strings.Contains(content, utf16BE(unwanted)) {
			t.Errorf("Spanish render should not contain %q", unwanted)
		}
	}
}

// TestRenderDefaultsToEnglish guards the historical default: without an
// explicit language the chrome stays in English.
func TestRenderDefaultsToEnglish(t *testing.T) {
	content := pdfContentStreams(t, renderSample(t, Options{}))
	wants := []string{
		"EDUCATION",
		"Aug 2020 \u2013 May 2024",
		"BS in Computer Science",
	}
	for _, w := range wants {
		if !strings.Contains(content, utf16BE(w)) {
			t.Errorf("default render is missing %q", w)
		}
	}
}

// TestEntryDatesAlignRight guards against the date being drawn at the left
// margin (overlapping the entry title) instead of right-aligned.
func TestEntryDatesAlignRight(t *testing.T) {
	out := renderSample(t, Options{})
	content := pdfContentStreams(t, out)

	date := utf16BE("Aug 2020 \u2013 May 2024")
	re := regexp.MustCompile(`BT ([\d.]+) [\d.]+ Td \(` + regexp.QuoteMeta(date) + `\)Tj`)
	m := re.FindStringSubmatch(content)
	if m == nil {
		t.Fatal("education date not found in rendered content")
	}
	x, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		t.Fatalf("parse date x: %v", err)
	}
	if x < 300 {
		t.Errorf("date x = %.2f; want right-aligned near the right margin", x)
	}
}

// contactOps maps each text-drawing operation in the document to the baselines
// (distance from the top, in points) at which it is drawn.
func contactOps(t *testing.T, content string) map[string][]float64 {
	t.Helper()
	re := regexp.MustCompile(`BT ([\d.]+) ([\d.]+) Td \((.*?)\)Tj ET`)
	ops := map[string][]float64{}
	for _, m := range re.FindAllStringSubmatch(content, -1) {
		y, err := strconv.ParseFloat(m[2], 64)
		if err != nil {
			t.Fatalf("parse y: %v", err)
		}
		ops[m[3]] = append(ops[m[3]], y)
	}
	return ops
}

// TestContactIcons verifies the header shows the expected icon glyphs and that
// each icon shares its baseline with the text it precedes.
func TestContactIcons(t *testing.T) {
	content := pdfContentStreams(t, renderSample(t, Options{}))
	ops := contactOps(t, content)

	pairs := []struct {
		rune rune
		text string
	}{
		{0xF32F, "jane@example.com"}, // envelope
		{0xF3E8, "Puebla, Mexico"},   // geo-alt
		{0xF3ED, "github.com/jane"},  // github
	}
	for _, p := range pairs {
		icon := utf16BE(string(p.rune))
		ys, ok := ops[icon]
		if !ok {
			t.Errorf("missing icon glyph U+%04X", p.rune)
			continue
		}
		textYs, ok := ops[utf16BE(p.text)]
		if !ok {
			t.Errorf("missing contact text %q", p.text)
			continue
		}
		// The icon is drawn slightly below the text baseline so its glyph centre
		// lines up with the cap-height centre of the text (optical centering).
		// The content stream measures distance from the bottom, so a smaller
		// value means lower on the page.
		drop := textYs[0] - ys[0]
		if drop < 0.6 || drop > 2.2 {
			t.Errorf("icon U+%04X drop = %.2fpt; want ~1.5pt of optical centering", p.rune, drop)
		}
	}
}

// TestContactIconsDisabled ensures "basics.icons: false" removes the glyphs.
func TestContactIconsDisabled(t *testing.T) {
	const doc = `
basics:
  name: "Jane Doe"
  icons: false
  email: "jane@example.com"
  location: "Puebla, Mexico"
  profiles:
    - network: GitHub
      url: "https://github.com/jane"
`
	content := pdfContentStreams(t, renderYAML(t, Options{}, doc))
	for _, r := range []rune{0xF32F, 0xF3E8, 0xF3ED, 0xF3EE} {
		if strings.Contains(content, utf16BE(string(r))) {
			t.Errorf("icon U+%04X should be absent when icons are disabled", r)
		}
	}
	if !strings.Contains(content, utf16BE("jane@example.com")) {
		t.Error("contact text should still be rendered")
	}
}

// TestContactWraps checks the contact line wraps gracefully instead of
// overflowing when many profiles are present.
func TestContactWraps(t *testing.T) {
	const doc = `
basics:
  name: "Jane Doe"
  email: "jane.doe@example-university-domain.example"
  location: "Benemerita Universidad Autonoma de Puebla, Puebla, Mexico"
  phone: "221-400-9504"
  profiles:
    - network: GitHub
      url: "https://github.com/jane"
    - network: LinkedIn
      url: "https://linkedin.com/in/jane"
    - network: Website
      url: "https://jane.dev"
    - network: Blog
      url: "https://blog.jane.dev"
    - network: Mastodon
      url: "https://mastodon.social/@jane"
    - network: GitLab
      url: "https://gitlab.com/jane"
`
	content := pdfContentStreams(t, renderYAML(t, Options{}, doc))
	ops := contactOps(t, content)
	baselines := map[float64]bool{}
	for _, text := range []string{"221-400-9504", "jane.doe@example-university-domain.example", "Benemerita Universidad Autonoma de Puebla, Puebla, Mexico", "github.com/jane", "linkedin.com/in/jane", "jane.dev", "blog.jane.dev", "mastodon.social/@jane", "gitlab.com/jane"} {
		ys, ok := ops[utf16BE(text)]
		if !ok {
			t.Fatalf("missing contact text %q", text)
		}
		baselines[ys[0]] = true
	}
	if len(baselines) < 2 {
		t.Errorf("expected the contact line to wrap onto multiple baselines, got %v", baselines)
	}
}

// utf16BE encodes s the way fpdf writes UTF-8 font strings: each rune as two
// big-endian bytes.
func utf16BE(s string) string {
	var b strings.Builder
	for _, r := range s {
		b.WriteByte(byte(r >> 8))
		b.WriteByte(byte(r))
	}
	return b.String()
}

// pdfContentStreams returns the concatenated zlib-decompressed content
// streams of a PDF document.
func pdfContentStreams(t *testing.T, pdf []byte) string {
	t.Helper()
	re := regexp.MustCompile(`(?s)stream\r?\n(.*?)endstream`)
	var b strings.Builder
	for _, m := range re.FindAllSubmatch(pdf, -1) {
		zr, err := zlib.NewReader(bytes.NewReader(m[1]))
		if err != nil {
			continue
		}
		dec, err := io.ReadAll(zr)
		zr.Close()
		if err != nil {
			continue
		}
		b.Write(dec)
	}
	return b.String()
}
