---
id: S28
title: Desktop-oriented layout with sidebar navigation
status: done
created: 2026-03-16
updated: 2026-03-16
---

# S28 — Desktop-oriented layout with sidebar navigation

## User story

As a **user**,
I want the app to use a sidebar navigation and full screen width
so that it feels like a proper desktop application and not a mobile web page.

## Acceptance criteria

- [ ] Top tab bar replaced with a vertical sidebar navigation on the left
- [ ] Sidebar includes icons + labels for: Chat, Dashboard, Context, Settings
- [ ] `max-width: 800px` constraint removed from `.app-shell` in App.svelte
- [ ] `max-width: 600px` constraint removed from `.settings` in Settings.svelte
- [ ] All content panels use full available width
- [ ] Sidebar collapses to icons on narrow viewports (< 768px)
- [ ] Active tab highlighted in sidebar
- [ ] App looks good at 1200px+ widths

## Technical notes

- Modify `App.svelte`: replace `.tab-bar` with `.sidebar`, change layout from column to row
- Add `'context'` to the `Tab` type union
- Remove `max-width` and `margin: 0 auto` from `.app-shell` (line 234) and `.settings` (line 285)
- Sidebar should be ~200px wide with dark background matching existing theme
- Use SVG icons for nav items (keep it simple, no external icon library)

## Tests required

- Visual: sidebar renders with all four nav items
- Visual: clicking each nav item switches content
- Visual: full width layout at desktop sizes

## Out of scope

- Responsive breakpoints below 768px (mobile support)
- Animated sidebar transitions
- User-configurable sidebar width

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-16 | draft | Created |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
