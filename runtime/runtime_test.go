package runtime

import (
	"runtime"
	"testing"
)

func TestPauses(t *testing.T) {
	p := NewPauses(1024)

	// Reset GC count
	p.Update()
	p.Snapshot()

	// Run GC once
	runtime.GC()
	p.Update()
	if count := len(p.Snapshot()); count != 1 {
		t.Fatalf("captured %d gc runs, expect 1", count)
	}

	// Run GC twice
	runtime.GC()
	runtime.GC()
	p.Update()
	if count := len(p.Snapshot()); count != 2 {
		t.Fatalf("captured %d gc runs, expected 2", count)
	}

	// Wraps GC counts
	for i := 0; i < 257; i++ {
		runtime.GC()
	}
	p.Update()
	if count := len(p.Snapshot()); count != 256 {
		t.Fatalf("captured %d gc runs, expected 256", count)
	}
}
