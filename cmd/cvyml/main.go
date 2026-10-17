// Command cvyml renders a YAML CV into a Harvard-style PDF.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/miguehm/cvyml/internal/model"
	"github.com/miguehm/cvyml/internal/render"
)

func main() {
	var (
		inFile = flag.String("in", "cv.yml", "ruta del archivo YAML de entrada")
		outPDF = flag.String("out", "cv.pdf", "ruta del PDF de salida")
		paper  = flag.String("paper", "letter", "tamaño de papel: letter o a4")
		margin = flag.Float64("margin", 0, "margen de página en mm (0 = predeterminado)")
		dump   = flag.Bool("dump", false, "imprime el CV parseado como YAML y sale")
	)
	flag.Parse()

	if err := run(*inFile, *outPDF, *paper, *margin, *dump); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(inFile, outPDF, paper string, margin float64, dump bool) error {
	data, err := os.ReadFile(inFile)
	if err != nil {
		return fmt.Errorf("no se pudo leer %q: %w", inFile, err)
	}

	cv, err := model.Parse(data)
	if err != nil {
		return err
	}

	if dump {
		return dumpCV(cv)
	}

	switch strings.ToLower(paper) {
	case "letter", "a4":
	default:
		return fmt.Errorf("papel no soportado %q (usa \"letter\" o \"a4\")", paper)
	}

	f, err := os.Create(outPDF)
	if err != nil {
		return fmt.Errorf("no se pudo crear %q: %w", outPDF, err)
	}
	defer f.Close()

	r := render.NewHarvard(render.Options{Paper: paper, Margin: margin})
	if err := r.Render(f, cv); err != nil {
		return fmt.Errorf("error al generar el PDF: %w", err)
	}
	if err := f.Close(); err != nil {
		return err
	}

	fmt.Printf("PDF generado: %s\n", outPDF)
	return nil
}

func dumpCV(cv *model.CV) error {
	enc := yaml.NewEncoder(os.Stdout)
	enc.SetIndent(2)
	if err := enc.Encode(cv); err != nil {
		return err
	}
	return errors.Join(enc.Close())
}
