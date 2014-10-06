package instruments

import (
	"fmt"
	"reflect"
	"testing"
	"testing/quick"
	"time"
)

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

func ExampleCounter() {
	counter := NewCounter()
	counter.Update(20)
	counter.Update(25)
	s := counter.Snapshot()
	fmt.Println(s)
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
	n := 10000
	total := int64((n * (n + 1)) / 2)
	t0 := time.Now().UnixNano()
	for i := 0; i < n; i++ {
		r.Update(int64(i))
	}
	expectedRate(total, r, t)
	time.Sleep(10 * time.Millisecond)
	expectedRate(total, r, t)
	s := r.Snapshot()
	t1 := time.Now().UnixNano()
	m := Ceil(float64(total) / (float64(t1-t0) * rateScale))
	if !tolerance(s, m, m/1000) {
		t.Errorf("snapshot should be the mean, wants %d, got %d", s, m)
	}
	if r.Snapshot() != 0 {
		t.Error("rate should be zero")
	}
}

func ExampleRate() {
	rate := NewRate()
	rate.Update(20)
	rate.Update(25)
	s := rate.Snapshot()
	fmt.Println(s)
}

var reservoirTests = []struct {
	updates  []int64
	snapshot []int64
}{
	{
		updates:  []int64{1},
		snapshot: []int64{1},
	},
	{
		updates:  []int64{1, -10, 23},
		snapshot: []int64{-10, 1, 23},
	},
	{
		updates:  []int64{1, -10, 23, 18},
		snapshot: []int64{-10, 1, 18},
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

func ExampleReservoir() {
	reservoir := NewReservoir(-1)
	reservoir.Update(12)
	reservoir.Update(54)
	reservoir.Update(34)
	s := reservoir.Snapshot()
	fmt.Println(Quantile(s, 0.99))
}

func TestGauge(t *testing.T) {
	g := NewGauge(1)
	g.Update(2)
	s := g.Snapshot()
	if s != 2 {
		t.Error("gauge didn't store new value")
	}
}

func ExampleGauge() {
	gauge := NewGauge(34)
	gauge.Update(35)
	s := gauge.Snapshot()
	fmt.Println(s)
}

func TestDerive(t *testing.T) {
	d := NewDerive(10)
	time.Sleep(10 * time.Millisecond)
	d.Update(15)
	if d.value != 15 {
		t.Error("previous value not updated")
	}
}

func ExampleDerive() {
	derive := NewDerive(34)
	derive.Update(56)
	derive.Update(78)
	s := derive.Snapshot()
	fmt.Println(s)
}

func TestTimer(t *testing.T) {
	tm := NewTimer(-1)
	tm.Time(func() { time.Sleep(50e6) })
	s := tm.Snapshot()
	if !tolerance(s[0], 50, 10) {
		t.Error("timer data is out of range")
	}
}

func ExampleTimer() {
	timer := NewTimer(-1)
	ts := time.Now()
	time.Sleep(10 * time.Second)
	timer.Update(time.Since(ts))
	s := timer.Snapshot()
	fmt.Println(Quantile(s, 0.99))
}

func ExampleTimer_Time() {
	timer := NewTimer(-1)
	timer.Time(func() {
		time.Sleep(10 * time.Second)
	})
	s := timer.Snapshot()
	fmt.Println(Quantile(s, 0.99))
}

func BenchmarkCounter(b *testing.B) {
	c := NewCounter()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Update(int64(i))
		c.Snapshot()
	}
}

func BenchmarkRate(b *testing.B) {
	r := NewRate()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Update(int64(i))
		r.Snapshot()
	}
}

func BenchmarkReservoir(b *testing.B) {
	r := NewReservoir(-1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Update(int64(i))
		r.Snapshot()
	}
}
