package context

import (
	"fmt"
	"strings"

	"coachlm/internal/storage"
)

// FormatGearBlock formats gear information for LLM context.
// Returns empty string if gear slice is nil or empty.
func FormatGearBlock(gear []*storage.Gear) string {
	if len(gear) == 0 {
		return ""
	}

	var lines []string
	lines = append(lines, "## Gear")

	for _, g := range gear {
		parts := []string{g.Name}

		brand := g.BrandName
		if g.ModelName != "" && g.ModelName != g.Name {
			if brand != "" {
				brand += " " + g.ModelName
			} else {
				brand = g.ModelName
			}
		}
		if brand != "" {
			parts = append(parts, fmt.Sprintf("(%s)", brand))
		}

		parts = append(parts, fmt.Sprintf("— %.0f km", g.Distance/1000.0))

		if g.IsPrimary {
			parts = append(parts, "[primary]")
		}

		lines = append(lines, "- "+strings.Join(parts, " "))
	}

	return strings.Join(lines, "\n")
}
