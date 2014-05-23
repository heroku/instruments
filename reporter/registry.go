package reporter

import (
	"sync"

	"github.com/heroku/instruments"
)

// Registry is a registry of all instruments.
type Registry struct {
	instruments map[string]interface{}
	m           sync.Mutex
}

// NewRegistry creates a new Register.
func NewRegistry() *Registry {
	return &Registry{
		instruments: make(map[string]interface{}),
	}
}

// Get returns an instrument from the Registry.
func (r *Registry) Get(name string) interface{} {
	r.m.Lock()
	defer r.m.Unlock()
	return r.instruments[name]
}

// Register registers a new instrument or return the existing one.
func (r *Registry) Register(name string, v interface{}) interface{} {
	r.m.Lock()
	defer r.m.Unlock()
	switch v.(type) {
	case instruments.Discrete, instruments.Sample:
		i, present := r.instruments[name]
		if present {
			return i
		}
		r.instruments[name] = v
		return v
	}
	return nil
}

// Unregister remove from the registry the instrument matching the given name.
func (r *Registry) Unregister(name string) {
	r.m.Lock()
	defer r.m.Unlock()
	delete(r.instruments, name)
}

// Snapshot returns and reset all instruments.
func (r *Registry) Snapshot() map[string]interface{} {
	r.m.Lock()
	defer r.m.Unlock()
	instruments := make(map[string]interface{}, len(r.instruments))
	for k, i := range r.instruments {
		instruments[k] = i
	}
	r.instruments = make(map[string]interface{})
	return instruments
}

// Instruments returns all instruments.
func (r *Registry) Instruments() map[string]interface{} {
	r.m.Lock()
	defer r.m.Unlock()
	instruments := make(map[string]interface{}, len(r.instruments))
	for k, i := range r.instruments {
		instruments[k] = i
	}
	return instruments
}

// Size returns the numbers of instruments in the registry.
func (r *Registry) Size() int {
	r.m.Lock()
	defer r.m.Unlock()
	return len(r.instruments)
}
