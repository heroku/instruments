package instruments

import (
	"sync"
	"time"
)

// Registry is a registry of all instruments.
type Registry struct {
	instruments instrumentsMap
	reporters   []Reporter
	prefix      string
	tags        []string
	errors      chan error
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

	reg := &Registry{
		instruments: newInstrumentsMap(0),
		prefix:      prefix,
		tags:        tags,
		errors:      make(chan error, 10),
		closing:     make(chan struct{}),
		closed:      make(chan error, 1),
	}
	go reg.loop(flushInterval)
	return reg
}

// Errors allows to subscribe to errors reported by the Registry.
//
func (r *Registry) Errors() <-chan error { return r.errors }

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

// Close flushes all pending data to reporters
// and releases resources.
func (r *Registry) Close() error {
	close(r.closing)
	return <-r.closed
}

func (r *Registry) reset() instrumentsMap {
	r.mutex.Lock()
	instruments := r.instruments
	r.instruments = newInstrumentsMap(0)
	r.mutex.Unlock()
	return instruments
}

func (r *Registry) flush() error {
	r.mutex.RLock()
	reporters := r.reporters
	r.mutex.RUnlock()

	for _, rep := range reporters {
		if err := rep.Prep(); err != nil {
			return err
		}
	}

	for metricID, val := range r.reset() {
		name, tags := SplitMetricID(metricID)
		name = r.prefix + name
		tags = append(r.tags, tags...)

		for _, rep := range reporters {
			var err error
			switch inst := val.(type) {
			case Discrete:
				err = rep.Discrete(name, tags, inst)
			case Sample:
				err = rep.Sample(name, tags, inst)
			}
			if err != nil {
				return err
			}
		}
	}

	for _, rep := range reporters {
		if err := rep.Flush(); err != nil {
			return err
		}
	}
	return nil
}

func (r *Registry) loop(flushInterval time.Duration) {
	flusher := time.NewTicker(flushInterval)
	defer flusher.Stop()

	for {
		select {
		case <-r.closing:
			// close errors channel
			close(r.errors)

			// flush unconsumed errors
			go func() {
				for _ = range r.errors {
				}
			}()

			// flush again
			r.closed <- r.flush()
			close(r.closed)
			return
		case <-flusher.C:
			if err := r.flush(); err != nil {
				r.handleError(err)
			}
		}
	}
}

func (r *Registry) handleError(err error) {
	select {
	case r.errors <- err:
	default:
	}
}

// --------------------------------------------------------------------

var instrumentsMapPool sync.Pool

type instrumentsMap map[string]interface{}

func newInstrumentsMap(size int) instrumentsMap {
	if v := instrumentsMapPool.Get(); v != nil {
		return v.(instrumentsMap)
	}
	return make(instrumentsMap, size)
}

func (m instrumentsMap) Release() {
	for k := range m {
		delete(m, k)
	}
	instrumentsMapPool.Put(m)
}
