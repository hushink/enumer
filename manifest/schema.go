package manifest

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// fileSchema mirrors the YAML manifest. Pointer fields distinguish "absent"
// from an explicit zero value.
type fileSchema struct {
	Version   int            `yaml:"version"`
	Package   string         `yaml:"package"`
	Aggregate bool           `yaml:"aggregate"`
	Output    string         `yaml:"output"`
	Defaults  defaultsSchema `yaml:"defaults"`
	Enums     []enumSchema   `yaml:"enums"`
	Targets   []targetSchema `yaml:"targets"`
}

type defaultsSchema struct {
	Features             []string     `yaml:"features"`
	DisableFeatures      []string     `yaml:"disableFeatures"`
	Naming               namingSchema `yaml:"naming"`
	Parse                parseSchema  `yaml:"parse"`
	Output               outputSchema `yaml:"output"`
	Comments             string       `yaml:"comments"`
	AllowDuplicateValues bool         `yaml:"allowDuplicateValues"`
}

type namingSchema struct {
	Prefix             string   `yaml:"prefix"`
	Identifier         string   `yaml:"identifier"`
	ExtraInitialisms   []string `yaml:"extraInitialisms"`
	DisableInitialisms bool     `yaml:"disableInitialisms"`
}

type parseSchema struct {
	CaseSensitive *bool `yaml:"caseSensitive"`
}

type outputSchema struct {
	Suffix string `yaml:"suffix"`
	Dir    string `yaml:"dir"`
}

type enumSchema struct {
	Name                 string        `yaml:"name"`
	Type                 string        `yaml:"type"`
	Doc                  string        `yaml:"doc"`
	Package              string        `yaml:"package"`
	Output               string        `yaml:"output"`
	Features             []string      `yaml:"features"`
	DisableFeatures      []string      `yaml:"disableFeatures"`
	Start                *numberScalar `yaml:"start"`
	AllowDuplicateValues *bool         `yaml:"allowDuplicateValues"`
	Values               []valueSchema `yaml:"values"`
}

type valueSchema struct {
	Name       string        `yaml:"name"`
	Value      *scalar       `yaml:"value"`
	Number     *numberScalar `yaml:"number"`
	Doc        string        `yaml:"doc"`
	Aliases    []string      `yaml:"aliases"`
	Deprecated deprecated    `yaml:"deprecated"`
	Node       yaml.Node     `yaml:"-"`
}

type targetSchema struct {
	Template string `yaml:"template"`
	Output   string `yaml:"output"`
	PerEnum  bool   `yaml:"perEnum"`
}

var valueKeys = map[string]struct{}{
	"name": {}, "value": {}, "number": {}, "doc": {}, "aliases": {}, "deprecated": {},
}

// scalar captures an arbitrary scalar node without interpreting it. The
// manifest type decides whether a value must be a string or an integer.
type scalar struct {
	Node yaml.Node
}

func (s *scalar) UnmarshalYAML(node *yaml.Node) error {
	node = resolveAlias(node)
	if node == nil || node.Kind != yaml.ScalarNode {
		line := 0
		if node != nil {
			line = node.Line
		}
		return fmt.Errorf("line %d: expected a scalar value", line)
	}
	s.Node = *node
	return nil
}

func (s *scalar) IsString() bool { return s != nil && s.Node.ShortTag() == "!!str" }
func (s *scalar) IsInt() bool    { return s != nil && s.Node.ShortTag() == "!!int" }

// numberScalar captures an integer literal as written.
type numberScalar struct {
	Text string
	Node yaml.Node
}

func (n *numberScalar) UnmarshalYAML(node *yaml.Node) error {
	node = resolveAlias(node)
	if node == nil || node.Kind != yaml.ScalarNode || node.ShortTag() != "!!int" {
		line := 0
		if node != nil {
			line = node.Line
		}
		return fmt.Errorf("line %d: expected an integer", line)
	}
	n.Text = node.Value
	n.Node = *node
	return nil
}

// deprecated accepts true, false, or a non-empty reason string.
type deprecated struct {
	Set    bool
	Reason string
	Node   yaml.Node
}

func (d *deprecated) UnmarshalYAML(node *yaml.Node) error {
	node = resolveAlias(node)
	if node == nil {
		return fmt.Errorf("empty deprecated value")
	}
	d.Node = *node
	switch node.ShortTag() {
	case "!!bool":
		var b bool
		if err := node.Decode(&b); err != nil {
			return err
		}
		if b {
			d.Set = true
			d.Reason = "Do not use."
		}
		return nil
	case "!!str":
		d.Set = true
		d.Reason = strings.TrimSpace(node.Value)
		if d.Reason == "" {
			d.Reason = "Do not use."
		}
		return nil
	default:
		return fmt.Errorf("line %d: deprecated must be a boolean or a string", node.Line)
	}
}

// UnmarshalYAML supports both the scalar shorthand ("- red") and the mapping
// form. The mapping form is validated manually because custom unmarshallers
// bypass yaml.Decoder.KnownFields.
func (v *valueSchema) UnmarshalYAML(node *yaml.Node) error {
	node = resolveAlias(node)
	if node == nil {
		return fmt.Errorf("empty enum value")
	}
	v.Node = *node
	switch node.Kind {
	case yaml.ScalarNode:
		if node.ShortTag() != "!!str" {
			return fmt.Errorf("line %d: an enum value must be a string or a mapping", node.Line)
		}
		v.Name = node.Value
		return nil
	case yaml.MappingNode:
		if err := checkKeys(node, valueKeys); err != nil {
			return err
		}
		type raw valueSchema
		var r raw
		if err := node.Decode(&r); err != nil {
			return err
		}
		*v = valueSchema(r)
		v.Node = *node
		return nil
	default:
		return fmt.Errorf("line %d: an enum value must be a string or a mapping", node.Line)
	}
}

func checkKeys(node *yaml.Node, allowed map[string]struct{}) error {
	if node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		key := node.Content[i]
		if key.Value == "<<" {
			continue
		}
		if _, ok := allowed[key.Value]; !ok {
			return fmt.Errorf("line %d: unknown field %q", key.Line, key.Value)
		}
	}
	return nil
}

func resolveAlias(node *yaml.Node) *yaml.Node {
	for i := 0; i < 100 && node != nil && node.Kind == yaml.AliasNode && node.Alias != nil; i++ {
		node = node.Alias
	}
	return node
}
