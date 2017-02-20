package datadog

import (
	"os"
	"time"

	"github.com/bsm/instruments"
)

var _ instruments.Reporter = (*Reporter)(nil)

var unixTime = func() int64 { return time.Now().Unix() }

// Reporter implements instruments.Reporter and simply logs metrics
type Reporter struct {
	// Client is a customisable reporter client
	Client *Client

	// Hostname can be customised.
	// Default: set via os.Hostname()
	Hostname string

	metrics   []Metric
	timestamp int64
	refs      map[string]int8
}

// New creates a new reporter.
func New(apiKey string) *Reporter {
	hostname, _ := os.Hostname()

	return &Reporter{
		Client:   NewClient(apiKey),
		Hostname: hostname,
		refs:     make(map[string]int8),
	}
}

// Prepare implements instruments.Reporter
func (r *Reporter) Prep() error {
	r.timestamp = unixTime()
	return nil
}

// Metric appends a new metric to the reporter. The value v must be either an
// int64 or float64, otherwise an error is returned
func (r *Reporter) Metric(name string, tags []string, v float32) {
	r.metrics = append(r.metrics, Metric{
		Name:   name,
		Points: [][2]interface{}{[2]interface{}{r.timestamp, v}},
		Tags:   tags,
		Host:   r.Hostname,
	})
}

// Discrete implements instruments.Reporter
func (r *Reporter) Discrete(name string, tags []string, val float64) error {
	metricID := instruments.MetricID(name, tags)
	r.refs[metricID] = 2
	r.Metric(name, tags, float32(val))
	return nil
}

// Sample implements instruments.Reporter
func (r *Reporter) Sample(name string, tags []string, dist instruments.Distribution) error {
	r.Metric(name+".p95", tags, float32(dist.Quantile(0.95)))
	r.Metric(name+".p99", tags, float32(dist.Quantile(0.99)))
	return nil
}

// Flush implements instruments.Reporter
func (r *Reporter) Flush() error {
	for metricID := range r.refs {
		if r.refs[metricID]--; r.refs[metricID] < 1 {
			name, tags := instruments.SplitMetricID(metricID)
			r.Metric(name, tags, 0)
			delete(r.refs, metricID)
		}
	}
	if len(r.metrics) != 0 {
		if err := r.Client.Post(r.metrics); err != nil {
			return err
		}
		r.metrics = r.metrics[:0]
	}
	return nil
}
