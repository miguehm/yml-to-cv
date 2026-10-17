package render

import "testing"

func TestSplitSegments(t *testing.T) {
	segs := splitSegments("Visit [my site](https://miguehm.com) now")
	if len(segs) != 3 {
		t.Fatalf("got %d segments, want 3: %+v", len(segs), segs)
	}
	if segs[0].Text != "Visit " || segs[0].URL != "" {
		t.Errorf("segment 0 = %+v", segs[0])
	}
	if segs[1].Text != "my site" || segs[1].URL != "https://miguehm.com" {
		t.Errorf("segment 1 = %+v", segs[1])
	}
	if segs[2].Text != " now" {
		t.Errorf("segment 2 = %+v", segs[2])
	}
}

func TestSplitSegmentsPlain(t *testing.T) {
	segs := splitSegments("no links here")
	if len(segs) != 1 || segs[0].Text != "no links here" || segs[0].URL != "" {
		t.Fatalf("got %+v", segs)
	}
	if hasURL(segs) {
		t.Fatal("plain text should not carry links")
	}
}
