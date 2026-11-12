package generator

// DefaultRegistry returns a registry with every built-in feature registered.
func DefaultRegistry() *Registry {
	r := NewRegistry()
	for _, f := range []Feature{
		stringFeature{},
		parseFeature{},
		validFeature{},
		mustParseFeature{},
		textFeature{},
		jsonFeature{},
		jsonNumberFeature{},
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
