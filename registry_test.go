package instruments

import (
	"reflect"
	"testing"
	"time"
)

// Testing extensions for Registry

func (r *Registry) GetInstruments() map[string]interface{}            { return r.instruments }
func (r *Registry) SetInstruments(instruments map[string]interface{}) { r.instruments = instruments }
func (r *Registry) Reset() int                                        { return len(r.reset()) }

// Mock reporter

type mockReported struct {
	Name  string
	Tags  []string
	Value float64
}

type mockReporter struct {
	Data    []mockReported
	Prepped bool
	Flushed map[string]float64
}

func (m *mockReporter) Prep() error {
	m.Prepped = true
	return nil
}

func (m *mockReporter) Flush() error {
	m.Flushed = make(map[string]float64, len(m.Data))
	for _, i := range m.Data {
		m.Flushed[MetricID(i.Name, i.Tags)] = i.Value
	}
	return nil
}

func (m *mockReporter) Discrete(name string, tags []string, inst Discrete) error {
	m.Data = append(m.Data, mockReported{
		Name:  name,
		Tags:  tags,
		Value: float64(inst.Snapshot()),
	})
	return nil
}

func (m *mockReporter) Sample(name string, tags []string, inst Sample) error {
	m.Data = append(m.Data, mockReported{
		Name:  name,
		Tags:  tags,
		Value: inst.Snapshot().Mean(),
	})
	return nil
}

// --------------------------------------------------------------------

func TestRegistry(t *testing.T) {
	r := New(time.Minute, "")
	defer r.Close()

	r.Register("foo", []string{"a", "b"}, NewRate())
	if r.Size() != 1 {
		t.Error("instrument not registered")
	}
	r.Unregister("foo", []string{"b", "a"})
	if r.Size() != 0 {
		t.Error("instrument not unregistered")
	}
}

func TestUnstartedRegistry(t *testing.T) {
	r := NewUnstarted("")
	if err := r.Close(); err != nil {
		t.Error("unexpected error on close", err)
	}
}

func TestRegistryNormalization(t *testing.T) {
	r := New(time.Minute, "")
	defer r.Close()

	r.Register("foo", []string{"a", "b"}, NewRate())
	r.Register("foo", []string{"b", "a"}, NewRate())
	r.Register("bar", []string{}, NewRate())
	r.Register("bar", nil, NewRate())
	if r.Size() != 2 {
		t.Error("incorrect normalization")
	}

	r.Unregister("foo", []string{"b", "a"})
	r.Unregister("bar", []string{"a"})
	if r.Size() != 1 {
		t.Error("incorrect normalization")
	}

	r.Unregister("bar", nil)
	if r.Size() != 0 {
		t.Error("incorrect normalization")
	}
}

func TestRegistryRegister(t *testing.T) {
	r := New(time.Minute, "")
	defer r.Close()

	r.Register("foo", nil, NewRate())
	if r := r.Get("foo", nil); r == nil {
		t.Error("instrument not returned")
	}
}

func TestRegistryFetch(t *testing.T) {
	r := New(time.Minute, "")
	defer r.Close()

	var ic, ir int
	nc := func() interface{} {
		ic++
		return NewCounter()
	}
	nr := func() interface{} {
		ir++
		return NewRate()
	}

	if _, ok := r.Fetch("foo", nil, nc).(*Counter); !ok {
		t.Error("unexpected instrument, expected counter")
	} else if ic != 1 {
		t.Error("expected NewCounter to have been called")
	}

	if _, ok := r.Fetch("foo", nil, nr).(*Counter); !ok {
		t.Error("unexpected instrument, expected counter")
	} else if ic != 1 {
		t.Error("expected NewCounter not to have been called again")
	}

	if _, ok := r.Fetch("foo", []string{"b", "a"}, nr).(*Rate); !ok {
		t.Error("unexpected instrument, expected rate")
	} else if ir != 1 {
		t.Error("expected NewRate to have been called")
	}

	if _, ok := r.Fetch("foo", []string{"a", "b"}, nc).(*Rate); !ok {
		t.Error("unexpected instrument, expected rate")
	} else if ir != 1 {
		t.Error("expected NewRate not to have been called again")
	}

	if _, ok := r.Fetch("foo", []string{"c"}, nr).(*Rate); !ok {
		t.Error("unexpected instrument, expected rate")
	} else if ir != 2 {
		t.Error("expected NewRate to have been called")
	}
}

func TestRegistryFlush(t *testing.T) {
	rep := new(mockReporter)
	reg := New(time.Minute, "myapp.", "a", "b")
	reg.Subscribe(rep)
	defer reg.Close()

	cnt1 := NewCounter()
	reg.Register("foo", []string{"c"}, cnt1)
	cnt1.Update(2)
	cnt1.Update(6)
	cnt1.Update(4)
	cnt1.Update(8)

	cnt2 := NewCounter()
	reg.Register("foo", []string{"d"}, cnt2)
	cnt2.Update(7)

	resv := NewReservoir(4)
	reg.Register("bar", []string{"d", "e"}, resv)
	resv.Update(2)
	resv.Update(6)
	resv.Update(4)
	resv.Update(8)

	if err := reg.Flush(); err != nil {
		t.Error("expected no error")
	}
	if !rep.Prepped {
		t.Errorf("expected reported to be prepped")
	}
	if exp := map[string]float64{
		"myapp.foo|a,b,c":   20,
		"myapp.foo|a,b,d":   7,
		"myapp.bar|a,b,d,e": 5,
	}; !reflect.DeepEqual(rep.Flushed, exp) {
		t.Errorf("want:\n%v\ngot:\n%v", exp, rep.Flushed)
	}
}

func TestRegistryReset(t *testing.T) {
	r := New(time.Minute, "")
	defer r.Close()

	r.Register("foo", nil, NewRate())
	if r.Size() != 1 {
		t.Error("instrument not registered")
	}
	if snapshot := r.reset(); len(snapshot) != 1 {
		t.Error("instrument not returned")
	}
	if r.Size() != 0 {
		t.Error("instrument not snapshoted")
	}
}
