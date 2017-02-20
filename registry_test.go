package instruments

import (
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Registry", func() {
	var subject *Registry

	var reporter *mockReporter
	var nCounters, nRates int

	newCounter := func() interface{} {
		nCounters++
		return NewCounter()
	}
	newRate := func() interface{} {
		nRates++
		return NewRate()
	}

	ginkgo.BeforeEach(func() {
		nCounters, nRates = 0, 0
		reporter = new(mockReporter)

		subject = New(time.Minute, "myapp.", "a", "b")
		subject.Subscribe(reporter)
	})

	ginkgo.AfterEach(func() {
		Expect(subject.Close()).To(Succeed())
	})

	ginkgo.It("should register/unregister", func() {
		subject.Register("foo", []string{"a", "b"}, NewRate())
		Expect(subject.Size()).To(Equal(1))

		subject.Unregister("foo", []string{"b", "a"})
		Expect(subject.Size()).To(Equal(0))

		subject.Register("foo", []string{"a", "b"}, NewRate())
		subject.Register("foo", []string{"b", "a"}, NewRate())
		Expect(subject.Size()).To(Equal(1))

		subject.Register("bar", []string{}, NewRate())
		subject.Register("bar", nil, NewRate())
		Expect(subject.Size()).To(Equal(2))

		subject.Unregister("foo", []string{"b", "a"})
		subject.Unregister("bar", []string{"a"})
		Expect(subject.Size()).To(Equal(1))

		subject.Unregister("bar", nil)
		Expect(subject.Size()).To(Equal(0))
	})

	ginkgo.It("should get pre-registered", func() {
		rate := NewRate()
		subject.Register("foo", []string{"a", "b"}, rate)
		Expect(subject.Get("foo", []string{"a", "b"})).To(Equal(rate))
		Expect(subject.Get("foo", []string{"b", "a"})).To(Equal(rate))
		Expect(subject.Get("foo", []string{"x"})).To(BeNil())
	})

	ginkgo.It("should fetch counters", func() {
		Expect(subject.Fetch("foo", nil, newCounter)).To(BeAssignableToTypeOf(&Counter{}))
		Expect(nCounters).To(Equal(1))

		Expect(subject.Fetch("foo", nil, newCounter)).To(BeAssignableToTypeOf(&Counter{}))
		Expect(nCounters).To(Equal(1))

		Expect(subject.Fetch("foo", []string{"a"}, newCounter)).To(BeAssignableToTypeOf(&Counter{}))
		Expect(nCounters).To(Equal(2))
	})

	ginkgo.It("should fetch rates", func() {
		Expect(subject.Fetch("foo", nil, newRate)).To(BeAssignableToTypeOf(&Rate{}))
		Expect(nRates).To(Equal(1))

		Expect(subject.Fetch("foo", nil, newRate)).To(BeAssignableToTypeOf(&Rate{}))
		Expect(nRates).To(Equal(1))

		Expect(subject.Fetch("foo", []string{"a"}, newRate)).To(BeAssignableToTypeOf(&Rate{}))
		Expect(nRates).To(Equal(2))
	})

	ginkgo.It("should flush", func() {
		// force-extend tags cap
		subject.tags = append(subject.tags, "x")[:2]

		cnt1 := NewCounter()
		subject.Register("foo", []string{"c", "d"}, cnt1)
		cnt1.Update(2)
		cnt1.Update(6)
		cnt1.Update(4)
		cnt1.Update(8)

		cnt2 := NewCounter()
		subject.Register("foo", []string{"e"}, cnt2)
		cnt2.Update(7)

		resv := NewReservoir()
		subject.Register("bar", []string{"f", "g"}, resv)
		resv.Update(2)
		resv.Update(6)
		resv.Update(4)
		resv.Update(8)

		cnt3 := NewCounter()
		subject.Register("|custom.foo", nil, cnt3)
		cnt3.Update(11)

		Expect(subject.Flush()).To(Succeed())
		Expect(reporter.Prepped).To(BeTrue())
		Expect(reporter.Flushed).To(Equal(map[string]float64{
			"myapp.foo|a,b,c,d": 20,
			"myapp.foo|a,b,e":   7,
			"myapp.bar|a,b,f,g": 5,
			"custom.foo|a,b":    11,
		}))
	})

	ginkgo.It("should reset", func() {
		subject.Register("foo", []string{"a", "b"}, NewRate())
		Expect(subject.Size()).To(Equal(1))

		snap := subject.reset()
		Expect(snap).To(HaveLen(1))
		Expect(subject.Size()).To(Equal(0))
	})

})

// --------------------------------------------------------------------

type mockReported struct {
	Name  string
	Tags  []string
	Value float64
}

type mockReporter struct {
	Data    []mockReported
	Prepped bool
	Flushed map[string]float64
}

func (m *mockReporter) Prep() error {
	m.Prepped = true
	return nil
}

func (m *mockReporter) Flush() error {
	m.Flushed = make(map[string]float64, len(m.Data))
	for _, i := range m.Data {
		m.Flushed[MetricID(i.Name, i.Tags)] = i.Value
	}
	return nil
}

func (m *mockReporter) Discrete(name string, tags []string, val float64) error {
	m.Data = append(m.Data, mockReported{
		Name:  name,
		Tags:  tags,
		Value: val,
	})
	return nil
}

func (m *mockReporter) Sample(name string, tags []string, dist Distribution) error {
	m.Data = append(m.Data, mockReported{
		Name:  name,
		Tags:  tags,
		Value: dist.Mean(),
	})
	return nil
}
