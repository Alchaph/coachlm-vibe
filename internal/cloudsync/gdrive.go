package cloudsync

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"time"
)

const (
	driveAPIBase   = "https://www.googleapis.com"
	driveFilesURL  = driveAPIBase + "/drive/v3/files"
	driveUploadURL = driveAPIBase + "/upload/drive/v3/files"
)

type GDriveConfig struct {
	AccessToken  string
	RefreshToken string
	TokenExpiry  time.Time
	ClientID     string
	ClientSecret string
	FolderName   string
}

type GDriveProvider struct {
	cfg    GDriveConfig
	client *http.Client
}

func NewGDrive(cfg GDriveConfig) (*GDriveProvider, error) {
	if cfg.AccessToken == "" {
		return nil, errors.New("gdrive: access token is required")
	}
	if cfg.FolderName == "" {
		cfg.FolderName = "CoachLM"
	}
	return &GDriveProvider{
		cfg:    cfg,
		client: &http.Client{Timeout: 60 * time.Second},
	}, nil
}

func (g *GDriveProvider) Name() string { return "Google Drive" }

func (g *GDriveProvider) Upload(ctx context.Context, key string, data []byte) error {
	fileID, err := g.findFileID(ctx, key)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return fmt.Errorf("gdrive upload: find existing: %w", err)
	}

	if fileID != "" {
		return g.updateFile(ctx, fileID, data)
	}
	return g.createFile(ctx, key, data)
}

func (g *GDriveProvider) Download(ctx context.Context, key string) ([]byte, error) {
	fileID, err := g.findFileID(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("gdrive download: %w", err)
	}

	url := fmt.Sprintf("%s/%s?alt=media", driveFilesURL, fileID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("gdrive download: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+g.cfg.AccessToken)

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gdrive download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("gdrive download: status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (g *GDriveProvider) LastModified(ctx context.Context, key string) (time.Time, error) {
	fileID, err := g.findFileID(ctx, key)
	if err != nil {
		return time.Time{}, err
	}

	url := fmt.Sprintf("%s/%s?fields=modifiedTime", driveFilesURL, fileID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return time.Time{}, fmt.Errorf("gdrive last modified: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+g.cfg.AccessToken)

	resp, err := g.client.Do(req)
	if err != nil {
		return time.Time{}, fmt.Errorf("gdrive last modified: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return time.Time{}, fmt.Errorf("gdrive last modified: status %d", resp.StatusCode)
	}

	var result struct {
		ModifiedTime string `json:"modifiedTime"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return time.Time{}, fmt.Errorf("gdrive last modified: parse: %w", err)
	}

	t, err := time.Parse(time.RFC3339Nano, result.ModifiedTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("gdrive last modified: parse time: %w", err)
	}
	return t, nil
}

type driveFileList struct {
	Files []driveFile `json:"files"`
}

type driveFile struct {
	ID string `json:"id"`
}

func (g *GDriveProvider) findFileID(ctx context.Context, key string) (string, error) {
	query := fmt.Sprintf("name='%s' and trashed=false", key)
	url := fmt.Sprintf("%s?q=%s&fields=files(id)&pageSize=1", driveFilesURL, query)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("gdrive find: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+g.cfg.AccessToken)

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("gdrive find: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("gdrive find: status %d", resp.StatusCode)
	}

	var list driveFileList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return "", fmt.Errorf("gdrive find: parse: %w", err)
	}

	if len(list.Files) == 0 {
		return "", ErrNotFound
	}
	return list.Files[0].ID, nil
}

func (g *GDriveProvider) createFile(ctx context.Context, key string, data []byte) error {
	metadata := map[string]string{"name": key}
	metaJSON, _ := json.Marshal(metadata)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	metaHeader := make(textproto.MIMEHeader)
	metaHeader.Set("Content-Type", "application/json; charset=UTF-8")
	metaPart, err := writer.CreatePart(metaHeader)
	if err != nil {
		return fmt.Errorf("gdrive create: metadata part: %w", err)
	}
	metaPart.Write(metaJSON)

	dataHeader := make(textproto.MIMEHeader)
	dataHeader.Set("Content-Type", "application/octet-stream")
	dataPart, err := writer.CreatePart(dataHeader)
	if err != nil {
		return fmt.Errorf("gdrive create: data part: %w", err)
	}
	dataPart.Write(data)
	writer.Close()

	url := driveUploadURL + "?uploadType=multipart"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &body)
	if err != nil {
		return fmt.Errorf("gdrive create: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+g.cfg.AccessToken)
	req.Header.Set("Content-Type", "multipart/related; boundary="+writer.Boundary())

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("gdrive create: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 300 {
		return fmt.Errorf("gdrive create: status %d", resp.StatusCode)
	}
	return nil
}

func (g *GDriveProvider) updateFile(ctx context.Context, fileID string, data []byte) error {
	url := fmt.Sprintf("%s/%s?uploadType=media", driveUploadURL, fileID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("gdrive update: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+g.cfg.AccessToken)
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("gdrive update: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 300 {
		return fmt.Errorf("gdrive update: status %d", resp.StatusCode)
	}
	return nil
}
