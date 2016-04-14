package instruments

import (
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
	"time"
)

func init() {
	rand.Seed(5)
}

func tolerance(value, control, tolerance int64) bool {
	if value > (control + tolerance) {
		return false
	}
	if value < (control - tolerance) {
		return false
	}
	return true
}

func count(values []int64) int64 {
	c := NewCounter()
	for _, v := range values {
		c.Update(v)
	}
	return c.Snapshot()
}

func reference(values []int64) (total int64) {
	for _, v := range values {
		total += v
	}
	return total
}

func TestCounter(t *testing.T) {
	// Yes, this is really close to testing golang "sync/atomic" package.
	if err := quick.CheckEqual(count, reference, nil); err != nil {
		t.Error(err)
	}
}

func expectedRate(total int64, r *Rate, t *testing.T) {
	x := calculateRate(total, r.time)
	v := calculateRate(r.count.count, r.time)
	if !tolerance(v, x, x/20) {
		t.Error("invalid rate")
	}
}

func calculateRate(c, t int64) int64 {
	now := time.Now().UnixNano()
	return Ceil(float64(c) / rateScale / float64(now-t))
}

func TestRate(t *testing.T) {
	r := NewRate()
	t0 := time.Now()
	n := 10000
	total := int64((n * (n + 1)) / 2)
	for i := 0; i < n; i++ {
		r.Update(int64(i))
	}
	expectedRate(total, r, t)
	time.Sleep(10 * time.Millisecond)
	expectedRate(total, r, t)

	s, d := r.Snapshot(), time.Since(t0)
	m := Ceil(float64(total) / (float64(d) * rateScale))
	if pm := m / 500; !tolerance(s, m, pm) {
		t.Errorf("snapshot should be the mean, wants %d, got %d (Δ%d ±%d)", s, m, s-m, pm)
	}
	if r.Snapshot() != 0 {
		t.Error("rate should be zero")
	}
}

var reservoirTests = []struct {
	updates  []int64
	snapshot SampleSlice
}{
	{
		updates:  []int64{1},
		snapshot: SampleSlice{1},
	},
	{
		updates:  []int64{1, -10, 23},
		snapshot: SampleSlice{-10, 1, 23},
	},
	{
		updates:  []int64{1, -10, 23, 18},
		snapshot: SampleSlice{-10, 1, 18},
	},
}

func TestReservoir(t *testing.T) {
	r := NewReservoir(3)
	for i, rt := range reservoirTests {
		for _, u := range rt.updates {
			r.Update(u)
		}
		s := r.Snapshot()
		if !reflect.DeepEqual(s, rt.snapshot) {
			t.Errorf("%d: wants %v got %v", i, rt.snapshot, s)
		}
	}
}

func TestGauge(t *testing.T) {
	g := NewGauge()
	g.Update(2)
	s := g.Snapshot()
	if s != 2 {
		t.Error("gauge didn't store new value")
	}
}

func TestDerive(t *testing.T) {
	d := NewDerive(10)
	time.Sleep(10 * time.Millisecond)
	d.Update(15)
	if d.value != 15 {
		t.Error("previous value not updated")
	}
}

func TestTimer(t *testing.T) {
	tm := NewTimer(-1)
	tm.Time(func() { time.Sleep(50e6) })
	s := tm.Snapshot()
	if !tolerance(s[0], 50, 10) {
		t.Error("timer data is out of range")
	}
}
