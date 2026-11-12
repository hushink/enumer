// Copyright (c) 2026 The enumer Authors
// SPDX-License-Identifier: MIT

package generator

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"

	"github.com/hushink/enumer/internal/diag"
	"github.com/hushink/enumer/model"
)

type typeDecl struct {
	underlying string
	file       string
	line       int
}

type typeDeclError struct {
	code string
	msg  string
}

func (e *typeDeclError) Error() string { return e.msg }

// resolveTypeMode turns auto into emit or existing by scanning the target
// package. The scan is syntax-only; enumer never type-checks packages.
func resolveTypeMode(e *model.EnumSpec, outputPath string) (model.TypeMode, diag.List) {
	if outputPath == "" {
		return model.TypeModeEmit, diag.List{errDiag("invalid-output", "enum output path is empty", e.Source.File, e.Source.Line, e.Source.Column, e.Name)}
	}
	dir := filepath.Dir(outputPath)
	decl, err := findTypeDecl(dir, e.Package, e.Name, outputPath)
	if err != nil {
		code := "type-detect"
		var te *typeDeclError
		if errors.As(err, &te) {
			code = te.code
		}
		return model.TypeModeEmit, diag.List{errDiag(code, err.Error(), e.Source.File, e.Source.Line, e.Source.Column, e.Name)}
	}

	mismatch := func(decl *typeDecl) diag.List {
		return diag.List{errDiag(
			"type-mismatch",
			fmt.Sprintf("type %s is declared with underlying type %s at %s:%d, want %s", e.Name, decl.underlying, decl.file, decl.line, e.Type.Name),
			e.Source.File, e.Source.Line, e.Source.Column, e.Name,
		)}
	}
	missing := diag.List{errDiag(
		"missing-type",
		fmt.Sprintf("typeMode existing requires type %s to be declared in %s", e.Name, dir),
		e.Source.File, e.Source.Line, e.Source.Column, e.Name,
	)}

	switch e.TypeMode {
	case model.TypeModeExisting:
		if decl == nil {
			return model.TypeModeExisting, missing
		}
		if decl.underlying != e.Type.Name {
			return model.TypeModeExisting, mismatch(decl)
		}
		return model.TypeModeExisting, nil
	case model.TypeModeEmit:
		if decl != nil {
			return model.TypeModeEmit, diag.List{errDiag(
				"type-exists",
				fmt.Sprintf("type %s is already declared at %s:%d; use typeMode auto or existing", e.Name, decl.file, decl.line),
				e.Source.File, e.Source.Line, e.Source.Column, e.Name,
			)}
		}
		return model.TypeModeEmit, nil
	default:
		if decl == nil {
			return model.TypeModeEmit, nil
		}
		if decl.underlying != e.Type.Name {
			return model.TypeModeExisting, mismatch(decl)
		}
		return model.TypeModeExisting, nil
	}
}

func findTypeDecl(dir, pkg, name, exclude string) (*typeDecl, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, &typeDeclError{code: "type-detect", msg: fmt.Sprintf("cannot scan %s: %v", dir, err)}
	}

	exclude = filepath.Clean(exclude)
	fset := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		if filepath.Clean(path) == exclude {
			continue
		}
		file, _ := parser.ParseFile(fset, path, nil, 0)
		if file == nil || file.Name == nil {
			continue
		}
		if pkg != "" && file.Name.Name != pkg {
			continue
		}
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || ts.Name == nil || ts.Name.Name != name {
					continue
				}
				line := fset.Position(ts.Pos()).Line
				if ts.TypeParams != nil {
					return nil, &typeDeclError{code: "type-generic", msg: fmt.Sprintf("type %s at %s:%d is generic; enums cannot use generic types", name, path, line)}
				}
				if ts.Assign.IsValid() {
					return nil, &typeDeclError{code: "type-alias", msg: fmt.Sprintf("type %s at %s:%d is an alias; declare a distinct type", name, path, line)}
				}
				return &typeDecl{underlying: types.ExprString(ts.Type), file: path, line: line}, nil
			}
		}
	}
	return nil, nil
}
