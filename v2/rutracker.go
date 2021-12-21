package rutracker

import (
	"errors"
	"net/http"
)

var (
	ErrBadResponse = errors.New("bad response")
	ErrNotFound    = errors.New("object not found")
)

type Client struct {
	httpClient *http.Client
}

func New(httpClient *http.Client) (*Client, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		httpClient: httpClient,
	}, nil
}
