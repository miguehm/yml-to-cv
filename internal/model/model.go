// Package model defines the CV data schema and parses it from YAML.
package model

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Section names recognised in the "section_order" field.
const (
	SectionEducation    = "education"
	SectionExperience   = "experience"
	SectionProjects     = "projects"
	SectionSkills       = "skills"
	SectionAchievements = "achievements"
)

// DefaultSectionOrder is used when "section_order" is not provided.
var DefaultSectionOrder = []string{
	SectionEducation,
	SectionExperience,
	SectionProjects,
	SectionSkills,
	SectionAchievements,
}

// CV is the root document.
type CV struct {
	Basics       Basics        `yaml:"basics"`
	SectionOrder []string      `yaml:"section_order,omitempty"`
	Education    []Education   `yaml:"education,omitempty"`
	Experience   []Experience  `yaml:"experience,omitempty"`
	Projects     []Project     `yaml:"projects,omitempty"`
	Skills       []Skill       `yaml:"skills,omitempty"`
	Achievements []Achievement `yaml:"achievements,omitempty"`
}

// Basics holds the header/contact information.
type Basics struct {
	Name     string    `yaml:"name"`
	Label    string    `yaml:"label,omitempty"`
	Location string    `yaml:"location,omitempty"`
	Email    string    `yaml:"email,omitempty"`
	Phone    string    `yaml:"phone,omitempty"`
	Summary  string    `yaml:"summary,omitempty"`
	Profiles []Profile `yaml:"profiles,omitempty"`
}

// Profile is a social network or personal link.
type Profile struct {
	Network  string `yaml:"network"`
	Username string `yaml:"username,omitempty"`
	URL      string `yaml:"url,omitempty"`
}

// Education is a single academic entry.
type Education struct {
	Institution string      `yaml:"institution"`
	Degree      string      `yaml:"degree,omitempty"`
	Area        string      `yaml:"area,omitempty"`
	Location    string      `yaml:"location,omitempty"`
	Start       PartialDate `yaml:"start,omitempty"`
	End         PartialDate `yaml:"end,omitempty"`
	Highlights  []string    `yaml:"highlights,omitempty"`
}

// Experience is a single work entry.
type Experience struct {
	Company    string      `yaml:"company"`
	Position   string      `yaml:"position"`
	Location   string      `yaml:"location,omitempty"`
	Start      PartialDate `yaml:"start,omitempty"`
	End        PartialDate `yaml:"end,omitempty"`
	Highlights []string    `yaml:"highlights,omitempty"`
}

// Project is a single personal or professional project.
type Project struct {
	Name       string      `yaml:"name"`
	URL        string      `yaml:"url,omitempty"`
	Date       PartialDate `yaml:"date,omitempty"`
	Summary    string      `yaml:"summary,omitempty"`
	Highlights []string    `yaml:"highlights,omitempty"`
}

// Skill is a labelled group of technologies or competencies.
type Skill struct {
	Label   string `yaml:"label"`
	Details string `yaml:"details"`
}

// Achievement is an award or recognition.
type Achievement struct {
	Name       string      `yaml:"name"`
	Date       PartialDate `yaml:"date,omitempty"`
	Highlights []string    `yaml:"highlights,omitempty"`
}

// Parse decodes and validates a CV from YAML. Unknown keys are reported as
// errors so typos in the input do not silently drop data.
func Parse(data []byte) (*CV, error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)

	var cv CV
	if err := dec.Decode(&cv); err != nil {
		return nil, fmt.Errorf("error al parsear el YAML: %w", err)
	}
	if err := cv.Validate(); err != nil {
		return nil, err
	}
	cv.normalize()
	return &cv, nil
}

// Validate checks the required fields and section order.
func (cv *CV) Validate() error {
	if strings.TrimSpace(cv.Basics.Name) == "" {
		return fmt.Errorf("el campo obligatorio \"basics.name\" está vacío")
	}
	for _, s := range cv.SectionOrder {
		switch s {
		case SectionEducation, SectionExperience, SectionProjects, SectionSkills, SectionAchievements:
		default:
			return fmt.Errorf("sección desconocida en \"section_order\": %q", s)
		}
	}
	if len(cv.SectionOrder) == 0 {
		cv.SectionOrder = append([]string(nil), DefaultSectionOrder...)
	}
	return nil
}

// normalize fills in derived/optional values after validation.
func (cv *CV) normalize() {
	for i := range cv.Basics.Profiles {
		cv.Basics.Profiles[i].URL = normalizeURL(cv.Basics.Profiles[i].URL)
	}
	for i := range cv.Projects {
		cv.Projects[i].URL = normalizeURL(cv.Projects[i].URL)
	}
	cv.Basics.Summary = strings.TrimSpace(cv.Basics.Summary)
}

// normalizeURL trims the value and adds an https:// scheme when the URL is a
// bare domain such as "miguehm.com".
func normalizeURL(u string) string {
	u = strings.TrimSpace(u)
	if u == "" {
		return ""
	}
	if strings.Contains(u, "://") || strings.HasPrefix(u, "mailto:") {
		return u
	}
	return "https://" + u
}

// HasSection reports whether the given section has any content.
func (cv *CV) HasSection(name string) bool {
	switch name {
	case SectionEducation:
		return len(cv.Education) > 0
	case SectionExperience:
		return len(cv.Experience) > 0
	case SectionProjects:
		return len(cv.Projects) > 0
	case SectionSkills:
		return len(cv.Skills) > 0
	case SectionAchievements:
		return len(cv.Achievements) > 0
	default:
		return false
	}
}
