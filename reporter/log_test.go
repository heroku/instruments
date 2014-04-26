package reporter

import "time"

func ExampleLog() {
	registry := NewRegistry()
	go Log("source", registry, time.Minute)
}
