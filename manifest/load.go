// Package manifest discovers, decodes, validates and normalizes enumer.yaml
// manifests into the normalized model.
package manifest

import (
	"bytes"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/hushink/enumer/internal/diag"
	"github.com/hushink/enumer/model"
	"gopkg.in/yaml.v3"
)

// Load reads, decodes, validates and normalizes one manifest.
//
// It returns a normalized manifest even when diagnostics contain errors, so
// callers can report every problem in a single pass. Generation must only
// proceed when List.HasErrors reports false.
func Load(path string) (*model.Manifest, diag.List) {
	display := filepath.Clean(path)
	data, err := os.ReadFile(display)
	if err != nil {
		return nil, diag.List{errDiag("manifest-read", err.Error(), display, 0, 0, "", "")}
	}
	abs, err := filepath.Abs(display)
	if err != nil {
		return nil, diag.List{errDiag("manifest-read", err.Error(), display, 0, 0, "", "")}
	}
	dir := filepath.Dir(abs)

	var schema fileSchema
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&schema); err != nil {
		return nil, decodeDiagnostics(display, err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, diag.List{errDiag("manifest-decode", "manifest contains more than one YAML document", display, 0, 0, "", "")}
		}
		return nil, decodeDiagnostics(display, err)
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, decodeDiagnostics(display, err)
	}
	enumPositions := enumNodes(&doc)

	version := schema.Version
	if version == 0 {
		version = 1
	}
	if version != 1 {
		return nil, diag.List{errDiag("unsupported-version", fmt.Sprintf("unsupported manifest version %d; supported version is 1", version), display, 1, 0, "", "")}
	}

	defaults, diags := resolveDefaults(&schema.Defaults, display)

	pkg, pkgDiags := resolvePackage(schema.Package, dir, display)
	if schema.Package == "" && pkg == "" && allEnumsHavePackage(schema.Enums) {
		pkg = schema.Enums[0].Package
		pkgDiags = nil
	}
	diags = append(diags, pkgDiags...)

	m := &model.Manifest{
		SchemaVersion: model.SchemaVersion,
		Source:        model.Source{File: display, Line: 1, Column: 1},
		Dir:           dir,
		Package:       pkg,
		Aggregate:     schema.Aggregate,
	}

	if !schema.Aggregate && schema.Output != "" {
		diags = append(diags, errDiag("unexpected-output", "output is only valid when aggregate is true", display, 1, 0, "", ""))
	}

	if len(schema.Enums) == 0 {
		diags = append(diags, errDiag("no-enums", "manifest declares no enums", display, 1, 0, "", ""))
	} else {
		for i := range schema.Enums {
			src := model.Source{File: display, Line: 1, Column: 1}
			if i < len(enumPositions) && enumPositions[i] != nil {
				src.Line = enumPositions[i].Line
				src.Column = enumPositions[i].Column
			}
			e, ediags := normalizeEnum(&schema.Enums[i], src, defaults, m, display)
			diags = append(diags, ediags...)
			m.Enums = append(m.Enums, e)
		}
	}

	for i := range schema.Targets {
		t, tdiags := normalizeTarget(schema.Targets[i], m, display)
		diags = append(diags, tdiags...)
		m.Targets = append(m.Targets, t)
	}

	if schema.Aggregate {
		out, odiags := resolveAggregateOutput(schema.Output, defaults, m, display)
		diags = append(diags, odiags...)
		m.Output = out
	}

	diags.Sort()
	return m, diags
}

var lineRe = regexp.MustCompile(`line (\d+):`)

func decodeDiagnostics(file string, err error) diag.List {
	var typeErr *yaml.TypeError
	if errors.As(err, &typeErr) {
		list := make(diag.List, 0, len(typeErr.Errors))
		for _, msg := range typeErr.Errors {
			list = append(list, yamlDiag(file, msg))
		}
		return list
	}
	return diag.List{yamlDiag(file, err.Error())}
}

func yamlDiag(file, msg string) diag.Diagnostic {
	d := diag.Diagnostic{
		Severity: diag.Error,
		Code:     "manifest-decode",
		Message:  msg,
		File:     file,
	}
	if m := lineRe.FindStringSubmatch(msg); len(m) == 2 {
		d.Line, _ = strconv.Atoi(m[1])
	}
	if strings.Contains(msg, "unknown field") || strings.Contains(msg, "not found in type") {
		d.Code = "unknown-field"
	}
	return d
}

func enumNodes(doc *yaml.Node) []*yaml.Node {
	root := doc
	if root != nil && root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		root = root.Content[0]
	}
	root = resolveAlias(root)
	if root == nil || root.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(root.Content); i += 2 {
		if root.Content[i].Value != "enums" {
			continue
		}
		seq := resolveAlias(root.Content[i+1])
		if seq == nil || seq.Kind != yaml.SequenceNode {
			return nil
		}
		nodes := make([]*yaml.Node, 0, len(seq.Content))
		for _, n := range seq.Content {
			nodes = append(nodes, resolveAlias(n))
		}
		return nodes
	}
	return nil
}

func resolvePackage(explicit, dir, display string) (string, diag.List) {
	inferred, inferErr := inferPackage(dir)
	if explicit != "" {
		if !token.IsIdentifier(explicit) || token.IsKeyword(explicit) {
			return explicit, diag.List{errDiag("invalid-package", fmt.Sprintf("%q is not a valid Go package name", explicit), display, 1, 0, "", "")}
		}
		if inferErr == nil && inferred != "" && inferred != explicit {
			return explicit, diag.List{errDiag("package-mismatch", fmt.Sprintf("package %q does not match existing package %q in %s", explicit, inferred, dir), display, 1, 0, "", "")}
		}
		return explicit, nil
	}
	if inferErr != nil || inferred == "" {
		return "", diag.List{errDiag("missing-package", "cannot infer package name; set package: in the manifest", display, 1, 0, "", "")}
	}
	return inferred, nil
}

func allEnumsHavePackage(enums []enumSchema) bool {
	if len(enums) == 0 {
		return false
	}
	for i := range enums {
		if enums[i].Package == "" {
			return false
		}
	}
	return true
}

func inferPackage(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	nonTest := map[string]bool{}
	testPkgs := map[string]bool{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		name, err := packageClause(filepath.Join(dir, entry.Name()))
		if err != nil || name == "" {
			continue
		}
		if strings.HasSuffix(entry.Name(), "_test.go") {
			testPkgs[name] = true
		} else {
			nonTest[name] = true
		}
	}
	if len(nonTest) == 1 {
		for name := range nonTest {
			return name, nil
		}
	}
	if len(nonTest) > 1 {
		return "", fmt.Errorf("multiple packages found in %s", dir)
	}
	if len(testPkgs) == 1 {
		for name := range testPkgs {
			return strings.TrimSuffix(name, "_test"), nil
		}
	}
	if len(testPkgs) > 1 {
		return "", fmt.Errorf("multiple test packages found in %s", dir)
	}
	return "", fmt.Errorf("no Go files found in %s", dir)
}

func packageClause(path string) (string, error) {
	f, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.PackageClauseOnly)
	if err != nil {
		return "", err
	}
	return f.Name.Name, nil
}

func normalizeTarget(ts targetSchema, m *model.Manifest, display string) (model.TargetSpec, diag.List) {
	var diags diag.List
	t := model.TargetSpec{
		Template: ts.Template,
		Output:   ts.Output,
		PerEnum:  ts.PerEnum,
		Source:   model.Source{File: display, Line: 1, Column: 1},
	}
	if ts.Template == "" {
		diags = append(diags, errDiag("missing-field", "target template is required", display, 1, 0, "", ""))
	} else {
		tpl := ts.Template
		if !filepath.IsAbs(tpl) {
			tpl = filepath.Join(m.Dir, tpl)
		}
		t.Template = filepath.Clean(tpl)
	}
	if ts.Output == "" {
		diags = append(diags, errDiag("missing-field", "target output is required", display, 1, 0, "", ""))
	}
	return t, diags
}

func resolveAggregateOutput(explicit string, d resolvedDefaults, m *model.Manifest, display string) (string, diag.List) {
	var diags diag.List
	pkg := ""
	for i := range m.Enums {
		e := &m.Enums[i]
		if i == 0 {
			pkg = e.Package
		} else if e.Package != pkg {
			diags = append(diags, errDiag("aggregate-package-mismatch", fmt.Sprintf("enum %s uses package %s but the aggregate file uses %s", e.Name, e.Package, pkg), e.Source.File, e.Source.Line, e.Source.Column, e.Name, ""))
		}
		if e.Output != "" {
			diags = append(diags, errDiag("unexpected-output", "enum output cannot be set when aggregate is true", e.Source.File, e.Source.Line, e.Source.Column, e.Name, ""))
		}
	}

	if explicit != "" {
		path, err := resolveOutputPath(explicit, m.Dir, "")
		if err != nil {
			diags = append(diags, errDiag("invalid-output", err.Error(), display, 1, 0, "", ""))
			return "", diags
		}
		return path, diags
	}

	stem := strings.TrimSuffix(filepath.Base(m.Source.File), filepath.Ext(m.Source.File))
	path, err := resolveOutputPath(stem+d.suffix, m.Dir, d.dir)
	if err != nil {
		diags = append(diags, errDiag("invalid-output", err.Error(), display, 1, 0, "", ""))
		return "", diags
	}
	return path, diags
}

func resolveOutputPath(output, manifestDir, defaultDir string) (string, error) {
	if output == "" {
		return "", fmt.Errorf("output path is empty")
	}
	if filepath.IsAbs(output) {
		return "", fmt.Errorf("output path %q must be relative", output)
	}
	p := output
	if defaultDir != "" {
		if filepath.IsAbs(defaultDir) {
			return "", fmt.Errorf("output dir %q must be relative", defaultDir)
		}
		p = filepath.Join(defaultDir, p)
	}
	if !strings.HasSuffix(p, ".go") {
		return "", fmt.Errorf("output path %q must end with .go", p)
	}
	abs, err := filepath.Abs(filepath.Join(manifestDir, p))
	if err != nil {
		return "", err
	}
	return abs, nil
}
