package context

import (
	"strings"
	"testing"

	"coachlm/internal/storage"
)

func TestFormatGearBlock_Nil(t *testing.T) {
	result := FormatGearBlock(nil)
	if result != "" {
		t.Errorf("expected empty string for nil gear, got %q", result)
	}
}

func TestFormatGearBlock_Empty(t *testing.T) {
	result := FormatGearBlock([]*storage.Gear{})
	if result != "" {
		t.Errorf("expected empty string for empty gear, got %q", result)
	}
}

func TestFormatGearBlock_SingleGear(t *testing.T) {
	gear := []*storage.Gear{
		{
			Name:      "Vaporfly",
			BrandName: "Nike",
			ModelName: "Vaporfly 4%",
			Distance:  350000.0,
			IsPrimary: true,
		},
	}

	result := FormatGearBlock(gear)

	if !strings.HasPrefix(result, "## Gear") {
		t.Error("missing heading")
	}
	if !strings.Contains(result, "Vaporfly") {
		t.Errorf("missing gear name, got:\n%s", result)
	}
	if !strings.Contains(result, "Nike Vaporfly 4%") {
		t.Errorf("missing brand+model, got:\n%s", result)
	}
	if !strings.Contains(result, "350 km") {
		t.Errorf("missing distance, got:\n%s", result)
	}
	if !strings.Contains(result, "[primary]") {
		t.Errorf("missing primary marker, got:\n%s", result)
	}
}

func TestFormatGearBlock_MultipleGear(t *testing.T) {
	gear := []*storage.Gear{
		{Name: "Race Shoes", BrandName: "Nike", Distance: 200000.0, IsPrimary: true},
		{Name: "Daily Trainer", BrandName: "ASICS", ModelName: "Gel Nimbus", Distance: 500000.0},
	}

	result := FormatGearBlock(gear)

	if !strings.Contains(result, "Race Shoes") {
		t.Error("missing first gear")
	}
	if !strings.Contains(result, "Daily Trainer") {
		t.Error("missing second gear")
	}
	if !strings.Contains(result, "(ASICS Gel Nimbus)") {
		t.Errorf("missing brand+model for second gear, got:\n%s", result)
	}
}

func TestFormatGearBlock_NoBrand(t *testing.T) {
	gear := []*storage.Gear{
		{Name: "Old Shoes", Distance: 800000.0},
	}

	result := FormatGearBlock(gear)

	if strings.Contains(result, "()") {
		t.Error("should not show empty parentheses when no brand")
	}
	if !strings.Contains(result, "800 km") {
		t.Errorf("missing distance, got:\n%s", result)
	}
}

func TestFormatGearBlock_ModelSameAsName(t *testing.T) {
	gear := []*storage.Gear{
		{Name: "Pegasus", BrandName: "Nike", ModelName: "Pegasus", Distance: 100000.0},
	}

	result := FormatGearBlock(gear)

	if strings.Contains(result, "Nike Pegasus") {
		// brand should be just "Nike" since model == name
		count := strings.Count(result, "Pegasus")
		if count > 2 {
			t.Errorf("model name duplicated with gear name, got:\n%s", result)
		}
	}
}
