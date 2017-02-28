package datadog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// DefaultURL is the default series URL the client sends metric data to
const DefaultURL = "https://app.datadoghq.com/api/v1/series"

type Client struct {
	apiKey string
	client *http.Client

	// URL is the series URL to push data to.
	// Default: DefaultURL
	URL string
}

// NewClient creates a new API client.
func NewClient(apiKey string) *Client {
	return &Client{
		client: &http.Client{},
		apiKey: apiKey,
		URL:    DefaultURL,
	}
}

// Post delivers a metrics snapshot to datadog
func (c *Client) Post(metrics []Metric) error {
	series := struct {
		Series []Metric `json:"series,omitempty"`
	}{metrics}

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(&series); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.URL+"?api_key="+c.apiKey, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("datadog: bad API response: %s", resp.Status)
	}
	return nil
}

// Metric represents a flushed metric
type Metric struct {
	Name   string           `json:"metric"`
	Points [][2]interface{} `json:"points"`
	Host   string           `json:"host,omitempty"`
	Tags   []string         `json:"tags,omitempty"`
}
