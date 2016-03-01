package datadog

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient(t *testing.T) {
	var method, query, body string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 1024)
		n, err := r.Body.Read(buf)
		if err != nil && err != io.EOF {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		method = r.Method
		query = r.URL.RawQuery
		body = strings.TrimSpace(string(buf[:n]))
	}))
	defer server.Close()

	client := NewClient("TEST_API_TOKEN")
	client.URL = server.URL
	err := client.Post([]*Metric{
		BuildMetric("m1", []string{"a", "b"}, 1414141414, 27),
		BuildMetric("m2", []string{"c"}, 1414141415, 0.8),
	})
	if err != nil {
		t.Fatal("unable to post metrics:", err)
	}

	if method != "POST" {
		t.Errorf("expected POST but received %s", method)
	}
	if query != "api_key=TEST_API_TOKEN" {
		t.Errorf("expected API token in query, but received %s", query)
	}
	expected := `{"series":[{"metric":"m1","points":[[1414141414,27]],"tags":["a","b"]},{"metric":"m2","points":[[1414141415,0.8]],"tags":["c"]}]}`
	if body != expected {
		t.Errorf("\nexpected: %s\n     got: %s", expected, body)
	}
}
