---
id: S21
title: Cloud storage sync
status: draft
created: 2026-03-16
updated: 2026-03-16
---

# S21 — Cloud storage sync

## User story

As a **runner**,
I want to **connect a cloud storage provider (self-hosted or Google Cloud) to automatically back up and sync my coaching context**
so that **my data is safe across devices and reinstalls without manual export/import steps**.

## Acceptance criteria

- [ ] Supported providers at launch: Google Drive (OAuth 2.0) and any S3-compatible endpoint (covers self-hosted MinIO, Backblaze B2, Cloudflare R2, AWS S3)
- [ ] Settings screen gains a "Cloud Sync" section: provider selector, credentials input, and connection test button
- [ ] Google Drive: OAuth 2.0 PKCE flow initiated from the app; tokens stored encrypted in SQLite (same mechanism as S01)
- [ ] S3-compatible: user provides endpoint URL, bucket name, access key ID, and secret access key; credentials stored encrypted in SQLite
- [ ] On successful connection, the app performs an initial upload of the full context export (S20 `.coachctx` format)
- [ ] Automatic sync triggers after every context-mutating operation (save insight, training summary update, profile change) — debounced 30 s to avoid thrashing
- [ ] Manual "Sync now" button in Settings triggers an immediate upload
- [ ] On app launch, if cloud sync is enabled, the app checks for a newer remote file and offers to pull it down (with diff summary: "Remote has 3 new insights since your last sync")
- [ ] Conflict resolution: remote timestamp wins by default; user can override to keep local
- [ ] Removing cloud connection clears stored credentials and disables auto-sync (does not delete the remote file)
- [ ] All sync operations run in a background goroutine; the UI is never blocked

## Technical notes

New package: `internal/cloudsync/` with a provider interface:
```go
type CloudProvider interface {
    Upload(ctx context.Context, key string, data []byte) error
    Download(ctx context.Context, key string) ([]byte, error)
    LastModified(ctx context.Context, key string) (time.Time, error)
    Name() string
}
```

Implementations:
- `internal/cloudsync/gdrive.go` — uses `golang.org/x/oauth2` + Google Drive REST API v3
- `internal/cloudsync/s3.go` — uses `github.com/aws/aws-sdk-go-v2/service/s3` (works with any S3-compatible endpoint via custom `EndpointResolver`)

The remote file key is `coachlm/context.coachctx`. Bucket / Drive folder is configurable.

Sync state table in SQLite: `cloud_sync_state` with columns `provider`, `last_synced_at`, `remote_etag`.

Google Drive OAuth redirect URI: `http://localhost:{random_port}/callback` — open a temporary HTTP listener, capture the auth code, exchange for tokens, close the listener.

Wails bindings in `app.go`:
- `ConnectGoogleDrive() error`
- `ConnectS3(endpoint, bucket, accessKey, secretKey string) error`
- `DisconnectCloud() error`
- `SyncNow() error`
- `GetSyncStatus() (SyncStatus, error)`

Depends on S20 (context export/import format) and S16 (settings screen).

Do not bundle service account credentials in the binary. The user supplies their own.

## Tests required

- Unit: provider interface contract for both implementations (mock HTTP), upload/download round-trip, conflict detection logic, debounce timer, credential encryption/decryption
- Integration: connect to a real MinIO instance (Docker-based test fixture) — upload, modify locally, detect remote is stale, pull
- Edge cases:
  - Network unreachable during auto-sync → error logged, retry queued for next sync trigger, UI shows last-synced timestamp
  - Invalid S3 credentials → connection test returns descriptive error, credentials not saved
  - Google Drive OAuth cancelled by user → no credentials stored, sync stays disabled
  - Remote file deleted externally → next sync re-uploads without error
  - Concurrent sync triggers (manual + auto) → only one upload in flight at a time (use a mutex or single-flight)
  - Very large context file (> 10 MB) → upload still completes, progress surfaced in UI

## Out of scope

- iCloud Drive, OneDrive, Dropbox
- Encrypted-at-rest remote file (can follow from S20 encrypted export)
- Syncing raw activity streams or FIT files (context only)
- Multi-device real-time collaboration / merge
- Paid cloud storage provisioning inside the app

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-16 | draft | Created |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
