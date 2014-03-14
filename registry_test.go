package instruments

import "testing"

func BenchmarkRegistry(b *testing.B) {
	r := NewRegistry()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Register("foo", NewRate())
	}
}

func BenchmarkInstruments(b *testing.B) {
	r := NewRegistry()
	r.Register("foo", NewRate())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Instruments()
	}
}

func TestRegistration(t *testing.T) {
	r := NewRegistry()
	r.Register("foo", NewRate())
	if registered := r.Instruments(); len(registered) != 1 {
		t.Error("instrument not registered")
	}
	r.Unregister("foo")
	if registered := r.Instruments(); len(registered) != 0 {
		t.Error("instrument not unregistered")
	}
}

func TestGetOrRegisterInstrument(t *testing.T) {
	r := NewRegistry()
	r.Register("foo", NewRate())
	i := r.Register("foo", NewGauge(0))
	if _, ok := i.(*Rate); !ok {
		t.Fatal("wrong instrument type")
	}
	registered := r.Instruments()
	if len(registered) != 1 {
		t.Fatal("registry should only have one instruments registered")
	}
	i, p := registered["foo"]
	if !p {
		t.Fatal("instrument not found")
	}
	if _, ok := i.(*Rate); !ok {
		t.Fatal("wrong instrument type")
	}
}

func TestGetInstrument(t *testing.T) {
	r := NewRegistry()
	r.Register("foo", NewRate())
	if r := r.Get("foo"); r == nil {
		t.Error("instrument not returned")
	}
}

func TestSnapshotInstruments(t *testing.T) {
	r := NewRegistry()
	r.Register("foo", NewRate())
	if r.Size() != 1 {
		t.Error("instrument not registered")
	}
	if snapshot := r.Snapshot(); len(snapshot) != 1 {
		t.Error("instrument not returned")
	}
	if r.Size() != 0 {
		t.Error("instrument not snapshoted")
	}
}
