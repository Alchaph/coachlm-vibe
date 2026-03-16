package storage

import (
	"bytes"
	"testing"
	"time"
)

func TestSaveAndGetTokensRoundTrip(t *testing.T) {
	db := newTestDB(t)

	accessToken := []byte("encrypted-access-token")
	refreshToken := []byte("encrypted-refresh-token")
	expiresAt := time.Now().Add(6 * time.Hour).Truncate(time.Second)

	if err := db.SaveTokens(accessToken, refreshToken, expiresAt); err != nil {
		t.Fatalf("SaveTokens: %v", err)
	}

	gotAccess, gotRefresh, gotExpires, err := db.GetTokens()
	if err != nil {
		t.Fatalf("GetTokens: %v", err)
	}

	if !bytes.Equal(gotAccess, accessToken) {
		t.Errorf("access token mismatch: got %x, want %x", gotAccess, accessToken)
	}
	if !bytes.Equal(gotRefresh, refreshToken) {
		t.Errorf("refresh token mismatch: got %x, want %x", gotRefresh, refreshToken)
	}
	if !gotExpires.Equal(expiresAt) {
		t.Errorf("expiresAt mismatch: got %v, want %v", gotExpires, expiresAt)
	}
}

func TestGetTokensEmptyDB(t *testing.T) {
	db := newTestDB(t)

	access, refresh, expiresAt, err := db.GetTokens()
	if err != nil {
		t.Fatalf("GetTokens on empty DB: %v", err)
	}
	if access != nil || refresh != nil {
		t.Errorf("expected nil tokens on empty DB, got access=%x refresh=%x", access, refresh)
	}
	if !expiresAt.IsZero() {
		t.Errorf("expected zero expiresAt on empty DB, got %v", expiresAt)
	}
}

func TestDeleteTokens(t *testing.T) {
	db := newTestDB(t)

	if err := db.SaveTokens([]byte("a"), []byte("r"), time.Now()); err != nil {
		t.Fatalf("SaveTokens: %v", err)
	}

	if err := db.DeleteTokens(); err != nil {
		t.Fatalf("DeleteTokens: %v", err)
	}

	access, refresh, _, err := db.GetTokens()
	if err != nil {
		t.Fatalf("GetTokens after delete: %v", err)
	}
	if access != nil || refresh != nil {
		t.Errorf("expected nil tokens after delete, got access=%x refresh=%x", access, refresh)
	}
}

func TestDeleteTokensEmptyDB(t *testing.T) {
	db := newTestDB(t)

	if err := db.DeleteTokens(); err != nil {
		t.Fatalf("DeleteTokens on empty DB should not error: %v", err)
	}
}

func TestSaveTokensUpsert(t *testing.T) {
	db := newTestDB(t)

	if err := db.SaveTokens([]byte("old-access"), []byte("old-refresh"), time.Now()); err != nil {
		t.Fatalf("first SaveTokens: %v", err)
	}

	newAccess := []byte("new-access")
	newRefresh := []byte("new-refresh")
	newExpires := time.Now().Add(12 * time.Hour).Truncate(time.Second)

	if err := db.SaveTokens(newAccess, newRefresh, newExpires); err != nil {
		t.Fatalf("second SaveTokens (upsert): %v", err)
	}

	gotAccess, gotRefresh, gotExpires, err := db.GetTokens()
	if err != nil {
		t.Fatalf("GetTokens after upsert: %v", err)
	}

	if !bytes.Equal(gotAccess, newAccess) {
		t.Errorf("access token not updated: got %x, want %x", gotAccess, newAccess)
	}
	if !bytes.Equal(gotRefresh, newRefresh) {
		t.Errorf("refresh token not updated: got %x, want %x", gotRefresh, newRefresh)
	}
	if !gotExpires.Equal(newExpires) {
		t.Errorf("expiresAt not updated: got %v, want %v", gotExpires, newExpires)
	}
}
