package renderer

import (
	"fmt"
	"os"

	"github.com/aymerick/raymond"
)

var ErrFileNotFound = fmt.Errorf("file not found")
var ErrTemplateRender = fmt.Errorf("template render error")

type TemplateRendererI interface {
	Render(data map[string]interface{}, templatePath string) (*string, error)
}

type TemplateRenderer struct {
}

func (s TemplateRenderer) Render(data map[string]interface{}, templatePath string) (*string, error) {
	tpl, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFileNotFound, err)
	}

	result, err := raymond.Render(string(tpl), data)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTemplateRender, err)
	}

	return &result, nil
}

func NewRenderer() TemplateRendererI {
	return &TemplateRenderer{}
}
