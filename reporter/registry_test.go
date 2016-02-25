package reporter

import (
	"testing"

	"github.com/bsm/instruments"
)

// SetInstruments sets instruments for testing
func (r *Registry) SetInstruments(instruments map[MetricID]interface{}) {
	r.instruments = instruments
}

func TestRegistration(t *testing.T) {
	r := NewRegistry()
	r.Register("foo", []string{"a", "b"}, instruments.NewRate())
	if registered := r.Instruments(); len(registered) != 1 {
		t.Error("instrument not registered")
	}
	r.Unregister("foo", []string{"b", "a"})
	if registered := r.Instruments(); len(registered) != 0 {
		t.Error("instrument not unregistered")
	}
}

func TestNormalization(t *testing.T) {
	r := NewRegistry()
	r.Register("foo", []string{"a", "b"}, instruments.NewRate())
	r.Register("foo", []string{"b", "a"}, instruments.NewRate())
	r.Register("bar", []string{}, instruments.NewRate())
	r.Register("bar", nil, instruments.NewRate())
	if registered := r.Instruments(); len(registered) != 2 {
		t.Error("incorrect normalization")
	}

	r.Unregister("foo", []string{"b", "a"})
	r.Unregister("bar", []string{"a"})
	if registered := r.Instruments(); len(registered) != 1 {
		t.Error("incorrect normalization")
	}

	r.Unregister("bar", nil)
	if registered := r.Instruments(); len(registered) != 0 {
		t.Error("incorrect normalization")
	}
}

func TestGetInstrument(t *testing.T) {
	r := NewRegistry()
	r.Register("foo", nil, instruments.NewRate())
	if r := r.Get("foo", nil); r == nil {
		t.Error("instrument not returned")
	}
}

func TestResetInstruments(t *testing.T) {
	r := NewRegistry()
	r.Register("foo", nil, instruments.NewRate())
	if r.Size() != 1 {
		t.Error("instrument not registered")
	}
	if snapshot := r.Reset(); len(snapshot) != 1 {
		t.Error("instrument not returned")
	}
	if r.Size() != 0 {
		t.Error("instrument not snapshoted")
	}
}
