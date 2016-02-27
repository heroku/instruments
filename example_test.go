package instruments_test

import (
	"fmt"
	"time"

	"github.com/bsm/instruments"
)

func ExampleCounter() {
	counter := instruments.NewCounter()
	counter.Update(20)
	counter.Update(25)
	fmt.Println(counter.Snapshot())
	// Output: 45
}

func ExampleRate() {
	rate := instruments.NewRate()
	rate.Update(20)
	rate.Update(25)
	fmt.Println(rate.Snapshot())
}

func ExampleReservoir() {
	reservoir := instruments.NewReservoir(-1)
	reservoir.Update(12)
	reservoir.Update(54)
	reservoir.Update(34)
	fmt.Println(reservoir.Snapshot().Quantile(0.99))
	// Output: 54
}

func ExampleGauge() {
	gauge := instruments.NewGauge(34)
	gauge.Update(35)
	fmt.Println(gauge.Snapshot())
	// Output: 35
}

func ExampleDerive() {
	derive := instruments.NewDerive(34)
	derive.Update(56)
	derive.Update(78)
	fmt.Println(derive.Snapshot())
}

func ExampleTimer() {
	timer := instruments.NewTimer(-1)
	ts := time.Now()
	time.Sleep(10 * time.Millisecond)
	timer.Since(ts)
	fmt.Println(timer.Snapshot().Quantile(0.99))
}

func ExampleTimer_Time() {
	timer := instruments.NewTimer(-1)
	timer.Time(func() {
		time.Sleep(10 * time.Millisecond)
	})
	fmt.Println(timer.Snapshot().Quantile(0.99))
}
