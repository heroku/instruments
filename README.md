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

## Composability

You can create custom instruments or compose new instruments form the built-in instruments as long as they implements the Sample or Discrete interfaces.

Registry also use theses interfaces, creating a custom Reporter should be trivial, for example:

```go
for k, m := range r.Instruments() {
  switch i := m.(type) {
  case instruments.Discrete:
    s := i.Snapshot()
    report(k, s)
  case instruments.Sample:
    s := instruments.Quantile(i.Snapshot(), 0.95)
    report(k, s)
  }
}
```` 