package instruments_test

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bsm/instruments"
	"github.com/bsm/instruments/logreporter"
)

func ExampleRegistry() {
	// Create new registry instance, flushing at minutely intervals
	registry := instruments.New(time.Minute, "myapp.")
	defer registry.Close()

	// Subscribe a reporter
	logger := log.New(os.Stdout, "", log.LstdFlags)
	registry.Subscribe(logreporter.New(logger))

	// Fetch a timer
	timer := registry.Timer("processing-time", []string{"tag1", "tag2"})

	// Measure something
	start := time.Now()
	time.Sleep(20 * time.Millisecond)
	timer.Since(start)
}

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
	reservoir := instruments.NewReservoir()
	reservoir.Update(12)
	reservoir.Update(54)
	reservoir.Update(34)
	fmt.Println(reservoir.Snapshot().Quantile(0.99))
	// Output: 54
}

func ExampleGauge() {
	gauge := instruments.NewGauge()
	gauge.Update(35.6)
	fmt.Println(gauge.Snapshot())
	// Output: 35.6
}

func ExampleDerive() {
	derive := instruments.NewDerive(34)
	derive.Update(56)
	derive.Update(78)
	fmt.Println(derive.Snapshot())
}

func ExampleTimer() {
	timer := instruments.NewTimer()
	ts := time.Now()
	time.Sleep(10 * time.Millisecond)
	timer.Since(ts)
	fmt.Println(timer.Snapshot().Quantile(0.99))
}
