package instruments

import (
	"math"
	"sort"
)

type int64array []int64

func (a int64array) Len() int {
	return len(a)
}

func (a int64array) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a int64array) Less(i, j int) bool {
	return a[i] < a[j]
}

func sorted(d int64array) {
	sort.Sort(d)
}

func isSorted(d int64array) bool {
	return sort.IsSorted(d)
}

func min(s, v int) int {
	if s <= v {
		return s
	} else {
		return v
	}
}

// Ceil returns the least integer value greater than or equal to x.
func Ceil(v float64) int64 {
	return int64(math.Ceil(v))
}

// Floor returns the greatest integer value less than or equal to x.
func Floor(v float64) int64 {
	return int64(math.Floor(v))
}
