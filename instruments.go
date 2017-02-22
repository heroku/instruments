/*
Package instruments allows you to collects metrics over discrete time intervals.

Collected metrics will only reflect observations from last time window only,
rather than including observations from prior windows, contrary to EWMA based metrics.

Instruments support two types of instruments:
Discrete instruments return a single value, and Sample instruments a value distribution.

Theses base instruments are available:

- Counter: holds a counter that can be incremented or decremented.
- Rate: tracks the rate of values per seconds.
- Reservoir: randomly samples values.
- Derive: tracks the rate of values based on the delta with previous value.
- Gauge: tracks last value.
- Timer: tracks durations.

You can create custom instruments or compose new instruments form the built-in
instruments as long as they implements the Sample or Discrete interfaces.
*/
package instruments

import (
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bsm/histogram"
)

// Discrete represents a single value instrument.
type Discrete interface {
	Snapshot() float64
}

// Sample represents a sample instrument.
type Sample interface {
	Snapshot() Distribution
}

// --------------------------------------------------------------------

// Counter holds a counter that can be incremented or decremented.
type Counter struct {
	count float64
}

// NewCounter creates a new counter instrument.
func NewCounter() *Counter {
	return new(Counter)
}

// Update adds v to the counter.
func (c *Counter) Update(v float64) {
	addFloat64(&c.count, v)
}

// Snapshot returns the current value and reset the counter.
func (c *Counter) Snapshot() float64 {
	return swapFloat64(&c.count, 0)
}

// --------------------------------------------------------------------

// Rate tracks the rate of values per second.
type Rate struct {
	time  int64
	unit  float64
	count Counter
}

// NewRate creates a new rate instrument.
func NewRate() *Rate {
	return NewRateScale(time.Second)
}

// NewRateScale creates a new rate instruments with the given unit.
func NewRateScale(d time.Duration) *Rate {
	return &Rate{
		time: time.Now().UnixNano(),
		unit: float64(d),
	}
}

// Update updates rate value.
func (r *Rate) Update(v float64) {
	r.count.Update(v)
}

// Snapshot returns the number of values per second since the last snapshot,
// and reset the count to zero.
func (r *Rate) Snapshot() float64 {
	now := time.Now().UnixNano()
	dur := now - atomic.SwapInt64(&r.time, now)
	return r.count.Snapshot() / float64(dur) * r.unit
}

// --------------------------------------------------------------------

// Derive tracks the rate of deltas per seconds.
type Derive struct {
	rate  Rate
	value uint64
}

// NewDerive creates a new derive instruments.
func NewDerive(v float64) *Derive {
	return NewDeriveScale(v, time.Second)
}

// NewDeriveScale creates a new derive instruments with the given unit.
func NewDeriveScale(v float64, d time.Duration) *Derive {
	return &Derive{
		value: math.Float64bits(v),
		rate: Rate{
			time: time.Now().UnixNano(),
			unit: float64(d),
		},
	}
}

// Update update rate value based on the stored previous value.
func (d *Derive) Update(v float64) {
	p := atomic.SwapUint64(&d.value, math.Float64bits(v))
	d.rate.Update(v - math.Float64frombits(p))
}

// Snapshot returns the number of values per seconds since the last snapshot,
// and reset the count to zero.
func (d *Derive) Snapshot() float64 {
	return d.rate.Snapshot()
}

// Reservoir tracks a sample of values.
type Reservoir struct {
	hist *histogram.Histogram
	m    sync.Mutex
}

// --------------------------------------------------------------------

// NewReservoir creates a new reservoir
func NewReservoir() *Reservoir {
	return &Reservoir{
		hist: newHistogram(defaultHistogramSize),
	}
}

// Update fills the sample randomly with given value,
// for reference, see: http://en.wikipedia.org/wiki/Reservoir_sampling
func (r *Reservoir) Update(v float64) {
	r.m.Lock()
	r.hist.Add(v)
	r.m.Unlock()
}

// Snapshot returns a Distribution
func (r *Reservoir) Snapshot() Distribution {
	h := newHistogram(defaultHistogramSize)
	r.m.Lock()
	h = r.hist.Copy(h)
	r.m.Unlock()
	return h
}

// --------------------------------------------------------------------

// Gauge tracks a value.
type Gauge struct {
	value uint64
}

// NewGauge creates a new Gauge
func NewGauge() *Gauge {
	return new(Gauge)
}

// Update updates the current stored value.
func (g *Gauge) Update(v float64) {
	atomic.StoreUint64(&g.value, math.Float64bits(v))
}

// Snapshot returns the current value.
func (g *Gauge) Snapshot() float64 {
	u := atomic.LoadUint64(&g.value)
	return math.Float64frombits(u)
}

// --------------------------------------------------------------------

// Timer tracks durations.
type Timer struct {
	r Reservoir
}

// NewTimer creates a new Timer with millisecond resolution
func NewTimer() *Timer {
	return &Timer{
		r: Reservoir{hist: newHistogram(defaultHistogramSize)},
	}
}

// Update adds duration to the sample in ms.
func (t *Timer) Update(d time.Duration) {
	t.r.Update(d.Seconds() * 1000)
}

// Snapshot returns durations distribution
func (t *Timer) Snapshot() Distribution {
	return t.r.Snapshot()
}

// Since records duration since the given start time.
func (t *Timer) Since(start time.Time) {
	t.Update(time.Since(start))
}
