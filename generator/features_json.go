// Copyright (c) 2026 The enumer Authors
// SPDX-License-Identifier: MIT

package generator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hushink/enumer/model"
)

type jsonFeature struct{}

func (jsonFeature) Name() string                   { return "json" }
func (jsonFeature) Requires() []string             { return []string{"parse", "string"} }
func (jsonFeature) Validate(*model.EnumSpec) error { return nil }

func (jsonFeature) Emit(e *model.EnumSpec) (Fragment, error) {
	var b strings.Builder
	b.WriteString("// MarshalJSON implements json.Marshaler.\n")
	fmt.Fprintf(&b, "func (x %s) MarshalJSON() ([]byte, error) {\n\treturn json.Marshal(x.String())\n}\n\n", e.Name)
	b.WriteString("// UnmarshalJSON implements json.Unmarshaler.\n")
	fmt.Fprintf(&b, "func (x *%s) UnmarshalJSON(data []byte) error {\n", e.Name)
	b.WriteString("\tif string(data) == \"null\" {\n\t\treturn nil\n\t}\n")
	b.WriteString("\tvar s string\n")
	b.WriteString("\tif err := json.Unmarshal(data, &s); err != nil {\n\t\treturn err\n\t}\n")
	fmt.Fprintf(&b, "\tv, err := Parse%s(s)\n", e.Name)
	b.WriteString("\tif err != nil {\n\t\treturn err\n\t}\n\t*x = v\n\treturn nil\n}\n")
	return Fragment{Code: b.String(), Imports: []model.Import{{Path: "encoding/json"}}}, nil
}

type jsonNumberFeature struct{}

func (jsonNumberFeature) Name() string       { return "json:number" }
func (jsonNumberFeature) Requires() []string { return []string{"parse"} }

func (jsonNumberFeature) Validate(e *model.EnumSpec) error {
	if e.IsString() {
		return fmt.Errorf("json:number requires a numeric enum")
	}
	return nil
}

func (jsonNumberFeature) Emit(e *model.EnumSpec) (Fragment, error) {
	var b strings.Builder
	b.WriteString("// MarshalJSON implements json.Marshaler.\n")
	fmt.Fprintf(&b, "func (x %s) MarshalJSON() ([]byte, error) {\n", e.Name)
	if isUnsigned(e) {
		b.WriteString("\treturn []byte(strconv.FormatUint(uint64(x), 10)), nil\n")
	} else {
		b.WriteString("\treturn []byte(strconv.FormatInt(int64(x), 10)), nil\n")
	}
	b.WriteString("}\n\n")

	b.WriteString("// UnmarshalJSON implements json.Unmarshaler.\n")
	fmt.Fprintf(&b, "func (x *%s) UnmarshalJSON(data []byte) error {\n", e.Name)
	b.WriteString("\tif string(data) == \"null\" {\n\t\treturn nil\n\t}\n")
	if isUnsigned(e) {
		fmt.Fprintf(&b, "\tn, err := strconv.ParseUint(string(data), 10, %d)\n", e.Type.Bits)
	} else {
		fmt.Fprintf(&b, "\tn, err := strconv.ParseInt(string(data), 10, %d)\n", e.Type.Bits)
	}
	b.WriteString("\tif err != nil {\n\t\treturn err\n\t}\n")
	fmt.Fprintf(&b, "\tv := %s(n)\n", e.Name)

	if e.HasFeature("flags") {
		b.WriteString("\tif !v.IsValid() {\n")
		fmt.Fprintf(&b, "\t\treturn fmt.Errorf(\"%%w: %%d\", ErrInvalid%s, n)\n", e.Name)
		b.WriteString("\t}\n")
	} else {
		uniq := uniqueValues(e)
		b.WriteString("\tswitch v {\n")
		if len(uniq) > 0 {
			b.WriteString("\tcase ")
			for i, value := range uniq {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(value.ConstName)
			}
			b.WriteString(":\n")
		}
		b.WriteString("\tdefault:\n")
		fmt.Fprintf(&b, "\t\treturn fmt.Errorf(\"%%w: %%d\", ErrInvalid%s, n)\n", e.Name)
		b.WriteString("\t}\n")
	}
	b.WriteString("\t*x = v\n\treturn nil\n}\n")

	return Fragment{
		Code:    b.String(),
		Imports: []model.Import{{Path: "fmt"}, {Path: "strconv"}},
	}, nil
}

type flagFeature struct{}

func (flagFeature) Name() string                   { return "flag" }
func (flagFeature) Requires() []string             { return []string{"parse", "string"} }
func (flagFeature) Validate(*model.EnumSpec) error { return nil }

func (flagFeature) Emit(e *model.EnumSpec) (Fragment, error) {
	var b strings.Builder
	b.WriteString("// Set implements flag.Value.\n")
	fmt.Fprintf(&b, "func (x *%s) Set(s string) error {\n", e.Name)
	fmt.Fprintf(&b, "\tv, err := Parse%s(s)\n", e.Name)
	b.WriteString("\tif err != nil {\n\t\treturn err\n\t}\n\t*x = v\n\treturn nil\n}\n\n")
	b.WriteString("// Type implements pflag.Value.\n")
	fmt.Fprintf(&b, "func (x %s) Type() string {\n\treturn %s\n}\n", e.Name, strconv.Quote(strings.ToLower(e.Name)))
	return Fragment{Code: b.String()}, nil
}
