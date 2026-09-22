package clients

import (
	"context"
	"fmt"
	"net/http"
)

type URLHealthChecker struct {
	url    string
	client *http.Client
}

func NewURLHealthChecker(url string, client *http.Client) *URLHealthChecker {
	return &URLHealthChecker{url: url, client: client}
}

func (c *URLHealthChecker) Check(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusInternalServerError || resp.StatusCode < http.StatusOK {
		return fmt.Errorf("dependency health returned %d", resp.StatusCode)
	}
	return nil
}
