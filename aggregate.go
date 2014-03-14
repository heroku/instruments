package instruments

import "math"

// Returns the nearest value to the given quantile.
func Quantile(v []int64, q float64) int64 {
	n := len(v)
	if n == 0 {
		return 0
	}
	m := Floor(float64(n) * q)
	i := min(n-1, int(m))
	return v[i]
}

// Returns the mean of the given sample.
func Mean(values []int64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	var sum int64
	for _, v := range values {
		sum += v
	}
	return float64(sum) / float64(len(values))
}

// Returns standard deviation of the given sample.
func StandardDeviation(v []int64) float64 {
	return math.Sqrt(Variance(v))
}

// Returns variance if the given sample.
func Variance(values []int64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	m := Mean(values)
	var sum float64
	for _, v := range values {
		d := float64(v) - m
		sum += d * d
	}
	return float64(sum) / float64(len(values))
}

// Returns maximun value of the given sample.
func Max(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}
	if isSorted(values) {
		return values[len(values)-1]
	}
	var max int64 = math.MinInt64
	for _, v := range values {
		if max < v {
			max = v
		}
	}
	return max
}

// Returns minimun value of the given sample.
func Min(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}
	if isSorted(values) {
		return values[0]
	}
	var min int64 = math.MaxInt64
	for _, v := range values {
		if min > v {
			min = v
		}
	}
	return min
}
