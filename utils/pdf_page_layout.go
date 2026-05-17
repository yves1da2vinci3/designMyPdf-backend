package utils

import (
	"math"
	"regexp"
	"strconv"
	"strings"
)

// CssPxPerMm matches frontend paperDimensions (96 CSS px per inch).
const CssPxPerMm = 96.0 / 25.4

// OrphanThreshold is the fraction of content-area height used for orphan/widow hints.
const OrphanThreshold = 0.2

type paperMm struct {
	widthMm  float64
	heightMm float64
}

var paperSizesMM = map[string]paperMm{
	"A1": {841, 1189},
	"A2": {594, 841},
	"A3": {420, 594},
	"A4": {210, 297},
	"A5": {148, 210},
	"A6": {105, 148},
}

var paddingTokenRe = regexp.MustCompile(`(?i)^([\d.]+|\.\d+)(px|rem|mm|cm|%)?$`)

// PaperViewportCssPixels returns page width/height in CSS pixels (portrait API path).
func PaperViewportCssPixels(formatTitle string) (widthPx, heightPx int) {
	key := strings.ToUpper(strings.TrimSpace(formatTitle))
	if key == "" {
		key = "A4"
	}
	d, ok := paperSizesMM[key]
	if !ok {
		d = paperSizesMM["A4"]
	}
	w := int(math.Round(d.widthMm * CssPxPerMm))
	h := int(math.Round(d.heightMm * CssPxPerMm))
	return w, h
}

// ParseVerticalPaddingPx parses CSS padding shorthand (top + bottom) into pixels.
func ParseVerticalPaddingPx(paddingCss string) int {
	t := strings.TrimSpace(strings.ToLower(paddingCss))
	if t == "" || t == "0" || t == "none" {
		return 0
	}

	toPx := func(token string) float64 {
		m := paddingTokenRe.FindStringSubmatch(strings.TrimSpace(token))
		if m == nil {
			return 0
		}
		n, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			return 0
		}
		unit := m[2]
		if unit == "" {
			unit = "px"
		}
		switch strings.ToLower(unit) {
		case "rem":
			return n * 16
		case "mm":
			return n * CssPxPerMm
		case "cm":
			return n * CssPxPerMm * 10
		case "%":
			return 0
		default:
			return n
		}
	}

	parts := strings.Fields(t)
	switch len(parts) {
	case 1:
		return int(math.Round(toPx(parts[0]) * 2))
	case 2:
		return int(math.Round(toPx(parts[0]) + toPx(parts[1])))
	case 3:
		return int(math.Round(toPx(parts[0]) + toPx(parts[2])))
	default:
		return int(math.Round(toPx(parts[0]) + toPx(parts[2])))
	}
}

// ContentAreaHeightPx is usable content height per page (page height minus vertical padding).
func ContentAreaHeightPx(formatTitle, paddingStored string) int {
	_, h := PaperViewportCssPixels(formatTitle)
	padCss := EffectivePdfContentPadding(paddingStored)
	verticalPad := ParseVerticalPaddingPx(padCss)
	out := h - verticalPad
	if out < 1 {
		return 1
	}
	return out
}

// ContentAreaWidthPx is the fixed .content width in CSS pixels.
func ContentAreaWidthPx(formatTitle string) int {
	w, _ := PaperViewportCssPixels(formatTitle)
	return w
}

// PdfPrintBreakCSS is shared with frontend pdfPageLayout.ts.
const PdfPrintBreakCSS = `
  .pdf-page-break-before {
    break-before: page;
    page-break-before: always;
  }
  .pdf-avoid-break-inside {
    break-inside: avoid;
    page-break-inside: avoid;
  }
  .pdf-keep-together {
    break-inside: avoid;
    page-break-inside: avoid;
  }
`

// PdfExportResetCSS avoids phantom full-viewport pages at print time.
const PdfExportResetCSS = `
  html, body {
    min-height: auto !important;
    height: auto !important;
  }
  .content {
    min-height: auto !important;
    height: auto !important;
  }
  .page-break {
    display: none !important;
  }
  [class~="min-h-screen"],
  [class~="h-screen"] {
    min-height: auto !important;
    height: auto !important;
  }
`
