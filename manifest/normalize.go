// Copyright (c) 2026 The enumer Authors
// SPDX-License-Identifier: MIT

package manifest

import (
	"fmt"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hushink/enumer/internal/diag"
	"github.com/hushink/enumer/internal/naming"
	"github.com/hushink/enumer/model"
)

type resolvedDefaults struct {
	features             []string
	disabled             []string
	prefixMode           string
	prefixValue          string
	identifier           model.IdentifierStyle
	initialisms          map[string]bool
	caseSensitive        bool
	suffix               string
	dir                  string
	comments             model.CommentsMode
	allowDuplicateValues bool
	typeMode             model.TypeMode
}

func resolveDefaults(ds *defaultsSchema, display string) (resolvedDefaults, diag.List) {
	d := resolvedDefaults{
		features:      []string{"string", "parse", "valid", "text", "names", "values"},
		prefixMode:    "type",
		identifier:    model.IdentifierCamel,
		initialisms:   naming.InitialismSet(nil, false),
		caseSensitive: true,
		suffix:        "_enum.go",
		comments:      model.CommentsAuto,
		typeMode:      model.TypeModeAuto,
	}
	var diags diag.List

	if ds.Features != nil {
		d.features = append([]string(nil), (*ds.Features)...)
	}
	if ds.DisableFeatures != nil {
		d.disabled = append([]string(nil), (*ds.DisableFeatures)...)
	}

	var prefix string
	if ds.Naming != nil && ds.Naming.Prefix != nil {
		prefix = *ds.Naming.Prefix
	}
	switch {
	case prefix == "" || prefix == "type":
		d.prefixMode = "type"
	case prefix == "none":
		d.prefixMode = "none"
	case strings.HasPrefix(prefix, "literal:"):
		d.prefixMode = "literal"
		d.prefixValue = strings.TrimPrefix(prefix, "literal:")
		if d.prefixValue == "" {
			diags = append(diags, errDiag("invalid-naming", "literal prefix must not be empty", display, 1, 0, "", ""))
		}
	default:
		diags = append(diags, errDiag("invalid-naming", fmt.Sprintf("unknown prefix %q; use type, none or literal:<prefix>", prefix), display, 1, 0, "", ""))
	}

	var identifier string
	var extraInitialisms []string
	var disableInitialisms bool
	if ds.Naming != nil {
		if ds.Naming.Identifier != nil {
			identifier = *ds.Naming.Identifier
		}
		if ds.Naming.ExtraInitialisms != nil {
			extraInitialisms = *ds.Naming.ExtraInitialisms
		}
		if ds.Naming.DisableInitialisms != nil {
			disableInitialisms = *ds.Naming.DisableInitialisms
		}
	}
	switch identifier {
	case "", "camel":
		d.identifier = model.IdentifierCamel
	case "raw":
		d.identifier = model.IdentifierRaw
	default:
		diags = append(diags, errDiag("invalid-naming", fmt.Sprintf("unknown identifier style %q; use camel or raw", identifier), display, 1, 0, "", ""))
	}
	d.initialisms = naming.InitialismSet(extraInitialisms, disableInitialisms)

	if ds.Parse != nil && ds.Parse.CaseSensitive != nil {
		d.caseSensitive = *ds.Parse.CaseSensitive
	}

	if ds.Output != nil {
		if ds.Output.Suffix != nil && *ds.Output.Suffix != "" {
			d.suffix = *ds.Output.Suffix
			if !strings.HasSuffix(d.suffix, ".go") {
				d.suffix += ".go"
			}
		}
		if ds.Output.Dir != nil {
			switch {
			case *ds.Output.Dir == "":
				d.dir = ""
			case filepath.IsAbs(*ds.Output.Dir):
				diags = append(diags, errDiag("invalid-output", "output.dir must be relative", display, 1, 0, "", ""))
			default:
				d.dir = filepath.Clean(*ds.Output.Dir)
			}
		}
	}

	if ds.Comments != nil {
		switch *ds.Comments {
		case "", "auto":
			d.comments = model.CommentsAuto
		case "none":
			d.comments = model.CommentsNone
		default:
			diags = append(diags, errDiag("invalid-comments", fmt.Sprintf("unknown comments mode %q; use auto or none", *ds.Comments), display, 1, 0, "", ""))
		}
	}

	if ds.AllowDuplicateValues != nil {
		d.allowDuplicateValues = *ds.AllowDuplicateValues
	}

	if ds.TypeMode != nil {
		mode, err := parseTypeMode(*ds.TypeMode)
		if err != nil {
			diags = append(diags, errDiag("invalid-type-mode", err.Error(), display, 1, 0, "", ""))
		} else {
			d.typeMode = mode
		}
	}

	return d, diags
}

func normalizeEnum(es *enumSchema, src model.Source, d resolvedDefaults, m *model.Manifest, display string) (model.EnumSpec, diag.List) {
	var diags diag.List
	doc := es.Doc
	if es.Comment != "" {
		doc = es.Comment
	}
	e := model.EnumSpec{
		Name:                 es.Name,
		Doc:                  doc,
		Source:               src,
		CaseSensitive:        d.caseSensitive,
		Comments:             d.comments,
		AllowDuplicateValues: d.allowDuplicateValues,
		TypeMode:             d.typeMode,
	}

	if es.Name == "" {
		diags = append(diags, errDiag("missing-field", "enum name is required", display, src.Line, src.Column, "", ""))
		return e, diags
	}
	if !token.IsIdentifier(es.Name) || token.IsKeyword(es.Name) {
		diags = append(diags, errDiag("invalid-identifier", fmt.Sprintf("%q is not a valid Go type name", es.Name), display, src.Line, src.Column, es.Name, ""))
	}
	if es.Comments != "" {
		switch es.Comments {
		case "auto":
			e.Comments = model.CommentsAuto
		case "none":
			e.Comments = model.CommentsNone
		default:
			diags = append(diags, errDiag("invalid-comments", fmt.Sprintf("unknown comments mode %q; use auto or none", es.Comments), display, src.Line, src.Column, es.Name, ""))
		}
	}

	typeName := es.Type
	if typeName == "" {
		typeName = "string"
	}
	typ, ok := model.LookupType(typeName)
	if !ok {
		diags = append(diags, errDiag("unsupported-type", fmt.Sprintf("unsupported enum type %q; supported types: %s", typeName, strings.Join(model.SupportedTypes(), ", ")), display, src.Line, src.Column, es.Name, ""))
		typ = model.Type{Name: "string", Kind: model.KindString}
	}
	e.Type = typ

	if es.Package != "" {
		if !token.IsIdentifier(es.Package) || token.IsKeyword(es.Package) {
			diags = append(diags, errDiag("invalid-package", fmt.Sprintf("%q is not a valid Go package name", es.Package), display, src.Line, src.Column, es.Name, ""))
		}
		e.Package = es.Package
	} else {
		e.Package = m.Package
	}

	if es.TypeMode != "" {
		mode, err := parseTypeMode(es.TypeMode)
		if err != nil {
			diags = append(diags, errDiag("invalid-type-mode", err.Error(), display, src.Line, src.Column, es.Name, ""))
		} else {
			e.TypeMode = mode
		}
	}

	features, disabled, fdiags := mergeFeatures(d.features, es.Features, d.disabled, es.DisableFeatures, display, src, es.Name)
	diags = append(diags, fdiags...)
	e.Features = features
	e.DisabledFeatures = disabled

	if es.AllowDuplicateValues != nil {
		e.AllowDuplicateValues = *es.AllowDuplicateValues
	}

	e.Naming = model.Naming{Identifier: d.identifier, Initialisms: d.initialisms}
	switch d.prefixMode {
	case "none":
		e.Naming.Prefix = ""
	case "literal":
		e.Naming.Prefix = d.prefixValue
	default:
		e.Naming.Prefix = es.Name
	}

	if typ.IsString() {
		if es.Start != nil {
			diags = append(diags, errDiag("unexpected-start", "start is only valid for numeric enums", display, es.Start.Node.Line, es.Start.Node.Column, es.Name, ""))
		}
		e.Start = ""
	} else {
		startText := "0"
		if es.Start != nil {
			startText = es.Start.Text
		}
		start, err := parseNumber(typ, startText)
		if err != nil {
			diags = append(diags, errDiag("invalid-number", fmt.Sprintf("invalid start value %q: %v", startText, err), display, src.Line, src.Column, es.Name, ""))
			e.Start = "0"
		} else {
			e.Start = start.canonical
		}
	}

	if !m.Aggregate {
		if es.Output != "" {
			p, err := resolveOutputPath(es.Output, m.Dir, "")
			if err != nil {
				diags = append(diags, errDiag("invalid-output", err.Error(), display, src.Line, src.Column, es.Name, ""))
			} else {
				e.Output = p
			}
		} else {
			base := naming.Snake(es.Name)
			if base == "" {
				base = "enum"
			}
			p, err := resolveOutputPath(base+d.suffix, m.Dir, d.dir)
			if err != nil {
				diags = append(diags, errDiag("invalid-output", err.Error(), display, src.Line, src.Column, es.Name, ""))
			} else {
				e.Output = p
			}
		}
	} else if es.Output != "" {
		diags = append(diags, errDiag("unexpected-output", "enum output cannot be set when aggregate is true", display, src.Line, src.Column, es.Name, ""))
	}
	if !m.Aggregate && m.Package != "" && es.Package != "" && es.Package != m.Package && es.Output == "" {
		diags = append(diags, errDiag("invalid-output", "enum package override requires an explicit output path", display, src.Line, src.Column, es.Name, ""))
	}

	values, vdiags := normalizeValues(es.Values, e, display)
	diags = append(diags, vdiags...)
	e.Values = values

	if len(es.Values) == 0 {
		diags = append(diags, errDiag("missing-values", "enum must declare at least one value", display, src.Line, src.Column, es.Name, ""))
	}

	return e, diags
}

func mergeFeatures(base, extra, disabledBase, disabledExtra []string, display string, src model.Source, enum string) ([]string, []string, diag.List) {
	var diags diag.List
	all := dedupeStrings(append(append([]string(nil), base...), extra...))
	disabled := dedupeStrings(append(append([]string(nil), disabledBase...), disabledExtra...))

	used := make(map[string]bool, len(disabled))
	out := make([]string, 0, len(all))
	for _, feature := range all {
		matched := ""
		for _, d := range disabled {
			if feature == d || strings.HasPrefix(feature, d+":") {
				matched = d
				break
			}
		}
		if matched != "" {
			used[matched] = true
			continue
		}
		out = append(out, feature)
	}
	baseDisabled := make(map[string]bool, len(disabledBase))
	for _, d := range disabledBase {
		baseDisabled[d] = true
	}
	for _, d := range disabledExtra {
		if !used[d] && !baseDisabled[d] {
			diags = append(diags, warnDiag("unused-feature-disable", fmt.Sprintf("feature %q is disabled but was not enabled", d), display, src.Line, src.Column, enum, ""))
		}
	}
	return out, disabled, diags
}

func normalizeValues(vs []valueSchema, e model.EnumSpec, display string) ([]model.ValueSpec, diag.List) {
	var diags diag.List
	values := make([]model.ValueSpec, 0, len(vs))
	constNames := make(map[string]model.Source, len(vs))
	usedTexts := make(map[string]model.Source, len(vs))
	usedNumbers := make(map[string]model.Source, len(vs))
	usedSQLNumbers := make(map[string]model.Source, len(vs))

	var cur number
	if e.Type.IsInteger() {
		start, err := parseNumber(e.Type, e.Start)
		if err != nil {
			start = number{signed: e.Type.Kind == model.KindSigned, canonical: "0", valid: true}
		}
		cur = start
	}

	for i := range vs {
		v := &vs[i]
		src := model.Source{File: display, Line: v.Node.Line, Column: v.Node.Column}
		if v.Name == "" {
			diags = append(diags, errDiag("missing-field", "enum value name is required", display, src.Line, src.Column, e.Name, ""))
			continue
		}

		constName := enumConstName(e, v.Name)
		if !token.IsIdentifier(constName) || token.IsKeyword(constName) {
			diags = append(diags, errDiag("invalid-identifier", fmt.Sprintf("%q is not a valid Go constant name", constName), display, src.Line, src.Column, e.Name, v.Name))
		} else if prev, ok := constNames[constName]; ok {
			diags = append(diags, errDiag("duplicate-name", fmt.Sprintf("constant %q is already declared at %s", constName, prev.String()), display, src.Line, src.Column, e.Name, v.Name))
		} else {
			constNames[constName] = src
		}

		valueDoc := v.Doc
		if v.Comment != "" {
			valueDoc = v.Comment
		}
		out := model.ValueSpec{
			Name:      v.Name,
			ConstName: constName,
			Doc:       valueDoc,
			Source:    src,
		}
		if v.Deprecated.Set {
			out.Deprecated = v.Deprecated.Reason
		}

		if e.Type.IsString() {
			text := v.Name
			if v.Value != nil {
				if !v.Value.IsString() {
					diags = append(diags, errDiag("invalid-value", fmt.Sprintf("string enum value %q must use a string value", v.Name), display, src.Line, src.Column, e.Name, v.Name))
				} else {
					text = v.Value.Node.Value
				}
			}
			out.Text = text

			if v.Number != nil {
				n, err := parseSQLNumber(v.Number.Text)
				if err != nil {
					diags = append(diags, errDiag("invalid-number", fmt.Sprintf("invalid number for value %q: %v", v.Name, err), display, v.Number.Node.Line, v.Number.Node.Column, e.Name, v.Name))
				} else {
					out.SQLNumber = n
					out.HasSQLNumber = true
					if e.HasFeature("sql:int") {
						if prev, ok := usedSQLNumbers[n]; ok {
							if !e.AllowDuplicateValues {
								diags = append(diags, errDiag("duplicate-value", fmt.Sprintf("duplicate sql number %s for value %q (first declared at %s); enable allowDuplicateValues to permit aliases", n, v.Name, prev.String()), display, v.Number.Node.Line, v.Number.Node.Column, e.Name, v.Name))
							}
						} else {
							usedSQLNumbers[n] = src
						}
					}
				}
				if !e.HasFeature("sql:int") {
					diags = append(diags, warnDiag("unused-number", fmt.Sprintf("value %q declares number but feature sql:int is not enabled", v.Name), display, v.Number.Node.Line, v.Number.Node.Column, e.Name, v.Name))
				}
			}
		} else {
			if v.Value != nil {
				if !v.Value.IsInt() {
					diags = append(diags, errDiag("invalid-value", fmt.Sprintf("numeric enum value %q must use an integer value", v.Name), display, src.Line, src.Column, e.Name, v.Name))
				} else {
					n, err := parseNumber(e.Type, v.Value.Node.Value)
					if err != nil {
						diags = append(diags, errDiag("invalid-number", fmt.Sprintf("invalid number for value %q: %v", v.Name, err), display, src.Line, src.Column, e.Name, v.Name))
					} else {
						cur = n
					}
				}
			} else if !cur.valid {
				diags = append(diags, errDiag("auto-increment-overflow", fmt.Sprintf("auto-increment overflows %s at value %q", e.Type.Name, v.Name), display, src.Line, src.Column, e.Name, v.Name))
			}
			if v.Number != nil {
				diags = append(diags, errDiag("unexpected-number", "number is only valid for string enums; use value for numeric enums", display, v.Number.Node.Line, v.Number.Node.Column, e.Name, v.Name))
			}
			out.Number = cur.canonical
			out.HasNumber = true
			out.Text = v.Name
			if prev, ok := usedNumbers[out.Number]; ok {
				if !e.AllowDuplicateValues {
					diags = append(diags, errDiag("duplicate-value", fmt.Sprintf("duplicate numeric value %s in enum %s (first declared at %s); enable allowDuplicateValues to permit aliases", out.Number, e.Name, prev.String()), display, src.Line, src.Column, e.Name, v.Name))
				}
			} else {
				usedNumbers[out.Number] = src
			}
			cur = cur.incremented(e.Type)
		}

		textKey := foldKey(out.Text, e.CaseSensitive)
		if prev, ok := usedTexts[textKey]; ok {
			diags = append(diags, errDiag("duplicate-text", fmt.Sprintf("duplicate text %q in enum %s (first declared at %s)", out.Text, e.Name, prev.String()), display, src.Line, src.Column, e.Name, v.Name))
		} else {
			usedTexts[textKey] = src
		}

		for _, alias := range v.Aliases {
			if alias == "" {
				diags = append(diags, errDiag("invalid-alias", fmt.Sprintf("value %q declares an empty alias", v.Name), display, src.Line, src.Column, e.Name, v.Name))
				continue
			}
			key := foldKey(alias, e.CaseSensitive)
			if prev, ok := usedTexts[key]; ok {
				diags = append(diags, errDiag("alias-collision", fmt.Sprintf("alias %q conflicts with an existing value or alias (first declared at %s)", alias, prev.String()), display, src.Line, src.Column, e.Name, v.Name))
				continue
			}
			usedTexts[key] = src
			out.Aliases = append(out.Aliases, alias)
		}

		values = append(values, out)
	}

	if e.Type.IsString() && e.HasFeature("sql:int") {
		for _, v := range values {
			if !v.HasSQLNumber {
				diags = append(diags, errDiag("missing-number", fmt.Sprintf("value %q must declare number when sql:int is enabled", v.Name), display, v.Source.Line, v.Source.Column, e.Name, v.Name))
			}
		}
	}

	return values, diags
}

func parseTypeMode(s string) (model.TypeMode, error) {
	switch s {
	case "", "auto":
		return model.TypeModeAuto, nil
	case "emit":
		return model.TypeModeEmit, nil
	case "existing":
		return model.TypeModeExisting, nil
	default:
		return model.TypeModeAuto, fmt.Errorf("unknown typeMode %q; use auto, emit or existing", s)
	}
}

func enumConstName(e model.EnumSpec, name string) string {
	if e.Naming.Identifier == model.IdentifierRaw {
		return e.Naming.Prefix + name
	}
	return e.Naming.Prefix + naming.Camel(name, e.Naming.Initialisms)
}

func foldKey(s string, caseSensitive bool) string {
	if caseSensitive {
		return s
	}
	return strings.ToLower(s)
}

func dedupeStrings(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

const (
	maxInt8   = 1<<7 - 1
	minInt8   = -1 << 7
	maxInt16  = 1<<15 - 1
	minInt16  = -1 << 15
	maxInt32  = 1<<31 - 1
	minInt32  = -1 << 31
	maxInt64  = 1<<63 - 1
	minInt64  = -1 << 63
	maxUint8  = 1<<8 - 1
	maxUint16 = 1<<16 - 1
	maxUint32 = 1<<32 - 1
	maxUint64 = 1<<64 - 1
)

type number struct {
	signed    bool
	i         int64
	u         uint64
	canonical string
	valid     bool
}

func parseNumber(t model.Type, raw string) (number, error) {
	if t.Kind == model.KindSigned {
		v, err := parseSignedLiteral(raw)
		if err != nil {
			return number{}, err
		}
		if !signedFits(t, v) {
			return number{}, fmt.Errorf("%d does not fit in %s", v, t.Name)
		}
		return number{signed: true, i: v, canonical: strconv.FormatInt(v, 10), valid: true}, nil
	}
	v, err := parseUnsignedLiteral(raw)
	if err != nil {
		return number{}, err
	}
	if !unsignedFits(t, v) {
		return number{}, fmt.Errorf("%d does not fit in %s", v, t.Name)
	}
	return number{signed: false, u: v, canonical: strconv.FormatUint(v, 10), valid: true}, nil
}

func (n number) incremented(t model.Type) number {
	if n.signed {
		if n.i == maxInt64 {
			return number{signed: true, valid: false}
		}
		v := n.i + 1
		return number{signed: true, i: v, canonical: strconv.FormatInt(v, 10), valid: signedFits(t, v)}
	}
	if n.u == maxUint64 {
		return number{signed: false, valid: false}
	}
	v := n.u + 1
	return number{signed: false, u: v, canonical: strconv.FormatUint(v, 10), valid: unsignedFits(t, v)}
}

func signedFits(t model.Type, v int64) bool {
	switch t.Bits {
	case 8:
		return v >= minInt8 && v <= maxInt8
	case 16:
		return v >= minInt16 && v <= maxInt16
	case 32:
		return v >= minInt32 && v <= maxInt32
	default:
		return true
	}
}

func unsignedFits(t model.Type, v uint64) bool {
	switch t.Bits {
	case 8:
		return v <= maxUint8
	case 16:
		return v <= maxUint16
	case 32:
		return v <= maxUint32
	default:
		return true
	}
}

func parseSQLNumber(raw string) (string, error) {
	v, err := parseSignedLiteral(raw)
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(v, 10), nil
}

func parseSignedLiteral(raw string) (int64, error) {
	body, base, negative, err := splitIntegerLiteral(raw)
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseUint(body, base, 64)
	if err != nil {
		return 0, err
	}
	if negative {
		if v > 1<<63 {
			return 0, fmt.Errorf("%q is out of range", raw)
		}
		if v == 1<<63 {
			return minInt64, nil
		}
		return -int64(v), nil
	}
	if v > maxInt64 {
		return 0, fmt.Errorf("%q is out of range", raw)
	}
	return int64(v), nil
}

func parseUnsignedLiteral(raw string) (uint64, error) {
	body, base, negative, err := splitIntegerLiteral(raw)
	if err != nil {
		return 0, err
	}
	if negative {
		return 0, fmt.Errorf("%q must not be negative", raw)
	}
	return strconv.ParseUint(body, base, 64)
}

func splitIntegerLiteral(raw string) (body string, base int, negative bool, err error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", 0, false, fmt.Errorf("empty integer")
	}
	s = strings.TrimPrefix(s, "+")
	if strings.HasPrefix(s, "-") {
		negative = true
		s = s[1:]
	}
	lower := strings.ToLower(s)
	base = 10
	switch {
	case strings.HasPrefix(lower, "0x"):
		base = 16
		s = s[2:]
	case strings.HasPrefix(lower, "0o"):
		base = 8
		s = s[2:]
	case strings.HasPrefix(lower, "0b"):
		base = 2
		s = s[2:]
	}
	s = strings.ReplaceAll(s, "_", "")
	if s == "" {
		return "", 0, false, fmt.Errorf("invalid integer %q", raw)
	}
	return s, base, negative, nil
}

func errDiag(code, msg, file string, line, col int, enum, value string) diag.Diagnostic {
	return diag.Diagnostic{
		Severity: diag.Error,
		Code:     code,
		Message:  msg,
		File:     file,
		Line:     line,
		Column:   col,
		Enum:     enum,
		Value:    value,
	}
}

func warnDiag(code, msg, file string, line, col int, enum, value string) diag.Diagnostic {
	return diag.Diagnostic{
		Severity: diag.Warning,
		Code:     code,
		Message:  msg,
		File:     file,
		Line:     line,
		Column:   col,
		Enum:     enum,
		Value:    value,
	}
}
