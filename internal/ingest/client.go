package ingest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type client struct{ httpc *http.Client }

func NewHTTPClient(timeout time.Duration) HTTPClient {
	return &http.Client{Timeout: timeout}
}

func getJSON(ctx context.Context, c HTTPClient, url string, v any) error {
	if url == "" {
		return errors.New("empty url")
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("non-2xx: %d body=%s", resp.StatusCode, string(b))
	}
	dec := json.NewDecoder(resp.Body)
	return dec.Decode(v)
}
