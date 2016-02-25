// Package instruments allows you to collects metrics over discrete time intervals.
//
// Collected metrics will only reflect observations from last time window only,
// rather than including observations from prior windows, contrary to EWMA based metrics.
//
// 	timer := instruments.NewTimer(-1)
//
//	registry := reporter.NewRegistry()
//	registry.Register("processing-time", timer)
//
//	go reporter.Log("process", registry, time.Minute)
//
//	timer.Time(func() {
//	  ...
//	})
//
// Instruments support two types of instruments:
// Discrete instruments return a single value, and Sample instruments a sorted array of values.
//
// Theses base instruments are available:
//
// - Counter: holds a counter that can be incremented or decremented.
//
// - Rate: tracks the rate of values per seconds.
//
// - Reservoir: randomly samples values.
//
// - Derive: tracks the rate of values based on the delta with previous value.
//
// - Gauge: tracks last value.
//
// - Timer: tracks durations.
//
// You can create custom instruments or compose new instruments form the built-in
// instruments as long as they implements the Sample or Discrete interfaces.
//
// Registry enforce the Discrete and Sample interfaces,
// creating a custom Reporter should be trivial, for example:
//
// 	for k, m := range registry.Instruments() {
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

// Discrete represents a single value instrument.
type Discrete interface {
	Snapshot() int64
}

// Sample represents a sample instrument.
type Sample interface {
	Snapshot() []int64
}

// Scale returns a conversion factor from one unit to another.
func Scale(o, d time.Duration) float64 {
	return float64(o) / float64(d)
}

// Counter holds a counter that can be incremented or decremented.
type Counter struct {
	count int64
}

// NewCounter creates a new counter instrument.
func NewCounter() *Counter {
	return new(Counter)
}

// Update adds v to the counter.
func (c *Counter) Update(v int64) {
	atomic.AddInt64(&c.count, v)
}

// Snapshot returns the current value and reset the counter.
func (c *Counter) Snapshot() int64 {
	return atomic.SwapInt64(&c.count, 0)
}

// Rate tracks the rate of values per second.
type Rate struct {
	time  int64
	unit  time.Duration
	count *Counter
}

// NewRate creates a new rate instrument.
func NewRate() *Rate {
	return NewRateScale(time.Second)
}

// NewRateScale creates a new rate instruments with the given unit.
func NewRateScale(d time.Duration) *Rate {
	return &Rate{
		time:  time.Now().UnixNano(),
		unit:  d,
		count: NewCounter(),
	}
}

// Update updates rate value.
func (r *Rate) Update(v int64) {
	r.count.Update(v)
}

// Snapshot returns the number of values per second since the last snapshot,
// and reset the count to zero.
func (r *Rate) Snapshot() int64 {
	now := time.Now().UnixNano()
	t := atomic.SwapInt64(&r.time, now)
	c := r.count.Snapshot()
	s := float64(c) / rateScale / float64(now-t)
	return Ceil(s * Scale(r.unit, time.Second))
}

// Derive tracks the rate of deltas per seconds.
type Derive struct {
	rate  *Rate
	value int64
}

// NewDerive creates a new derive instruments.
func NewDerive(v int64) *Derive {
	return &Derive{
		value: v,
		rate:  NewRate(),
	}
}

// NewDeriveScale creates a new derive instruments with the given unit.
func NewDeriveScale(v int64, d time.Duration) *Derive {
	return &Derive{
		value: v,
		rate:  NewRateScale(d),
	}
}

// Update update rate value based on the stored previous value.
func (d *Derive) Update(v int64) {
	p := atomic.SwapInt64(&d.value, v)
	d.rate.Update(v - p)
}

// Snapshot returns the number of values per seconds since the last snapshot,
// and reset the count to zero.
func (d *Derive) Snapshot() int64 {
	return d.rate.Snapshot()
}

// Reservoir tracks a sample of values.
type Reservoir struct {
	size   int64
	values []int64
	m      sync.Mutex
}

const defaultReservoirSize = 1028

// NewReservoir creates a new reservoir of the given size.
// If size is negative, it will create a sample of DefaultReservoirSize size.
func NewReservoir(size int64) *Reservoir {
	if size <= 0 {
		size = defaultReservoirSize
	}
	return &Reservoir{
		values: make([]int64, size),
	}
}

// Update fills the sample randomly with given value,
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

// Snapshot returns sample as a sorted array.
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

// Gauge tracks a value.
type Gauge struct {
	value int64
}

// NewGauge creates a new Gauge with the given value.
func NewGauge(v int64) *Gauge {
	return &Gauge{
		value: v,
	}
}

// Update updates the current stored value.
func (g *Gauge) Update(v int64) {
	atomic.StoreInt64(&g.value, v)
}

// Snapshot returns the current value.
func (g *Gauge) Snapshot() int64 {
	return atomic.LoadInt64(&g.value)
}

// Timer tracks durations.
type Timer struct {
	r *Reservoir
}

// NewTimer creates a new Timer with the given sample size.
func NewTimer(size int64) *Timer {
	return &Timer{
		r: NewReservoir(size),
	}
}

// Update adds duration to the sample in ms.
func (t *Timer) Update(d time.Duration) {
	v := Floor(d.Seconds() * 1000)
	t.r.Update(v)
}

// Snapshot returns durations sample as a sorted array.
func (t *Timer) Snapshot() []int64 {
	return t.r.Snapshot()
}

// Since records duration since the given start time.
func (t *Timer) Since(start time.Time) {
	t.Update(time.Since(start))
}

// Time records given function execution time.
func (t *Timer) Time(f func()) {
	ts := time.Now()
	f()
	t.Update(time.Since(ts))
}
