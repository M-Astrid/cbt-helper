package renderer

import (
	"bytes"

	wkhtmltopdf "github.com/SebastiaanKlippert/go-wkhtmltopdf"
)

type PDFRendererI interface {
	Render(html string) ([]byte, error)
}

type PDFRenderer struct {
}

func (s PDFRenderer) Render(html string) ([]byte, error) {
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return nil, err
	}

	page := wkhtmltopdf.NewPageReader(bytes.NewReader([]byte(html)))
	pdfg.AddPage(page)

	err = pdfg.Create()
	if err != nil {
		return nil, err
	}

	return pdfg.Bytes(), nil
}

func NewPDFRenderer() PDFRendererI {
	return &PDFRenderer{}
}
