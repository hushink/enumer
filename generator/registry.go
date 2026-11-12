// Copyright (c) 2026 The enumer Authors
// SPDX-License-Identifier: MIT

package generator

// DefaultRegistry returns a registry with every built-in feature registered.
func DefaultRegistry() *Registry {
	r := NewRegistry()
	for _, f := range []Feature{
		stringFeature{},
		parseFeature{},
		validFeature{},
		flagsFeature{},
		mustParseFeature{},
		textFeature{},
		jsonFeature{},
		jsonNumberFeature{},
		yamlFeature{},
		sqlFeature{},
		sqlIntFeature{},
		sqlNullFeature{},
		flagFeature{},
		namesFeature{},
		valuesFeature{},
		ptrFeature{},
	} {
		if err := r.Register(f); err != nil {
			panic(err)
		}
	}
	return r
}
