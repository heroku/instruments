package instruments

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

// Testing extensions for Registry

func (r *Registry) GetInstruments() map[string]interface{}            { return r.instruments }
func (r *Registry) SetInstruments(instruments map[string]interface{}) { r.instruments = instruments }
func (r *Registry) Reset() int                                        { return len(r.reset()) }

// Mock reporter

type mockReporter map[string]float64

func (m mockReporter) Flush() error {
	m["flush.called"] = 1
	return nil
}

func (m mockReporter) Discrete(name string, tags []string, inst Discrete) error {
	key := fmt.Sprintf("%s|%s", name, strings.Join(tags, ","))
	m[key] = float64(inst.Snapshot())
	return nil
}

func (m mockReporter) Sample(name string, tags []string, inst Sample) error {
	key := fmt.Sprintf("%s|%s", name, strings.Join(tags, ","))
	m[key] = inst.Snapshot().Mean()
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
}

func TestRegistryFlush(t *testing.T) {
	rep := make(mockReporter)
	reg := New(time.Minute, "myapp.", "a", "b")
	reg.Subscribe(rep)
	defer reg.Close()

	cntr := NewCounter()
	reg.Register("foo", []string{"c"}, cntr)
	cntr.Update(2)
	cntr.Update(6)
	cntr.Update(4)
	cntr.Update(8)

	resv := NewReservoir(4)
	reg.Register("bar", []string{"d", "e"}, resv)
	resv.Update(2)
	resv.Update(6)
	resv.Update(4)
	resv.Update(8)

	if err := reg.flush(); err != nil {
		t.Error("expected no error")
	}
	exp := mockReporter{
		"myapp.foo|a,b,c":   20,
		"myapp.bar|a,b,d,e": 5,
		"flush.called":      1,
	}
	if !reflect.DeepEqual(rep, exp) {
		t.Errorf("wants %v got %v", exp, rep)
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
