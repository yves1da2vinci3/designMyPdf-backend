package utils

import (
	"regexp"
	"strings"
)

// DefaultPdfContentPadding matches the Next.js preview / export default.
const DefaultPdfContentPadding = "2rem"

var pdfContentPaddingValueRe = regexp.MustCompile(`^(?i)(\d+(\.\d+)?|\.\d+)(px|rem|mm|cm|%)$`)

// IsValidPdfContentPadding returns true for empty (use default), "0"/"none", or a safe CSS length.
func IsValidPdfContentPadding(s string) bool {
	t := strings.TrimSpace(s)
	if t == "" {
		return true
	}
	low := strings.ToLower(t)
	if low == "0" || low == "none" {
		return true
	}
	return pdfContentPaddingValueRe.MatchString(strings.TrimSpace(s))
}

// EffectivePdfContentPadding returns a CSS padding value safe to inject into a stylesheet.
func EffectivePdfContentPadding(stored string) string {
	t := strings.TrimSpace(stored)
	if t == "" {
		return DefaultPdfContentPadding
	}
	low := strings.ToLower(t)
	if low == "0" || low == "none" {
		return "0"
	}
	if pdfContentPaddingValueRe.MatchString(t) {
		return t
	}
	return DefaultPdfContentPadding
}
