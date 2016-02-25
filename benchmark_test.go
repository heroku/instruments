package instruments_test

import (
	"testing"

	"github.com/bsm/instruments"
)

func BenchmarkCounter(b *testing.B) {
	c := instruments.NewCounter()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Update(int64(i))
		c.Snapshot()
	}
}

func BenchmarkRate(b *testing.B) {
	r := instruments.NewRate()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Update(int64(i))
		r.Snapshot()
	}
}

func BenchmarkReservoir(b *testing.B) {
	r := instruments.NewReservoir(-1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Update(int64(i))
		r.Snapshot()
	}
}
