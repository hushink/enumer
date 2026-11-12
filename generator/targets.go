package generator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"github.com/hushink/enumer/internal/diag"
	"github.com/hushink/enumer/internal/naming"
	"github.com/hushink/enumer/model"
)

// TargetData is the template context for a custom target.
type TargetData struct {
	Manifest *model.Manifest
	Enums    []*model.EnumSpec
	Enum     *model.EnumSpec
}

func renderTargets(m *model.Manifest) ([]Output, diag.List) {
	if len(m.Targets) == 0 {
		return nil, nil
	}

	var diags diag.List
	var outputs []Output
	funcs := templateFuncs()
	enums := enumPtrs(m.Enums)

	for _, target := range m.Targets {
		data, err := os.ReadFile(target.Template)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Code:     "template-read",
				Message:  err.Error(),
				File:     target.Source.File,
				Line:     target.Source.Line,
				Column:   target.Source.Column,
			})
			continue
		}
		tmpl, err := template.New(filepath.Base(target.Template)).Funcs(funcs).Parse(string(data))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Code:     "template-parse",
				Message:  err.Error(),
				File:     target.Template,
			})
			continue
		}
		outTmpl, err := template.New("output").Funcs(funcs).Parse(target.Output)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Code:     "template-output",
				Message:  err.Error(),
				File:     target.Source.File,
				Line:     target.Source.Line,
				Column:   target.Source.Column,
			})
			continue
		}

		renderOne := func(enum *model.EnumSpec) {
			ctx := TargetData{Manifest: m, Enums: enums, Enum: enum}

			var content bytes.Buffer
			if err := tmpl.Execute(&content, ctx); err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Code:     "template-exec",
					Message:  err.Error(),
					File:     target.Template,
				})
				return
			}

			var renderedPath bytes.Buffer
			if err := outTmpl.Execute(&renderedPath, ctx); err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Code:     "template-output",
					Message:  err.Error(),
					File:     target.Source.File,
					Line:     target.Source.Line,
					Column:   target.Source.Column,
				})
				return
			}
			outPath := strings.TrimSpace(renderedPath.String())
			if outPath == "" {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Code:     "template-output",
					Message:  "target output path is empty",
					File:     target.Source.File,
					Line:     target.Source.Line,
					Column:   target.Source.Column,
				})
				return
			}
			if filepath.IsAbs(outPath) {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Code:     "invalid-output",
					Message:  "target output path must be relative",
					File:     target.Source.File,
					Line:     target.Source.Line,
					Column:   target.Source.Column,
				})
				return
			}
			abs, err := filepath.Abs(filepath.Join(m.Dir, outPath))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Code:     "invalid-output",
					Message:  err.Error(),
					File:     target.Source.File,
					Line:     target.Source.Line,
					Column:   target.Source.Column,
				})
				return
			}
			outputs = append(outputs, Output{
				Path:    abs,
				Package: m.Package,
				Content: content.Bytes(),
				Kind:    "target",
			})
		}

		if target.PerEnum {
			for _, e := range enums {
				renderOne(e)
			}
		} else {
			renderOne(nil)
		}
	}

	return outputs, diags
}

func enumPtrs(enums []model.EnumSpec) []*model.EnumSpec {
	out := make([]*model.EnumSpec, 0, len(enums))
	for i := range enums {
		out = append(out, &enums[i])
	}
	return out
}

func templateFuncs() template.FuncMap {
	initialisms := naming.InitialismSet(nil, false)
	return template.FuncMap{
		"camel": func(s string) string { return naming.Camel(s, initialisms) },
		"snake": naming.Snake,
		"kebab": naming.Kebab,
		"lower": strings.ToLower,
		"upper": strings.ToUpper,
		"title": naming.Title,
		"quote": quoteValue,
		"join":  func(sep string, ss []string) string { return strings.Join(ss, sep) },
		"json":  func(v any) (string, error) { b, err := json.Marshal(v); return string(b), err },
		"jsonPretty": func(v any) (string, error) {
			b, err := json.MarshalIndent(v, "", "  ")
			return string(b), err
		},
		"default": func(fallback, value any) any {
			if isZeroValue(value) {
				return fallback
			}
			return value
		},
	}
}

func quoteValue(v any) string {
	if s, ok := v.(string); ok {
		return strconv.Quote(s)
	}
	return strconv.Quote(fmt.Sprint(v))
}

func isZeroValue(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array:
		return rv.Len() == 0
	default:
		return rv.IsZero()
	}
}
