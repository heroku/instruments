package instruments

import (
	"log"
	"os"
	"sync"
	"time"
)

type Logger interface {
	Printf(string, ...interface{})
}

// Registry is a registry of all instruments.
type Registry struct {
	Logger      Logger
	instruments map[string]interface{}
	reporters   []Reporter
	prefix      string
	tags        []string
	closing     chan struct{}
	closed      chan error
	mutex       sync.RWMutex
}

// New creates a new Registry with a flushInterval at which metrics
// are reported to the subscribed Reporter instances, a custom prefix
// which is prepended to every metric name and default tags.
// Default: 60s
//
// You should call/defer Close() on exit to flush all
// accummulated data and release all resources.
func New(flushInterval time.Duration, prefix string, tags ...string) *Registry {
	if flushInterval < time.Second {
		flushInterval = time.Minute
	}

	r := &Registry{
		Logger:      log.New(os.Stderr, "instruments: ", log.LstdFlags),
		instruments: make(map[string]interface{}),
		prefix:      prefix,
		tags:        tags,
		closing:     make(chan struct{}),
		closed:      make(chan error, 1),
	}
	go r.loop(flushInterval)
	return r
}

// New creates a new Registry without a background flush thread.
func NewUnstarted(prefix string, tags ...string) *Registry {
	return &Registry{
		instruments: make(map[string]interface{}),
		prefix:      prefix,
		tags:        tags,
	}
}

// Subscribe attaches a reporter to the Registry.
func (r *Registry) Subscribe(rep Reporter) {
	r.mutex.Lock()
	r.reporters = append(r.reporters, rep)
	r.mutex.Unlock()
}

// Get returns an instrument from the Registry.
func (r *Registry) Get(name string, tags []string) interface{} {
	key := MetricID(name, tags)
	r.mutex.RLock()
	v := r.instruments[key]
	r.mutex.RUnlock()
	return v
}

// Register registers a new instrument.
func (r *Registry) Register(name string, tags []string, v interface{}) {
	switch v.(type) {
	case Discrete, Sample:
		key := MetricID(name, tags)
		r.mutex.Lock()
		r.instruments[key] = v
		r.mutex.Unlock()
	}
}

// Unregister remove from the registry the instrument matching the given name/tags
func (r *Registry) Unregister(name string, tags []string) {
	key := MetricID(name, tags)
	r.mutex.Lock()
	delete(r.instruments, key)
	r.mutex.Unlock()
}

// Fetch returns an instrument from the Registry or creates a new one
// using the provided factory.
func (r *Registry) Fetch(name string, tags []string, factory func() interface{}) interface{} {
	key := MetricID(name, tags)

	r.mutex.RLock()
	v, ok := r.instruments[key]
	r.mutex.RUnlock()
	if ok {
		return v
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if v, ok = r.instruments[key]; !ok {
		switch v = factory(); v.(type) {
		case Discrete, Sample:
			r.instruments[key] = v
		}
	}
	return v
}

// Size returns the numbers of instruments in the registry.
func (r *Registry) Size() int {
	r.mutex.RLock()
	size := len(r.instruments)
	r.mutex.RUnlock()
	return size
}

// Flush performs a manual flush to all subscribed reporters.
// This method is usually called by a background thread
// every flushInterval, specified in New()
func (r *Registry) Flush() error {
	r.mutex.RLock()
	reporters := r.reporters
	rtags := r.tags
	r.mutex.RUnlock()

	for _, rep := range reporters {
		if err := rep.Prep(); err != nil {
			return err
		}
	}

	for metricID, val := range r.reset() {
		name, tags := SplitMetricID(metricID)
		if len(name) > 0 && name[0] == '|' {
			name = name[1:]
		} else {
			name = r.prefix + name
		}
		tags = append(tags, rtags...)

		switch inst := val.(type) {
		case Discrete:
			val := inst.Snapshot()
			for _, rep := range reporters {
				if err := rep.Discrete(name, tags, val); err != nil {
					return err
				}
			}
		case Sample:
			val := inst.Snapshot()
			for _, rep := range reporters {
				if err := rep.Sample(name, tags, val); err != nil {
					return err
				}
			}
			releaseDistribution(val)
		}
	}

	for _, rep := range reporters {
		if err := rep.Flush(); err != nil {
			return err
		}
	}
	return nil
}

// Tags returns global registry tags
func (r *Registry) Tags() []string {
	r.mutex.RLock()
	tags := r.tags
	r.mutex.RUnlock()
	return tags
}

// SetTags allows to set tags
func (r *Registry) SetTags(tags ...string) {
	r.mutex.Lock()
	r.tags = tags
	r.mutex.Unlock()
}

// AddTags allows to add tags
func (r *Registry) AddTags(tags ...string) {
	r.mutex.Lock()
	r.tags = append(r.tags, tags...)
	r.mutex.Unlock()
}

// Close flushes all pending data to reporters
// and releases resources.
func (r *Registry) Close() error {
	if r.closing == nil {
		return nil
	}
	close(r.closing)
	return <-r.closed
}

func (r *Registry) reset() map[string]interface{} {
	r.mutex.Lock()
	instruments := r.instruments
	r.instruments = make(map[string]interface{})
	r.mutex.Unlock()
	return instruments
}

func (r *Registry) loop(flushInterval time.Duration) {
	flusher := time.NewTicker(flushInterval)
	defer flusher.Stop()

	for {
		select {
		case <-r.closing:
			r.closed <- r.Flush()
			close(r.closed)
			return
		case <-flusher.C:
			if err := r.Flush(); err != nil {
				r.logf("flush error: %s", err.Error())
			}
		}
	}
}

func (r *Registry) logf(s string, v ...interface{}) {
	if r.Logger != nil {
		r.Logger.Printf(s, v...)
	}
}
