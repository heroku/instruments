package runtime_test

import (
	"fmt"
	"time"

	"github.com/bsm/instruments/runtime"
)

func ExamplePauses() {
	pauses := runtime.NewPauses(512)
	time.Sleep(time.Minute)
	pauses.Update()
	fmt.Println(pauses.Snapshot().Quantile(0.95))
}
