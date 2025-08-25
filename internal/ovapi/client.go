package ovapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sony/gobreaker"
)

const (
	baseURL = "http://v0.ovapi.nl"
)

// Client is a client for the OVapi.
type Client struct {
	httpClient *http.Client
	cb         *gobreaker.CircuitBreaker
}

// NewClient creates a new OVapi client.
func NewClient() *Client {
	st := gobreaker.Settings{
		Name:        "OVapi",
		MaxRequests: 5,
		Interval:    10 * time.Second,
		Timeout:     5 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 3
		},
	}

	return &Client{
		httpClient: &http.Client{},
		cb:         gobreaker.NewCircuitBreaker(st),
	}
}

// GetDepartures fetches departures for a specific stop.
func (c *Client) GetDepartures(stopCode string) (map[string]interface{}, error) {
	body, err := c.cb.Execute(func() (interface{}, error) {
		url := fmt.Sprintf("%s/tpc/%s", baseURL, stopCode)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		var data map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return nil, err
		}

		return data, nil
	})

	if err != nil {
		return nil, err
	}

	return body.(map[string]interface{}), nil
}