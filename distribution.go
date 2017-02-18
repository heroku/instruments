package instruments

import (
	"math"
	"sort"
	"sync"
)

// Distribution is returned by Sample snapshots
type Distribution interface {
	// Add adds a new value
	Add(v float64)
	// Count returns the number of observations
	Count() int
	// Min returns the minimum observed value
	Min() float64
	// Max returns the maximum observed value
	Max() float64
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

func (b histogramBin) Sum() float64 { return math.Abs(b.w) * b.v }

// A histogram data struct, heavily inspired by Ben-Haim's and Yom-Tov's
// "A Streaming Parallel Decision Tree Algorithm"
// paper http://www.jmlr.org/papers/volume11/ben-haim10a/ben-haim10a.pdf
type histogram struct {
	bins []histogramBin
	size int
	cnt  int

	min, max float64
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

func (h *histogram) Count() int   { return h.cnt }
func (h *histogram) Min() float64 { return h.min }
func (h *histogram) Max() float64 { return h.max }

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
		return h.min
	} else if q == 1.0 {
		return h.max
	}

	delta := q * float64(h.cnt)
	pos := 0
	for w0 := 0.0; pos < len(h.bins); pos++ {
		w1 := math.Abs(h.bins[pos].w) / 2.0
		if delta-w1-w0 < 0 {
			break
		}
		delta -= (w1 + w0)
		w0 = w1
	}

	switch pos {
	case 0: // lower bound
		return h.resolve(histogramBin{v: h.min, w: 0}, h.bins[pos], delta)
	case len(h.bins): // upper bound
		return h.resolve(h.bins[pos-1], histogramBin{v: h.max, w: 0}, delta)
	default:
		return h.resolve(h.bins[pos-1], h.bins[pos], delta)
	}
}

func (h *histogram) Add(v float64) {
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

func (h *histogram) resolve(b1, b2 histogramBin, delta float64) float64 {
	w1, w2 := b1.w, b2.w

	// return if both bins are exact (unmerged)
	if w1 > 0 && w2 > 0 {
		return b2.v
	}

	// normalise
	w1, w2 = math.Abs(w1), math.Abs(w2)

	// calculate multiplier
	var z float64
	if w1 == w2 {
		z = delta / w1
	} else {
		a := 2 * (w2 - w1)
		b := 2 * w1
		z = (math.Sqrt(b*b+4*a*delta) - b) / a
	}
	return b1.v + (b2.v-b1.v)*z
}

func (h *histogram) insert(v float64) {
	pos := h.search(v)

	if pos < len(h.bins) && h.bins[pos].v == v {
		h.bins[pos].w += math.Copysign(1, h.bins[pos].w)
		return
	}

	if pos < len(h.bins) {
		h.bins = h.bins[:len(h.bins)+1]
		copy(h.bins[pos+1:], h.bins[pos:])
	} else {
		h.bins = h.bins[:len(h.bins)+1]
	}

	h.bins[pos].w = 1
	h.bins[pos].v = v
}

func (h *histogram) prune() {
	if len(h.bins) <= h.size {
		return
	}

	delta := math.MaxFloat64
	pos := 0
	for i := 0; i < len(h.bins)-1; i++ {
		b1, b2 := h.bins[i], h.bins[i+1]
		if x := b2.v - b1.v; x < delta {
			pos, delta = i, x
		}
	}

	b1, b2 := h.bins[pos], h.bins[pos+1]
	w := math.Abs(b1.w) + math.Abs(b2.w)
	v := (b1.Sum() + b2.Sum()) / w
	h.bins[pos+1].w = -w
	h.bins[pos+1].v = v
	h.bins = h.bins[:pos+copy(h.bins[pos:], h.bins[pos+1:])]
}

func (h *histogram) search(v float64) int {
	return sort.Search(len(h.bins), func(i int) bool { return h.bins[i].v >= v })
}
