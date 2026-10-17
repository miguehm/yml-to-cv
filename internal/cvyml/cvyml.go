// Package cvyml is the platform-independent core of the tool: it turns YAML CV
// data into a PDF.
//
// It performs no file or terminal I/O of its own; callers pass in the input
// bytes and an io.Writer for the output. That keeps the core usable both from
// the native CLI (cmd/cvyml) and from a WebAssembly build (cmd/cvyml-wasm).
package cvyml

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/miguehm/cvyml/internal/model"
	"github.com/miguehm/cvyml/internal/render"
)

// Paper sizes accepted by the renderer.
const (
	PaperLetter = "letter"
	PaperA4     = "a4"
)

// Options configures a render. Zero values fall back to the Harvard defaults.
type Options struct {
	// Paper selects the page size: "letter" (default) or "a4".
	Paper string
	// Margin is the page margin in millimetres; zero uses the default.
	Margin float64
}

// Parse decodes and validates a CV from its YAML representation.
func Parse(data []byte) (*model.CV, error) {
	return model.Parse(data)
}

// Render parses data as YAML and writes the resulting PDF to w.
func Render(w io.Writer, data []byte, opts Options) error {
	paper, err := normalizePaper(opts.Paper)
	if err != nil {
		return err
	}

	cv, err := model.Parse(data)
	if err != nil {
		return err
	}

	r := render.NewHarvard(render.Options{Paper: paper, Margin: opts.Margin})
	if err := r.Render(w, cv); err != nil {
		return fmt.Errorf("error al generar el PDF: %w", err)
	}
	return nil
}

// Dump parses data, then writes the normalised CV back out as indented YAML.
func Dump(w io.Writer, data []byte) error {
	cv, err := model.Parse(data)
	if err != nil {
		return err
	}
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	if err := enc.Encode(cv); err != nil {
		return err
	}
	return errors.Join(enc.Close())
}

// normalizePaper validates and canonicalises the requested paper size.
func normalizePaper(paper string) (string, error) {
	switch p := strings.ToLower(strings.TrimSpace(paper)); p {
	case "", PaperLetter:
		return PaperLetter, nil
	case PaperA4:
		return PaperA4, nil
	default:
		return "", fmt.Errorf("papel no soportado %q (usa %q o %q)", paper, PaperLetter, PaperA4)
	}
}
