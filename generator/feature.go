// Package generator turns normalized manifests into Go source files and
// custom template outputs.
package generator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hushink/enumer/model"
)

// Feature is one independently switchable code generation capability.
//
// Implementations must be deterministic and must only emit imports they use.
// The renderer sorts features by registry order, deduplicates imports and runs
// gofmt over the assembled file.
//
// Custom features are responsible for avoiding package-level symbol collisions;
// the built-in collision checks only know about built-in features.
type Feature interface {
	// Name is the feature name as used in manifests, for example "sql".
	Name() string
	// Requires lists feature names that must also be enabled.
	Requires() []string
	// Validate checks enum-specific constraints.
	Validate(*model.EnumSpec) error
	// Emit returns the generated source fragment and its imports.
	Emit(*model.EnumSpec) (Fragment, error)
}

// Fragment is the source contributed by one feature.
type Fragment struct {
	Imports []model.Import
	Code    string
}

// Registry holds all known features.
type Registry struct {
	features map[string]Feature
	order    []string
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{features: map[string]Feature{}}
}

// Register adds a feature. Registering the same name twice is an error.
func (r *Registry) Register(f Feature) error {
	if f == nil {
		return fmt.Errorf("generator: nil feature")
	}
	name := f.Name()
	if name == "" {
		return fmt.Errorf("generator: feature with empty name")
	}
	if _, exists := r.features[name]; exists {
		return fmt.Errorf("generator: feature %q is already registered", name)
	}
	r.features[name] = f
	r.order = append(r.order, name)
	return nil
}

// Lookup finds a feature by name.
func (r *Registry) Lookup(name string) (Feature, bool) {
	f, ok := r.features[name]
	return f, ok
}

// Names returns all registered names in alphabetical order.
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.features))
	for name := range r.features {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// resolve validates the configured feature set, expands dependencies and
// returns features in registration order.
func (r *Registry) resolve(e *model.EnumSpec) ([]Feature, error) {
	configured := make(map[string]bool, len(e.Features))
	for _, name := range e.Features {
		if name == "" {
			return nil, fmt.Errorf("feature name must not be empty")
		}
		if _, ok := r.features[name]; !ok {
			return nil, fmt.Errorf("unknown feature %q; available features: %s", name, strings.Join(r.Names(), ", "))
		}
		configured[name] = true
	}

	if configured["json"] && configured["json:number"] {
		return nil, fmt.Errorf("features json and json:number are mutually exclusive")
	}
	if configured["sql"] && configured["sql:int"] {
		return nil, fmt.Errorf("features sql and sql:int are mutually exclusive")
	}
	if configured["json:number"] && e.IsString() {
		return nil, fmt.Errorf("feature json:number requires a numeric enum")
	}
	if configured["sql:int"] && !e.IsString() {
		return nil, fmt.Errorf("feature sql:int requires a string enum")
	}
	if configured["sql:null"] {
		bases := 0
		if configured["sql"] {
			bases++
		}
		if configured["sql:int"] {
			bases++
		}
		if bases != 1 {
			return nil, fmt.Errorf("feature sql:null requires exactly one of sql or sql:int")
		}
	}

	resolved := make(map[string]bool, len(configured))
	var add func(string) error
	add = func(name string) error {
		if resolved[name] {
			return nil
		}
		if isDisabledFeature(name, e.DisabledFeatures) {
			return fmt.Errorf("feature %q is required but disabled", name)
		}
		f, ok := r.features[name]
		if !ok {
			return fmt.Errorf("unknown feature %q; available features: %s", name, strings.Join(r.Names(), ", "))
		}
		resolved[name] = true
		for _, req := range f.Requires() {
			if err := add(req); err != nil {
				return fmt.Errorf("%s requires %s: %w", name, req, err)
			}
		}
		return nil
	}
	for name := range configured {
		if err := add(name); err != nil {
			return nil, err
		}
	}

	features := make([]Feature, 0, len(resolved))
	for _, name := range r.order {
		if !resolved[name] {
			continue
		}
		f := r.features[name]
		if err := f.Validate(e); err != nil {
			return nil, fmt.Errorf("feature %s: %w", name, err)
		}
		features = append(features, f)
	}
	return features, nil
}

func isDisabledFeature(name string, disabled []string) bool {
	for _, d := range disabled {
		if name == d || strings.HasPrefix(name, d+":") {
			return true
		}
	}
	return false
}
