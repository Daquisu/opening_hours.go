package openinghours

import (
	"testing"
	"time"
)

func TestGetOpenDuration_SimpleTimeRange(t *testing.T) {
	// Test: Simple time range "09:00-17:00" over one day should return 8 hours open, 0 unknown
	oh, err := New("09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)

	openDuration, unknownDuration := oh.GetOpenDuration(from, to)

	expectedOpen := 8 * time.Hour
	expectedUnknown := time.Duration(0)

	if openDuration != expectedOpen {
		t.Errorf("GetOpenDuration with '09:00-17:00' over 1 day: got open duration %v, want %v", openDuration, expectedOpen)
	}

	if unknownDuration != expectedUnknown {
		t.Errorf("GetOpenDuration with '09:00-17:00' over 1 day: got unknown duration %v, want %v", unknownDuration, expectedUnknown)
	}
}

func TestGetOpenDuration_WeekdayConstraint(t *testing.T) {
	// Test: Weekday constraint "Mo-Fr 09:00-17:00" over a week should return 40 hours open
	oh, err := New("Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Jan 15, 2024 is Monday
	// Jan 22, 2024 is the following Monday
	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 22, 0, 0, 0, 0, time.UTC)

	openDuration, unknownDuration := oh.GetOpenDuration(from, to)

	expectedOpen := 40 * time.Hour // 5 days * 8 hours
	expectedUnknown := time.Duration(0)

	if openDuration != expectedOpen {
		t.Errorf("GetOpenDuration with 'Mo-Fr 09:00-17:00' over 1 week: got open duration %v, want %v", openDuration, expectedOpen)
	}

	if unknownDuration != expectedUnknown {
		t.Errorf("GetOpenDuration with 'Mo-Fr 09:00-17:00' over 1 week: got unknown duration %v, want %v", unknownDuration, expectedUnknown)
	}
}

func TestGetOpenDuration_AlwaysOpen(t *testing.T) {
	// Test: "24/7" over one day should return 24 hours open
	oh, err := New("24/7")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)

	openDuration, unknownDuration := oh.GetOpenDuration(from, to)

	expectedOpen := 24 * time.Hour
	expectedUnknown := time.Duration(0)

	if openDuration != expectedOpen {
		t.Errorf("GetOpenDuration with '24/7' over 1 day: got open duration %v, want %v", openDuration, expectedOpen)
	}

	if unknownDuration != expectedUnknown {
		t.Errorf("GetOpenDuration with '24/7' over 1 day: got unknown duration %v, want %v", unknownDuration, expectedUnknown)
	}
}

func TestGetOpenDuration_UnknownState(t *testing.T) {
	// Test: Unknown state "Mo-Fr 09:00-17:00 unknown" should return 0 open and 8 hours unknown for one weekday
	oh, err := New("Mo-Fr 09:00-17:00 unknown")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Jan 15, 2024 is Monday
	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)

	openDuration, unknownDuration := oh.GetOpenDuration(from, to)

	expectedOpen := time.Duration(0)
	expectedUnknown := 8 * time.Hour

	if openDuration != expectedOpen {
		t.Errorf("GetOpenDuration with 'Mo-Fr 09:00-17:00 unknown' over 1 day: got open duration %v, want %v", openDuration, expectedOpen)
	}

	if unknownDuration != expectedUnknown {
		t.Errorf("GetOpenDuration with 'Mo-Fr 09:00-17:00 unknown' over 1 day: got unknown duration %v, want %v", unknownDuration, expectedUnknown)
	}
}

func TestGetOpenDuration_PartialDay(t *testing.T) {
	// Test: Partial day calculation - from 10:00 to 15:00 with "09:00-17:00" should return 5 hours
	oh, err := New("09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	from := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 15, 15, 0, 0, 0, time.UTC)

	openDuration, unknownDuration := oh.GetOpenDuration(from, to)

	expectedOpen := 5 * time.Hour
	expectedUnknown := time.Duration(0)

	if openDuration != expectedOpen {
		t.Errorf("GetOpenDuration with '09:00-17:00' from 10:00 to 15:00: got open duration %v, want %v", openDuration, expectedOpen)
	}

	if unknownDuration != expectedUnknown {
		t.Errorf("GetOpenDuration with '09:00-17:00' from 10:00 to 15:00: got unknown duration %v, want %v", unknownDuration, expectedUnknown)
	}
}

func TestGetOpenDuration_AlwaysClosed(t *testing.T) {
	// Test: Off hours - "off" should return 0 open, 0 unknown
	oh, err := New("off")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)

	openDuration, unknownDuration := oh.GetOpenDuration(from, to)

	expectedOpen := time.Duration(0)
	expectedUnknown := time.Duration(0)

	if openDuration != expectedOpen {
		t.Errorf("GetOpenDuration with 'off' over 1 day: got open duration %v, want %v", openDuration, expectedOpen)
	}

	if unknownDuration != expectedUnknown {
		t.Errorf("GetOpenDuration with 'off' over 1 day: got unknown duration %v, want %v", unknownDuration, expectedUnknown)
	}
}

func TestGetOpenDuration_MultipleTimeRanges(t *testing.T) {
	// Test: Multiple time ranges "08:00-12:00,14:00-18:00" over one day
	oh, err := New("08:00-12:00,14:00-18:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)

	openDuration, unknownDuration := oh.GetOpenDuration(from, to)

	expectedOpen := 8 * time.Hour // 4 hours + 4 hours
	expectedUnknown := time.Duration(0)

	if openDuration != expectedOpen {
		t.Errorf("GetOpenDuration with '08:00-12:00,14:00-18:00' over 1 day: got open duration %v, want %v", openDuration, expectedOpen)
	}

	if unknownDuration != expectedUnknown {
		t.Errorf("GetOpenDuration with '08:00-12:00,14:00-18:00' over 1 day: got unknown duration %v, want %v", unknownDuration, expectedUnknown)
	}
}

func TestGetOpenDuration_MidnightSpanning(t *testing.T) {
	// Test: Time range spanning midnight "22:00-02:00"
	oh, err := New("22:00-02:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)

	openDuration, unknownDuration := oh.GetOpenDuration(from, to)

	expectedOpen := 4 * time.Hour // 2 hours (22:00-00:00) + 2 hours (00:00-02:00)
	expectedUnknown := time.Duration(0)

	if openDuration != expectedOpen {
		t.Errorf("GetOpenDuration with '22:00-02:00' over 1 day: got open duration %v, want %v", openDuration, expectedOpen)
	}

	if unknownDuration != expectedUnknown {
		t.Errorf("GetOpenDuration with '22:00-02:00' over 1 day: got unknown duration %v, want %v", unknownDuration, expectedUnknown)
	}
}

func TestGetOpenDuration_PartiallyOpenAndClosed(t *testing.T) {
	// Test: Query range that spans both open and closed periods
	oh, err := New("09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// From 8:00 to 18:00 - includes 1 hour closed before, 8 hours open, 1 hour closed after
	from := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 15, 18, 0, 0, 0, time.UTC)

	openDuration, unknownDuration := oh.GetOpenDuration(from, to)

	expectedOpen := 8 * time.Hour
	expectedUnknown := time.Duration(0)

	if openDuration != expectedOpen {
		t.Errorf("GetOpenDuration with '09:00-17:00' from 08:00 to 18:00: got open duration %v, want %v", openDuration, expectedOpen)
	}

	if unknownDuration != expectedUnknown {
		t.Errorf("GetOpenDuration with '09:00-17:00' from 08:00 to 18:00: got unknown duration %v, want %v", unknownDuration, expectedUnknown)
	}
}

func TestGetOpenDuration_WeekendDays(t *testing.T) {
	// Test: Weekday constraint over weekend - should return 0
	oh, err := New("Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Jan 20, 2024 is Saturday
	// Jan 21, 2024 is Sunday
	from := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 22, 0, 0, 0, 0, time.UTC)

	openDuration, unknownDuration := oh.GetOpenDuration(from, to)

	expectedOpen := time.Duration(0)
	expectedUnknown := time.Duration(0)

	if openDuration != expectedOpen {
		t.Errorf("GetOpenDuration with 'Mo-Fr 09:00-17:00' over weekend: got open duration %v, want %v", openDuration, expectedOpen)
	}

	if unknownDuration != expectedUnknown {
		t.Errorf("GetOpenDuration with 'Mo-Fr 09:00-17:00' over weekend: got unknown duration %v, want %v", unknownDuration, expectedUnknown)
	}
}

func TestGetOpenDuration_MixedOpenAndUnknown(t *testing.T) {
	// Test: Multiple rules with both open and unknown states
	oh, err := New("08:00-12:00; 14:00-18:00 unknown")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)

	openDuration, unknownDuration := oh.GetOpenDuration(from, to)

	expectedOpen := 4 * time.Hour    // 08:00-12:00
	expectedUnknown := 4 * time.Hour // 14:00-18:00

	if openDuration != expectedOpen {
		t.Errorf("GetOpenDuration with mixed open/unknown: got open duration %v, want %v", openDuration, expectedOpen)
	}

	if unknownDuration != expectedUnknown {
		t.Errorf("GetOpenDuration with mixed open/unknown: got unknown duration %v, want %v", unknownDuration, expectedUnknown)
	}
}

func TestGetOpenDuration_ZeroDuration(t *testing.T) {
	// Test: from and to are the same time - should return 0 for both
	oh, err := New("09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	from := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	openDuration, unknownDuration := oh.GetOpenDuration(from, to)

	expectedOpen := time.Duration(0)
	expectedUnknown := time.Duration(0)

	if openDuration != expectedOpen {
		t.Errorf("GetOpenDuration with zero duration: got open duration %v, want %v", openDuration, expectedOpen)
	}

	if unknownDuration != expectedUnknown {
		t.Errorf("GetOpenDuration with zero duration: got unknown duration %v, want %v", unknownDuration, expectedUnknown)
	}
}

func TestGetOpenDuration_ComplexWeekdayRule(t *testing.T) {
	// Test: "Mo-Fr 08:00-18:00; Sa 10:00-14:00"
	oh, err := New("Mo-Fr 08:00-18:00; Sa 10:00-14:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Jan 15, 2024 is Monday
	// Jan 22, 2024 is the following Monday
	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 22, 0, 0, 0, 0, time.UTC)

	openDuration, unknownDuration := oh.GetOpenDuration(from, to)

	// 5 weekdays * 10 hours + 1 Saturday * 4 hours = 54 hours
	expectedOpen := 54 * time.Hour
	expectedUnknown := time.Duration(0)

	if openDuration != expectedOpen {
		t.Errorf("GetOpenDuration with complex weekday rule: got open duration %v, want %v", openDuration, expectedOpen)
	}

	if unknownDuration != expectedUnknown {
		t.Errorf("GetOpenDuration with complex weekday rule: got unknown duration %v, want %v", unknownDuration, expectedUnknown)
	}
}

func TestGetOpenDuration_WithOffRule(t *testing.T) {
	// Test: "08:00-18:00; 12:00-14:00 off" - lunch break
	oh, err := New("08:00-18:00; 12:00-14:00 off")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)

	openDuration, unknownDuration := oh.GetOpenDuration(from, to)

	// 10 hours total - 2 hours off = 8 hours open
	expectedOpen := 8 * time.Hour
	expectedUnknown := time.Duration(0)

	if openDuration != expectedOpen {
		t.Errorf("GetOpenDuration with off rule: got open duration %v, want %v", openDuration, expectedOpen)
	}

	if unknownDuration != expectedUnknown {
		t.Errorf("GetOpenDuration with off rule: got unknown duration %v, want %v", unknownDuration, expectedUnknown)
	}
}
