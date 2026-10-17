package model

import "testing"

func TestParseDate(t *testing.T) {
	cases := []struct {
		in               string
		year, month, day int
		wantErr          bool
	}{
		{"", 0, 0, 0, false},
		{"2020", 2020, 0, 0, false},
		{"2020-08", 2020, 8, 0, false},
		{"2020-08-11", 2020, 8, 11, false},
		{"2020-13", 0, 0, 0, true},
		{"2024-02-31", 0, 0, 0, true},
		{"2020-08-11-05", 0, 0, 0, true},
		{"garbage", 0, 0, 0, true},
	}
	for _, c := range cases {
		got, err := ParseDate(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("ParseDate(%q) expected error", c.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseDate(%q) unexpected error: %v", c.in, err)
			continue
		}
		if got.Year != c.year || got.Month != c.month || got.Day != c.day {
			t.Errorf("ParseDate(%q) = %+v, want {%d %d %d}", c.in, got, c.year, c.month, c.day)
		}
	}
}

func TestFormatPoint(t *testing.T) {
	cases := map[string]string{
		"2020":       "2020",
		"2020-08":    "Aug 2020",
		"2022-08-11": "Aug 11, 2022",
	}
	for in, want := range cases {
		d, err := ParseDate(in)
		if err != nil {
			t.Fatalf("ParseDate(%q): %v", in, err)
		}
		if got := d.FormatPoint(); got != want {
			t.Errorf("FormatPoint(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFormatRange(t *testing.T) {
	must := func(s string) PartialDate {
		d, err := ParseDate(s)
		if err != nil {
			t.Fatalf("ParseDate(%q): %v", s, err)
		}
		return d
	}
	cases := []struct {
		start, end string
		want       string
	}{
		{"2020-08", "2024-05", "Aug 2020 \u2013 May 2024"},
		{"2020-08", "", "Aug 2020 \u2013 Present"},
		{"2025-06", "2025-06", "Jun 2025"},
		{"2020", "2024", "2020 \u2013 2024"},
		{"", "", ""},
	}
	for _, c := range cases {
		if got := FormatRange(must(c.start), must(c.end)); got != c.want {
			t.Errorf("FormatRange(%q,%q) = %q, want %q", c.start, c.end, got, c.want)
		}
	}
}
