package instruments

import (
	"fmt"
	"time"
)

// Counter fetches an instrument from the registry or creates a new one.
//
// If another instrument type is already registered with the same
// name/tags, a blank one will be returned and an error
// will be logged to the Errors() channel.
func (r *Registry) Counter(name string, tags []string) *Counter {
	return r.fetchCounter(name, tags, newCounter)
}

func newCounter() interface{} { return NewCounter() }

// Rate fetches an instrument from the registry or creates a new one.
//
// If another instrument type is already registered with the same
// name/tags, a blank one will be returned and an error
// will be logged to the Errors() channel.
func (r *Registry) Rate(name string, tags []string) *Rate {
	return r.fetchRate(name, tags, newRate)
}

func newRate() interface{} { return NewRate() }

// RateScale fetches an instrument from the registry or creates a new one
// with a custom scale.
//
// If another instrument type is already registered with the same
// name/tags, a blank one will be returned and an error
// will be logged to the Errors() channel.
func (r *Registry) RateScale(name string, tags []string, d time.Duration) *Rate {
	factory := func() interface{} { return NewRateScale(d) }
	return r.fetchRate(name, tags, factory)
}

// Derive fetches an instrument from the registry or creates a new one.
//
// If another instrument type is already registered with the same
// name/tags, a blank one will be returned and an error
// will be logged to the Errors() channel.
func (r *Registry) Derive(name string, tags []string, v int64) *Derive {
	factory := func() interface{} { return NewDerive(v) }
	return r.fetchDerive(name, tags, factory)
}

// DeriveScale fetches an instrument from the registry or creates a new one
// with a custom scale.
//
// If another instrument type is already registered with the same
// name/tags, a blank one will be returned and an error
// will be logged to the Errors() channel.
func (r *Registry) DeriveScale(name string, tags []string, v int64, d time.Duration) *Derive {
	factory := func() interface{} { return NewDeriveScale(v, d) }
	return r.fetchDerive(name, tags, factory)
}

// Reservoir fetches an instrument from the registry or creates a new one.
//
// If another instrument type is already registered with the same
// name/tags, a blank one will be returned and an error
// will be logged to the Errors() channel.
func (r *Registry) Reservoir(name string, tags []string, size int64) *Reservoir {
	factory := func() interface{} { return NewReservoir(size) }
	return r.fetchReservoir(name, tags, factory)
}

// Gauge fetches an instrument from the registry or creates a new one.
//
// If another instrument type is already registered with the same
// name/tags, a blank one will be returned and an error
// will be logged to the Errors() channel.
func (r *Registry) Gauge(name string, tags []string, size int64) *Gauge {
	factory := func() interface{} { return NewGauge(size) }
	return r.fetchGauge(name, tags, factory)
}

// Timer fetches an instrument from the registry or creates a new one.
//
// If another instrument type is already registered with the same
// name/tags, a blank one will be returned and an error
// will be logged to the Errors() channel.
func (r *Registry) Timer(name string, tags []string, size int64) *Timer {
	factory := func() interface{} { return NewTimer(size) }
	return r.fetchTimer(name, tags, factory)
}

// --------------------------------------------------------------------

func (r *Registry) fetchCounter(name string, tags []string, factory func() interface{}) *Counter {
	v := r.Fetch(name, tags, factory)
	if i, ok := v.(*Counter); ok {
		return i
	}
	r.handleFetchError("counter", name, tags, v)
	return factory().(*Counter)
}

func (r *Registry) fetchRate(name string, tags []string, factory func() interface{}) *Rate {
	v := r.Fetch(name, tags, factory)
	if i, ok := v.(*Rate); ok {
		return i
	}
	r.handleFetchError("rate", name, tags, v)
	return factory().(*Rate)
}

func (r *Registry) fetchDerive(name string, tags []string, factory func() interface{}) *Derive {
	v := r.Fetch(name, tags, factory)
	if i, ok := v.(*Derive); ok {
		return i
	}
	r.handleFetchError("derive", name, tags, v)
	return factory().(*Derive)
}

func (r *Registry) fetchReservoir(name string, tags []string, factory func() interface{}) *Reservoir {
	v := r.Fetch(name, tags, factory)
	if i, ok := v.(*Reservoir); ok {
		return i
	}
	r.handleFetchError("reservoir", name, tags, v)
	return factory().(*Reservoir)
}

func (r *Registry) fetchGauge(name string, tags []string, factory func() interface{}) *Gauge {
	v := r.Fetch(name, tags, factory)
	if i, ok := v.(*Gauge); ok {
		return i
	}
	r.handleFetchError("gauge", name, tags, v)
	return factory().(*Gauge)
}

func (r *Registry) fetchTimer(name string, tags []string, factory func() interface{}) *Timer {
	v := r.Fetch(name, tags, factory)
	if i, ok := v.(*Timer); ok {
		return i
	}
	r.handleFetchError("timer", name, tags, v)
	return factory().(*Timer)
}

func (r *Registry) handleFetchError(kind, name string, tags []string, inst interface{}) {
	key := MetricID(name, tags)
	err := fmt.Errorf("instruments: expected a %s at '%s', found a stored %T", kind, key, inst)
	r.handleError(err)
}
