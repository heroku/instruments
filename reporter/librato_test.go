package reporter

import "time"

func ExampleLibrato() {
	registry := NewRegistry()
	go Librato("account@librato.com", "<token>", "source", registry, time.Minute)
}
