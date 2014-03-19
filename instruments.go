// Instruments allows you to collects metrics over discrete time intervals.
//
// Collected metrics will only reflect observations from last time window only,
// rather than including observations from prior windows, contrary to EWMA based metrics.
//
// 	timer := instruments.NewTimer(-1)
//
//	registry := instruments.NewRegistry()
//	registry.Register("processing-time", timer)
//
//	go reporter.Log("process", registry, time.Minute)
//
//	timer.Time(func() {
//	  ...
//	})
//
// You can create custom instruments or compose new instruments form the built-in
// instruments as long as they implements the Sample or Discrete interfaces.
//
// Registry also use theses interfaces, creating a custom Reporter should be trivial, for example:
//
// 	for k, m := range r.Instruments() {
// 	 	switch i := m.(type) {
// 	 	case instruments.Discrete:
// 	 	 	s := i.Snapshot()
// 	 	 	report(k, s)
// 	 	case instruments.Sample:
// 	 	 	s := instruments.Quantile(i.Snapshot(), 0.95)
// 	 	 	report(k, s)
// 	 	}
//	}
//
package instruments

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

const rateScale = 1e-9

// Represents a single value instrument.
type Discrete interface {
	Snapshot() int64
}

// Represents a sample instrument.
type Sample interface {
	Snapshot() []int64
}

// Track the rate of values per second.
type Rate struct {
	count int64
	time  int64
}

// Create a new rate instrument.
func NewRate() *Rate {
	return &Rate{
		time: time.Now().UnixNano(),
	}
}

// Update rate value.
func (r *Rate) Update(v int64) {
	atomic.AddInt64(&r.count, v)
}

// Return the number of values per second since the last snapshot,
// and reset the count to zero.
func (r *Rate) Snapshot() int64 {
	now := time.Now().UnixNano()
	t := atomic.SwapInt64(&r.time, now)
	c := atomic.SwapInt64(&r.count, 0)
	s := float64(c) / rateScale / float64(now-t)
	return Ceil(s)
}

// Track the rate of deltas per seconds.
type Derive struct {
	rate  *Rate
	value int64
}

// Create a new derive instruments.
func NewDerive(v int64) *Derive {
	return &Derive{
		value: v,
		rate:  NewRate(),
	}
}

// Update rate value based on the stored previous value.
func (d *Derive) Update(v int64) {
	p := atomic.SwapInt64(&d.value, v)
	d.rate.Update(v - p)
}

// Return the number of values per seconds since the last snapshot,
// and reset the count to zero.
func (d *Derive) Snapshot() int64 {
	return d.rate.Snapshot()
}

// Track a sample of values.
type Reservoir struct {
	size   int64
	values []int64
	m      sync.Mutex
}

const defaultReservoirSize = 1028

// Create a new reservoir of the given size.
// If size is negative, it will create a sample of DefaultReservoirSize size.
func NewReservoir(size int64) *Reservoir {
	if size <= 0 {
		size = defaultReservoirSize
	}
	return &Reservoir{
		values: make([]int64, size),
	}
}

// Fill the sample randomly with given value,
// for reference, see: http://en.wikipedia.org/wiki/Reservoir_sampling
func (r *Reservoir) Update(v int64) {
	r.m.Lock()
	defer r.m.Unlock()
	s := atomic.AddInt64(&r.size, 1)
	if int(s) <= len(r.values) {
		// Not full
		r.values[s-1] = v
	} else {
		// Full
		l := rand.Int63n(s)
		if int(l) < len(r.values) {
			r.values[l] = v
		}
	}
}

// Return sample as a sorted array.
func (r *Reservoir) Snapshot() []int64 {
	r.m.Lock()
	defer r.m.Unlock()
	s := atomic.SwapInt64(&r.size, 0)
	v := make([]int64, min(int(s), len(r.values)))
	copy(v, r.values)
	r.values = make([]int64, cap(r.values))
	sorted(v)
	return v
}

// Tracks a value.
type Gauge struct {
	value int64
}

// Creates a new Gauge with the given value.
func NewGauge(v int64) *Gauge {
	return &Gauge{
		value: v,
	}
}

// Update the current stored value.
func (g *Gauge) Update(v int64) {
	atomic.StoreInt64(&g.value, v)
}

// Return the current value.
func (g *Gauge) Snapshot() int64 {
	return atomic.LoadInt64(&g.value)
}

// Tracks durations.
type Timer struct {
	r *Reservoir
}

// Create a new Timer with the given sample size.
func NewTimer(size int64) *Timer {
	return &Timer{
		r: NewReservoir(size),
	}
}

// Add duration to the sample in ms.
func (t *Timer) Update(d time.Duration) {
	v := Floor(d.Seconds() * 1000)
	t.r.Update(v)
}

// Returns durations sample as a sorted array.
func (t *Timer) Snapshot() []int64 {
	return t.r.Snapshot()
}

// Records given function execution time.
func (t *Timer) Time(f func()) {
	ts := time.Now()
	f()
	t.Update(time.Since(ts))
}
