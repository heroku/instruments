package instruments

import "testing"

var quantileTests = []struct {
	values   []int64
	quantile float64
	value    int64
}{
	{
		values:   []int64{6, 7, 15, 36, 39, 40, 41, 43, 43, 47, 49},
		quantile: 0.75,
		value:    43,
	},
	{
		values:   []int64{6, 7, 15, 36, 39, 40, 41, 43, 43, 47, 49},
		quantile: 0.25,
		value:    15,
	},
	{
		values:   []int64{6, 7, 15, 36, 39, 40, 41, 43, 43, 47, 49},
		quantile: 0.5,
		value:    40,
	},
	{
		values:   []int64{},
		quantile: 0.95,
		value:    0,
	},
}

func TestQuantile(t *testing.T) {
	for i, qt := range quantileTests {
		q := Quantile(qt.values, qt.quantile)
		if q != qt.value {
			t.Errorf("%d: wrong value returned for quantile %g, got %d wants %d", i, qt.quantile, q, qt.value)
		}
	}
}

func TestMean(t *testing.T) {
	m := Mean([]int64{1, 2, 3, 4})
	if m != 2.5 {
		t.Error("wrong calculated mean")
	}
}

var minMaxTests = []struct {
	values []int64
	min    int64
	max    int64
}{
	{
		values: []int64{},
		min:    0,
		max:    0,
	},
	{
		values: []int64{6, 7, 15, 36, 39, 40, 41, 43, 43, 47, 49},
		min:    6,
		max:    49,
	},
	{
		values: []int64{15, 6, 36, 39, 7, 40, 43, 43, 47, 49},
		min:    6,
		max:    49,
	},
}

func TestMinMax(t *testing.T) {
	for i, mm := range minMaxTests {
		min := Min(mm.values)
		if min != mm.min {
			t.Errorf("%d: expected min value %d got %d", i, mm.min, min)
		}
		max := Max(mm.values)
		if max != mm.max {
			t.Errorf("%d: expected max value %d got %d", i, mm.max, max)
		}
	}
}
