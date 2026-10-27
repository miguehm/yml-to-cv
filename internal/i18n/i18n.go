// Package i18n provides the small translation catalogue used for the text the
// renderer emits itself ("chrome"): section headings, the open-ended date
// marker, the degree connector and the month names.
//
// User content coming from the YAML (names, summaries, highlights, ...) is
// never translated: it is always rendered exactly as written.
package i18n

import (
	"fmt"
	"strings"
)

// Language identifies a supported UI language.
type Language string

// Supported languages.
const (
	English Language = "en"
	Spanish Language = "es"
)

// Parse normalises a language tag and returns the matching Language. An empty
// string means English, which keeps the historical behaviour of the tool.
func Parse(s string) (Language, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", string(English):
		return English, nil
	case string(Spanish):
		return Spanish, nil
	default:
		return "", fmt.Errorf("idioma no soportado %q (usa %q o %q)", s, English, Spanish)
	}
}

// Catalog holds every translated string the renderer needs.
type Catalog struct {
	// Education, Experience, Projects, Skills and Achievements are the section
	// headings, keyed by the constants in the model package.
	Education    string
	Experience   string
	Projects     string
	Skills       string
	Achievements string

	// Present is the marker for an open-ended date range.
	Present string
	// DegreeConnector joins a degree with its area, e.g. "BS in Computer
	// Science".
	DegreeConnector string

	// MonthNames are the month abbreviations used in rendered dates.
	MonthNames [12]string
	// DayFirst selects the day-first date layout ("11 Aug 2022") over the
	// month-first one ("Aug 11, 2022").
	DayFirst bool
}

// FormatDate renders a single date with the catalog's month name and layout.
// A zero day renders "Aug 2020"; a zero month renders "2020".
func (c Catalog) FormatDate(day int, month string, year int) string {
	switch {
	case year == 0:
		return ""
	case month == "":
		return fmt.Sprintf("%d", year)
	case day == 0:
		return month + " " + fmt.Sprintf("%d", year)
	case c.DayFirst:
		return fmt.Sprintf("%d %s %d", day, month, year)
	default:
		return fmt.Sprintf("%s %d, %d", month, day, year)
	}
}

// For returns the catalog for the given language, defaulting to English for
// unknown values so the renderer never fails on a zero Language.
func For(lang Language) Catalog {
	if lang == Spanish {
		return spanishCatalog
	}
	return englishCatalog
}

var englishCatalog = Catalog{
	Education:       "Education",
	Experience:      "Experience",
	Projects:        "Projects",
	Skills:          "Skills",
	Achievements:    "Achievements",
	Present:         "Present",
	DegreeConnector: "in",
	MonthNames: [12]string{
		"Jan", "Feb", "Mar", "Apr", "May", "Jun",
		"Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
	},
	DayFirst: false,
}

var spanishCatalog = Catalog{
	Education:       "Educación",
	Experience:      "Experiencia",
	Projects:        "Proyectos",
	Skills:          "Habilidades",
	Achievements:    "Logros",
	Present:         "Actualidad",
	DegreeConnector: "en",
	MonthNames: [12]string{
		"ene", "feb", "mar", "abr", "may", "jun",
		"jul", "ago", "sep", "oct", "nov", "dic",
	},
	DayFirst: true,
}
