package instruments

import (
	"math"
	"math/rand"
	"sort"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Distribution", func() {
	blnk := hist()
	std := hist(39, 15, 43, 7, 43, 36, 47, 6, 40, 49, 41)

	DescribeTable("Quantile",
		func(d Distribution, q float64, x float64) {
			Expect(d.Quantile(q)).To(BeNumerically("~", x, 0.1))
		},

		Entry("blank", blnk, 0.95, 0.0),
		Entry("0%", std, 0.0, 6.0),
		Entry("25%", std, 0.25, 19.6),
		Entry("50%", std, 0.5, 39.8),
		Entry("75%", std, 0.75, 44.3),
		Entry("95%", std, 0.95, 47.2),
		Entry("100%", std, 1.0, 49.0),
		Entry("bad input", std, -1.0, 0.0),
	)

	// inspired by https://github.com/aaw/histosketch/commit/d8284aa#diff-11101c92fbb1d58ccf30ca49764bf202R180
	// released into the public domain
	ginkgo.It("should accurately predict quantile", func() {
		N := 10000
		Q := []float64{0.0001, 0.001, 0.01, 0.1, 0.25, 0.35, 0.65, 0.75, 0.9, 0.99, 0.999, 0.9999}

		for seed := 0; seed < 10; seed++ {
			r := rand.New(rand.NewSource(int64(seed)))
			s := newHistogram(16)   // sketch
			x := make([]float64, N) // exact

			for i := 0; i < N; i++ {
				num := r.NormFloat64()
				s.Add(num)
				x[i] = num
			}
			sort.Float64s(x)

			for _, q := range Q {
				sQ := s.Quantile(q)
				xQ := x[int(float64(len(x))*q)]
				re := math.Abs((sQ - xQ) / xQ)

				Expect(re).To(BeNumerically("<", 0.09),
					"s.Quantile(%v) (got %.3f, want %.3f with seed = %v)", q, sQ, xQ, seed,
				)
			}
		}
	})

	ginkgo.It("should calc mean", func() {
		Expect(blnk.Mean()).To(Equal(0.0))
		Expect(std.Mean()).To(BeNumerically("~", 33.27, 0.01))
	})

	ginkgo.It("should calc count", func() {
		Expect(blnk.Count()).To(Equal(0))
		Expect(std.Count()).To(Equal(11))
	})

	ginkgo.It("should calc min", func() {
		Expect(blnk.Min()).To(Equal(float64(0)))
		Expect(std.Min()).To(Equal(float64(6)))
	})

	ginkgo.It("should calc max", func() {
		Expect(blnk.Max()).To(Equal(float64(0)))
		Expect(std.Max()).To(Equal(float64(49)))
	})

	ginkgo.It("should add", func() {
		h := std.(*histogram)
		Expect(h.bins).To(HaveLen(4))
		Expect(h.bins).To(HaveCap(5))
		Expect(h.bins).To(Equal([]histogramBin{
			{w: -2, v: 6.5},
			{w: 1, v: 15},
			{w: -4, v: 39},
			{w: -4, v: 45.5},
		}))
	})

})

func hist(vv ...float64) Distribution {
	h := newHistogram(4)
	for _, v := range vv {
		h.Add(v)
	}
	return h
}
