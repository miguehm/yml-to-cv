package i18n

import "testing"

func TestParse(t *testing.T) {
	cases := []struct {
		in   string
		want Language
		ok   bool
	}{
		{"", English, true},
		{"en", English, true},
		{"EN", English, true},
		{" en ", English, true},
		{"es", Spanish, true},
		{"ES", Spanish, true},
		{"Es", Spanish, true},
		{" fr ", "", false},
		{"spanish", "", false},
		{"es-MX", "", false},
	}
	for _, c := range cases {
		got, err := Parse(c.in)
		if c.ok {
			if err != nil {
				t.Errorf("Parse(%q) error = %v", c.in, err)
				continue
			}
			if got != c.want {
				t.Errorf("Parse(%q) = %q, want %q", c.in, got, c.want)
			}
		} else if err == nil {
			t.Errorf("Parse(%q) = %q, want error", c.in, got)
		}
	}
}

func TestForDefaultsToEnglish(t *testing.T) {
	cat := For("")
	if cat.Education != "Education" {
		t.Errorf("For(\"\").Education = %q, want %q", cat.Education, "Education")
	}
	if cat.Present != "Present" {
		t.Errorf("For(\"\").Present = %q, want %q", cat.Present, "Present")
	}
	if cat.DegreeConnector != "in" {
		t.Errorf("For(\"\").DegreeConnector = %q, want %q", cat.DegreeConnector, "in")
	}
	if cat.DayFirst {
		t.Error("For(\"\").DayFirst = true, want false")
	}
}

func TestSpanishCatalog(t *testing.T) {
	cat := For(Spanish)
	cases := []struct{ name, got, want string }{
		{"Education", cat.Education, "Educación"},
		{"Experience", cat.Experience, "Experiencia"},
		{"Projects", cat.Projects, "Proyectos"},
		{"Skills", cat.Skills, "Habilidades"},
		{"Achievements", cat.Achievements, "Logros"},
		{"Present", cat.Present, "Actualidad"},
		{"DegreeConnector", cat.DegreeConnector, "en"},
		{"MonthNames[7]", cat.MonthNames[7], "ago"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("Spanish %s = %q, want %q", c.name, c.got, c.want)
		}
	}
	if !cat.DayFirst {
		t.Error("Spanish DayFirst = false, want true")
	}
}

func TestFormatDate(t *testing.T) {
	cases := []struct {
		cat   Catalog
		day   int
		month string
		year  int
		want  string
	}{
		{englishCatalog, 0, "", 2020, "2020"},
		{englishCatalog, 0, "Aug", 2020, "Aug 2020"},
		{englishCatalog, 11, "Aug", 2022, "Aug 11, 2022"},
		{spanishCatalog, 0, "", 2020, "2020"},
		{spanishCatalog, 0, "ago", 2020, "ago 2020"},
		{spanishCatalog, 11, "ago", 2022, "11 ago 2022"},
		{spanishCatalog, 1, "ene", 0, ""},
	}
	for _, c := range cases {
		if got := c.cat.FormatDate(c.day, c.month, c.year); got != c.want {
			t.Errorf("FormatDate(%d, %q, %d) = %q, want %q", c.day, c.month, c.year, got, c.want)
		}
	}
}
