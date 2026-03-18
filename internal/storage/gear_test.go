package storage

import (
	"testing"
)

func sampleGear(externalID string) *Gear {
	return &Gear{
		ExternalID:  externalID,
		Name:        "Pegasus 40",
		BrandName:   "Nike",
		ModelName:   "Pegasus",
		Description: "Daily trainer",
		Distance:    500000,
		IsPrimary:   true,
		IsRetired:   false,
	}
}

func TestSaveAndGetGear(t *testing.T) {
	db := newTestDB(t)
	g := sampleGear("g12345")

	if err := db.SaveGear(g); err != nil {
		t.Fatalf("SaveGear: %v", err)
	}

	got, err := db.GetGearByExternalID("g12345")
	if err != nil {
		t.Fatalf("GetGearByExternalID: %v", err)
	}
	if got == nil {
		t.Fatal("GetGearByExternalID returned nil after save")
	}

	if got.ExternalID != g.ExternalID {
		t.Errorf("ExternalID = %q, want %q", got.ExternalID, g.ExternalID)
	}
	if got.Name != g.Name {
		t.Errorf("Name = %q, want %q", got.Name, g.Name)
	}
	if got.BrandName != g.BrandName {
		t.Errorf("BrandName = %q, want %q", got.BrandName, g.BrandName)
	}
	if got.ModelName != g.ModelName {
		t.Errorf("ModelName = %q, want %q", got.ModelName, g.ModelName)
	}
	if got.Description != g.Description {
		t.Errorf("Description = %q, want %q", got.Description, g.Description)
	}
	if got.Distance != g.Distance {
		t.Errorf("Distance = %f, want %f", got.Distance, g.Distance)
	}
	if got.IsPrimary != g.IsPrimary {
		t.Errorf("IsPrimary = %v, want %v", got.IsPrimary, g.IsPrimary)
	}
	if got.IsRetired != g.IsRetired {
		t.Errorf("IsRetired = %v, want %v", got.IsRetired, g.IsRetired)
	}
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if got.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}

func TestSaveGearUpdate(t *testing.T) {
	db := newTestDB(t)
	g := sampleGear("g12345")

	if err := db.SaveGear(g); err != nil {
		t.Fatalf("first SaveGear: %v", err)
	}

	first, err := db.GetGearByExternalID("g12345")
	if err != nil {
		t.Fatalf("first GetGearByExternalID: %v", err)
	}

	g.Distance = 750000
	g.Name = "Pegasus 40 worn"
	if err := db.SaveGear(g); err != nil {
		t.Fatalf("second SaveGear: %v", err)
	}

	second, err := db.GetGearByExternalID("g12345")
	if err != nil {
		t.Fatalf("second GetGearByExternalID: %v", err)
	}

	if second.Distance != 750000 {
		t.Errorf("Distance after update = %f, want 750000", second.Distance)
	}
	if second.Name != "Pegasus 40 worn" {
		t.Errorf("Name after update = %q, want 'Pegasus 40 worn'", second.Name)
	}
	if !second.CreatedAt.Equal(first.CreatedAt) {
		t.Errorf("CreatedAt changed after update: %v → %v", first.CreatedAt, second.CreatedAt)
	}
}

func TestGetGearNotFound(t *testing.T) {
	db := newTestDB(t)

	got, err := db.GetGearByExternalID("g_nonexistent")
	if err != nil {
		t.Fatalf("GetGearByExternalID for unknown id: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for unknown external id, got %+v", got)
	}
}

func TestListGear(t *testing.T) {
	db := newTestDB(t)

	items := []*Gear{
		{ExternalID: "g001", Name: "Shoe A", Distance: 200000, IsPrimary: false, IsRetired: false},
		{ExternalID: "g002", Name: "Shoe B", Distance: 800000, IsPrimary: true, IsRetired: false},
		{ExternalID: "g003", Name: "Shoe C", Distance: 500000, IsPrimary: false, IsRetired: false},
		{ExternalID: "g004", Name: "Shoe D (retired)", Distance: 1200000, IsPrimary: false, IsRetired: true},
	}
	for _, item := range items {
		if err := db.SaveGear(item); err != nil {
			t.Fatalf("SaveGear %s: %v", item.ExternalID, err)
		}
	}

	list, err := db.ListGear()
	if err != nil {
		t.Fatalf("ListGear: %v", err)
	}

	if len(list) != 3 {
		t.Fatalf("ListGear returned %d items, want 3 (retired excluded)", len(list))
	}

	if list[0].ExternalID != "g002" {
		t.Errorf("first item ExternalID = %q, want g002 (highest distance)", list[0].ExternalID)
	}
	if list[1].ExternalID != "g003" {
		t.Errorf("second item ExternalID = %q, want g003", list[1].ExternalID)
	}
	if list[2].ExternalID != "g001" {
		t.Errorf("third item ExternalID = %q, want g001", list[2].ExternalID)
	}

	for _, g := range list {
		if g.IsRetired {
			t.Errorf("ListGear returned retired gear %q", g.ExternalID)
		}
	}
}

func TestListGearEmpty(t *testing.T) {
	db := newTestDB(t)

	list, err := db.ListGear()
	if err != nil {
		t.Fatalf("ListGear on empty db: %v", err)
	}
	if list == nil {
		t.Error("ListGear should return empty slice, not nil")
	}
	if len(list) != 0 {
		t.Errorf("ListGear returned %d items on empty db, want 0", len(list))
	}
}

func TestSaveGearNil(t *testing.T) {
	db := newTestDB(t)

	err := db.SaveGear(nil)
	if err == nil {
		t.Error("SaveGear(nil) should return an error")
	}
}
