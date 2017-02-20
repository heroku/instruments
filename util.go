package instruments

import (
	"math"
	"sort"
	"strings"
	"sync/atomic"
	"unsafe"
)

// MetricID takes a name and tags and generates a consistent
// metric identifier
func MetricID(name string, tags []string) string {
	if len(tags) == 0 {
		return name
	}
	sort.Strings(tags)
	return name + "|" + strings.Join(tags, ",")
}

// SplitMetricID takes a metric ID ans splits it into
// name and tags
func SplitMetricID(metricID string) (name string, tags []string) {
	if metricID == "" {
		return "", nil
	}

	pos := strings.LastIndexByte(metricID, '|')
	if pos > 0 && pos < len(metricID)-1 {
		return metricID[:pos], strings.Split(metricID[pos+1:], ",")
	}
	return metricID, nil
}

func addFloat64(val *float64, delta float64) (new float64) {
	ptr := (*uint64)(unsafe.Pointer(val))

	for {
		old := *val
		new = old + delta
		if atomic.CompareAndSwapUint64(ptr, math.Float64bits(old), math.Float64bits(new)) {
			break
		}
	}
	return
}

func swapFloat64(val *float64, new float64) (old float64) {
	ptr := (*uint64)(unsafe.Pointer(val))
	newb := math.Float64bits(new)

	for {
		oldb := *ptr
		old = math.Float64frombits(oldb)
		if atomic.CompareAndSwapUint64(ptr, oldb, newb) {
			break
		}
	}
	return
}
