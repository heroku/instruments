package datadog

import (
	"time"

	"github.com/bsm/instruments"
)

var _ instruments.Reporter = (*Reporter)(nil)

// Reporter implements instruments.Reporter and simply logs metrics
type Reporter struct {
	// Client is a customisable reporter client
	Client *Client

	metrics   []*Metric
	timestamp int64
}

// New creates a new reporter.
func New(apiKey string) *Reporter {
	return &Reporter{Client: NewClient(apiKey)}
}

// Prepare implements instruments.Reporter
func (r *Reporter) Prep() error {
	r.timestamp = time.Now().Unix()
	return nil
}

// Discrete implements instruments.Reporter
func (r *Reporter) Discrete(name string, tags []string, inst instruments.Discrete) error {
	r.metrics = append(r.metrics,
		BuildMetric(name, tags, r.timestamp, inst.Snapshot()),
	)
	return nil
}

// Sample implements instruments.Reporter
func (r *Reporter) Sample(name string, tags []string, inst instruments.Sample) error {
	s := inst.Snapshot()
	r.metrics = append(r.metrics,
		BuildMetric(name+".p95", tags, r.timestamp, s.Quantile(0.95)),
		BuildMetric(name+".p99", tags, r.timestamp, s.Quantile(0.99)),
	)
	return nil
}

// Flush implements instruments.Reporter
func (r *Reporter) Flush() error {
	if err := r.Client.Post(r.metrics); err != nil {
		return err
	}
	r.metrics = r.metrics[:0]
	return nil
}
