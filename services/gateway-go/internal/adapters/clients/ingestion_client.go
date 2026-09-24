package clients

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type IngestionClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewIngestionClient(baseURL string, httpClient *http.Client) *IngestionClient {
	return &IngestionClient{baseURL: strings.TrimRight(baseURL, "/"), httpClient: httpClient}
}

func (c *IngestionClient) Proxy(ctx context.Context, method, path, contentType string, body io.Reader) (int, []byte, error) {
	request, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return 0, nil, err
	}
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return 0, nil, err
	}
	defer response.Body.Close()
	payload, err := io.ReadAll(response.Body)
	if err != nil {
		return 0, nil, err
	}
	if response.StatusCode >= http.StatusInternalServerError {
		return response.StatusCode, nil, fmt.Errorf("ingestion API returned status %d", response.StatusCode)
	}
	return response.StatusCode, payload, nil
}
