package cloudsync

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var ErrNotFound = errors.New("cloudsync: key not found")

type S3Config struct {
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	Region    string
}

type S3Provider struct {
	cfg    S3Config
	client *http.Client
}

func NewS3(cfg S3Config) (*S3Provider, error) {
	if cfg.Endpoint == "" {
		return nil, errors.New("s3: endpoint is required")
	}
	if cfg.Bucket == "" {
		return nil, errors.New("s3: bucket is required")
	}
	if cfg.AccessKey == "" {
		return nil, errors.New("s3: access key is required")
	}
	if cfg.SecretKey == "" {
		return nil, errors.New("s3: secret key is required")
	}
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}
	return &S3Provider{
		cfg:    cfg,
		client: &http.Client{Timeout: 60 * time.Second},
	}, nil
}

func (s *S3Provider) Name() string { return "S3" }

func (s *S3Provider) Upload(ctx context.Context, key string, data []byte) error {
	url := fmt.Sprintf("%s/%s/%s", s.cfg.Endpoint, s.cfg.Bucket, key)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("s3 upload: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.SetBasicAuth(s.cfg.AccessKey, s.cfg.SecretKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("s3 upload: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 300 {
		return fmt.Errorf("s3 upload: status %d", resp.StatusCode)
	}
	return nil
}

func (s *S3Provider) Download(ctx context.Context, key string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s/%s", s.cfg.Endpoint, s.cfg.Bucket, key)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("s3 download: build request: %w", err)
	}
	req.SetBasicAuth(s.cfg.AccessKey, s.cfg.SecretKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("s3 download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("s3 download: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("s3 download: read body: %w", err)
	}
	return data, nil
}

type s3ListResult struct {
	XMLName  xml.Name   `xml:"ListBucketResult"`
	Contents []s3Object `xml:"Contents"`
}

type s3Object struct {
	Key          string `xml:"Key"`
	LastModified string `xml:"LastModified"`
}

func (s *S3Provider) LastModified(ctx context.Context, key string) (time.Time, error) {
	url := fmt.Sprintf("%s/%s?prefix=%s&max-keys=1", s.cfg.Endpoint, s.cfg.Bucket, key)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return time.Time{}, fmt.Errorf("s3 last modified: build request: %w", err)
	}
	req.SetBasicAuth(s.cfg.AccessKey, s.cfg.SecretKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return time.Time{}, fmt.Errorf("s3 last modified: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return time.Time{}, fmt.Errorf("s3 last modified: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return time.Time{}, fmt.Errorf("s3 last modified: read body: %w", err)
	}

	var result s3ListResult
	if err := xml.Unmarshal(body, &result); err != nil {
		return time.Time{}, fmt.Errorf("s3 last modified: parse xml: %w", err)
	}

	for _, obj := range result.Contents {
		if obj.Key == key {
			t, err := time.Parse(time.RFC3339Nano, obj.LastModified)
			if err != nil {
				t, err = time.Parse("2006-01-02T15:04:05.000Z", obj.LastModified)
				if err != nil {
					return time.Time{}, fmt.Errorf("s3 last modified: parse time %q: %w", obj.LastModified, err)
				}
			}
			return t, nil
		}
	}

	return time.Time{}, ErrNotFound
}
