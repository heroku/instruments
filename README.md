# Instruments

Instruments allows you to collects metrics over discrete time intervals.

## Installation

Download and install:

```
$ go get github.com/heroku/instruments
```

Add it to your code:

```go
import "github.com/heroku/instruments"
```

## Usage

```go
timer := instruments.NewTimer(-1)

registry := instruments.NewRegistry()
registry.Register("processing-time", timer)

go reporter.Log("process", registry, time.Minute)

timer.Time(func() {
  ...
})
```
