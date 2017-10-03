package datadog

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// DefaultURL is the default series URL the client sends metric data to
const DefaultURL = "https://app.datadoghq.com/api/v1/series"

type Client struct {
	apiKey string
	client *http.Client

	// URL is the series URL to push data to.
	// Default: DefaultURL
	URL string

	bfs, zws sync.Pool
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

	buf := c.buffer()
	defer c.bfs.Put(buf)

	bfz := c.zWriter(buf)
	defer c.zws.Put(bfz)
	defer bfz.Close()

	if err := json.NewEncoder(bfz).Encode(&series); err != nil {
		return err
	}
	if err := bfz.Flush(); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.URL+"?api_key="+c.apiKey, buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "deflate")

	for i := 1; i < 4; i++ {
		code, err := c.post(req)
		if err == nil || code == http.StatusForbidden || code == http.StatusUnauthorized {
			return err
		}
		time.Sleep(time.Duration(i) * 200 * time.Millisecond)
	}
	return nil
}

func (c *Client) post(req *http.Request) (int, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted, http.StatusNoContent:
		return resp.StatusCode, nil
	}
	return resp.StatusCode, fmt.Errorf("datadog: bad API response: %s", resp.Status)
}

func (c *Client) buffer() *bytes.Buffer {
	if v := c.bfs.Get(); v != nil {
		b := v.(*bytes.Buffer)
		b.Reset()
		return b
	}
	return new(bytes.Buffer)
}

func (c *Client) zWriter(w io.Writer) *zlib.Writer {
	if v := c.zws.Get(); v != nil {
		z := v.(*zlib.Writer)
		z.Reset(w)
		return z
	}
	return zlib.NewWriter(w)
}

// Metric represents a flushed metric
type Metric struct {
	Name   string           `json:"metric"`
	Points [][2]interface{} `json:"points"`
	Host   string           `json:"host,omitempty"`
	Tags   []string         `json:"tags,omitempty"`
}
