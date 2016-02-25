package reporter

import (
	"sync"

	"github.com/heroku/instruments"
)

// DefaultRegistry is the default registry.
var DefaultRegistry = NewRegistry()

// Register a new instruments in the default registry.
func Register(name string, v interface{}) interface{} {
	return DefaultRegistry.Register(name, v)
}

// Get returns the named instruments from the default registry.
func Get(name string) interface{} {
	return DefaultRegistry.Get(name)
}

// Unregister remove the named instruments from the default registry.
func Unregister(name string) {
	DefaultRegistry.Unregister(name)
}

// Snapshot returns all instruments and reset the default registry.
func Snapshot() map[string]interface{} {
	return DefaultRegistry.Snapshot()
}

// Registry is a registry of all instruments.
type Registry struct {
	instruments map[string]interface{}
	m           sync.RWMutex
}

// NewRegistry creates a new Register.
func NewRegistry() *Registry {
	return &Registry{
		instruments: make(map[string]interface{}),
	}
}

// Get returns an instrument from the Registry.
func (r *Registry) Get(name string) interface{} {
	r.m.RLock()
	defer r.m.RUnlock()
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
	r.m.RLock()
	defer r.m.RUnlock()
	instruments := make(map[string]interface{}, len(r.instruments))
	for k, i := range r.instruments {
		instruments[k] = i
	}
	return instruments
}

// Size returns the numbers of instruments in the registry.
func (r *Registry) Size() int {
	r.m.RLock()
	defer r.m.RUnlock()
	return len(r.instruments)
}
