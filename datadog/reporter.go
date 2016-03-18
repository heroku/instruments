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

	current, stale map[string]struct{}
}

// New creates a new reporter.
func New(apiKey string) *Reporter {
	return &Reporter{
		Client:  NewClient(apiKey),
		current: make(map[string]struct{}),
		stale:   make(map[string]struct{}),
	}
}

// Prepare implements instruments.Reporter
func (r *Reporter) Prep() error {
	r.timestamp = time.Now().Unix()
	return nil
}

// Discrete implements instruments.Reporter
func (r *Reporter) Discrete(name string, tags []string, inst instruments.Discrete) error {
	metricID := instruments.MetricID(name, tags)
	r.current[metricID] = struct{}{}
	delete(r.stale, metricID)

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
	r.zeroFillStale()
	if err := r.Client.Post(r.metrics); err != nil {
		return err
	}
	r.metrics = r.metrics[:0]
	r.stale, r.current = r.current, r.stale
	return nil
}

// Datadog has the annoying habbit of using the last reported value
// for alerts, which is particularly bad when you set a minimum threshold
// on e.g. errors. You may end up with a short-term spike which then stops
// being reported as errors go back to zero. To compensate, we cache reported
// metrics and zero-fill (once) them if they stop being reported.
func (r *Reporter) zeroFillStale() {
	now := time.Now().Unix()
	for metricID := range r.stale {
		name, tags := instruments.SplitMetricID(metricID)
		r.metrics = append(r.metrics, BuildMetric(name, tags, now, 0))
		delete(r.stale, metricID)
	}
}
