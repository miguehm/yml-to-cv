//go:build js && wasm

// Command cvyml-wasm exposes the cvyml renderer to JavaScript under the global
// `cvyml` object. Build it with:
//
//	GOOS=js GOARCH=wasm go build -o cvyml.wasm ./cmd/cvyml-wasm
//
// Load it from a page alongside the matching Go wasm_exec.js, then call:
//
//	const result = cvyml.render(yamlText, {paper: "a4", margin: 15, lang: "es"});
//	if (result.error) throw new Error(result.error);
//	const pdf = result.pdf; // Uint8Array
//
// render and parse never throw: they always return an object with either the
// payload field set or `error` describing the failure.
package main

import (
	"bytes"
	"syscall/js"

	"github.com/miguehm/yml-to-cv/internal/cvyml"
)

func main() {
	api := js.Global().Get("Object").New()
	api.Set("render", js.FuncOf(render))
	api.Set("parse", js.FuncOf(parse))
	js.Global().Set("cvyml", api)

	// Keep the Go runtime alive so the exported functions stay callable.
	select {}
}

// render(data[, options]) -> {pdf, error}
//
// options may be a paper-size string ("letter" or "a4") or an object of the
// form {paper: string, margin: number, lang: string}.
func render(_ js.Value, args []js.Value) any {
	if len(args) == 0 || args[0].Type() != js.TypeString {
		return fail("se esperaba el YAML como primer argumento (string)")
	}

	opts := optionsFrom(args, 1)

	var buf bytes.Buffer
	if err := cvyml.Render(&buf, []byte(args[0].String()), opts); err != nil {
		return fail(err.Error())
	}
	return succeed(bytesField("pdf", buf.Bytes()))
}

// parse(data) -> {yaml, error}
//
// It returns the validated, normalised CV re-encoded as YAML.
func parse(_ js.Value, args []js.Value) any {
	if len(args) == 0 || args[0].Type() != js.TypeString {
		return fail("se esperaba el YAML como primer argumento (string)")
	}

	var buf bytes.Buffer
	if err := cvyml.Dump(&buf, []byte(args[0].String())); err != nil {
		return fail(err.Error())
	}
	return succeed(bytesField("yaml", buf.Bytes()))
}

// optionsFrom extracts render options from an optional argument at index i.
func optionsFrom(args []js.Value, i int) cvyml.Options {
	var opts cvyml.Options
	if len(args) <= i {
		return opts
	}
	switch v := args[i]; v.Type() {
	case js.TypeString:
		opts.Paper = v.String()
	case js.TypeObject:
		if p := v.Get("paper"); p.Type() == js.TypeString {
			opts.Paper = p.String()
		}
		if m := v.Get("margin"); m.Type() == js.TypeNumber {
			opts.Margin = m.Float()
		}
		if l := v.Get("lang"); l.Type() == js.TypeString {
			opts.Lang = l.String()
		}
	}
	return opts
}

// succeed wraps a result object produced by the field builder.
func succeed(fields func(js.Value)) js.Value {
	o := js.Global().Get("Object").New()
	o.Set("error", js.Null())
	fields(o)
	return o
}

// fail returns a result object carrying an error message and a null payload.
func fail(msg string) js.Value {
	o := js.Global().Get("Object").New()
	o.Set("error", msg)
	o.Set("pdf", js.Null())
	o.Set("yaml", js.Null())
	return o
}

// bytesField copies src into a fresh Uint8Array stored under name.
func bytesField(name string, src []byte) func(js.Value) {
	return func(o js.Value) {
		arr := js.Global().Get("Uint8Array").New(len(src))
		js.CopyBytesToJS(arr, src)
		o.Set(name, arr)
	}
}
