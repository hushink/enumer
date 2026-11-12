// Copyright (c) 2026 The enumer Authors
// SPDX-License-Identifier: MIT

package generator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hushink/enumer/model"
)

type flagsFeature struct{}

func (flagsFeature) Name() string       { return "flags" }
func (flagsFeature) Requires() []string { return []string{"string", "parse", "valid"} }

func (flagsFeature) Validate(e *model.EnumSpec) error {
	if e.Type.Kind != model.KindUnsigned {
		return fmt.Errorf("flags requires an unsigned integer type")
	}
	seen := make(map[string]string, len(e.Values))
	zeroSeen := false
	for i := range e.Values {
		v := &e.Values[i]
		n, err := strconv.ParseUint(v.Number, 10, 64)
		if err != nil {
			return fmt.Errorf("value %q has an invalid number %q", v.Name, v.Number)
		}
		if n == 0 {
			if zeroSeen {
				return fmt.Errorf("flags may declare at most one zero value")
			}
			zeroSeen = true
			continue
		}
		if n&(n-1) != 0 {
			return fmt.Errorf("value %q must be zero or a power of two, got %s", v.Name, v.Number)
		}
		if prev, ok := seen[v.Number]; ok {
			return fmt.Errorf("duplicate flag value %s for %q and %q", v.Number, prev, v.Name)
		}
		seen[v.Number] = v.Name
	}
	return nil
}

func (flagsFeature) Emit(e *model.EnumSpec) (Fragment, error) {
	var nonzero []*model.ValueSpec
	hasZero := false
	zeroText := ""
	for i := range e.Values {
		v := &e.Values[i]
		if v.Number == "0" {
			hasZero = true
			zeroText = v.Text
			continue
		}
		nonzero = append(nonzero, v)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "var _%sFlags = []%s{\n", e.Name, e.Name)
	for _, v := range nonzero {
		fmt.Fprintf(&b, "\t%s,\n", v.ConstName)
	}
	b.WriteString("}\n\n")

	fmt.Fprintf(&b, "var _%sFlagNames = map[%s]string{\n", e.Name, e.Name)
	for _, v := range nonzero {
		fmt.Fprintf(&b, "\t%s: %s,\n", v.ConstName, strconv.Quote(v.Text))
	}
	b.WriteString("}\n\n")

	b.WriteString("// String implements fmt.Stringer.\n")
	fmt.Fprintf(&b, "func (x %s) String() string {\n", e.Name)
	if hasZero {
		fmt.Fprintf(&b, "\tif x == 0 {\n\t\treturn %s\n\t}\n", strconv.Quote(zeroText))
	}
	fmt.Fprintf(&b, "\tparts := make([]string, 0, len(_%sFlags))\n", e.Name)
	fmt.Fprintf(&b, "\tknown := %s(0)\n", e.Name)
	fmt.Fprintf(&b, "\tfor _, flag := range _%sFlags {\n", e.Name)
	b.WriteString("\t\tif x&flag == flag {\n")
	fmt.Fprintf(&b, "\t\t\tparts = append(parts, _%sFlagNames[flag])\n", e.Name)
	b.WriteString("\t\t\tknown |= flag\n\t\t}\n\t}\n")
	fmt.Fprintf(&b, "\tif unknown := x &^ known; unknown != 0 {\n\t\tparts = append(parts, fmt.Sprintf(%q, uint64(unknown)))\n\t}\n", e.Name+"(%d)")
	b.WriteString("\tif len(parts) == 0 {\n")
	fmt.Fprintf(&b, "\t\treturn fmt.Sprintf(%q, uint64(x))\n\t}\n", e.Name+"(%d)")
	b.WriteString("\treturn strings.Join(parts, \"|\")\n}\n\n")

	b.WriteString("// Has reports whether all bits of flag are set in x.\n")
	fmt.Fprintf(&b, "func (x %s) Has(flag %s) bool {\n\treturn x&flag == flag\n}\n\n", e.Name, e.Name)
	b.WriteString("// With returns x with flag set.\n")
	fmt.Fprintf(&b, "func (x %s) With(flag %s) %s {\n\treturn x | flag\n}\n\n", e.Name, e.Name, e.Name)
	b.WriteString("// Clear returns x with flag cleared.\n")
	fmt.Fprintf(&b, "func (x %s) Clear(flag %s) %s {\n\treturn x &^ flag\n}\n\n", e.Name, e.Name, e.Name)
	b.WriteString("// Toggle returns x with flag toggled.\n")
	fmt.Fprintf(&b, "func (x %s) Toggle(flag %s) %s {\n\treturn x ^ flag\n}\n\n", e.Name, e.Name, e.Name)

	b.WriteString("// Flags returns the individual flags set in x.\n")
	fmt.Fprintf(&b, "func (x %s) Flags() []%s {\n", e.Name, e.Name)
	fmt.Fprintf(&b, "\tout := make([]%s, 0, len(_%sFlags))\n", e.Name, e.Name)
	fmt.Fprintf(&b, "\tfor _, flag := range _%sFlags {\n", e.Name)
	b.WriteString("\t\tif x&flag == flag {\n\t\t\tout = append(out, flag)\n\t\t}\n\t}\n\treturn out\n}\n")

	return Fragment{
		Code:    b.String(),
		Imports: []model.Import{{Path: "fmt"}, {Path: "strings"}},
	}, nil
}
