package instruments

import (
	"reflect"
	"testing"
)

func TestMetricID(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		metricID string
	}{
		{"counter", []string{"a", "b"}, "counter|a,b"},
		{"counter", []string{"b", "a"}, "counter|a,b"},
		{"counter", nil, "counter"},
		{"counter", []string{}, "counter"},
	}

	for _, test := range tests {
		if metricID := MetricID(test.name, test.tags); metricID != test.metricID {
			t.Errorf("want %q got %q", test.metricID, metricID)
		}
	}
}

func TestSplitMetricID(t *testing.T) {
	tests := []struct {
		metricID string
		name     string
		tags     []string
	}{
		{"counter|a,b", "counter", []string{"a", "b"}},
		{"counter", "counter", nil},
	}

	for _, test := range tests {
		name, tags := SplitMetricID(test.metricID)
		if name != test.name {
			t.Errorf("want %q got %q", test.name, name)
		}
		if !reflect.DeepEqual(tags, test.tags) {
			t.Errorf("want %v got %v", test.tags, tags)
		}
	}
}
