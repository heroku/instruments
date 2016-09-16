package datadog

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/bsm/instruments"
)

func init() {
	unixTime = func() int64 { return 1414141414 }
}

func TestReporter(t *testing.T) {
	cnt := instruments.NewCounter()
	testReporter(func(rep *Reporter, body *bytes.Buffer) {
		assertNoError(t, rep.Prep())
		assertNoError(t, rep.Discrete("cnt", []string{"a", "b"}, cnt))
		assertNoError(t, rep.Discrete("cnt", []string{"a", "c"}, cnt))
		assertNoError(t, rep.Discrete("cnt", []string{"b", "c"}, cnt))
		assertNoError(t, rep.Flush())
		assertJSON(t, body.String(), `{"series":[
			{"metric":"cnt","points":[[1414141414,0]],"tags":["a","b"],"host":"test.host"},
			{"metric":"cnt","points":[[1414141414,0]],"tags":["a","c"],"host":"test.host"},
			{"metric":"cnt","points":[[1414141414,0]],"tags":["b","c"],"host":"test.host"}
		]}`)
	})
}

func TestReporter_Flush(t *testing.T) {
	cnt1a := instruments.NewCounter()
	cnt1a.Update(3)
	cnt2a := instruments.NewCounter()
	cnt2a.Update(7)
	cnt2b := instruments.NewCounter()
	cnt2b.Update(5)
	tmr1a := instruments.NewTimer(5)
	tmr1a.Update(time.Second)

	testReporter(func(rep *Reporter, body *bytes.Buffer) {
		// First flush
		assertNoError(t, rep.Prep())
		assertNoError(t, rep.Discrete("cnt1", []string{"a"}, cnt1a))
		assertNoError(t, rep.Discrete("cnt2", []string{"a"}, cnt2a))
		assertNoError(t, rep.Sample("tmr1", []string{"a"}, tmr1a))
		assertNoError(t, rep.Flush())
		assertJSON(t, body.String(), `{"series":[
			{"metric":"cnt1","points":[[1414141414,3]],"tags":["a"],"host":"test.host"},
			{"metric":"cnt2","points":[[1414141414,7]],"tags":["a"],"host":"test.host"},
			{"metric":"tmr1.p95","points":[[1414141414,1000]],"tags":["a"],"host":"test.host"},
			{"metric":"tmr1.p99","points":[[1414141414,1000]],"tags":["a"],"host":"test.host"}
		]}`)

		// Second flush
		cnt1a.Update(2)
		assertNoError(t, rep.Prep())
		assertNoError(t, rep.Discrete("cnt1", []string{"a"}, cnt1a))
		assertNoError(t, rep.Discrete("cnt2", []string{"b"}, cnt2b))
		assertNoError(t, rep.Flush())
		assertJSON(t, body.String(), `{"series":[
			{"metric":"cnt1","points":[[1414141414,2]],"tags":["a"],"host":"test.host"},
			{"metric":"cnt2","points":[[1414141414,5]],"tags":["b"],"host":"test.host"},
			{"metric":"cnt2","points":[[1414141414,0]],"tags":["a"],"host":"test.host"}
		]}`)

		// Third flush
		cnt2b.Update(9)
		assertNoError(t, rep.Discrete("cnt2", []string{"b"}, cnt2b))
		assertNoError(t, rep.Prep())
		assertNoError(t, rep.Flush())
		assertJSON(t, body.String(), `{"series":[
			{"metric":"cnt2","points":[[1414141414,9]],"tags":["b"],"host":"test.host"},
			{"metric":"cnt1","points":[[1414141414,0]],"tags":["a"],"host":"test.host"}
		]}`)

		// Final flush
		assertNoError(t, rep.Prep())
		assertNoError(t, rep.Flush())
		assertJSON(t, body.String(), `{"series":[
			{"metric":"cnt2","points":[[1414141414,0]],"tags":["b"],"host":"test.host"}
		]}`)
	})
}

func testReporter(cb func(*Reporter, *bytes.Buffer)) {
	body := &bytes.Buffer{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		body.Reset()
		if _, err := io.Copy(body, r.Body); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusAccepted)
		}
	}))
	defer server.Close()

	rep := New("BOGUS")
	rep.Client.URL = server.URL
	rep.Hostname = "test.host"
	cb(rep, body)
}

func assertJSON(t *testing.T, have, want string) {
	var h, w map[string]interface{}
	if err := json.Unmarshal([]byte(have), &h); err != nil {
		t.Fatal("unable to decode 'have' JSON", err)
	}
	if err := json.Unmarshal([]byte(want), &w); err != nil {
		t.Fatal("unable to decode 'want' JSON", err)
	}

	if !reflect.DeepEqual(h, w) {
		norm := strings.NewReplacer(" ", "", "\t", "", "\n", "")
		t.Errorf("want:\n%s\nhave:\n%s", norm.Replace(want), norm.Replace(have))
	}
}

func assertNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatal("wanted no error, but got", err.Error())
	}
}
