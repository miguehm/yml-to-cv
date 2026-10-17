package model

import "testing"

func TestParseValid(t *testing.T) {
	const doc = `
basics:
  name: "Jane Doe"
  email: "jane@example.com"
  profiles:
    - network: GitHub
      url: "github.com/jane"
education:
  - institution: "Example University"
    degree: "BS"
    area: "Computer Science"
    start: "2020-08"
    end: "2024-05"
skills:
  - label: "Languages"
    details: "Go, Python"
`
	cv, err := Parse([]byte(doc))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if cv.Basics.Name != "Jane Doe" {
		t.Errorf("name = %q", cv.Basics.Name)
	}
	if got := cv.Basics.Profiles[0].URL; got != "https://github.com/jane" {
		t.Errorf("normalized profile URL = %q", got)
	}
	if got := FormatRange(cv.Education[0].Start, cv.Education[0].End); got != "Aug 2020 \u2013 May 2024" {
		t.Errorf("education range = %q", got)
	}
	if len(cv.SectionOrder) != len(DefaultSectionOrder) {
		t.Errorf("default section order not applied: %v", cv.SectionOrder)
	}
}

func TestParseUnknownField(t *testing.T) {
	const doc = `
basics:
  name: "Jane Doe"
  emial: "typo@example.com"
`
	if _, err := Parse([]byte(doc)); err == nil {
		t.Fatal("expected error for unknown field, got nil")
	}
}

func TestParseMissingName(t *testing.T) {
	if _, err := Parse([]byte("basics:\n  email: a@b.com\n")); err == nil {
		t.Fatal("expected error for missing name, got nil")
	}
}

func TestParseInvalidDate(t *testing.T) {
	const doc = `
basics:
  name: "Jane Doe"
education:
  - institution: "X"
    start: "2020-13"
`
	if _, err := Parse([]byte(doc)); err == nil {
		t.Fatal("expected error for invalid month, got nil")
	}
}

func TestParseUnknownSection(t *testing.T) {
	const doc = `
basics:
  name: "Jane Doe"
section_order: [nope]
`
	if _, err := Parse([]byte(doc)); err == nil {
		t.Fatal("expected error for unknown section, got nil")
	}
}

func TestNormalizeURL(t *testing.T) {
	cases := map[string]string{
		"":                  "",
		"miguehm.com":       "https://miguehm.com",
		"https://x.dev":     "https://x.dev",
		"mailto:a@b.com":    "mailto:a@b.com",
		"  https://x.dev  ": "https://x.dev",
	}
	for in, want := range cases {
		if got := normalizeURL(in); got != want {
			t.Errorf("normalizeURL(%q) = %q, want %q", in, got, want)
		}
	}
}
