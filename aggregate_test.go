package instruments

import (
	"testing"
	"testing/quick"
)

var quantileTests = []struct {
	slice    SampleSlice
	quantile float64
	value    int64
}{
	{
		slice:    SampleSlice{6, 7, 15, 36, 39, 40, 41, 43, 43, 47, 49},
		quantile: 0.75,
		value:    43,
	},
	{
		slice:    SampleSlice{6, 7, 15, 36, 39, 40, 41, 43, 43, 47, 49},
		quantile: 0.25,
		value:    15,
	},
	{
		slice:    SampleSlice{6, 7, 15, 36, 39, 40, 41, 43, 43, 47, 49},
		quantile: 0.5,
		value:    40,
	},
	{
		slice:    SampleSlice{},
		quantile: 0.95,
		value:    0,
	},
}

func TestQuantile(t *testing.T) {
	for i, qt := range quantileTests {
		q := qt.slice.Quantile(qt.quantile)
		if q != qt.value {
			t.Errorf("%d: wrong value returned for quantile %g, got %d wants %d", i, qt.quantile, q, qt.value)
		}
	}
}

func TestMean(t *testing.T) {
	m := SampleSlice{1, 2, 3, 4}.Mean()
	if m != 2.5 {
		t.Error("wrong calculated mean")
	}
}

var minMaxTests = []struct {
	slice SampleSlice
	min   int64
	max   int64
}{
	{
		slice: SampleSlice{},
		min:   0,
		max:   0,
	},
	{
		slice: SampleSlice{6, 7, 15, 36, 39, 40, 41, 43, 43, 47, 49},
		min:   6,
		max:   49,
	},
	{
		slice: SampleSlice{15, 6, 36, 39, 7, 40, 43, 43, 47, 49},
		min:   6,
		max:   49,
	},
}

func TestMinMax(t *testing.T) {
	for i, mm := range minMaxTests {
		min := mm.slice.Min()
		if min != mm.min {
			t.Errorf("%d: expected min value %d got %d", i, mm.min, min)
		}
		max := mm.slice.Max()
		if max != mm.max {
			t.Errorf("%d: expected max value %d got %d", i, mm.max, max)
		}
	}
}

func TestMin(t *testing.T) {
	min := func(slice SampleSlice) bool {
		min := slice.Min()
		for _, v := range slice {
			if v < min {
				return false
			}
		}
		return true
	}
	if err := quick.Check(min, nil); err != nil {
		t.Error(err)
	}
}

func TestMax(t *testing.T) {
	max := func(slice SampleSlice) bool {
		max := slice.Max()
		for _, v := range slice {
			if v > max {
				return false
			}
		}
		return true
	}
	if err := quick.Check(max, nil); err != nil {
		t.Error(err)
	}
}
