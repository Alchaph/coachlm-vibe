---
id: S49
title: Seamless Strava login — remove client ID/secret from UI
status: done
created: 2026-03-18
updated: 2026-03-18
---

# S49 — Seamless Strava login — remove client ID/secret from UI

## User story

As a **runner setting up CoachLM for the first time**,
I want to **click a single "Connect Strava" button** that opens the Strava website in my browser
so that **I can authorise the app without ever seeing or pasting API credentials**.

## Research findings

### Why users currently have to paste credentials

Strava's OAuth2 flow requires a registered `client_id` and `client_secret`.
These belong to the **application developer** (i.e. CoachLM), not to the user.
The current UI exposes them because the app was never shipped with a registered
Strava application — it assumed every user would register their own.

### The correct approach

Register CoachLM as a Strava application once (at
`https://www.strava.com/settings/api`).  Strava will issue a permanent
`client_id` (a public integer) and `client_secret` (a secret string).
Inject both into the binary at build time exactly like the Gemini API key:

```
-ldflags "-X coachlm/internal/strava.builtinClientID=12345
          -X coachlm/internal/strava.builtinClientSecret=abc..."
```

The user flow then becomes:

1. User clicks **Connect Strava** — no form fields.
2. The Go backend opens `https://www.strava.com/oauth/authorize?client_id=...`
   in the system browser (via `wailsRuntime.BrowserOpenURL`).
3. The user logs in on Strava and approves access.
4. Strava redirects to `http://localhost:9876/callback?code=...`.
5. The Go backend exchanges the code for tokens, saves them encrypted, and
   emits a `strava:auth:complete` event to the frontend.
6. The frontend updates its connected state without a page reload.

This is already how `StartStravaAuth` works mechanically — the only change is
that the credentials are no longer entered by the user.

### Strava redirect URI whitelist

Strava requires the `redirect_uri` to match a domain registered on the app.
`localhost` and `127.0.0.1` are explicitly whitelisted by Strava for all apps,
so `http://localhost:9876/callback` is valid without any additional
configuration.

### Local development without a registered app

Developers without the injected credentials can still supply them via
environment variables:

```
STRAVA_CLIENT_ID=12345 STRAVA_CLIENT_SECRET=abc... ./CoachLM
```

If neither build-time injection nor environment variables are present, the
"Connect Strava" button is replaced by a "Not available in this build" note.
The fallback manual entry fields from the current UI are removed entirely.

## Acceptance criteria

- [ ] `internal/strava/oauth.go` exports two package-level `var` strings:
      `builtinClientID` and `builtinClientSecret` (default `""`), injectable
      via `-ldflags`
- [ ] `app.go` reads credentials in this priority order:
      (1) build-time ldflags vars, (2) `STRAVA_CLIENT_ID` /
      `STRAVA_CLIENT_SECRET` env vars, (3) empty → disable Strava connect button
- [ ] `Settings.svelte` Strava section: remove Client ID and Client Secret
      input fields entirely; show only a status badge and Connect/Disconnect
      button
- [ ] `Onboarding.svelte` step "Connect Strava": remove credential input
      fields; show only a "Connect Strava" button (or a disabled note if no
      credentials are present in the build)
- [ ] `storage.Settings` struct drops `StravaClientID` and `StravaClientSecret`
      fields; corresponding SQL columns are left in the DB (migration is
      additive-only) but no longer read or written by the app
- [ ] `app.go` `SaveSettingsData` / `GetSettingsData` no longer handle
      `stravaClientId` / `stravaClientSecret`; `SettingsData` struct drops
      those fields
- [ ] `StartStravaAuth` in `app.go` uses the built-in / env credentials instead
      of reading them from the settings row
- [ ] `GetStravaAuthStatus` still works unchanged
- [ ] `DisconnectStrava` still works unchanged
- [ ] If credentials are absent, `StartStravaAuth` returns a clear error:
      `"strava: no client credentials available in this build"`
- [ ] Release workflow (`release.yml`) documents (in a comment) the two new
      ldflags vars `builtinClientID` and `builtinClientSecret`; actual secrets
      are added manually to GitHub Actions by the maintainer
- [ ] `go test ./...` passes
- [ ] E2e: `settings.spec.ts` — no `#strava-client-id` or
      `#strava-client-secret` inputs; Connect/Disconnect button present
- [ ] E2e: `onboarding.spec.ts` — Connect Strava step has no credential inputs;
      "Connect Strava" button is visible
- [ ] E2e: `s49-strava-seamless-login.spec.ts` — new spec covering the
      connect flow end-to-end against the mock

## Technical notes

### Files to change

| File | Change |
|---|---|
| `internal/strava/oauth.go` | Add `var builtinClientID = ""` and `var builtinClientSecret = ""` at package level |
| `app.go` (`resolveStravaCredentials`) | New helper; returns `(clientID, clientSecret string, ok bool)` checking ldflags → env → empty |
| `app.go` (`StartStravaAuth`) | Replace settings-row credential read with `resolveStravaCredentials()` |
| `app.go` (`SettingsData`) | Remove `StravaClientID`, `StravaClientSecret` fields |
| `app.go` (`GetSettingsData`, `SaveSettingsData`) | Stop reading/writing those fields |
| `internal/storage/settings.go` | Remove `StravaClientID`, `StravaClientSecret` from `Settings` struct, `SaveSettings`, `GetSettings` |
| `frontend/src/Settings.svelte` | Remove credential inputs from Strava section; keep status badge + Connect/Disconnect |
| `frontend/src/Onboarding.svelte` | Remove credential inputs from Connect Strava step |
| `frontend/e2e/mocks/wails.ts` | Drop `stravaClientId` / `stravaClientSecret` from `DEFAULT_SETTINGS`; add `GetStravaCredentialsAvailable: () => mockAsync(true)` |
| `frontend/e2e/settings.spec.ts` | Remove credential-field tests; add test for absence of those fields |
| `frontend/e2e/onboarding.spec.ts` | Update Connect Strava step assertions |
| `.github/workflows/release.yml` | Add comment about `STRAVA_CLIENT_ID` / `STRAVA_CLIENT_SECRET` secrets |

### New app.go helper

```go
func resolveStravaCredentials() (clientID, clientSecret string, ok bool) {
    id := strava.BuiltinClientID
    secret := strava.BuiltinClientSecret
    if id == "" {
        id = os.Getenv("STRAVA_CLIENT_ID")
    }
    if secret == "" {
        secret = os.Getenv("STRAVA_CLIENT_SECRET")
    }
    return id, secret, id != "" && secret != ""
}
```

### Expose credentials-available status to frontend

Add a new binding `GetStravaCredentialsAvailable() bool` so the frontend can
show a disabled state when credentials are absent (developer build).

### DB migration

The `strava_client_id` and `strava_client_secret` columns in the `settings`
table are **not** dropped — SQLite `ALTER TABLE DROP COLUMN` requires SQLite
3.35+, which is not guaranteed.  The columns are simply ignored going forward.
No new migration entry is needed.

### Build-time injection (release.yml)

```yaml
- name: Build
  run: wails build ${{ matrix.build-flags }}
    -ldflags "-X coachlm/internal/llm.builtinGeminiApiKey=${{ secrets.GEMINI_API_KEY }}
              -X coachlm/internal/strava.builtinClientID=${{ secrets.STRAVA_CLIENT_ID }}
              -X coachlm/internal/strava.builtinClientSecret=${{ secrets.STRAVA_CLIENT_SECRET }}"
```

(Actual secret values must be added by the maintainer in GitHub → Settings →
Secrets and variables → Actions.)

## Tests required

- Unit: `TestResolveStravaCredentials_LDFlags` — ldflags var set → returns them
- Unit: `TestResolveStravaCredentials_Env` — env vars set → returns them
- Unit: `TestResolveStravaCredentials_Empty` — nothing set → ok = false
- Unit: `TestStartStravaAuth_NoCredentials` — returns the expected error string
- E2e: `settings.spec.ts` — `#strava-client-id` and `#strava-client-secret`
  do not exist in the DOM
- E2e: `onboarding.spec.ts` — Connect Strava step: no credential inputs, button
  is visible
- E2e: `s49-strava-seamless-login.spec.ts` — mock `StartStravaAuth` resolves,
  status badge updates to "Connected"

## Out of scope

- Registering the CoachLM Strava application (must be done manually by the
  maintainer before the first official release that includes this story)
- Storing or displaying the athlete's Strava username/avatar after auth
- Revoking/deauthorising from within the app (Strava deauthorise endpoint)
- Any changes to activity sync logic

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-18 | draft | Created — research complete, implementation ready to begin after S48 |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
