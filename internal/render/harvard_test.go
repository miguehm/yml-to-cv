package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/miguehm/cvyml/internal/model"
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

func renderSample(t *testing.T, opts Options) []byte {
	t.Helper()
	cv, err := model.Parse([]byte(sampleYAML))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	var buf bytes.Buffer
	if err := NewHarvard(opts).Render(&buf, cv); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	return buf.Bytes()
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
