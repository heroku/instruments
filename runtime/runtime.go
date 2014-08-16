// Runtime provides runtime instrumentations
package runtime

import (
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/heroku/instruments"
)

type Allocated struct {
	g   *instruments.Gauge
	mem runtime.MemStats
	m   sync.Mutex
}

func NewAllocated() *Allocated {
	return &Allocated{
		g: instruments.NewGauge(0),
	}
}

func (a *Allocated) Update() {
	a.m.Lock()
	defer a.m.Unlock()

	runtime.ReadMemStats(&a.mem)
	a.g.Update(int64(a.mem.Alloc))
}

func (a *Allocated) Snapshot() int64 {
	return a.Snapshot()
}

type HeapAllocated struct {
	g   *instruments.Gauge
	mem runtime.MemStats
	m   sync.Mutex
}

func NewHeapAllocated() *HeapAllocated {
	return &HeapAllocated{
		g: instruments.NewGauge(0),
	}
}

func (ha *HeapAllocated) Update() {
	ha.m.Lock()
	defer ha.m.Unlock()

	runtime.ReadMemStats(&ha.mem)
	ha.g.Update(int64(ha.mem.HeapAlloc))
}

func (ha *HeapAllocated) Snapshot() int64 {
	return ha.g.Snapshot()
}

type StackInUse struct {
	g   *instruments.Gauge
	mem runtime.MemStats
	m   sync.Mutex
}

func NewStackInUse() *StackInUse {
	return &StackInUse{
		g: instruments.NewGauge(0),
	}
}

func (su *StackInUse) Update() {
	su.m.Lock()
	defer su.m.Unlock()

	runtime.ReadMemStats(&su.mem)
	su.g.Update(int64(su.mem.StackInuse))
}

func (su *StackInUse) Snapshot() int64 {
	return su.g.Snapshot()
}

type Goroutine struct {
	g *instruments.Gauge
}

func NewGoroutine() *Goroutine {
	return &Goroutine{
		g: instruments.NewGauge(0),
	}
}

func (gr *Goroutine) Update() {
	gr.g.Update(int64(runtime.NumGoroutine()))
}

func (gr *Goroutine) Snapshot() int64 {
	return gr.Snapshot()
}

type Cgo struct {
	g *instruments.Gauge
}

func NewCgo() *Cgo {
	return &Cgo{
		g: instruments.NewGauge(0),
	}
}

func (c *Cgo) Update() {
	c.g.Update(runtime.NumCgoCall())
}

func (c *Cgo) Snapshot() int64 {
	return c.Snapshot()
}

type Frees struct {
	d   *instruments.Derive
	mem runtime.MemStats
	m   sync.Mutex
}

func NewFrees() *Frees {
	return &Frees{
		d: instruments.NewDerive(0),
	}
}

func (f *Frees) Update() {
	f.m.Lock()
	defer f.m.Unlock()

	runtime.ReadMemStats(&f.mem)
	f.d.Update(int64(f.mem.Frees))
}

func (f *Frees) Snapshot() int64 {
	return f.Snapshot()
}

type Lookups struct {
	d   *instruments.Derive
	mem runtime.MemStats
	m   sync.Mutex
}

func NewLookups() *Lookups {
	return &Lookups{
		d: instruments.NewDerive(0),
	}
}

func (l *Lookups) Update() {
	l.m.Lock()
	defer l.m.Unlock()

	runtime.ReadMemStats(&l.mem)
	l.d.Update(int64(l.mem.Lookups))
}

func (l *Lookups) Snapshot() int64 {
	return l.Snapshot()
}

type Mallocs struct {
	d   *instruments.Derive
	mem runtime.MemStats
	m   sync.Mutex
}

func NewMallocs() *Mallocs {
	return &Mallocs{
		d: instruments.NewDerive(0),
	}
}

func (m *Mallocs) Update() {
	m.m.Lock()
	defer m.m.Unlock()

	runtime.ReadMemStats(&m.mem)
	m.d.Update(int64(m.mem.Mallocs))
}

func (m *Mallocs) Snapshot() int64 {
	return m.Snapshot()
}

type NumGC struct {
	d   *instruments.Derive
	mem runtime.MemStats
	m   sync.Mutex
}

func NewNumGC() *NumGC {
	return &NumGC{
		d: instruments.NewDerive(0),
	}
}

func (ng *NumGC) Update() {
	ng.m.Lock()
	defer ng.m.Unlock()

	runtime.ReadMemStats(&ng.mem)
	ng.d.Update(int64(ng.mem.NumGC))
}

func (ng *NumGC) Snapshot() int64 {
	return ng.Snapshot()
}

type Pauses struct {
	r   *instruments.Reservoir
	n   uint32
	mem runtime.MemStats
	m   sync.Mutex
}

func NewPauses(size int64) *Pauses {
	return &Pauses{
		r: instruments.NewReservoir(size),
	}
}

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

func (p *Pauses) Snapshot() []int64 {
	return p.r.Snapshot()
}

type TotalPause struct {
	g   *instruments.Gauge
	mem runtime.MemStats
	m   sync.Mutex
}

func NewTotalPause() *TotalPause {
	return &TotalPause{
		g: instruments.NewGauge(0),
	}
}

func (tp *TotalPause) Update() {
	tp.m.Lock()
	defer tp.m.Unlock()

	runtime.ReadMemStats(&tp.mem)
	tp.g.Update(int64(tp.mem.PauseTotalNs))
}

func (tp *TotalPause) Snapshot() int64 {
	return tp.Snapshot()
}
