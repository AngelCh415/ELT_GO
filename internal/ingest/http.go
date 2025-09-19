package ingest

import (
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"time"
)

func GetJSONWithRetry(c HTTPClient, url string, dst any) error {
	var lastErr error
	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest(http.MethodGet, url, nil)
		resp, err := c.Do(req)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			defer resp.Body.Close()
			return json.NewDecoder(resp.Body).Decode(dst)
		}
		if err != nil {
			lastErr = err
		} else {
			lastErr = errors.New(resp.Status)
			resp.Body.Close()
		}
		// backoff exponencial + jitter
		sleep := time.Duration((1<<i)*100) * time.Millisecond
		sleep += time.Duration(rand.Intn(150)) * time.Millisecond
		time.Sleep(sleep)
	}
	return lastErr
}
