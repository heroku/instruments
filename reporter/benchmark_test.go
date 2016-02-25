package reporter_test

import (
	"fmt"
	"testing"

	"github.com/bsm/instruments"
	"github.com/bsm/instruments/reporter"
)

func BenchmarkRegistry(b *testing.B) {
	r := reporter.NewRegistry()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Register("foo", nil, instruments.NewRate())
	}
}

func BenchmarkInstruments(b *testing.B) {
	n := 10000
	r := reporter.NewRegistry()
	for i := 0; i < n; i++ {
		r.Register(fmt.Sprintf("foo.%d", i), nil, instruments.NewRate())
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Instruments()
	}
}

func BenchmarkSnapshot(b *testing.B) {
	n := 10000
	s := reporter.NewRegistry()
	for i := 0; i < n; i++ {
		s.Register(fmt.Sprintf("foo.%d", i), nil, instruments.NewRate())
	}
	m := s.Instruments()
	r := reporter.NewRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.SetInstruments(m)
		if count := len(r.Snapshot()); count != n {
			b.Fatal("snapshot returned unexpected count:", count)
		}
	}
}
