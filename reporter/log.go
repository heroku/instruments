// Package reporter provides default reporting functionnality.
package reporter

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/heroku/instruments"
)

// Log logs metrics using logfmt every given duration.
func Log(source string, r *Registry, d time.Duration) {
	for {
		time.Sleep(d)
		var parts []string
		for k, m := range r.Instruments() {
			switch i := m.(type) {
			case instruments.Discrete:
				s := i.Snapshot()
				parts = append(parts, fmt.Sprintf("sample#%s=%d", k, s))
			case instruments.Sample:
				s := instruments.Quantile(i.Snapshot(), 0.95)
				parts = append(parts, fmt.Sprintf("sample#%s=%d", k, s))
			}
		}
		log.Println(fmt.Sprintf("source=%s", source), strings.Join(parts, " "))
	}
}
