package reporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/heroku/instruments"
)

const defaultLibratoURL = "https://metrics-api.librato.com/v1/metrics"

type batch struct {
	Gauges      []map[string]interface{} `json:"gauges,omitempty"`
	MeasureTime int64                    `json:"measure_time"`
	Source      string                   `json:"source"`
}

type client struct {
	Email string
	Token string
}

func (c *client) URL() string {
	uri, found := os.LookupEnv("LIBRATO_API_URL")
	if !found {
		return defaultLibratoURL
	}

	return uri
}

func (c *client) Post(b batch) error {
	if len(b.Gauges) == 0 {
		return nil
	}

	body, err := json.Marshal(b)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprint(c.URL()), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.Email, c.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("librato: failed to communicate %d / %s", resp.StatusCode, resp.Status)
	}

	return nil
}

// Librato logs metrics to librato every given duration.
func Librato(email, token, source string, r *Registry, d time.Duration) {
	client := &client{
		Email: email,
		Token: token,
	}

	for now := range time.Tick(d) {
		b := batch{
			Source:      source,
			Gauges:      []map[string]interface{}{},
			MeasureTime: now.Unix(),
		}
		for k, m := range r.Instruments() {
			var s int64
			switch i := m.(type) {
			case instruments.Discrete:
				s = i.Snapshot()
			case instruments.Sample:
				s = instruments.Quantile(i.Snapshot(), 0.95)
			}
			b.Gauges = append(b.Gauges, map[string]interface{}{
				"name":   k,
				"value":  float64(s),
				"period": d.Seconds(),
			})
		}

		err := client.Post(b)
		if err != nil {
			log.Println(err)
		}
	}
}
