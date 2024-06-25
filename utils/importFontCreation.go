package utils

import (
	"fmt"
	"net/url"
	"strings"
)

// ImportFontCreation generates a Google Fonts URL and returns it as an HTML link element string
func ImportFontCreation(fonts []string) string {
	if len(fonts) == 0 {
		return ""
	}

	var fontUrl strings.Builder
	encodedFont := url.QueryEscape(fonts[0])
	fontUrl.WriteString(fmt.Sprintf("https://fonts.googleapis.com/css2?family=%s:wght@100;200;300;400;500;600;700;800;900", encodedFont))

	for _, font := range fonts[1:] {
		encodedFont := url.QueryEscape(font)
		fontUrl.WriteString(fmt.Sprintf("&display=swap&family=%s", encodedFont))
	}

	return fmt.Sprintf(`<link key="font-import" rel="stylesheet" href="%s" />`, fontUrl.String())
}
