package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// HTTPClient wraps http.Client with rate limiting, retries, and logging
// that all provider adapters can share.
type HTTPClient struct {
	client      *http.Client
	baseURL     string
	rateLimiter *RateLimiter
	headers     map[string]string
	mu          sync.RWMutex
}

// NewHTTPClient creates an HTTP client for API calls.
func NewHTTPClient(baseURL string, timeoutSec int, rateRPS int) *HTTPClient {
	if timeoutSec <= 0 {
		timeoutSec = 30
	}
	if rateRPS <= 0 {
		rateRPS = 5
	}

	return &HTTPClient{
		client: &http.Client{
			Timeout: time.Duration(timeoutSec) * time.Second,
		},
		baseURL:     baseURL,
		rateLimiter: NewRateLimiter(rateRPS),
		headers:     make(map[string]string),
	}
}

// SetHeader sets a persistent header for all requests.
func (c *HTTPClient) SetHeader(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.headers[key] = value
}

// Get performs a GET request and decodes the JSON response.
func (c *HTTPClient) Get(ctx context.Context, path string, params url.Values, result interface{}) error {
	return c.do(ctx, http.MethodGet, path, params, nil, result)
}

// Post performs a POST request with a JSON body and decodes the response.
func (c *HTTPClient) Post(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.do(ctx, http.MethodPost, path, nil, body, result)
}

func (c *HTTPClient) do(ctx context.Context, method, path string, params url.Values, body interface{}, result interface{}) error {
	// Rate limit
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limiter: %w", err)
	}

	// Build URL
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return fmt.Errorf("parse URL: %w", err)
	}
	if params != nil {
		u.RawQuery = params.Encode()
	}

	// Build body
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	// Set headers
	c.mu.RLock()
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	c.mu.RUnlock()

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Execute with retry
	var resp *http.Response
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err = c.client.Do(req)
		if err == nil && resp.StatusCode < 500 {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		if attempt < maxRetries-1 {
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			log.Warn().
				Str("method", method).
				Str("url", u.String()).
				Int("attempt", attempt+1).
				Dur("backoff", backoff).
				Msg("Retrying request")
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}
	}
	if err != nil {
		return fmt.Errorf("HTTP %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	// Read body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d from %s %s: %s", resp.StatusCode, method, path, string(respBody))
	}

	// Decode
	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("decode response from %s %s: %w (body: %s)", method, path, err, string(respBody[:min(len(respBody), 200)]))
		}
	}

	return nil
}

// ── Rate Limiter ──

type RateLimiter struct {
	ticker *time.Ticker
	tokens chan struct{}
}

func NewRateLimiter(rps int) *RateLimiter {
	rl := &RateLimiter{
		ticker: time.NewTicker(time.Second / time.Duration(rps)),
		tokens: make(chan struct{}, rps),
	}
	// Fill initial tokens
	for i := 0; i < rps; i++ {
		rl.tokens <- struct{}{}
	}
	// Refill tokens
	go func() {
		for range rl.ticker.C {
			select {
			case rl.tokens <- struct{}{}:
			default:
			}
		}
	}()
	return rl
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-rl.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
