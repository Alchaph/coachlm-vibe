package storage

import (
	"fmt"
	"sync"
	"testing"
)

func TestCreateSessionAndGetSession(t *testing.T) {
	db := newTestDB(t)

	session, err := db.CreateSession("ses-001")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if session.ID != "ses-001" {
		t.Errorf("ID = %q, want %q", session.ID, "ses-001")
	}
	if session.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}

	got, err := db.GetSession("ses-001")
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if got == nil {
		t.Fatal("expected session, got nil")
	}
	if got.ID != session.ID {
		t.Errorf("ID = %q, want %q", got.ID, session.ID)
	}
}

func TestGetSessionNotFound(t *testing.T) {
	db := newTestDB(t)

	got, err := db.GetSession("nonexistent")
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil session, got %+v", got)
	}
}

func TestCreateSessionEmptyID(t *testing.T) {
	db := newTestDB(t)

	cases := []string{"", "   ", "\t\n"}
	for _, id := range cases {
		_, err := db.CreateSession(id)
		if err == nil {
			t.Errorf("expected error for empty session id %q", id)
		}
	}
}

func TestCreateSessionDuplicateID(t *testing.T) {
	db := newTestDB(t)

	_, err := db.CreateSession("dup-1")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	_, err = db.CreateSession("dup-1")
	if err == nil {
		t.Fatal("expected error creating duplicate session")
	}
}

func TestListSessions(t *testing.T) {
	db := newTestDB(t)

	ids := []string{"ses-a", "ses-b", "ses-c"}
	for _, id := range ids {
		if _, err := db.CreateSession(id); err != nil {
			t.Fatalf("CreateSession(%q): %v", id, err)
		}
	}

	sessions, err := db.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(sessions) != 3 {
		t.Fatalf("len = %d, want 3", len(sessions))
	}

	// DESC order: newest first — all have same timestamp so fallback is id DESC
	if sessions[0].ID != "ses-c" {
		t.Errorf("sessions[0].ID = %q, want %q", sessions[0].ID, "ses-c")
	}
}

func TestListSessionsEmpty(t *testing.T) {
	db := newTestDB(t)

	sessions, err := db.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if sessions == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(sessions))
	}
}

func TestSaveMessageAndGetMessages(t *testing.T) {
	db := newTestDB(t)

	if _, err := db.CreateSession("ses-msg"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	msg, err := db.SaveMessage("ses-msg", "user", "Hello coach")
	if err != nil {
		t.Fatalf("SaveMessage: %v", err)
	}
	if msg.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if msg.SessionID != "ses-msg" {
		t.Errorf("SessionID = %q, want %q", msg.SessionID, "ses-msg")
	}
	if msg.Role != "user" {
		t.Errorf("Role = %q, want %q", msg.Role, "user")
	}
	if msg.Content != "Hello coach" {
		t.Errorf("Content = %q, want %q", msg.Content, "Hello coach")
	}
	if msg.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}

	msgs, err := db.GetMessages("ses-msg")
	if err != nil {
		t.Fatalf("GetMessages: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("len = %d, want 1", len(msgs))
	}
	if msgs[0].Content != "Hello coach" {
		t.Errorf("round-trip content mismatch: got %q", msgs[0].Content)
	}
}

func TestSaveMessageAllRoles(t *testing.T) {
	db := newTestDB(t)

	if _, err := db.CreateSession("ses-roles"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	roles := []string{"user", "assistant", "system"}
	for _, role := range roles {
		_, err := db.SaveMessage("ses-roles", role, "content for "+role)
		if err != nil {
			t.Errorf("SaveMessage with role %q: %v", role, err)
		}
	}

	msgs, err := db.GetMessages("ses-roles")
	if err != nil {
		t.Fatalf("GetMessages: %v", err)
	}
	if len(msgs) != 3 {
		t.Fatalf("len = %d, want 3", len(msgs))
	}
}

func TestSaveMessageInvalidRole(t *testing.T) {
	db := newTestDB(t)

	if _, err := db.CreateSession("ses-bad-role"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	invalidRoles := []string{"", "admin", "bot", "USER", "Assistant"}
	for _, role := range invalidRoles {
		_, err := db.SaveMessage("ses-bad-role", role, "content")
		if err == nil {
			t.Errorf("expected error for invalid role %q", role)
		}
	}
}

func TestSaveMessageEmptyContent(t *testing.T) {
	db := newTestDB(t)

	if _, err := db.CreateSession("ses-empty"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	cases := []string{"", "   ", "\t\n"}
	for _, content := range cases {
		_, err := db.SaveMessage("ses-empty", "user", content)
		if err == nil {
			t.Errorf("expected error for empty content %q", content)
		}
	}
}

func TestSaveMessageEmptySessionID(t *testing.T) {
	db := newTestDB(t)

	_, err := db.SaveMessage("", "user", "content")
	if err == nil {
		t.Fatal("expected error for empty session id")
	}
}

func TestSaveMessageForeignKeyViolation(t *testing.T) {
	db := newTestDB(t)

	_, err := db.SaveMessage("nonexistent-session", "user", "hello")
	if err == nil {
		t.Fatal("expected error for nonexistent session (FK violation)")
	}
}

func TestGetMessagesOrdering(t *testing.T) {
	db := newTestDB(t)

	if _, err := db.CreateSession("ses-order"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	contents := []string{"first", "second", "third"}
	for _, c := range contents {
		if _, err := db.SaveMessage("ses-order", "user", c); err != nil {
			t.Fatalf("SaveMessage(%q): %v", c, err)
		}
	}

	msgs, err := db.GetMessages("ses-order")
	if err != nil {
		t.Fatalf("GetMessages: %v", err)
	}
	if len(msgs) != 3 {
		t.Fatalf("len = %d, want 3", len(msgs))
	}
	for i, want := range contents {
		if msgs[i].Content != want {
			t.Errorf("msgs[%d].Content = %q, want %q", i, msgs[i].Content, want)
		}
	}
}

func TestGetMessagesEmptySession(t *testing.T) {
	db := newTestDB(t)

	if _, err := db.CreateSession("ses-empty-msgs"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	msgs, err := db.GetMessages("ses-empty-msgs")
	if err != nil {
		t.Fatalf("GetMessages: %v", err)
	}
	if msgs == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(msgs) != 0 {
		t.Errorf("expected 0 messages, got %d", len(msgs))
	}
}

func TestGetMessagesNonexistentSession(t *testing.T) {
	db := newTestDB(t)

	msgs, err := db.GetMessages("does-not-exist")
	if err != nil {
		t.Fatalf("GetMessages: %v", err)
	}
	if msgs == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(msgs) != 0 {
		t.Errorf("expected 0 messages, got %d", len(msgs))
	}
}

func TestDeleteSession(t *testing.T) {
	db := newTestDB(t)

	if _, err := db.CreateSession("ses-del"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if _, err := db.SaveMessage("ses-del", "user", "hello"); err != nil {
		t.Fatalf("SaveMessage: %v", err)
	}

	if err := db.DeleteSession("ses-del"); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}

	got, err := db.GetSession("ses-del")
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if got != nil {
		t.Error("expected nil after delete")
	}

	msgs, err := db.GetMessages("ses-del")
	if err != nil {
		t.Fatalf("GetMessages: %v", err)
	}
	if len(msgs) != 0 {
		t.Errorf("expected 0 messages after cascade delete, got %d", len(msgs))
	}
}

func TestDeleteSessionNotFound(t *testing.T) {
	db := newTestDB(t)

	err := db.DeleteSession("nonexistent")
	if err == nil {
		t.Fatal("expected error deleting non-existent session")
	}
}

func TestManyMessages(t *testing.T) {
	db := newTestDB(t)

	if _, err := db.CreateSession("ses-bulk"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	const n = 1100
	for i := 0; i < n; i++ {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		_, err := db.SaveMessage("ses-bulk", role, fmt.Sprintf("msg-%d", i))
		if err != nil {
			t.Fatalf("SaveMessage #%d: %v", i, err)
		}
	}

	msgs, err := db.GetMessages("ses-bulk")
	if err != nil {
		t.Fatalf("GetMessages: %v", err)
	}
	if len(msgs) != n {
		t.Errorf("len = %d, want %d", len(msgs), n)
	}
	if msgs[0].Content != "msg-0" {
		t.Errorf("first message = %q, want %q", msgs[0].Content, "msg-0")
	}
	if msgs[n-1].Content != fmt.Sprintf("msg-%d", n-1) {
		t.Errorf("last message = %q, want %q", msgs[n-1].Content, fmt.Sprintf("msg-%d", n-1))
	}
}

func TestConcurrentWrites(t *testing.T) {
	db := newTestDB(t)

	if _, err := db.CreateSession("ses-concurrent"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	const goroutines = 20
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := db.SaveMessage("ses-concurrent", "user", fmt.Sprintf("concurrent-%d", i))
			if err != nil {
				errs <- fmt.Errorf("goroutine %d: %w", i, err)
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent write error: %v", err)
	}

	msgs, err := db.GetMessages("ses-concurrent")
	if err != nil {
		t.Fatalf("GetMessages: %v", err)
	}
	if len(msgs) != goroutines {
		t.Errorf("len = %d, want %d", len(msgs), goroutines)
	}
}

func TestConcurrentSessionCreation(t *testing.T) {
	db := newTestDB(t)

	const goroutines = 20
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := db.CreateSession(fmt.Sprintf("concurrent-ses-%d", i))
			if err != nil {
				errs <- fmt.Errorf("goroutine %d: %w", i, err)
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent session create error: %v", err)
	}

	sessions, err := db.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(sessions) != goroutines {
		t.Errorf("len = %d, want %d", len(sessions), goroutines)
	}
}

func TestMultipleSessionsIsolation(t *testing.T) {
	db := newTestDB(t)

	if _, err := db.CreateSession("ses-A"); err != nil {
		t.Fatalf("CreateSession A: %v", err)
	}
	if _, err := db.CreateSession("ses-B"); err != nil {
		t.Fatalf("CreateSession B: %v", err)
	}

	if _, err := db.SaveMessage("ses-A", "user", "message for A"); err != nil {
		t.Fatalf("SaveMessage A: %v", err)
	}
	if _, err := db.SaveMessage("ses-B", "user", "message for B"); err != nil {
		t.Fatalf("SaveMessage B: %v", err)
	}

	msgsA, err := db.GetMessages("ses-A")
	if err != nil {
		t.Fatalf("GetMessages A: %v", err)
	}
	if len(msgsA) != 1 {
		t.Fatalf("session A: len = %d, want 1", len(msgsA))
	}
	if msgsA[0].Content != "message for A" {
		t.Errorf("session A message = %q, want %q", msgsA[0].Content, "message for A")
	}

	msgsB, err := db.GetMessages("ses-B")
	if err != nil {
		t.Fatalf("GetMessages B: %v", err)
	}
	if len(msgsB) != 1 {
		t.Fatalf("session B: len = %d, want 1", len(msgsB))
	}
	if msgsB[0].Content != "message for B" {
		t.Errorf("session B message = %q, want %q", msgsB[0].Content, "message for B")
	}
}

func TestDeleteSessionCascadesMessages(t *testing.T) {
	db := newTestDB(t)

	if _, err := db.CreateSession("ses-cascade"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	for i := 0; i < 5; i++ {
		if _, err := db.SaveMessage("ses-cascade", "user", fmt.Sprintf("msg-%d", i)); err != nil {
			t.Fatalf("SaveMessage %d: %v", i, err)
		}
	}

	if err := db.DeleteSession("ses-cascade"); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}

	msgs, err := db.GetMessages("ses-cascade")
	if err != nil {
		t.Fatalf("GetMessages: %v", err)
	}
	if len(msgs) != 0 {
		t.Errorf("expected 0 messages after cascade delete, got %d", len(msgs))
	}
}

func TestDeleteOneSessionKeepsOthers(t *testing.T) {
	db := newTestDB(t)

	if _, err := db.CreateSession("keep"); err != nil {
		t.Fatalf("CreateSession keep: %v", err)
	}
	if _, err := db.CreateSession("delete"); err != nil {
		t.Fatalf("CreateSession delete: %v", err)
	}

	if _, err := db.SaveMessage("keep", "user", "keeper message"); err != nil {
		t.Fatalf("SaveMessage keep: %v", err)
	}
	if _, err := db.SaveMessage("delete", "user", "doomed message"); err != nil {
		t.Fatalf("SaveMessage delete: %v", err)
	}

	if err := db.DeleteSession("delete"); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}

	sessions, err := db.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].ID != "keep" {
		t.Errorf("remaining session = %q, want %q", sessions[0].ID, "keep")
	}

	msgs, err := db.GetMessages("keep")
	if err != nil {
		t.Fatalf("GetMessages keep: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message in kept session, got %d", len(msgs))
	}
}

func TestSaveMessageUnicodeContent(t *testing.T) {
	db := newTestDB(t)

	if _, err := db.CreateSession("ses-unicode"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	content := "🏃‍♂️ Running at 4:30/km pace — très bien! 日本語テスト"
	msg, err := db.SaveMessage("ses-unicode", "assistant", content)
	if err != nil {
		t.Fatalf("SaveMessage: %v", err)
	}
	if msg.Content != content {
		t.Errorf("content = %q, want %q", msg.Content, content)
	}

	msgs, err := db.GetMessages("ses-unicode")
	if err != nil {
		t.Fatalf("GetMessages: %v", err)
	}
	if len(msgs) != 1 || msgs[0].Content != content {
		t.Errorf("round-trip unicode mismatch")
	}
}

func TestSaveMessageLargeContent(t *testing.T) {
	db := newTestDB(t)

	if _, err := db.CreateSession("ses-large"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	large := make([]byte, 100_000)
	for i := range large {
		large[i] = 'x'
	}
	content := string(large)

	msg, err := db.SaveMessage("ses-large", "assistant", content)
	if err != nil {
		t.Fatalf("SaveMessage: %v", err)
	}
	if len(msg.Content) != 100_000 {
		t.Errorf("content length = %d, want 100000", len(msg.Content))
	}
}
