package instruments

// Reporter describes the interface every reporter must follow.
// See logreporter package as an example.
type Reporter interface {
	// Discrete accepts a Discrete instrument with name and (sorted) tags
	Discrete(name string, tags []string, inst Discrete) error
	// Sample accepts a Sample instrument with name and (sorted) tags
	Sample(name string, tags []string, inst Sample) error
	// Flush is called at the end of each reporting cycle, which
	// allows reporters to safely buffer data and emit it to
	// backend as a bulk.
	Flush() error
}
