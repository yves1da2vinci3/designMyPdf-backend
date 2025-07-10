package utils

import (
	"strings"
)

// FontCssCreation generates CSS styles for setting font-family
func FontCssCreation(fonts []string) string {
	if len(fonts) == 0 {
		return ""
	}
	return strings.TrimSpace(`
		body {
			font-family: '` + fonts[0] + `', sans-serif;
		}
	`)
}
