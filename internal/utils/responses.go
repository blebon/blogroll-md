package utils

import (
	"context"
	"net/http"

	"github.com/blebon/blogroll-md/internal/client"
)

func Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return client.Client.Do(req)
}
