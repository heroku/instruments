package instruments

import (
	"sync"

	"github.com/bsm/histogram"
)

// Distribution is returned by Sample snapshots
type Distribution interface {
	// Count returns the number of observations
	Count() int
	// Min returns the minimum observed value
	Min() float64
	// Max returns the maximum observed value
	Max() float64
	// Sum returns the sum
	Sum() float64
	// Mean returns the mean
	Mean() float64
	// Quantile returns the quantile for a given q (0..1)
	Quantile(q float64) float64
	// Variance returns the variance
	Variance() float64
}

// --------------------------------------------------------------------

const defaultHistogramSize = 20

var histogramPool sync.Pool

func newHistogram(sz int) (h *histogram.Histogram) {
	if v := histogramPool.Get(); v != nil {
		h = v.(*histogram.Histogram)
	} else {
		h = new(histogram.Histogram)
	}
	h.Reset(sz)
	return
}

func releaseHistogram(h *histogram.Histogram) {
	histogramPool.Put(h)
}
