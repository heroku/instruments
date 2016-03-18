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
	parts := strings.SplitN(metricID, "|", 2)
	if len(parts) != 2 || parts[1] == "" {
		return parts[0], nil
	}
	return parts[0], strings.Split(parts[1], ",")
}
