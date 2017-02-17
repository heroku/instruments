package instruments

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("MetricID", func() {

	DescribeTable("should assemble",
		func(name string, tags []string, x string) {
			m := MetricID(name, tags)
			Expect(string(m)).To(Equal(x))
		},

		Entry("", "counter", []string{"a", "b"}, "counter|a,b"),
		Entry("", "counter", []string{"b", "a"}, "counter|a,b"),
		Entry("", "counter", nil, "counter"),
		Entry("", "counter", []string{}, "counter"),
	)

	DescribeTable("should split",
		func(metricID string, xn string, xt []string) {
			name, tags := SplitMetricID(metricID)
			Expect(name).To(Equal(xn))
			Expect(tags).To(Equal(xt))
		},

		Entry("", "counter|a,b", "counter", []string{"a", "b"}),
		Entry("", "|counter|a,b", "|counter", []string{"a", "b"}),
		Entry("", "counter", "counter", nil),
	)

})
