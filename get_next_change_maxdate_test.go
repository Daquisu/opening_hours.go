package openinghours

import (
	"testing"
	"time"
)

func TestGetNextChangeWithMaxDate_BasicLimit(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Start on Monday at 08:00, next change is at 09:00
	start := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC) // Monday

	// With maxdate after the expected change, should return the change
	maxdate := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	change := oh.GetNextChangeWithMaxDate(start, maxdate)

	expected := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	if !change.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, change)
	}
}

func TestGetNextChangeWithMaxDate_LimitBeforeChange(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Start on Monday at 08:00, next change would be at 09:00
	start := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC) // Monday

	// With maxdate before the expected change, should return zero time
	maxdate := time.Date(2024, 1, 15, 8, 30, 0, 0, time.UTC)
	change := oh.GetNextChangeWithMaxDate(start, maxdate)

	if !change.IsZero() {
		t.Errorf("expected zero time when maxdate is before next change, got %v", change)
	}
}

func TestGetNextChangeWithMaxDate_ExactlyAtChange(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Start on Monday at 08:00, next change is at 09:00
	start := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC)

	// With maxdate exactly at the change, should return the change
	maxdate := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	change := oh.GetNextChangeWithMaxDate(start, maxdate)

	expected := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	if !change.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, change)
	}
}

func TestGetNextChangeWithMaxDate_24_7(t *testing.T) {
	oh, err := New("24/7")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	start := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC)
	maxdate := time.Date(2024, 12, 31, 23, 59, 0, 0, time.UTC)
	change := oh.GetNextChangeWithMaxDate(start, maxdate)

	// 24/7 never changes, should return zero time
	if !change.IsZero() {
		t.Errorf("expected zero time for 24/7 within maxdate, got %v", change)
	}
}

func TestGetNextChangeWithMaxDate_AlwaysClosed(t *testing.T) {
	oh, err := New("off")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	start := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC)
	maxdate := time.Date(2024, 12, 31, 23, 59, 0, 0, time.UTC)
	change := oh.GetNextChangeWithMaxDate(start, maxdate)

	// Always closed, should return zero time
	if !change.IsZero() {
		t.Errorf("expected zero time for always closed within maxdate, got %v", change)
	}
}

func TestGetNextChangeWithMaxDate_MultipleChanges(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-12:00,14:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Start on Monday at 10:00 (during first open period)
	// Next change is at 12:00 (closing)
	start := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	// Maxdate after the first change but before the second
	maxdate := time.Date(2024, 1, 15, 13, 0, 0, 0, time.UTC)
	change := oh.GetNextChangeWithMaxDate(start, maxdate)

	expected := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	if !change.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, change)
	}
}

func TestGetNextChangeWithMaxDate_LongRange(t *testing.T) {
	oh, err := New("Su 10:00-14:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Start on Monday at 10:00, next change is Sunday at 10:00
	start := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC) // Monday

	// Maxdate is on Wednesday, before the next Sunday
	maxdate := time.Date(2024, 1, 17, 10, 0, 0, 0, time.UTC) // Wednesday
	change := oh.GetNextChangeWithMaxDate(start, maxdate)

	// No change before Wednesday
	if !change.IsZero() {
		t.Errorf("expected zero time when next change is beyond maxdate, got %v", change)
	}
}

func TestGetNextChangeWithMaxDate_CrossesWeek(t *testing.T) {
	oh, err := New("Su 10:00-14:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Start on Monday at 10:00, next change is Sunday at 10:00
	start := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC) // Monday

	// Maxdate includes the next Sunday
	maxdate := time.Date(2024, 1, 22, 0, 0, 0, 0, time.UTC) // Next Monday
	change := oh.GetNextChangeWithMaxDate(start, maxdate)

	// Next change is Sunday Jan 21 at 10:00
	expected := time.Date(2024, 1, 21, 10, 0, 0, 0, time.UTC)
	if !change.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, change)
	}
}
