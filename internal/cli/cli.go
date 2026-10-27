// Package cli implements the native command-line interface for cvyml: flag
// parsing, file I/O and the process exit code. The actual work lives in the
// platform-independent core package, internal/cvyml.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/miguehm/yml-to-cv/internal/cvyml"
)

// Config holds the parsed command-line options.
type Config struct {
	// In is the input YAML path.
	In string
	// Out is the output PDF path.
	Out string
	// Paper selects the page size: "letter" or "a4".
	Paper string
	// Margin is the page margin in millimetres; zero uses the default.
	Margin float64
	// Lang selects the language of the PDF chrome: "en" or "es".
	Lang string
	// Dump prints the parsed CV as YAML and exits without rendering a PDF.
	Dump bool
}

// Main parses args, runs the command and returns a process exit code.
func Main(args []string, stdout, stderr io.Writer) int {
	cfg, err := Parse(args, stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintln(stderr, "error:", err)
		return 2
	}
	if err := Run(cfg, stdout); err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}
	return 0
}

// Parse builds a Config from command-line arguments. Usage is written to stderr.
func Parse(args []string, stderr io.Writer) (Config, error) {
	var cfg Config

	fs := flag.NewFlagSet("cvyml", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.StringVar(&cfg.In, "in", "cv.yml", "ruta del archivo YAML de entrada")
	fs.StringVar(&cfg.Out, "out", "cv.pdf", "ruta del PDF de salida")
	fs.StringVar(&cfg.Paper, "paper", "letter", "tamaño de papel: letter o a4")
	fs.Float64Var(&cfg.Margin, "margin", 0, "margen de página en mm (0 = predeterminado)")
	fs.StringVar(&cfg.Lang, "lang", "en", "idioma de los títulos y fechas del PDF: en o es")
	fs.BoolVar(&cfg.Dump, "dump", false, "imprime el CV parseado como YAML y sale")
	fs.Usage = func() {
		fmt.Fprintf(stderr, "Uso: cvyml [opciones]\n\nOpciones:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Run executes the command described by cfg, writing progress or dump output to
// stdout.
func Run(cfg Config, stdout io.Writer) error {
	data, err := os.ReadFile(cfg.In)
	if err != nil {
		return fmt.Errorf("no se pudo leer %q: %w", cfg.In, err)
	}

	if cfg.Dump {
		return cvyml.Dump(stdout, data)
	}

	f, err := os.Create(cfg.Out)
	if err != nil {
		return fmt.Errorf("no se pudo crear %q: %w", cfg.Out, err)
	}
	defer f.Close()

	if err := cvyml.Render(f, data, cvyml.Options{Paper: cfg.Paper, Margin: cfg.Margin, Lang: cfg.Lang}); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	fmt.Fprintf(stdout, "PDF generado: %s\n", cfg.Out)
	return nil
}
