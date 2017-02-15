package instruments

import (
	"sort"
	"strings"
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
