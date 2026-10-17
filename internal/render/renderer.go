// Package render turns a parsed CV model into a PDF document.
package render

import (
	"io"
	"time"

	"github.com/miguehm/yml-to-cv/internal/model"
)

// Options configures the PDF output. Zero values fall back to sensible
// Harvard-style defaults.
type Options struct {
	// Paper selects the page size: "letter" (default) or "a4".
	Paper string
	// Margin is the page margin in millimetres.
	Margin float64
	// BodySize is the base font size in points.
	BodySize float64
	// CreationDate is written into the PDF metadata. A zero value uses a fixed
	// date so that repeated runs produce byte-identical output.
	CreationDate time.Time
}

// Renderer renders a CV to PDF.
type Renderer interface {
	Render(w io.Writer, cv *model.CV) error
}

// NewHarvard returns a Renderer that lays out the CV in the classic,
// single-column Harvard style.
func NewHarvard(opts Options) Renderer {
	return &harvardRenderer{opts: opts}
}
