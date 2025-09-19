// (c) Copyright IBM Corp. 2025
package instana

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpanCategoryString(t *testing.T) {
	tests := []struct {
		name     string
		category spanCategory
		expected string
	}{
		{
			name:     "Logging category",
			category: logging,
			expected: "logging",
		},
		{
			name:     "Unknown category",
			category: unknown,
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.category.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestTracerOptionsDisableAllCategories(t *testing.T) {
	opts := &TracerOptions{}
	opts.DisableAllCategories()

	expectedCategories := []spanCategory{logging}

	// Check if all categories are disabled
	for _, category := range expectedCategories {
		if !opts.Disable[category.String()] {
			t.Errorf("Category %s should be disabled", category)
		}
	}

	// Check if the map has the correct size
	if len(opts.Disable) != len(expectedCategories) {
		t.Errorf("Expected %d disabled categories, got %d", len(expectedCategories), len(opts.Disable))
	}
}

func TestGetSpanCategory(t *testing.T) {
	tests := []struct {
		name        string
		operation   string
		expectedCat spanCategory
	}{
		{
			name:        "Log span",
			operation:   string(LogSpanType),
			expectedCat: logging,
		},
		{
			name:        "Unknown span type",
			operation:   "unknown-span-type",
			expectedCat: unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			span := &spanS{
				Operation: tt.operation,
			}

			result := span.getSpanCategory()

			if result != tt.expectedCat {
				t.Errorf("Expected category %s, got %s", tt.expectedCat, result)
			}
		})
	}
}

func TestSpanCategoryEnabled(t *testing.T) {
	tests := []struct {
		name     string
		category spanCategory
		disable  map[string]bool
		expected bool
	}{
		{
			name:     "Category enabled when no categories are disabled",
			category: logging,
			disable:  map[string]bool{},
			expected: true,
		},
		{
			name:     "Category disabled when specifically disabled",
			category: logging,
			disable:  map[string]bool{"logging": true},
			expected: false,
		},
		{
			name:     "Unknown category always enabled",
			category: unknown,
			disable:  map[string]bool{"logging": true},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			InitCollector(&Options{
				Tracer: TracerOptions{
					Disable: tc.disable,
				},
			})
			defer ShutdownCollector()

			result := tc.category.Enabled()

			assert.Equal(t, tc.expected, result)
		})
	}
}
