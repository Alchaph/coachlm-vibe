---
id: S45
title: Fix export context functionality
status: draft
created: 2026-03-17
updated: 2026-03-17
---

# S45 — Fix export context functionality

## User story

As a **runner**,
I want **to export my coaching context to a file**
so that **I can backup my data and import it on another device**.

## Problem

When the user clicks the "Export Context" button in the Settings screen, an error occurs. The frontend calls `ExportContext()` without a file path argument, but the Go backend function requires a `filePath` parameter. The function signature is:

```go
func (a *App) ExportContext(filePath string) error
```

The frontend implementation in `Settings.svelte` (line 65) is:

```javascript
async function exportContext() {
  try {
    await ExportContext()  // Missing filePath argument
    showFeedback('Context exported successfully', 'success')
  } catch (e: any) {
    showFeedback(e?.message || 'Failed to export context', 'error')
  }
}
```

This causes the Wails binding to fail because the required `filePath` parameter is not provided. The error pops up on screen with no helpful context to the user.

## Acceptance criteria

- [ ] Export button triggers a native file save dialog using Wails runtime
- [ ] File dialog defaults to a sensible filename (e.g., `coach-context-YYYY-MM-DD.coachctx`)
- [ ] File dialog filters to `.coachctx` extension
- [ ] If user cancels the dialog, no error is shown (silent failure is OK for cancel)
- [ ] Selected file path is passed to the Go backend's `ExportContext(filePath string)` function
- [ ] Success message is shown after successful export
- [ ] Error message is shown if export fails (with meaningful error details)
- [ ] The export process shows feedback to the user during the operation

## Technical notes

**File location**: `frontend/src/Settings.svelte`

**Required Wails runtime import**:
```typescript
import { SaveFileDialog } from '../wailsjs/runtime/runtime.js'
```

**Implementation approach**:
1. Before calling `ExportContext`, invoke `SaveFileDialog()` with appropriate options
2. Pass the returned filePath to `ExportContext(filePath)`
3. Handle null/empty filePath (user canceled) gracefully

**Example Wails SaveFileDialog usage**:
```typescript
const filePath = await SaveFileDialog({
  defaultFilename: `coach-context-${new Date().toISOString().split('T')[0]}.coachctx`,
  filters: [
    { name: 'CoachLM Context', pattern: '*.coachctx' }
  ]
})
if (filePath) {
  await ExportContext(filePath)
  // ... handle success
}
```

The app.go already imports `wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"` at line 22, so the runtime should be available for frontend dialogs.

**Related files**:
- `frontend/src/Settings.svelte` - Add file dialog before export
- `app.go` - `ExportContext` function at line 596 (no changes needed)
- `internal/exportimport/exportimport.go` - Export logic (no changes needed)

## Tests required

- Manual test: Click "Export Context" → File save dialog appears → Select location → File is created
- Manual test: Click "Export Context" → Cancel dialog → No error shown
- Manual test: Export to read-only location → Error message appears
- Edge case: Export when database is empty → File is created with valid empty envelope

## Out of scope

- Export format changes (the export format is correct; the bug is only about triggering the export)
- Import functionality (this is working correctly)
- Encryption of exported files (future enhancement)
- Progress bar for large exports

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-17 | draft | Created |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
