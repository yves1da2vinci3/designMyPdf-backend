package utils

import (
	"github.com/aymerick/raymond"
)

func RenderTemplate(htmlContent string, data interface{}) (string, error) {
	// Parse and execute the template with data using Handlebars
	rendered, err := raymond.Render(string(htmlContent), data)
	if err != nil {
		return "", err
	}

	return rendered, nil
}
