package openinghours

import (
	"testing"
	"time"
)

func TestFallback_BasicUnknownWithFallback(t *testing.T) {
	// Test basic fallback behavior: primary group has unknown state, fallback provides actual hours
	oh, err := New("Mo-Fr 09:00-17:00 unknown || Mo-Fr 10:00-16:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		name    string
		hour    int
		minute  int
		wantState bool
		desc    string
	}{
		{
			name:    "primary unknown, fallback open",
			hour:    12,
			minute:  0,
			wantState: true,
			desc:    "Mo 12:00: primary is unknown, fallback is open -> should be open",
		},
		{
			name:    "primary closed, fallback closed",
			hour:    18,
			minute:  0,
			wantState: false,
			desc:    "Mo 18:00: primary is closed, fallback is closed -> should be closed",
		},
		{
			name:    "primary unknown, fallback closed",
			hour:    9,
			minute:  30,
			wantState: false,
			desc:    "Mo 09:30: primary is unknown, fallback is closed -> should be closed",
		},
		{
			name:    "primary unknown, fallback open",
			hour:    15,
			minute:  0,
			wantState: true,
			desc:    "Mo 15:00: primary is unknown, fallback is open -> should be open",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTime := time.Date(2024, 1, 15, tt.hour, tt.minute, 0, 0, time.UTC) // Monday
			got := oh.GetState(testTime)
			if got != tt.wantState {
				t.Errorf("%s: got state %v, want %v", tt.desc, got, tt.wantState)
			}
		})
	}
}

func TestFallback_OpenDoesNotUseFallback(t *testing.T) {
	// Test fallback behavior when primary matches vs when primary doesn't match
	oh, err := New("Mo-Fr 09:00-17:00; Sa 10:00-14:00 || 24/7")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		name      string
		dayOfWeek time.Weekday
		hour      int
		minute    int
		wantState bool
		desc      string
	}{
		{
			name:      "primary open",
			dayOfWeek: time.Monday,
			hour:      10,
			minute:    0,
			wantState: true,
			desc:      "Mo 10:00: primary is open -> should be open (fallback not checked)",
		},
		{
			name:      "no primary match, fallback applies",
			dayOfWeek: time.Sunday,
			hour:      10,
			minute:    0,
			wantState: true,
			desc:      "Su 10:00: no primary matches -> fallback (24/7) is checked -> open",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate the date for the desired weekday
			// Start from a known Monday (2024-01-15)
			baseDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC) // Monday
			daysToAdd := int(tt.dayOfWeek - baseDate.Weekday())
			if daysToAdd < 0 {
				daysToAdd += 7
			}
			testDate := baseDate.AddDate(0, 0, daysToAdd)
			testTime := time.Date(testDate.Year(), testDate.Month(), testDate.Day(),
				tt.hour, tt.minute, 0, 0, time.UTC)

			got := oh.GetState(testTime)
			if got != tt.wantState {
				t.Errorf("%s: got state %v, want %v", tt.desc, got, tt.wantState)
			}
		})
	}
}

func TestFallback_UnknownComment(t *testing.T) {
	// Test that comments from unknown state in primary are preserved
	oh, err := New("Mo-Fr 09:00-17:00 unknown \"call for hours\" || Mo-Fr 10:00-16:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	testTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC) // Monday 12:00

	// GetComment should return the comment from the primary group
	comment := oh.GetComment(testTime)
	if comment != "call for hours" {
		t.Errorf("Mo 12:00: GetComment should return 'call for hours' (from primary), got '%s'", comment)
	}

	// GetState should return true (from fallback)
	if !oh.GetState(testTime) {
		t.Errorf("Mo 12:00: GetState should return true (from fallback)")
	}
}

func TestFallback_MultipleFallbackGroups(t *testing.T) {
	// Test chaining multiple fallback groups
	oh, err := New("Mo-Fr unknown || Mo-Fr 09:00-17:00 unknown || 24/7")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Monday 10:00
	// - First group: Mo-Fr unknown -> state is unknown
	// - Second group: Mo-Fr 09:00-17:00 unknown -> state is unknown (10:00 is in range)
	// - Third group: 24/7 -> state is open
	// Expected: should be open (third group is checked because first two return unknown)
	testTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC) // Monday 10:00

	if !oh.GetState(testTime) {
		t.Errorf("Mo 10:00: first is unknown, second is unknown, third is open -> should be open, got closed")
	}
}

func TestFallback_NoFallback(t *testing.T) {
	// Test that when no || is present, the rule works as normal
	oh, err := New("Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Monday 10:00 - should be open
	testTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	if !oh.GetState(testTime) {
		t.Errorf("Mo 10:00: should be open")
	}

	// Monday 18:00 - should be closed
	closedTime := time.Date(2024, 1, 15, 18, 0, 0, 0, time.UTC)
	if oh.GetState(closedTime) {
		t.Errorf("Mo 18:00: should be closed")
	}
}
