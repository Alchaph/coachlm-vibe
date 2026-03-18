package strava

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	maxRetries       = 3
	defaultRetryWait = 60 * time.Second
	backoffBase      = time.Second
)

// RateLimitedClient wraps an *http.Client and transparently handles Strava
// API rate limiting. It tracks the 15-minute and daily usage counters from
// X-RateLimit-Limit / X-RateLimit-Usage response headers, retries on HTTP 429
// (respecting Retry-After), and retries on 5xx responses with exponential backoff.
type RateLimitedClient struct {
	client     *http.Client
	mu         sync.Mutex
	limit15Min int
	usage15Min int
	limitDaily int
	usageDaily int
}

// NewRateLimitedClient creates a RateLimitedClient wrapping the given HTTP
// client. If client is nil, a default client with a 10-second timeout is used.
func NewRateLimitedClient(client *http.Client) *RateLimitedClient {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &RateLimitedClient{client: client}
}

// RateLimitUsage returns a snapshot of the current rate limit state.
func (c *RateLimitedClient) RateLimitUsage() (limit15Min, usage15Min, limitDaily, usageDaily int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.limit15Min, c.usage15Min, c.limitDaily, c.usageDaily
}

// Do sends an HTTP request, retrying on 429 (with Retry-After backoff) and
// 5xx (with exponential backoff). Rate limit headers are parsed and stored
// from each response. Returns the final response on success, or an error after
// exhausting all retries.
func (c *RateLimitedClient) Do(req *http.Request) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := c.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("http request: %w", err)
		}

		c.updateRateLimits(resp)

		switch {
		case resp.StatusCode == http.StatusTooManyRequests:
			resp.Body.Close()
			if attempt == maxRetries {
				lastErr = fmt.Errorf("rate limited after %d retries", maxRetries)
				break
			}
			wait := parseRetryAfter(resp.Header.Get("Retry-After"))
			time.Sleep(wait)
			// Clone the request for the next attempt — the body has been consumed.
			req = cloneRequest(req)
			continue

		case resp.StatusCode >= 500:
			resp.Body.Close()
			if attempt == maxRetries {
				lastErr = fmt.Errorf("server error HTTP %d after %d retries", resp.StatusCode, maxRetries)
				break
			}
			backoff := backoffBase * (1 << uint(attempt)) // 1s, 2s, 4s
			time.Sleep(backoff)
			req = cloneRequest(req)
			continue

		default:
			return resp, nil
		}
	}

	return nil, lastErr
}

// updateRateLimits parses X-RateLimit-Limit and X-RateLimit-Usage headers and
// stores the values. Header format: "15min_value,daily_value".
func (c *RateLimitedClient) updateRateLimits(resp *http.Response) {
	limitHeader := resp.Header.Get("X-RateLimit-Limit")
	usageHeader := resp.Header.Get("X-RateLimit-Usage")

	limit15, limitDay := parsePair(limitHeader)
	usage15, usageDay := parsePair(usageHeader)

	c.mu.Lock()
	defer c.mu.Unlock()
	if limitHeader != "" {
		c.limit15Min = limit15
		c.limitDaily = limitDay
	}
	if usageHeader != "" {
		c.usage15Min = usage15
		c.usageDaily = usageDay
	}
}

// parsePair splits a "a,b" header value into two ints. Returns (0, 0) on error.
func parsePair(h string) (int, int) {
	parts := strings.SplitN(h, ",", 2)
	if len(parts) != 2 {
		return 0, 0
	}
	a, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	b, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil {
		return 0, 0
	}
	return a, b
}

// parseRetryAfter parses the Retry-After header (seconds as integer string).
// Returns defaultRetryWait if the header is absent or invalid.
func parseRetryAfter(header string) time.Duration {
	if header == "" {
		return defaultRetryWait
	}
	secs, err := strconv.Atoi(strings.TrimSpace(header))
	if err != nil || secs < 0 {
		return defaultRetryWait
	}
	return time.Duration(secs) * time.Second
}

// cloneRequest returns a shallow clone of req so we can re-send it.
// Only safe for requests with no body (GET/DELETE). The original context is preserved.
func cloneRequest(req *http.Request) *http.Request {
	clone := req.Clone(req.Context())
	return clone
}
