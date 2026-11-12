// Copyright (c) 2026 The enumer Authors
// SPDX-License-Identifier: MIT

package generator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hushink/enumer/model"
)

type namesFeature struct{}

func (namesFeature) Name() string                   { return "names" }
func (namesFeature) Requires() []string             { return nil }
func (namesFeature) Validate(*model.EnumSpec) error { return nil }

func (namesFeature) Emit(e *model.EnumSpec) (Fragment, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "// %sNames returns the names of all %s values in declaration order.\n", e.Name, e.Name)
	fmt.Fprintf(&b, "func %sNames() []string {\n", e.Name)
	b.WriteString("\treturn []string{\n")
	for i := range e.Values {
		fmt.Fprintf(&b, "\t\t%s,\n", strconv.Quote(e.Values[i].Text))
	}
	b.WriteString("\t}\n}\n")
	return Fragment{Code: b.String()}, nil
}

type valuesFeature struct{}

func (valuesFeature) Name() string                   { return "values" }
func (valuesFeature) Requires() []string             { return nil }
func (valuesFeature) Validate(*model.EnumSpec) error { return nil }

func (valuesFeature) Emit(e *model.EnumSpec) (Fragment, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "// %sValues returns all %s values in declaration order.\n", e.Name, e.Name)
	fmt.Fprintf(&b, "func %sValues() []%s {\n", e.Name, e.Name)
	b.WriteString("\treturn []" + e.Name + "{\n")
	for i := range e.Values {
		fmt.Fprintf(&b, "\t\t%s,\n", e.Values[i].ConstName)
	}
	b.WriteString("\t}\n}\n")
	return Fragment{Code: b.String()}, nil
}

type ptrFeature struct{}

func (ptrFeature) Name() string                   { return "ptr" }
func (ptrFeature) Requires() []string             { return nil }
func (ptrFeature) Validate(*model.EnumSpec) error { return nil }

func (ptrFeature) Emit(e *model.EnumSpec) (Fragment, error) {
	var b strings.Builder
	b.WriteString("// Ptr returns a pointer to x.\n")
	fmt.Fprintf(&b, "func (x %s) Ptr() *%s {\n\treturn &x\n}\n", e.Name, e.Name)
	return Fragment{Code: b.String()}, nil
}
