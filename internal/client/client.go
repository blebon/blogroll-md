package client

import (
	"net/http"
	"time"
)

func init() {
	Client = &http.Client{
		Timeout: 10 * time.Second,
	}
}

//mockery:generate: true
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var Client HTTPClient
