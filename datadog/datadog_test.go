package datadog

import (
	"bytes"
	"compress/zlib"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Client", func() {
	var subject *Client

	var server *httptest.Server
	var last *mockServerRequest

	ginkgo.BeforeEach(func() {
		last = new(mockServerRequest)
		server = newMockServer(last)

		subject = NewClient("TEST_API_TOKEN")
		subject.URL = server.URL
	})

	ginkgo.AfterEach(func() {
		server.Close()
	})

	ginkgo.It("should post metrics", func() {
		err := subject.Post([]Metric{
			{Name: "m1", Points: [][2]interface{}{{1414141414, 27}}, Tags: []string{"a", "b"}},
			{Name: "m2", Points: [][2]interface{}{{1414141415, 0.8}}, Tags: []string{"c"}},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(last.Method).To(Equal("POST"))
		Expect(last.URL.RawQuery).To(Equal("api_key=TEST_API_TOKEN"))
		Expect(last.Body.Bytes()).To(MatchJSON(`{
			"series": [
				{"metric":"m1", "points":[[1414141414,27]],  "tags":["a","b"]},
				{"metric":"m2", "points":[[1414141415,0.8]], "tags":["c"]}
			]
		}`))

	})

})

// --------------------------------------------------------------------

func TestSuite(t *testing.T) {
	RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "instruments/datadog")
}

func init() {
	unixTime = func() int64 { return 1414141414 }
}

type mockServerRequest struct {
	Method string
	URL    *url.URL
	Body   bytes.Buffer
}

func newMockServer(last *mockServerRequest) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		last.Method = r.Method
		last.URL = r.URL
		last.Body.Reset()

		z, err := zlib.NewReader(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer z.Close()

		if _, err := io.Copy(&last.Body, z); err != nil && err != io.ErrUnexpectedEOF {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}))
}
