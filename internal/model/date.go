package model

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// monthNames holds the English month abbreviations used in the rendered output.
var monthNames = [...]string{
	"Jan", "Feb", "Mar", "Apr", "May", "Jun",
	"Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
}

// PartialDate represents a date with optional precision. A zero value means the
// date is unknown/absent, which is useful for open-ended ranges ("Present").
type PartialDate struct {
	Year  int
	Month int // 1-12, or 0 when unknown
	Day   int // 1-31, or 0 when unknown
}

// ParseDate parses a date in the supported formats: "YYYY", "YYYY-MM" and
// "YYYY-MM-DD". An empty string yields a zero PartialDate.
func ParseDate(s string) (PartialDate, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return PartialDate{}, nil
	}

	parts := strings.Split(s, "-")
	if len(parts) > 3 {
		return PartialDate{}, fmt.Errorf("fecha inválida %q: formato esperado YYYY, YYYY-MM o YYYY-MM-DD", s)
	}

	year, err := strconv.Atoi(parts[0])
	if err != nil || year < 1000 || year > 9999 {
		return PartialDate{}, fmt.Errorf("fecha inválida %q: año no válido", s)
	}

	d := PartialDate{Year: year}
	if len(parts) > 1 {
		month, err := strconv.Atoi(parts[1])
		if err != nil || month < 1 || month > 12 {
			return PartialDate{}, fmt.Errorf("fecha inválida %q: mes no válido", s)
		}
		d.Month = month
		if len(parts) > 2 {
			day, err := strconv.Atoi(parts[2])
			if err != nil || day < 1 || day > 31 {
				return PartialDate{}, fmt.Errorf("fecha inválida %q: día no válido", s)
			}
			// Reject impossible combinations such as 2024-02-31.
			if t := time.Date(d.Year, time.Month(d.Month), day, 0, 0, 0, 0, time.UTC); t.Day() != day {
				return PartialDate{}, fmt.Errorf("fecha inválida %q: día fuera de rango para el mes", s)
			}
			d.Day = day
		}
	}
	return d, nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (d *PartialDate) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return fmt.Errorf("la fecha debe ser una cadena (YYYY, YYYY-MM o YYYY-MM-DD): %w", err)
	}
	pd, err := ParseDate(s)
	if err != nil {
		return err
	}
	*d = pd
	return nil
}

// MarshalYAML implements yaml.Marshaler, rendering the ISO form.
func (d PartialDate) MarshalYAML() (interface{}, error) {
	switch {
	case d.Year == 0:
		return "", nil
	case d.Day != 0:
		return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day), nil
	case d.Month != 0:
		return fmt.Sprintf("%04d-%02d", d.Year, d.Month), nil
	default:
		return strconv.Itoa(d.Year), nil
	}
}

// IsZero reports whether no date information is present.
func (d PartialDate) IsZero() bool {
	return d.Year == 0 && d.Month == 0 && d.Day == 0
}

// monthYear renders "Aug 2020", or just "2020" when the month is unknown.
func (d PartialDate) monthYear() string {
	if d.Year == 0 {
		return ""
	}
	if d.Month == 0 {
		return strconv.Itoa(d.Year)
	}
	return monthNames[d.Month-1] + " " + strconv.Itoa(d.Year)
}

// FormatPoint renders a single date, including the day when known:
// "Aug 11, 2022" or "Aug 2022" or "2020".
func (d PartialDate) FormatPoint() string {
	if d.Year == 0 {
		return ""
	}
	if d.Day != 0 {
		return fmt.Sprintf("%s %d, %d", monthNames[d.Month-1], d.Day, d.Year)
	}
	return d.monthYear()
}

// FormatRange renders a start/end range. A missing end is rendered as
// "Present"; identical start and end collapse to a single date.
func FormatRange(start, end PartialDate) string {
	switch {
	case start.IsZero() && end.IsZero():
		return ""
	case start.IsZero():
		return end.FormatPoint()
	case end.IsZero():
		return start.monthYear() + " \u2013 Present"
	case start == end:
		return start.monthYear()
	case start.Year == end.Year && start.Month == 0 && end.Month == 0:
		return strconv.Itoa(start.Year)
	default:
		return start.monthYear() + " \u2013 " + end.monthYear()
	}
}
