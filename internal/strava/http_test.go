package strava

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimitedClient429Retry(t *testing.T) {
	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewRateLimitedClient(server.Client())

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if attempt != 2 {
		t.Errorf("attempt = %d, want 2 (one 429, one success)", attempt)
	}
}

func TestRateLimitedClient5xxBackoff(t *testing.T) {
	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewRateLimitedClient(server.Client())

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if attempt != 3 {
		t.Errorf("attempt = %d, want 3 (two 500s, one success)", attempt)
	}
}

func TestRateLimitedClientHeaderParsing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Limit", "100,1000")
		w.Header().Set("X-RateLimit-Usage", "42,350")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewRateLimitedClient(server.Client())

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	limit15, usage15, limitDay, usageDay := c.RateLimitUsage()
	if limit15 != 100 {
		t.Errorf("limit15Min = %d, want 100", limit15)
	}
	if limitDay != 1000 {
		t.Errorf("limitDaily = %d, want 1000", limitDay)
	}
	if usage15 != 42 {
		t.Errorf("usage15Min = %d, want 42", usage15)
	}
	if usageDay != 350 {
		t.Errorf("usageDaily = %d, want 350", usageDay)
	}
}

func TestRateLimitedClientMaxRetriesExceeded(t *testing.T) {
	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	c := NewRateLimitedClient(server.Client())

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
	_, err := c.Do(req)
	if err == nil {
		t.Fatal("expected error after max retries, got nil")
	}

	wantAttempts := maxRetries + 1
	if attempt != wantAttempts {
		t.Errorf("attempt = %d, want %d", attempt, wantAttempts)
	}
}

func TestRateLimitedClientNilClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewRateLimitedClient(nil)
	if c.client == nil {
		t.Fatal("expected non-nil client after NewRateLimitedClient(nil)")
	}
}

func TestRateLimitedClientMaxRetries5xx(t *testing.T) {
	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	c := NewRateLimitedClient(server.Client())

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
	_, err := c.Do(req)
	if err == nil {
		t.Fatal("expected error after max 5xx retries, got nil")
	}

	wantAttempts := maxRetries + 1
	if attempt != wantAttempts {
		t.Errorf("attempt = %d, want %d", attempt, wantAttempts)
	}
}
