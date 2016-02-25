// Runtime provides runtime instrumentations
// around memory usage, goroutine and cgo calls.
package runtime

import (
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/bsm/instruments"
)

// Allocated collects the number of bytes allocated and still in use.
type Allocated struct {
	g   *instruments.Gauge
	mem runtime.MemStats
	m   sync.Mutex
}

// NewAllocated creates a new Allocated.
func NewAllocated() *Allocated {
	return &Allocated{
		g: instruments.NewGauge(0),
	}
}

// Update updates the number of bytes allocated and still in use.
func (a *Allocated) Update() {
	a.m.Lock()
	defer a.m.Unlock()

	runtime.ReadMemStats(&a.mem)
	a.g.Update(int64(a.mem.Alloc))
}

// Snapshot returns the current number of bytes allocated and still in use.
func (a *Allocated) Snapshot() int64 {
	return a.g.Snapshot()
}

// Heap collects the number of bytes allocated and still in use in the heap.
type Heap struct {
	g   *instruments.Gauge
	mem runtime.MemStats
	m   sync.Mutex
}

// NewHeap creates a new Heap.
func NewHeap() *Heap {
	return &Heap{
		g: instruments.NewGauge(0),
	}
}

// Update updates the number of bytes allocated and still in use in the heap.
func (ha *Heap) Update() {
	ha.m.Lock()
	defer ha.m.Unlock()

	runtime.ReadMemStats(&ha.mem)
	ha.g.Update(int64(ha.mem.HeapAlloc))
}

// Snapshot returns the current number of bytes allocated and still in use in the heap.
func (ha *Heap) Snapshot() int64 {
	return ha.g.Snapshot()
}

// Stack collects the number of bytes used now in the stack.
type Stack struct {
	g   *instruments.Gauge
	mem runtime.MemStats
	m   sync.Mutex
}

// NewStack creates a new Stack.
func NewStack() *Stack {
	return &Stack{
		g: instruments.NewGauge(0),
	}
}

// Update updates the number of bytes allocated and still in use in the stack.
func (s *Stack) Update() {
	s.m.Lock()
	defer s.m.Unlock()

	runtime.ReadMemStats(&s.mem)
	s.g.Update(int64(s.mem.StackInuse))
}

// Snapshot returns the current number of bytes allocated and still in use in the stack.
func (s *Stack) Snapshot() int64 {
	return s.g.Snapshot()
}

// Goroutine collects the number of existing goroutines.
type Goroutine struct {
	g *instruments.Gauge
}

// NewGoroutine creats a new Goroutine.
func NewGoroutine() *Goroutine {
	return &Goroutine{
		g: instruments.NewGauge(0),
	}
}

// Update udpates the number of existing goroutines.
func (gr *Goroutine) Update() {
	gr.g.Update(int64(runtime.NumGoroutine()))
}

// Snapshot returns the current number of existing goroutines
func (gr *Goroutine) Snapshot() int64 {
	return gr.g.Snapshot()
}

// Cgo collects the number of cgo calls made by the current process.
type Cgo struct {
	g *instruments.Gauge
}

// NewCgo creats a new Cgo.
func NewCgo() *Cgo {
	return &Cgo{
		g: instruments.NewGauge(0),
	}
}

// Update updates the number of cgo calls made by the current process.
func (c *Cgo) Update() {
	c.g.Update(runtime.NumCgoCall())
}

// Snapshot returns the current number of cgo calls made.
func (c *Cgo) Snapshot() int64 {
	return c.g.Snapshot()
}

// Frees collects the number of frees.
type Frees struct {
	d   *instruments.Derive
	mem runtime.MemStats
	m   sync.Mutex
}

// NewFrees creates a Frees.
func NewFrees() *Frees {
	return &Frees{
		d: instruments.NewDerive(0),
	}
}

// Update updates the number of frees.
func (f *Frees) Update() {
	f.m.Lock()
	defer f.m.Unlock()

	runtime.ReadMemStats(&f.mem)
	f.d.Update(int64(f.mem.Frees))
}

// Snapshot returns the number of frees.
func (f *Frees) Snapshot() int64 {
	return f.d.Snapshot()
}

// Lookups collects the number of pointer lookups.
type Lookups struct {
	d   *instruments.Derive
	mem runtime.MemStats
	m   sync.Mutex
}

// NewLookups creates a new Lookups.
func NewLookups() *Lookups {
	return &Lookups{
		d: instruments.NewDerive(0),
	}
}

// Update updates the number of pointer lookups.
func (l *Lookups) Update() {
	l.m.Lock()
	defer l.m.Unlock()

	runtime.ReadMemStats(&l.mem)
	l.d.Update(int64(l.mem.Lookups))
}

// Snapshot returns the number of pointer lookups.
func (l *Lookups) Snapshot() int64 {
	return l.d.Snapshot()
}

// Mallocs collects the number of mallocs.
type Mallocs struct {
	d   *instruments.Derive
	mem runtime.MemStats
	m   sync.Mutex
}

// NewMallocs creates a new Mallocs.
func NewMallocs() *Mallocs {
	return &Mallocs{
		d: instruments.NewDerive(0),
	}
}

// Update updates the number of mallocs.
func (m *Mallocs) Update() {
	m.m.Lock()
	defer m.m.Unlock()

	runtime.ReadMemStats(&m.mem)
	m.d.Update(int64(m.mem.Mallocs))
}

// Snapshot returns the number of mallocs.
func (m *Mallocs) Snapshot() int64 {
	return m.d.Snapshot()
}

// Pauses collects pauses times.
type Pauses struct {
	r   *instruments.Reservoir
	n   uint32
	mem runtime.MemStats
	m   sync.Mutex
}

// NewPauses creates a new Pauses.
func NewPauses(size int64) *Pauses {
	return &Pauses{
		r: instruments.NewReservoir(size),
	}
}

// Update updates GC pauses times.
func (p *Pauses) Update() {
	p.m.Lock()
	defer p.m.Unlock()

	runtime.ReadMemStats(&p.mem)
	numGC := atomic.SwapUint32(&p.n, p.mem.NumGC)
	i := numGC % uint32(len(p.mem.PauseNs))
	j := p.mem.NumGC % uint32(len(p.mem.PauseNs))
	if p.mem.NumGC-numGC >= uint32(len(p.mem.PauseNs)) {
		for i = 0; i < uint32(len(p.mem.PauseNs)); i++ {
			p.r.Update(int64(p.mem.PauseNs[i]))
		}
	} else {
		if i > j {
			for ; i < uint32(len(p.mem.PauseNs)); i++ {
				p.r.Update(int64(p.mem.PauseNs[i]))
			}
			i = 0
		}
		for ; i < j; i++ {
			p.r.Update(int64(p.mem.PauseNs[i]))
		}
	}
}

// Snapshot returns a sample of GC pauses times.
func (p *Pauses) Snapshot() instruments.SampleSlice {
	return p.r.Snapshot()
}
