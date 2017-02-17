package instruments

import (
	"math"
	"sort"
	"sync"
)

// Distribution is returned by Sample snapshots
type Distribution interface {
	// Add adds a new value
	Add(v int64)
	// Count returns the number of observations
	Count() int
	// Min returns the minimum observed value
	Min() int64
	// Max returns the maximum observed value
	Max() int64
	// Mean returns the mean
	Mean() float64
	// Clone copies the distribution
	Clone() Distribution
	// Quantile returns the quantile for a given q (0..1)
	Quantile(q float64) float64
	// Variance returns the variance
	Variance() float64
	// Release releases the Distribution to a pool where it can be recycled.
	// Use this with caution! Once released the object must not be accessed
	// by your code again.
	Release()
}

// --------------------------------------------------------------------

var histogramPool sync.Pool

type histogramBin struct {
	w float64 // weight
	v float64 // value
}

func (b histogramBin) Sum() float64 { return b.w * b.v }

// A histogram data struct, heavily inspired by Ben-Haim's and Yom-Tov's
// "A Streaming Parallel Decision Tree Algorithm"
// paper http://www.jmlr.org/papers/volume11/ben-haim10a/ben-haim10a.pdf
type histogram struct {
	bins []histogramBin
	size int
	cnt  int

	min, max int64
}

func newHistogram(sz int) *histogram {
	if v := histogramPool.Get(); v != nil {
		if h := v.(*histogram); sz < cap(h.bins) {
			h.Reset(sz)
			return h
		}
	}
	return &histogram{
		bins: make([]histogramBin, 0, sz+1),
		size: sz,
	}
}

func (h *histogram) Reset(sz int) {
	h.bins = h.bins[:0]
	h.size = sz
	h.min = 0
	h.max = 0
	h.cnt = 0
}

func (h *histogram) Clone() Distribution {
	x := newHistogram(h.size)
	x.bins = x.bins[:len(h.bins)]
	copy(x.bins, h.bins)
	x.min = h.min
	x.max = h.max
	x.cnt = h.cnt
	return x
}
func (h *histogram) Release() { histogramPool.Put(h) }

func (h *histogram) Count() int { return h.cnt }
func (h *histogram) Min() int64 { return h.min }
func (h *histogram) Max() int64 { return h.max }

func (h *histogram) Mean() float64 {
	if h.cnt == 0 {
		return 0.0
	}

	var sum float64
	for _, b := range h.bins {
		sum += b.Sum()
	}
	return sum / float64(h.cnt)
}

func (h *histogram) Variance() float64 {
	if h.cnt == 0 {
		return 0.0
	}

	var vv float64
	mean := h.Mean()
	for _, b := range h.bins {
		delta := mean - b.v
		vv += delta * delta * b.w
	}
	return vv / float64(h.cnt)
}

func (h *histogram) Quantile(q float64) float64 {
	if h.cnt == 0 || q < 0.0 || q > 1.0 {
		return 0.0
	} else if q == 0.0 {
		return float64(h.min)
	} else if q == 1.0 {
		return float64(h.max)
	}

	gap := q * float64(h.cnt)
	pos := 0
	for w0 := 0.0; pos < len(h.bins); pos++ {
		w1 := h.bins[pos].w / 2.0
		if gap-w1-w0 < 0 {
			break
		}
		gap -= (w1 + w0)
		w0 = w1
	}

	// if we have hit the lower bound
	if pos == 0 {
		b := h.bins[pos]
		a := 2 * b.w
		m := float64(h.min)
		return m + (b.v-m)*math.Sqrt(4*a*gap)/a
	}

	// if we are in between two bins
	if pos < len(h.bins) {
		return h.bins[pos-1].v
	}

	// if we have hit the upper bound
	b := h.bins[pos-1]
	a := 2 * b.w
	return b.v + (float64(h.max)-b.v)*(1-math.Sqrt(a*a-4*a*gap)/a)
}

func (h *histogram) Add(v int64) {
	if h.cnt == 0 || v < h.min {
		h.min = v
	}
	if h.cnt == 0 || v > h.max {
		h.max = v
	}

	h.insert(v)
	h.cnt++

	h.prune()
}

func (h *histogram) getbins(i int) (x, y histogramBin) {
	if i == 0 {
		x, y = histogramBin{v: float64(h.min), w: 0.0}, h.bins[0]
	} else if i == len(h.bins) {
		x, y = h.bins[len(h.bins)-1], histogramBin{v: float64(h.max), w: 0.0}
	} else {
		x, y = h.bins[i-1], h.bins[i]
	}
	return
}

func (h *histogram) insert(v int64) {
	pos := h.search(v)
	fv := float64(v)

	if pos < len(h.bins) && h.bins[pos].v == fv {
		h.bins[pos].w += 1
		return
	}

	if pos < len(h.bins) {
		h.bins = h.bins[:len(h.bins)+1]
		copy(h.bins[pos+1:], h.bins[pos:])
	} else {
		h.bins = h.bins[:len(h.bins)+1]
	}

	h.bins[pos].w = 1
	h.bins[pos].v = fv
}

func (h *histogram) prune() {
	if len(h.bins) <= h.size {
		return
	}

	pos := 0
	delta := math.MaxFloat64
	for i := 0; i < len(h.bins)-1; i++ {
		b1, b2 := h.bins[i], h.bins[i+1]
		if x := b2.v - b1.v; x < delta {
			pos, delta = i, x
		}
	}

	b1, b2 := h.bins[pos], h.bins[pos+1]
	w := b1.w + b2.w
	v := (b1.Sum() + b2.Sum()) / w
	h.bins[pos+1].w = w
	h.bins[pos+1].v = v
	h.bins = h.bins[:pos+copy(h.bins[pos:], h.bins[pos+1:])]
}

func (h *histogram) search(v int64) int {
	fv := float64(v)
	return sort.Search(len(h.bins), func(i int) bool { return h.bins[i].v >= fv })
}
