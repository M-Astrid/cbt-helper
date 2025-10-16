package renderer

import (
	"time"

	"github.com/M-Astrid/cbt-helper/internal/domain/entity"
)

type SmerRendererI interface {
	RenderPDF(smers []*entity.SMEREntry, startDate time.Time, endDate time.Time) ([]byte, error)
	RenderMessageSingle(smer *entity.SMEREntry) (*string, error)
}

type SmerRenderer struct {
	pdfRenderer      PDFRendererI
	templateRenderer TemplateRendererI
}

func (s SmerRenderer) RenderPDF(smers []*entity.SMEREntry, startDate time.Time, endDate time.Time) ([]byte, error) {
	html, err := s.templateRenderer.Render(map[string]interface{}{
		"entries":   smers,
		"startDate": startDate.Format("2006.01.02"),
		"endDate":   endDate.Format("2006.01.02"),
	}, SmerHTMLTemplatePath)
	if err != nil {
		return nil, err
	}
	return s.pdfRenderer.Render(*html)
}

func (s SmerRenderer) RenderMessageSingle(smer *entity.SMEREntry) (*string, error) {
	txt, err := s.templateRenderer.Render(map[string]interface{}{
		"entry": smer,
	}, SmerSingleMessageTemplatePath)
	if err != nil {
		return nil, err
	}
	return txt, nil
}

func NewSmerRenderer() SmerRendererI {
	return &SmerRenderer{
		pdfRenderer:      NewPDFRenderer(),
		templateRenderer: &TemplateRenderer{},
	}
}
