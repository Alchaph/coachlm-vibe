package strava

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func testKey() []byte {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		panic(err)
	}
	return key
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := testKey()
	plaintext := []byte("super-secret-access-token")

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Fatalf("round-trip mismatch: got %q, want %q", decrypted, plaintext)
	}
}

func TestEncryptDecryptEmptyPlaintext(t *testing.T) {
	key := testKey()
	plaintext := []byte("")

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Fatalf("round-trip mismatch for empty plaintext")
	}
}

func TestDecryptWithWrongKeyFails(t *testing.T) {
	key1 := testKey()
	key2 := testKey()

	ciphertext, err := Encrypt([]byte("secret"), key1)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	_, err = Decrypt(ciphertext, key2)
	if err == nil {
		t.Fatal("expected error decrypting with wrong key")
	}
}

func TestEncryptProducesDifferentCiphertextEachTime(t *testing.T) {
	key := testKey()
	plaintext := []byte("same-input")

	ct1, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt 1: %v", err)
	}

	ct2, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt 2: %v", err)
	}

	if bytes.Equal(ct1, ct2) {
		t.Fatal("two encryptions of same plaintext produced identical ciphertext")
	}
}

func TestEncryptRejectsInvalidKeySize(t *testing.T) {
	_, err := Encrypt([]byte("data"), []byte("short-key"))
	if err == nil {
		t.Fatal("expected error with invalid key size")
	}
}

func TestDecryptRejectsTruncatedCiphertext(t *testing.T) {
	key := testKey()
	_, err := Decrypt([]byte("too-short"), key)
	if err == nil {
		t.Fatal("expected error with truncated ciphertext")
	}
}

func newTestOAuthServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *OAuthClient) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	client := NewOAuthClient("test-client-id", "test-secret", "http://localhost/callback", testKey())
	client.tokenURL = srv.URL
	client.httpClient = srv.Client()

	return srv, client
}

func TestExchangeSuccess(t *testing.T) {
	_, client := newTestOAuthServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if got := r.FormValue("grant_type"); got != "authorization_code" {
			t.Errorf("grant_type = %q, want authorization_code", got)
		}
		if got := r.FormValue("code"); got != "test-auth-code" {
			t.Errorf("code = %q, want test-auth-code", got)
		}

		json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "access-123",
			"refresh_token": "refresh-456",
			"expires_at":    time.Now().Add(6 * time.Hour).Unix(),
		})
	})

	tp, err := client.Exchange(context.Background(), "test-auth-code")
	if err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	if tp.AccessToken != "access-123" {
		t.Errorf("AccessToken = %q, want access-123", tp.AccessToken)
	}
	if tp.RefreshToken != "refresh-456" {
		t.Errorf("RefreshToken = %q, want refresh-456", tp.RefreshToken)
	}
	if tp.ExpiresAt.Before(time.Now()) {
		t.Error("ExpiresAt should be in the future")
	}
}

func TestExchangeInvalidGrant(t *testing.T) {
	_, client := newTestOAuthServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"message": "Bad Request",
			"errors":  []map[string]string{{"resource": "Application", "field": "code", "code": "invalid"}},
		})
	})

	_, err := client.Exchange(context.Background(), "revoked-code")
	if err == nil {
		t.Fatal("expected error for invalid grant")
	}
}

func TestExchangeServerError(t *testing.T) {
	_, client := newTestOAuthServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := client.Exchange(context.Background(), "code")
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestRefreshSuccess(t *testing.T) {
	_, client := newTestOAuthServer(t, func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if got := r.FormValue("grant_type"); got != "refresh_token" {
			t.Errorf("grant_type = %q, want refresh_token", got)
		}
		if got := r.FormValue("refresh_token"); got != "old-refresh" {
			t.Errorf("refresh_token = %q, want old-refresh", got)
		}

		json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "new-access",
			"refresh_token": "new-refresh",
			"expires_at":    time.Now().Add(6 * time.Hour).Unix(),
		})
	})

	tp, err := client.Refresh(context.Background(), "old-refresh")
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if tp.AccessToken != "new-access" {
		t.Errorf("AccessToken = %q, want new-access", tp.AccessToken)
	}
	if tp.RefreshToken != "new-refresh" {
		t.Errorf("RefreshToken = %q, want new-refresh", tp.RefreshToken)
	}
}

func TestIsExpired(t *testing.T) {
	client := NewOAuthClient("", "", "", testKey())

	if !client.IsExpired(time.Now().Add(-1 * time.Hour)) {
		t.Error("past time should be expired")
	}
	if client.IsExpired(time.Now().Add(1 * time.Hour)) {
		t.Error("future time should not be expired")
	}
}

func TestExchangeContextCancellation(t *testing.T) {
	_, client := newTestOAuthServer(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(5 * time.Second):
		case <-r.Context().Done():
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "should-not-get-this",
			"refresh_token": "nope",
			"expires_at":    time.Now().Add(6 * time.Hour).Unix(),
		})
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.Exchange(ctx, "code")
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestAuthURL(t *testing.T) {
	client := NewOAuthClient("my-client-id", "secret", "http://localhost:8080/callback", testKey())
	u := client.AuthURL()

	if u == "" {
		t.Fatal("AuthURL returned empty string")
	}
	for _, want := range []string{"client_id=my-client-id", "redirect_uri=", "response_type=code", "scope=activity"} {
		if !contains(u, want) {
			t.Errorf("AuthURL missing %q in %q", want, u)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestEncryptDecryptToken(t *testing.T) {
	client := NewOAuthClient("", "", "", testKey())
	token := "my-secret-token-value"

	encrypted, err := client.EncryptToken(token)
	if err != nil {
		t.Fatalf("EncryptToken: %v", err)
	}

	decrypted, err := client.DecryptToken(encrypted)
	if err != nil {
		t.Fatalf("DecryptToken: %v", err)
	}

	if decrypted != token {
		t.Fatalf("DecryptToken = %q, want %q", decrypted, token)
	}
}
