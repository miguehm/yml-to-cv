package icons

import (
	"encoding/binary"
	"testing"

	assetIcons "github.com/miguehm/yml-to-cv/assets/icons"
)

// TestEveryIconHasGlyph makes sure each registered code point exists in the
// embedded font, so a typo cannot silently render an empty box.
func TestEveryIconHasGlyph(t *testing.T) {
	cmap := parseCmap(t, assetIcons.BootstrapIcons)
	for name, r := range runes {
		if cmap[r] == 0 {
			t.Errorf("icon %q (U+%04X) has no glyph in the embedded font", name, r)
		}
	}
}

func TestIsName(t *testing.T) {
	if !IsName("github") || !IsName(" geo-alt ") {
		t.Error("known icon names should be accepted")
	}
	if IsName("nope") || IsName("") {
		t.Error("unknown icon names should be rejected")
	}
}

func TestForNetwork(t *testing.T) {
	if ForNetwork("GitHub") != GitHub {
		t.Error("GitHub should map to the github icon")
	}
	if ForNetwork("Personal Blog") != Globe {
		t.Error("unknown networks should fall back to the globe")
	}
}

// parseCmap returns the set of code points mapped to a non-zero glyph by the
// TrueType font, supporting cmap formats 4 and 12.
func parseCmap(t *testing.T, font []byte) map[rune]uint16 {
	t.Helper()
	u16 := func(off int) uint16 { return binary.BigEndian.Uint16(font[off:]) }
	u32 := func(off int) uint32 { return binary.BigEndian.Uint32(font[off:]) }

	numTables := int(u16(4))
	cmapOff := 0
	for i := 0; i < numTables; i++ {
		rec := 12 + i*16
		if string(font[rec:rec+4]) == "cmap" {
			cmapOff = int(u32(rec + 8))
		}
	}
	if cmapOff == 0 {
		t.Fatal("embedded font has no cmap table")
	}

	best, bestFormat := 0, 0
	n := int(u16(cmapOff + 2))
	for i := 0; i < n; i++ {
		rec := cmapOff + 4 + i*8
		sub := cmapOff + int(u32(rec+4))
		switch u16(sub) {
		case 12:
			best, bestFormat = sub, 12
		case 4:
			if bestFormat != 12 {
				best, bestFormat = sub, 4
			}
		}
	}
	if best == 0 {
		t.Fatal("embedded font has no supported cmap subtable")
	}

	glyphs := map[rune]uint16{}
	if bestFormat == 12 {
		groups := int(u32(best + 12))
		for i := 0; i < groups; i++ {
			g := best + 16 + i*12
			start, end, gid := rune(u32(g)), rune(u32(g+4)), u32(g+8)
			for c := start; c <= end; c++ {
				glyphs[c] = uint16(gid + uint32(c-start))
			}
		}
		return glyphs
	}

	segCount := int(u16(best+6)) / 2
	endOff := best + 14
	startOff := endOff + segCount*2 + 2
	deltaOff := startOff + segCount*2
	rangeOff := deltaOff + segCount*2
	for seg := 0; seg < segCount; seg++ {
		end := rune(u16(endOff + seg*2))
		start := rune(u16(startOff + seg*2))
		delta := u16(deltaOff + seg*2)
		ro := u16(rangeOff + seg*2)
		for c := start; c <= end; c++ {
			var gid uint16
			if ro == 0 {
				gid = uint16(uint32(c) + uint32(delta))
			} else {
				addr := rangeOff + seg*2 + int(ro) + int(c-start)*2
				if addr+2 > len(font) {
					continue
				}
				gid = u16(addr)
				if gid != 0 {
					gid = uint16(uint32(gid) + uint32(delta))
				}
			}
			if gid != 0 {
				glyphs[c] = gid
			}
		}
	}
	return glyphs
}
