package datadog

import (
	"net/http/httptest"

	"github.com/bsm/instruments"
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Reporter", func() {
	var subject *Reporter

	var server *httptest.Server
	var last *mockServerRequest

	ginkgo.BeforeEach(func() {
		last = new(mockServerRequest)
		server = newMockServer(last)

		subject = New("BOGUS")
		subject.Client.URL = server.URL
		subject.Hostname = "test.host"
	})

	ginkgo.AfterEach(func() {
		server.Close()
	})

	ginkgo.It("should support reporter cycle", func() {
		Expect(subject.Prep()).To(Succeed())
		Expect(subject.Discrete("cnt", []string{"a", "b"}, 0)).To(Succeed())
		Expect(subject.Discrete("cnt", []string{"a", "c"}, 1)).To(Succeed())
		Expect(subject.Discrete("cnt", []string{"b", "c"}, 2)).To(Succeed())
		Expect(subject.Flush()).To(Succeed())

		Expect(last.Body.Bytes()).To(MatchJSON(`{
			"series":[
				{"metric":"cnt","points":[[1414141414,0]],"tags":["a","b"],"host":"test.host"},
				{"metric":"cnt","points":[[1414141414,1]],"tags":["a","c"],"host":"test.host"},
				{"metric":"cnt","points":[[1414141414,2]],"tags":["b","c"],"host":"test.host"}
			]
		}`))
	})

	ginkgo.It("should expire old metrics after flush", func() {
		Expect(subject.Prep()).To(Succeed())
		Expect(subject.Discrete("cnt1", []string{"a"}, 3)).To(Succeed())
		Expect(subject.Discrete("cnt2", []string{"a"}, 7)).To(Succeed())
		Expect(subject.Sample("tmr1", []string{"a"}, mockDistribution{})).To(Succeed())
		Expect(subject.Flush()).To(Succeed())

		// after 1st flush
		Expect(last.Body.Bytes()).To(MatchJSON(`{
			"series":[
				{"metric":"cnt1","points":[[1414141414,3]],"tags":["a"],"host":"test.host"},
				{"metric":"cnt2","points":[[1414141414,7]],"tags":["a"],"host":"test.host"},
				{"metric":"tmr1.p95","points":[[1414141414,100.1]],"tags":["a"],"host":"test.host"},
				{"metric":"tmr1.p99","points":[[1414141414,100.1]],"tags":["a"],"host":"test.host"}
			]
		}`))

		Expect(subject.Prep()).To(Succeed())
		Expect(subject.Discrete("cnt1", []string{"a"}, 2)).To(Succeed())
		Expect(subject.Discrete("cnt2", []string{"b"}, 5)).To(Succeed())
		Expect(subject.Flush()).To(Succeed())

		// after 2nd flush
		Expect(last.Body.Bytes()).To(MatchJSON(`{
			"series":[
				{"metric":"cnt1","points":[[1414141414,2]],"tags":["a"],"host":"test.host"},
				{"metric":"cnt2","points":[[1414141414,5]],"tags":["b"],"host":"test.host"},
				{"metric":"cnt2","points":[[1414141414,0]],"tags":["a"],"host":"test.host"}
			]
		}`))

		Expect(subject.Prep()).To(Succeed())
		Expect(subject.Discrete("cnt2", []string{"b"}, 9)).To(Succeed())
		Expect(subject.Flush()).To(Succeed())

		// after 3nd flush
		Expect(last.Body.Bytes()).To(MatchJSON(`{
			"series":[
				{"metric":"cnt2","points":[[1414141414,9]],"tags":["b"],"host":"test.host"},
				{"metric":"cnt1","points":[[1414141414,0]],"tags":["a"],"host":"test.host"}
			]
		}`))

		Expect(subject.Prep()).To(Succeed())
		Expect(subject.Flush()).To(Succeed())

		// after 4th flush
		Expect(last.Body.Bytes()).To(MatchJSON(`{
			"series":[
				{"metric":"cnt2","points":[[1414141414,0]],"tags":["b"],"host":"test.host"}
			]
		}`))
	})

})

type mockDistribution struct {
	instruments.Distribution
}

func (mockDistribution) Quantile(_ float64) float64 { return 100.1 }
