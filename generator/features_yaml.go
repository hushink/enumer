package generator

import (
	"fmt"
	"strings"

	"github.com/hushink/enumer/model"
)

type yamlFeature struct{}

func (yamlFeature) Name() string                   { return "yaml" }
func (yamlFeature) Requires() []string             { return []string{"parse", "string"} }
func (yamlFeature) Validate(*model.EnumSpec) error { return nil }

func (yamlFeature) Emit(e *model.EnumSpec) (Fragment, error) {
	var b strings.Builder
	b.WriteString("// MarshalYAML implements yaml.Marshaler.\n")
	fmt.Fprintf(&b, "func (x %s) MarshalYAML() (any, error) {\n\treturn x.String(), nil\n}\n\n", e.Name)
	b.WriteString("// UnmarshalYAML implements yaml.Unmarshaler.\n")
	fmt.Fprintf(&b, "func (x *%s) UnmarshalYAML(value *yaml.Node) error {\n", e.Name)
	b.WriteString("\tvar s string\n")
	b.WriteString("\tif err := value.Decode(&s); err != nil {\n\t\treturn err\n\t}\n")
	fmt.Fprintf(&b, "\tv, err := Parse%s(s)\n", e.Name)
	b.WriteString("\tif err != nil {\n\t\treturn err\n\t}\n\t*x = v\n\treturn nil\n}\n")
	return Fragment{
		Code:    b.String(),
		Imports: []model.Import{{Path: "gopkg.in/yaml.v3"}},
	}, nil
}
