package instruments

import (
	"math/rand"
	"testing"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Instruments", func() {

	DescribeTable("Reservoir",
		func(vv []float64, x float64) {
			i := NewReservoir()
			for _, v := range vv {
				i.Update(v)
			}
			Expect(i.Snapshot().Mean()).To(BeNumerically("~", x, 0.1))
		},
		Entry("single", []float64{1}, 1.0),
		Entry("a few", []float64{1, -10, 23}, 4.7),
	)

	ginkgo.It("should update counters", func() {
		c := NewCounter()
		c.Update(7)
		c.Update(12)
		Expect(c.Snapshot()).To(Equal(19.0))
		Expect(c.Snapshot()).To(Equal(0.0))

		for i := 1; i < 100; i++ {
			c.Update(float64(i))
		}
		Expect(c.Snapshot()).To(Equal(4950.0))
	})

	ginkgo.It("should update gauges", func() {
		g := NewGauge()
		g.Update(7)
		g.Update(12)
		Expect(g.Snapshot()).To(Equal(12.0))
	})

	ginkgo.It("should update derives", func() {
		d := NewDerive(10)
		d.Update(7)
		time.Sleep(10 * time.Millisecond)
		d.Update(12)
		Expect(d.Snapshot()).To(BeNumerically("~", 200, 20))
	})

	ginkgo.It("should update rates", func() {
		r := NewRate()
		for i := 0; i < 100; i++ {
			r.Update(float64(i))
		}
		Expect(r.Snapshot()).To(BeNumerically(">", 1e6))
		Eventually(r.Snapshot).Should(BeNumerically("<", 1e5))
	})

	ginkgo.It("should update timers", func() {
		t := NewTimer()
		for i := 0; i < 100; i++ {
			t.Update(time.Second * time.Duration(i))
		}
		Expect(t.Snapshot().Mean()).To(Equal(49500.0))
	})

})

// --------------------------------------------------------------------

func init() {
	rand.Seed(5)
}

func TestSuite(t *testing.T) {
	RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "instruments")
}

// --------------------------------------------------------------------

// test exports

func (r *Registry) GetInstruments() map[string]interface{}            { return r.instruments }
func (r *Registry) SetInstruments(instruments map[string]interface{}) { r.instruments = instruments }
func (r *Registry) Reset() int                                        { return len(r.reset()) }

func ReleaseDistribution(d Distribution) {
	releaseDistribution(d)
}
