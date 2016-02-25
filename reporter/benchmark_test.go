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
	r := reporter.NewRegistry()
	r.Register("foo", nil, instruments.NewRate())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Instruments()
	}
}

func BenchmarkSnapshot(b *testing.B) {
	n := 100000
	r := reporter.NewRegistry()
	for i := 0; i < n; i++ {
		r.Register(fmt.Sprintf("foo.%d", i), nil, instruments.NewRate())
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.Instruments()
	}
}
