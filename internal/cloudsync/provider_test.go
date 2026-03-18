package cloudsync

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewS3_ValidConfig(t *testing.T) {
	p, err := NewS3(S3Config{
		Endpoint:  "http://localhost:9000",
		Bucket:    "test",
		AccessKey: "key",
		SecretKey: "secret",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name() != "S3" {
		t.Errorf("Name() = %q, want %q", p.Name(), "S3")
	}
}

func TestNewS3_MissingFields(t *testing.T) {
	tests := []struct {
		name string
		cfg  S3Config
	}{
		{"missing endpoint", S3Config{Bucket: "b", AccessKey: "a", SecretKey: "s"}},
		{"missing bucket", S3Config{Endpoint: "http://x", AccessKey: "a", SecretKey: "s"}},
		{"missing access key", S3Config{Endpoint: "http://x", Bucket: "b", SecretKey: "s"}},
		{"missing secret key", S3Config{Endpoint: "http://x", Bucket: "b", AccessKey: "a"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewS3(tt.cfg)
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestNewS3_DefaultRegion(t *testing.T) {
	p, err := NewS3(S3Config{
		Endpoint:  "http://localhost:9000",
		Bucket:    "test",
		AccessKey: "key",
		SecretKey: "secret",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.cfg.Region != "us-east-1" {
		t.Errorf("Region = %q, want %q", p.cfg.Region, "us-east-1")
	}
}

func TestS3_UploadDownloadRoundTrip(t *testing.T) {
	store := map[string][]byte{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Path

		switch r.Method {
		case http.MethodPut:
			data, _ := io.ReadAll(r.Body)
			store[key] = data
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			if data, ok := store[key]; ok {
				w.Write(data)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}
	}))
	defer server.Close()

	p, err := NewS3(S3Config{
		Endpoint:  server.URL,
		Bucket:    "testbucket",
		AccessKey: "key",
		SecretKey: "secret",
	})
	if err != nil {
		t.Fatalf("NewS3: %v", err)
	}

	ctx := context.Background()
	payload := []byte(`{"test":"data"}`)

	if err := p.Upload(ctx, "mykey", payload); err != nil {
		t.Fatalf("Upload: %v", err)
	}

	got, err := p.Download(ctx, "mykey")
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if string(got) != string(payload) {
		t.Errorf("Download = %q, want %q", got, payload)
	}
}

func TestS3_DownloadNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	p, _ := NewS3(S3Config{
		Endpoint:  server.URL,
		Bucket:    "b",
		AccessKey: "a",
		SecretKey: "s",
	})

	_, err := p.Download(context.Background(), "nonexistent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestS3_UploadServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	p, _ := NewS3(S3Config{
		Endpoint:  server.URL,
		Bucket:    "b",
		AccessKey: "a",
		SecretKey: "s",
	})

	err := p.Upload(context.Background(), "k", []byte("data"))
	if err == nil {
		t.Error("expected error for 500 status")
	}
}

func TestS3_LastModified(t *testing.T) {
	ts := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result := s3ListResult{
			Contents: []s3Object{
				{Key: "coachlm/context.coachctx", LastModified: ts.Format(time.RFC3339Nano)},
			},
		}
		data, _ := xml.Marshal(result)
		w.Write(data)
	}))
	defer server.Close()

	p, _ := NewS3(S3Config{
		Endpoint:  server.URL,
		Bucket:    "b",
		AccessKey: "a",
		SecretKey: "s",
	})

	got, err := p.LastModified(context.Background(), "coachlm/context.coachctx")
	if err != nil {
		t.Fatalf("LastModified: %v", err)
	}
	if !got.Equal(ts) {
		t.Errorf("LastModified = %v, want %v", got, ts)
	}
}

func TestS3_LastModifiedNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result := s3ListResult{Contents: nil}
		data, _ := xml.Marshal(result)
		w.Write(data)
	}))
	defer server.Close()

	p, _ := NewS3(S3Config{
		Endpoint:  server.URL,
		Bucket:    "b",
		AccessKey: "a",
		SecretKey: "s",
	})

	_, err := p.LastModified(context.Background(), "nonexistent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestS3_BasicAuthSent(t *testing.T) {
	var gotUser, gotPass string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser, gotPass, _ = r.BasicAuth()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	p, _ := NewS3(S3Config{
		Endpoint:  server.URL,
		Bucket:    "b",
		AccessKey: "mykey",
		SecretKey: "mysecret",
	})

	_ = p.Upload(context.Background(), "k", []byte("data"))

	if gotUser != "mykey" || gotPass != "mysecret" {
		t.Errorf("auth = (%q, %q), want (%q, %q)", gotUser, gotPass, "mykey", "mysecret")
	}
}

func TestNewGDrive_ValidConfig(t *testing.T) {
	p, err := NewGDrive(GDriveConfig{AccessToken: "token123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name() != "Google Drive" {
		t.Errorf("Name() = %q, want %q", p.Name(), "Google Drive")
	}
}

func TestNewGDrive_MissingToken(t *testing.T) {
	_, err := NewGDrive(GDriveConfig{})
	if err == nil {
		t.Error("expected error for missing access token")
	}
}

func TestNewGDrive_DefaultFolder(t *testing.T) {
	p, _ := NewGDrive(GDriveConfig{AccessToken: "token"})
	if p.cfg.FolderName != "CoachLM" {
		t.Errorf("FolderName = %q, want %q", p.cfg.FolderName, "CoachLM")
	}
}
