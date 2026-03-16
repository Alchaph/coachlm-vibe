---
id: S18
title: GitHub Actions release pipeline
status: in-progress
created: 2026-03-16
updated: 2026-03-16
---

# S18 — GitHub Actions release pipeline

## User story

As a **maintainer**,
I want to **tag a commit and have GitHub Actions automatically build and publish a release**
so that **users can download signed, versioned binaries for macOS, Windows, and Linux without manual steps**.

## Acceptance criteria

- [ ] Workflow triggers on `v*` semver tags (e.g. `v1.2.3`) pushed to `main`
- [ ] Matrix build covers: `ubuntu-latest`, `windows-latest`, `macos-latest`
- [ ] Each runner installs Go, Node, and the Wails CLI, then builds the app with `wails build -production`
- [ ] macOS binary is code-signed when `APPLE_SIGNING_CERT` secret is present; build succeeds (unsigned) when absent
- [ ] Windows binary is optionally signed when `WINDOWS_SIGNING_CERT` + `WINDOWS_SIGNING_PASSWORD` secrets are present
- [ ] A GitHub Release is created automatically with the tag name as title and the compiled binaries attached as assets
- [ ] Release notes are auto-generated from commit messages since the previous tag using `actions/release-drafter` or equivalent
- [ ] Workflow fails fast and reports a clear error if the build breaks on any platform
- [ ] A separate `ci.yml` workflow runs `go test ./...` and `go vet ./...` on every pull request to `main`

## Technical notes

Wails v2 build command: `wails build -production -platform <GOOS/GOARCH>`.
Output binary paths differ per OS — document the exact artifact paths for the upload step.

Secrets required (configured in GitHub repo settings, never in code):
- `APPLE_SIGNING_CERT` — base64-encoded `.p12`
- `APPLE_SIGNING_PASSWORD`
- `APPLE_TEAM_ID`
- `WINDOWS_SIGNING_CERT` — base64-encoded `.pfx`
- `WINDOWS_SIGNING_PASSWORD`

Workflow files live in `.github/workflows/`:
- `release.yml` — tag-triggered build + publish
- `ci.yml` — PR validation

Use `actions/upload-artifact` within the matrix jobs and `actions/download-artifact` in the final publish job to collect all platform binaries before creating the release. This avoids race conditions.

Cache Go modules and npm packages with `actions/cache` to keep build times under 10 minutes.

Do not embed secrets or tokens in workflow YAML. Use `${{ secrets.NAME }}` only.

## Tests required

- Unit: n/a (infrastructure, not Go code)
- Integration: push a test tag to a fork and verify the workflow produces all three binaries and creates a draft release
- Edge cases:
  - Tag pushed without signing secrets present → build succeeds, release published as unsigned
  - Build failure on one platform → workflow exits non-zero, no release created
  - Duplicate tag pushed → workflow should not overwrite an existing release

## Out of scope

- Homebrew tap or package manager publishing
- Auto-update mechanism inside the app
- Docker image or web deployment
- Notarization (macOS Gatekeeper) — can be added in a follow-up story

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-16 | draft | Created |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
