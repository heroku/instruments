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
		func(vv []int64, x Distribution) {
			i := NewReservoir(4)
			for _, v := range vv {
				i.Update(v)
			}
			Expect(i.Snapshot()).To(Equal(x))
		},
		Entry("blank", []int64{}, mockDist()),
		Entry("single", []int64{1}, mockDist(1)),
		Entry("a few", []int64{1, -10, 23}, mockDist(-10, 1, 23)),
	)

	ginkgo.It("should update counters", func() {
		c := NewCounter()
		c.Update(7)
		c.Update(12)
		Expect(c.Snapshot()).To(Equal(int64(19)))
	})

	ginkgo.It("should update gauges", func() {
		g := NewGauge()
		g.Update(7)
		g.Update(12)
		Expect(g.Snapshot()).To(Equal(int64(12)))
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
			r.Update(int64(i))
		}
		Expect(r.Snapshot()).To(BeNumerically(">", 1e6))
		Eventually(r.Snapshot).Should(BeNumerically("<", 1e5))
	})

	ginkgo.It("should update timers", func() {
		t := NewTimer(-1)
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
