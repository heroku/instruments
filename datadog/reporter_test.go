package datadog

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bsm/instruments"
)

func init() {
	unixTime = func() int64 { return 1414141414 }
}

func TestReporter(t *testing.T) {
	postBody := &bytes.Buffer{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		postBody.Reset()
		if _, err := io.Copy(postBody, r.Body); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusAccepted)
		}
	}))
	defer server.Close()

	cnt1 := instruments.NewCounter()
	cnt2 := instruments.NewCounter()
	tmr1 := instruments.NewTimer(5)

	rep := New("BOGUS")
	rep.Client.URL = server.URL

	// First flush
	cnt1.Update(3)
	cnt2.Update(7)
	tmr1.Update(time.Second)
	assertNoError(t, rep.Prep())
	assertNoError(t, rep.Discrete("cnt1", []string{"a"}, cnt1))
	assertNoError(t, rep.Discrete("cnt2", []string{"a"}, cnt2))
	assertNoError(t, rep.Sample("tmr1", []string{"a"}, tmr1))
	assertNoError(t, rep.Flush())

	if have, want := normJSON.Replace(postBody.String()), normJSON.Replace(`{"series":[
		{"metric":"cnt1","points":[[1414141414,3]],"tags":["a"]},
		{"metric":"cnt2","points":[[1414141414,7]],"tags":["a"]},
		{"metric":"tmr1.p95","points":[[1414141414,1000]],"tags":["a"]},
		{"metric":"tmr1.p99","points":[[1414141414,1000]],"tags":["a"]}
	]}`); have != want {
		t.Fatalf("want:\n%s\nhave:\n%s", want, have)
	}

	// Second flush
	cnt1.Update(2)
	cnt2.Update(9)
	assertNoError(t, rep.Prep())
	assertNoError(t, rep.Discrete("cnt1", []string{"a"}, cnt1))
	assertNoError(t, rep.Discrete("cnt2", []string{"b"}, cnt2))
	assertNoError(t, rep.Flush())

	if have, want := normJSON.Replace(postBody.String()), normJSON.Replace(`{"series":[
		{"metric":"cnt1","points":[[1414141414,2]],"tags":["a"]},
		{"metric":"cnt2","points":[[1414141414,9]],"tags":["b"]},
		{"metric":"cnt2","points":[[1414141414,0]],"tags":["a"]}
	]}`); have != want {
		t.Fatalf("want:\n%s\nhave:\n%s", want, have)
	}

	// Third flush
	assertNoError(t, rep.Prep())
	assertNoError(t, rep.Flush())

	if have, want := normJSON.Replace(postBody.String()), normJSON.Replace(`{"series":[
		{"metric":"cnt1","points":[[1414141414,0]],"tags":["a"]},
		{"metric":"cnt2","points":[[1414141414,0]],"tags":["b"]}
	]}`); have != want {
		t.Fatalf("want:\n%s\nhave:\n%s", want, have)
	}
}

var normJSON = strings.NewReplacer(" ", "", "\n", "", "\t", "")

func assertNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatal("wanted no error, but got", err.Error())
	}
}
