package utils

import (
	"math"
	"testing"
)

func TestParseVerticalPaddingPx(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{"empty", "", 0},
		{"zero", "0", 0},
		{"none", "none", 0},
		{"single rem", "2rem", 64},
		{"two values", "1rem 2rem", 48},
		{"px single", "10px", 20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseVerticalPaddingPx(tt.in)
			if got != tt.want {
				t.Errorf("ParseVerticalPaddingPx(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestContentAreaHeightPx_A4DefaultPadding(t *testing.T) {
	pageH := int(math.Round(297.0 * CssPxPerMm))
	verticalPad := 64 // 2rem top + bottom
	want := pageH - verticalPad
	if want < 1 {
		want = 1
	}
	got := ContentAreaHeightPx("A4", "")
	if got != want {
		t.Errorf("ContentAreaHeightPx(A4, default pad) = %d, want %d", got, want)
	}
}

func TestContentAreaWidthPx_A4(t *testing.T) {
	want := int(math.Round(210.0 * CssPxPerMm))
	got := ContentAreaWidthPx("A4")
	if got != want {
		t.Errorf("ContentAreaWidthPx(A4) = %d, want %d", got, want)
	}
}

func TestPaperViewportCssPixels_normalizesCase(t *testing.T) {
	w1, h1 := PaperViewportCssPixels("a4")
	w2, h2 := PaperViewportCssPixels("A4")
	if w1 != w2 || h1 != h2 {
		t.Errorf("case normalization failed: a4=%dx%d A4=%dx%d", w1, h1, w2, h2)
	}
}
