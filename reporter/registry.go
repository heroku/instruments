package reporter

import (
	"sort"
	"strings"
	"sync"

	"github.com/bsm/instruments"
)

type MetricID string

func newMetricID(name string, tags []string) MetricID {
	if len(tags) == 0 {
		return MetricID(name)
	}
	sort.Strings(tags)
	return MetricID(name + "|" + strings.Join(tags, ","))
}

// Name returns the metric name
func (s MetricID) Name() string {
	if s == "" {
		return ""
	}
	return strings.SplitN(string(s), "|", 2)[0]
}

// Tags returns the tags
func (s MetricID) Tags() []string {
	if s == "" {
		return nil
	}
	parts := strings.SplitN(string(s), "|", 2)
	if len(parts) != 2 || parts[1] == "" {
		return nil
	}
	return strings.Split(parts[1], ",")
}

// --------------------------------------------------------------------

var instrumentsMapPool sync.Pool

// InstrumentsMap is a named map of instruments returned
// by Registry.Reset()
type InstrumentsMap map[MetricID]interface{}

func newInstrumentsMap(size int) InstrumentsMap {
	if v := instrumentsMapPool.Get(); v != nil {
		return v.(InstrumentsMap)
	}
	return make(InstrumentsMap, size)
}

// Release can be called to release the resources of an
// InstrumentsMap and allow the system to recycle the memory.
// DO NOT use the map after calling this function!
func (m InstrumentsMap) Release() {
	for k := range m {
		delete(m, k)
	}
	instrumentsMapPool.Put(m)
}

// --------------------------------------------------------------------

// DefaultRegistry is the default registry.
var DefaultRegistry = NewRegistry()

// Register a new instrument in the default registry.
func Register(name string, tags []string, v interface{}) {
	DefaultRegistry.Register(name, tags, v)
}

// Get returns the named instruments from the default registry.
func Get(name string, tags []string) interface{} {
	return DefaultRegistry.Get(name, tags)
}

// Unregister remove the named instruments from the default registry.
func Unregister(name string, tags []string) {
	DefaultRegistry.Unregister(name, tags)
}

// Reset resets the default registry and returns all instruments.
func Reset() InstrumentsMap {
	return DefaultRegistry.Reset()
}

// --------------------------------------------------------------------

// Registry is a registry of all instruments.
type Registry struct {
	instruments InstrumentsMap
	m           sync.RWMutex
}

// NewRegistry creates a new Register.
func NewRegistry() *Registry {
	return &Registry{
		instruments: newInstrumentsMap(0),
	}
}

// Get returns an instrument from the Registry.
func (r *Registry) Get(name string, tags []string) interface{} {
	key := newMetricID(name, tags)
	r.m.RLock()
	v := r.instruments[key]
	r.m.RUnlock()
	return v
}

// Register registers a new instrument.
func (r *Registry) Register(name string, tags []string, v interface{}) {
	switch v.(type) {
	case instruments.Discrete, instruments.Sample:
		key := newMetricID(name, tags)
		r.m.Lock()
		r.instruments[key] = v
		r.m.Unlock()
	}
}

// Unregister remove from the registry the instrument matching the given name/tags
func (r *Registry) Unregister(name string, tags []string) {
	key := newMetricID(name, tags)
	r.m.Lock()
	delete(r.instruments, key)
	r.m.Unlock()
}

// Reset resets and returns all instruments.
func (r *Registry) Reset() InstrumentsMap {
	r.m.Lock()
	instruments := r.instruments
	r.instruments = newInstrumentsMap(0)
	r.m.Unlock()
	return instruments
}

// Instruments returns all named instruments without resetting them.
func (r *Registry) Instruments() InstrumentsMap {
	r.m.RLock()
	instruments := newInstrumentsMap(len(r.instruments))
	for k, i := range r.instruments {
		instruments[k] = i
	}
	r.m.RUnlock()
	return instruments
}

// Size returns the numbers of instruments in the registry.
func (r *Registry) Size() int {
	r.m.RLock()
	size := len(r.instruments)
	r.m.RUnlock()
	return size
}
