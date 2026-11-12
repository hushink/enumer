package generator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hushink/enumer/model"
)

type stringFeature struct{}

func (stringFeature) Name() string                   { return "string" }
func (stringFeature) Requires() []string             { return nil }
func (stringFeature) Validate(*model.EnumSpec) error { return nil }

func (stringFeature) Emit(e *model.EnumSpec) (Fragment, error) {
	var b strings.Builder
	b.WriteString("// String implements fmt.Stringer.\n")
	fmt.Fprintf(&b, "func (x %s) String() string {\n", e.Name)
	if e.IsString() {
		b.WriteString("\treturn string(x)\n")
	} else {
		b.WriteString("\tswitch x {\n")
		for _, v := range uniqueValues(e) {
			fmt.Fprintf(&b, "\tcase %s:\n\t\treturn %s\n", v.ConstName, strconv.Quote(v.Text))
		}
		b.WriteString("\t}\n")
		if isUnsigned(e) {
			fmt.Fprintf(&b, "\treturn fmt.Sprintf(%q, uint64(x))\n", e.Name+"(%d)")
		} else {
			fmt.Fprintf(&b, "\treturn fmt.Sprintf(%q, int64(x))\n", e.Name+"(%d)")
		}
	}
	b.WriteString("}\n")

	frag := Fragment{Code: b.String()}
	if !e.IsString() {
		frag.Imports = []model.Import{{Path: "fmt"}}
	}
	return frag, nil
}

type parseFeature struct{}

func (parseFeature) Name() string                   { return "parse" }
func (parseFeature) Requires() []string             { return nil }
func (parseFeature) Validate(*model.EnumSpec) error { return nil }

func (parseFeature) Emit(e *model.EnumSpec) (Fragment, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "// ErrInvalid%s is returned when a value is not valid for %s.\n", e.Name, e.Name)
	fmt.Fprintf(&b, "var ErrInvalid%s = errors.New(%q)\n\n", e.Name, "invalid "+e.Name)

	fmt.Fprintf(&b, "var _%sValue = map[string]%s{\n", e.Name, e.Name)
	seen := map[string]bool{}
	for i := range e.Values {
		v := &e.Values[i]
		texts := make([]string, 0, 1+len(v.Aliases))
		texts = append(texts, v.Text)
		texts = append(texts, v.Aliases...)
		for _, text := range texts {
			key := mapKey(e, text)
			if seen[key] {
				continue
			}
			seen[key] = true
			fmt.Fprintf(&b, "\t%s: %s,\n", strconv.Quote(key), v.ConstName)
		}
	}
	b.WriteString("}\n\n")

	fmt.Fprintf(&b, "// Parse%s converts a string to %s.\n", e.Name, e.Name)
	fmt.Fprintf(&b, "func Parse%s(name string) (%s, error) {\n", e.Name, e.Name)
	lookup := "name"
	if !e.CaseSensitive {
		lookup = "strings.ToLower(name)"
	}
	fmt.Fprintf(&b, "\tif x, ok := _%sValue[%s]; ok {\n\t\treturn x, nil\n\t}\n", e.Name, lookup)
	fmt.Fprintf(&b, "\treturn %s, fmt.Errorf(\"%%w: %%q\", ErrInvalid%s, name)\n", zeroLiteral(e), e.Name)
	b.WriteString("}\n")

	frag := Fragment{
		Code:    b.String(),
		Imports: []model.Import{{Path: "errors"}, {Path: "fmt"}},
	}
	if !e.CaseSensitive {
		frag.Imports = append(frag.Imports, model.Import{Path: "strings"})
	}
	return frag, nil
}

type validFeature struct{}

func (validFeature) Name() string                   { return "valid" }
func (validFeature) Requires() []string             { return []string{"parse"} }
func (validFeature) Validate(*model.EnumSpec) error { return nil }

func (validFeature) Emit(e *model.EnumSpec) (Fragment, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "// IsValid reports whether x is a valid %s.\n", e.Name)
	fmt.Fprintf(&b, "func (x %s) IsValid() bool {\n", e.Name)
	if e.IsString() {
		lookup := "string(x)"
		if !e.CaseSensitive {
			lookup = "strings.ToLower(string(x))"
		}
		fmt.Fprintf(&b, "\t_, ok := _%sValue[%s]\n\treturn ok\n", e.Name, lookup)
	} else {
		uniq := uniqueValues(e)
		if len(uniq) == 0 {
			b.WriteString("\treturn false\n")
		} else {
			b.WriteString("\tswitch x {\n\tcase ")
			for i, v := range uniq {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(v.ConstName)
			}
			b.WriteString(":\n\t\treturn true\n\t}\n\treturn false\n")
		}
	}
	b.WriteString("}\n")

	frag := Fragment{Code: b.String()}
	if e.IsString() && !e.CaseSensitive {
		frag.Imports = []model.Import{{Path: "strings"}}
	}
	return frag, nil
}

type mustParseFeature struct{}

func (mustParseFeature) Name() string                   { return "mustparse" }
func (mustParseFeature) Requires() []string             { return []string{"parse"} }
func (mustParseFeature) Validate(*model.EnumSpec) error { return nil }

func (mustParseFeature) Emit(e *model.EnumSpec) (Fragment, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "// MustParse%s is like Parse%s but panics on error.\n", e.Name, e.Name)
	fmt.Fprintf(&b, "func MustParse%s(name string) %s {\n", e.Name, e.Name)
	fmt.Fprintf(&b, "\tx, err := Parse%s(name)\n", e.Name)
	b.WriteString("\tif err != nil {\n\t\tpanic(err)\n\t}\n\treturn x\n}\n")
	return Fragment{Code: b.String()}, nil
}

type textFeature struct{}

func (textFeature) Name() string                   { return "text" }
func (textFeature) Requires() []string             { return []string{"parse", "string"} }
func (textFeature) Validate(*model.EnumSpec) error { return nil }

func (textFeature) Emit(e *model.EnumSpec) (Fragment, error) {
	var b strings.Builder
	b.WriteString("// MarshalText implements encoding.TextMarshaler.\n")
	fmt.Fprintf(&b, "func (x %s) MarshalText() ([]byte, error) {\n\treturn []byte(x.String()), nil\n}\n\n", e.Name)
	b.WriteString("// UnmarshalText implements encoding.TextUnmarshaler.\n")
	fmt.Fprintf(&b, "func (x *%s) UnmarshalText(text []byte) error {\n", e.Name)
	fmt.Fprintf(&b, "\tv, err := Parse%s(string(text))\n", e.Name)
	b.WriteString("\tif err != nil {\n\t\treturn err\n\t}\n\t*x = v\n\treturn nil\n}\n")
	return Fragment{Code: b.String()}, nil
}
