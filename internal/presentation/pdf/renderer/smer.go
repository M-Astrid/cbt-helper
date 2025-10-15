package renderer

import (
	"github.com/M-Astrid/cbt-helper/internal/domain/entity"
)

type SmerRendererI interface {
	Render(smers []*entity.SMEREntry) ([]byte, error)
}

type SmerRenderer struct {
	pdfRenderer  PDFRendererI
	htmlRenderer HTMLRendererI
}

func (s SmerRenderer) Render(smers []*entity.SMEREntry) ([]byte, error) {
	html, err := s.htmlRenderer.Render(map[string]interface{}{
		"entries": smers,
	})
	if err != nil {
		return nil, err
	}
	return s.pdfRenderer.Render(*html)
}

func NewSmerRenderer() SmerRendererI {
	return &SmerRenderer{
		pdfRenderer:  NewPDFRenderer(),
		htmlRenderer: NewRenderer(SmerTemplatePath),
	}
}
