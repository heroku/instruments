package instruments

import (
	"testing"
	"time"
)

// Testing extensions for Registry

func (r *Registry) GetInstruments() map[string]interface{}            { return r.instruments }
func (r *Registry) SetInstruments(instruments map[string]interface{}) { r.instruments = instruments }
func (r *Registry) Reset() int                                        { return len(r.reset()) }

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

func TestRegistry_reset(t *testing.T) {
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
