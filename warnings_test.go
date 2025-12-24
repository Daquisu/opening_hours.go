package openinghours

import (
	"testing"
)

func TestWarnings_NoWarnings(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	warnings := oh.GetWarnings()
	if len(warnings) != 0 {
		t.Errorf("expected no warnings for valid input, got %d warnings: %v", len(warnings), warnings)
	}
}

func TestWarnings_ShortTimeFormat(t *testing.T) {
	oh, err := New("10-12")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	warnings := oh.GetWarnings()
	if len(warnings) == 0 {
		t.Errorf("expected warning about abbreviated time format, got no warnings")
	}

	// Check if any warning mentions abbreviated or short time format
	foundTimeFormatWarning := false
	for _, w := range warnings {
		if containsAny(w, []string{"abbreviated", "short", "time format", "10-12"}) {
			foundTimeFormatWarning = true
			break
		}
	}

	if !foundTimeFormatWarning {
		t.Errorf("expected warning about abbreviated time format, got warnings: %v", warnings)
	}
}

func TestWarnings_RedundantRule(t *testing.T) {
	oh, err := New("24/7; Mo 10:00-12:00 off")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	warnings := oh.GetWarnings()
	if len(warnings) == 0 {
		t.Errorf("expected warning about redundant 24/7 rule, got no warnings")
	}

	// Check if any warning mentions redundant or 24/7
	foundRedundantWarning := false
	for _, w := range warnings {
		if containsAny(w, []string{"redundant", "24/7", "unnecessary"}) {
			foundRedundantWarning = true
			break
		}
	}

	if !foundRedundantWarning {
		t.Errorf("expected warning about redundant 24/7 when there are other rules, got warnings: %v", warnings)
	}
}

func TestWarnings_OverlappingRanges(t *testing.T) {
	oh, err := New("10:00-14:00,12:00-16:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	warnings := oh.GetWarnings()
	if len(warnings) == 0 {
		t.Errorf("expected warning about overlapping time ranges, got no warnings")
	}

	// Check if any warning mentions overlapping
	foundOverlapWarning := false
	for _, w := range warnings {
		if containsAny(w, []string{"overlap", "overlapping"}) {
			foundOverlapWarning = true
			break
		}
	}

	if !foundOverlapWarning {
		t.Errorf("expected warning about overlapping time ranges, got warnings: %v", warnings)
	}
}

func TestWarnings_EmptyComment(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-17:00 \"\"")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	warnings := oh.GetWarnings()
	if len(warnings) == 0 {
		t.Errorf("expected warning about empty comment, got no warnings")
	}

	// Check if any warning mentions empty comment
	foundEmptyCommentWarning := false
	for _, w := range warnings {
		if containsAny(w, []string{"empty comment", "comment is empty", "empty string"}) {
			foundEmptyCommentWarning = true
			break
		}
	}

	if !foundEmptyCommentWarning {
		t.Errorf("expected warning about empty comment, got warnings: %v", warnings)
	}
}

func TestWarnings_MultipleWarnings(t *testing.T) {
	// Input with multiple issues:
	// 1. Overlapping time ranges: 10:00-14:00 and 12:00-16:00
	// 2. Empty comment: ""
	oh, err := New("10:00-14:00,12:00-16:00 \"\"")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	warnings := oh.GetWarnings()
	if len(warnings) < 2 {
		t.Errorf("expected at least 2 warnings for input with multiple issues, got %d warnings: %v", len(warnings), warnings)
	}

	// Check for overlapping ranges warning
	foundOverlapWarning := false
	for _, w := range warnings {
		if containsAny(w, []string{"overlap", "overlapping"}) {
			foundOverlapWarning = true
			break
		}
	}

	if !foundOverlapWarning {
		t.Errorf("expected warning about overlapping time ranges in multiple warnings test, got warnings: %v", warnings)
	}

	// Check for empty comment warning
	foundEmptyCommentWarning := false
	for _, w := range warnings {
		if containsAny(w, []string{"empty comment", "comment is empty", "empty string"}) {
			foundEmptyCommentWarning = true
			break
		}
	}

	if !foundEmptyCommentWarning {
		t.Errorf("expected warning about empty comment in multiple warnings test, got warnings: %v", warnings)
	}
}

func TestWarnings_ValidComplexInput(t *testing.T) {
	// Complex but valid input should produce no warnings
	oh, err := New("Mo-Fr 09:00-17:00; Sa 10:00-14:00; PH off")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	warnings := oh.GetWarnings()
	if len(warnings) != 0 {
		t.Errorf("expected no warnings for valid complex input, got %d warnings: %v", len(warnings), warnings)
	}
}

func TestWarnings_RedundantRuleWithMultipleRules(t *testing.T) {
	// 24/7 with multiple other rules should warn about redundancy
	oh, err := New("24/7; Mo 09:00-17:00; Fr 09:00-15:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	warnings := oh.GetWarnings()
	if len(warnings) == 0 {
		t.Errorf("expected warning about redundant 24/7 with multiple other rules, got no warnings")
	}

	foundRedundantWarning := false
	for _, w := range warnings {
		if containsAny(w, []string{"redundant", "24/7", "unnecessary"}) {
			foundRedundantWarning = true
			break
		}
	}

	if !foundRedundantWarning {
		t.Errorf("expected warning about redundant 24/7, got warnings: %v", warnings)
	}
}

func TestWarnings_MultipleOverlappingRanges(t *testing.T) {
	// Multiple overlapping time ranges
	oh, err := New("08:00-12:00,10:00-14:00,13:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	warnings := oh.GetWarnings()
	if len(warnings) == 0 {
		t.Errorf("expected warnings about overlapping time ranges, got no warnings")
	}

	foundOverlapWarning := false
	for _, w := range warnings {
		if containsAny(w, []string{"overlap", "overlapping"}) {
			foundOverlapWarning = true
			break
		}
	}

	if !foundOverlapWarning {
		t.Errorf("expected warning about overlapping time ranges, got warnings: %v", warnings)
	}
}

func TestWarnings_NoOverlapAdjacentRanges(t *testing.T) {
	// Adjacent but non-overlapping ranges should not warn
	oh, err := New("08:00-12:00,12:00-16:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	warnings := oh.GetWarnings()

	// Check that there's no overlap warning (adjacent ranges are OK)
	for _, w := range warnings {
		if containsAny(w, []string{"overlap", "overlapping"}) {
			t.Errorf("did not expect overlap warning for adjacent ranges, got: %v", warnings)
		}
	}
}

// Helper function to check if a string contains any of the given substrings (case-insensitive)
func containsAny(s string, substrs []string) bool {
	lower := toLower(s)
	for _, substr := range substrs {
		if contains(lower, toLower(substr)) {
			return true
		}
	}
	return false
}

// Helper function for case-insensitive string matching
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c = c + ('a' - 'A')
		}
		result[i] = c
	}
	return string(result)
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
