package datadog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// DefaultURL is the default series URL the client sends metric data to
const DefaultURL = "https://app.datadoghq.com/api/v1/series"

type Client struct {
	apiKey string

	// Hostname can be customised.
	// Default: set via os.Hostname()
	Hostname string

	// URL is the series URL to push data to.
	// Default: DefaultURL
	URL string
}

// NewClient creates a new API client.
func NewClient(apiKey string) *Client {
	hostname, _ := os.Hostname()

	return &Client{
		apiKey: apiKey,

		Hostname: hostname,
		URL:      DefaultURL,
	}
}

// Post delivers a metrics snapshot to datadog
func (c *Client) Post(metrics []*Metric) error {
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(&postMessage{metrics})
	if err != nil {
		return err
	}

	url := c.URL + "?api_key=" + c.apiKey
	resp, err := http.Post(url, "application/json", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("datadog: bad API response: %s", resp.Status)
	}
	return nil
}

type postMessage struct {
	Series []*Metric `json:"series,omitempty"`
}

// --------------------------------------------------------------------

// Metric represents a flushed metric
type Metric struct {
	Name   string           `json:"metric"`
	Points [][2]interface{} `json:"points"`
	Host   string           `json:"host,omitempty"`
	Tags   []string         `json:"tags,omitempty"`
}

// BuildMetric builds a metric record
func BuildMetric(name string, tags []string, ts int64, val interface{}) *Metric {
	return &Metric{
		Name:   name,
		Points: [][2]interface{}{[2]interface{}{ts, val}},
		Tags:   tags,
	}
}
