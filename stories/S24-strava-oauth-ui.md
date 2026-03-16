---
id: S24
title: Strava OAuth login UI
status: done
created: 2026-03-16
updated: 2026-03-16
---

# S24 — Strava OAuth login UI

## User story

As a **runner**,
I want to **connect my Strava account from the Settings screen**
so that **my activities sync automatically without manual configuration**.

## Acceptance criteria

- [ ] "Strava Connection" section in Settings.svelte
- [ ] "Connect Strava" button that opens the Strava OAuth authorization URL in the system browser
- [ ] Local HTTP callback server to receive the OAuth redirect with authorization code
- [ ] Automatic token exchange after receiving the authorization code
- [ ] Encrypted token storage in SQLite (reuses existing token storage)
- [ ] Connection status displayed: "Connected" with disconnect option, or "Not connected" with connect button
- [ ] "Disconnect" button that deletes stored tokens
- [ ] Wails bindings: `StartStravaAuth()`, `GetStravaAuthStatus()`, `DisconnectStrava()`
- [ ] Input fields for Strava Client ID and Client Secret (stored encrypted alongside LLM keys)
- [ ] Error handling: network failure, user denied access, invalid credentials

## Technical notes

The OAuth backend already exists (`internal/strava/oauth.go`). This story adds:
1. Wails bindings in `app.go` for the OAuth flow
2. A local HTTP server (e.g., `localhost:9876/callback`) to catch the OAuth redirect
3. Strava section in `Settings.svelte`
4. Storage for Strava client credentials (client ID + secret) — extend the settings table

`StartStravaAuth` flow:
1. Read client ID/secret from settings
2. Create OAuthClient
3. Start local HTTP server on `localhost:9876/callback`
4. Open `AuthURL()` in system browser via `runtime.BrowserOpenURL`
5. Wait for callback with auth code
6. Exchange code for tokens
7. Encrypt and store tokens
8. Return success/error

The settings table needs two new columns: `strava_client_id` and `strava_client_secret` (encrypted).

## Tests required

- Unit: StartStravaAuth validates credentials exist, GetStravaAuthStatus returns correct state
- Unit: callback server starts and stops cleanly
- Integration: full OAuth mock flow (mock Strava token endpoint)
- Edge cases: user cancels auth, duplicate connect, tokens expired

## Out of scope

Webhook setup, activity sync trigger, background sync — those exist in S02/S03

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-16 | draft | Created |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
