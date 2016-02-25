package instruments

import "math"

func min(s, v int) int {
	if s <= v {
		return s
	}
	return v
}

// Ceil returns the least integer value greater than or equal to x.
func Ceil(v float64) int64 {
	return int64(math.Ceil(v))
}

// Floor returns the greatest integer value less than or equal to x.
func Floor(v float64) int64 {
	return int64(math.Floor(v))
}
