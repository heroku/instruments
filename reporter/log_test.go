package reporter

import (
	"time"

	"github.com/heroku/instruments"
)

func ExampleLog() {
	registry := instruments.NewRegistry()
	go Log("source", registry, time.Minute)
}
