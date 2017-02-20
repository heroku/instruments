package instruments

// Reporter describes the interface every reporter must follow.
// See logreporter package as an example.
type Reporter interface {
	// Prep is called at the beginning of each reporting cycle, which
	// allows reporters to prepare for next data snapshot.
	Prep() error
	// Discrete accepts a numeric value with name and (sorted) tags
	Discrete(name string, tags []string, value float64) error
	// Sample accepts a sampled distribution with name and (sorted) tags
	Sample(name string, tags []string, dist Distribution) error
	// Flush is called at the end of each reporting cycle, which
	// allows reporters to safely buffer data and emit it to
	// backend as a bulk.
	Flush() error
}
