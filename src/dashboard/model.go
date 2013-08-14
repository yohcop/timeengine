package dashboard

type Dashboard struct {
	// Name.
	// Immutable, but not stored anywhere anyways.
	name string

	// Definition. Currently just a string.
	// It actually is an array of Graph serialized with JSON.
	G []byte
}

type Graph struct {
	Name    string
	Targets []string
}
