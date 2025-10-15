package renderer

import (
	"fmt"
	"os"

	"github.com/aymerick/raymond"
)

var ErrFileNotFound = fmt.Errorf("file not found")
var ErrTemplateRender = fmt.Errorf("template render error")

type HTMLRendererI interface {
	Render(data map[string]interface{}) (*string, error)
}

type HTMLRenderer struct {
	templatePath string
}

func (s HTMLRenderer) Render(data map[string]interface{}) (*string, error) {
	tpl, err := os.ReadFile(s.templatePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFileNotFound, err)
	}

	result, err := raymond.Render(string(tpl), data)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTemplateRender, err)
	}

	return &result, nil
}

func NewRenderer(path string) HTMLRendererI {
	return &HTMLRenderer{templatePath: path}
}
