package instruments

import (
	"math"
	"sort"
	"sync"
)

// SampleSlice are returned by Sample.Snapshot. These are simple
// int64 slices which extensions for simpler calculations:
type SampleSlice []int64

func (s SampleSlice) Len() int           { return len(s) }
func (s SampleSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s SampleSlice) Less(i, j int) bool { return s[i] < s[j] }
func (s SampleSlice) Sort()              { sort.Sort(s) }
func (s SampleSlice) IsSorted() bool     { return sort.IsSorted(s) }

// Min returns minimun value of the given sample.
func (s SampleSlice) Min() int64 {
	if len(s) == 0 {
		return 0
	}
	if s.IsSorted() {
		return s[0]
	}
	min := s[0]
	for i := 1; i < len(s); i++ {
		v := s[i]
		if min > v {
			min = v
		}
	}
	return min
}

// Max returns maximum value of the given sample.
func (s SampleSlice) Max() int64 {
	if len(s) == 0 {
		return 0
	}
	if s.IsSorted() {
		return s[len(s)-1]
	}
	max := s[0]
	for i := 1; i < len(s); i++ {
		v := s[i]
		if max < v {
			max = v
		}
	}
	return max
}

// Mean returns the mean of the given sample.
func (s SampleSlice) Mean() float64 {
	if len(s) == 0 {
		return 0.0
	}
	var sum int64
	for _, v := range s {
		sum += v
	}
	return float64(sum) / float64(len(s))
}

// Quantile returns the nearest value to the given quantile.
func (s SampleSlice) Quantile(q float64) int64 {
	n := len(s)
	if n == 0 {
		return 0
	}
	m := Floor(float64(n) * q)
	i := min(n-1, int(m))
	return s[i]
}

// Variance returns variance if the given sample.
func (s SampleSlice) Variance() float64 {
	if len(s) == 0 {
		return 0.0
	}
	m := s.Mean()
	var sum float64
	for _, v := range s {
		d := float64(v) - m
		sum += d * d
	}
	return float64(sum) / float64(len(s))
}

// StandardDeviation returns standard deviation of the given sample.
func (s SampleSlice) StandardDeviation() float64 {
	return math.Sqrt(s.Variance())
}

// Release releases the SampleSlice to a pool where it can be recycled.
// Use this with caution! Once released the object must not be accessed
// by your code again.
func (s SampleSlice) Release() { sampleSlicePool.Put(s) }

// --------------------------------------------------------------------

var sampleSlicePool sync.Pool

func makeSampleSlice(size int) SampleSlice {
	if v := sampleSlicePool.Get(); v != nil {
		s := v.(SampleSlice)
		if size <= cap(s) {
			return s[:size]
		}
	}
	return make(SampleSlice, size)
}
