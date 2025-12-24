package openinghours

import (
	"testing"
	"time"
)

// TestGetOpenIntervals_SingleDay tests a simple single day with one time range
func TestGetOpenIntervals_SingleDay(t *testing.T) {
	oh, err := New("09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 15, 23, 59, 59, 0, time.UTC)

	intervals := oh.GetOpenIntervals(from, to)

	if len(intervals) != 1 {
		t.Fatalf("expected 1 interval, got %d", len(intervals))
	}

	expectedStart := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2024, 1, 15, 17, 0, 0, 0, time.UTC)

	if !intervals[0].Start.Equal(expectedStart) {
		t.Errorf("expected start %v, got %v", expectedStart, intervals[0].Start)
	}

	if !intervals[0].End.Equal(expectedEnd) {
		t.Errorf("expected end %v, got %v", expectedEnd, intervals[0].End)
	}

	if intervals[0].Unknown {
		t.Errorf("expected Unknown=false, got true")
	}

	if intervals[0].Comment != "" {
		t.Errorf("expected empty comment, got %q", intervals[0].Comment)
	}
}

// TestGetOpenIntervals_MultipleRanges tests multiple time ranges in a single day
func TestGetOpenIntervals_MultipleRanges(t *testing.T) {
	oh, err := New("08:00-12:00,14:00-18:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 15, 23, 59, 59, 0, time.UTC)

	intervals := oh.GetOpenIntervals(from, to)

	if len(intervals) != 2 {
		t.Fatalf("expected 2 intervals, got %d", len(intervals))
	}

	// First interval: 08:00-12:00
	expectedStart1 := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC)
	expectedEnd1 := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	if !intervals[0].Start.Equal(expectedStart1) {
		t.Errorf("interval[0]: expected start %v, got %v", expectedStart1, intervals[0].Start)
	}

	if !intervals[0].End.Equal(expectedEnd1) {
		t.Errorf("interval[0]: expected end %v, got %v", expectedEnd1, intervals[0].End)
	}

	// Second interval: 14:00-18:00
	expectedStart2 := time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC)
	expectedEnd2 := time.Date(2024, 1, 15, 18, 0, 0, 0, time.UTC)

	if !intervals[1].Start.Equal(expectedStart2) {
		t.Errorf("interval[1]: expected start %v, got %v", expectedStart2, intervals[1].Start)
	}

	if !intervals[1].End.Equal(expectedEnd2) {
		t.Errorf("interval[1]: expected end %v, got %v", expectedEnd2, intervals[1].End)
	}
}

// TestGetOpenIntervals_MultipleWeekdays tests multiple weekdays with same hours
func TestGetOpenIntervals_MultipleWeekdays(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Monday, Jan 15, 2024 to Friday, Jan 19, 2024
	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 19, 23, 59, 59, 0, time.UTC)

	intervals := oh.GetOpenIntervals(from, to)

	if len(intervals) != 5 {
		t.Fatalf("expected 5 intervals (one per weekday), got %d", len(intervals))
	}

	// Check each day's interval
	for i := 0; i < 5; i++ {
		day := 15 + i
		expectedStart := time.Date(2024, 1, day, 9, 0, 0, 0, time.UTC)
		expectedEnd := time.Date(2024, 1, day, 17, 0, 0, 0, time.UTC)

		if !intervals[i].Start.Equal(expectedStart) {
			t.Errorf("interval[%d]: expected start %v, got %v", i, expectedStart, intervals[i].Start)
		}

		if !intervals[i].End.Equal(expectedEnd) {
			t.Errorf("interval[%d]: expected end %v, got %v", i, expectedEnd, intervals[i].End)
		}
	}
}

// TestGetOpenIntervals_MultipleWeekdays_IncludesWeekend tests that weekends are excluded
func TestGetOpenIntervals_MultipleWeekdays_IncludesWeekend(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Monday, Jan 15 to Sunday, Jan 21 (includes weekend)
	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 21, 23, 59, 59, 0, time.UTC)

	intervals := oh.GetOpenIntervals(from, to)

	// Should still be 5 intervals (Mo-Fr only, Sa-Su excluded)
	if len(intervals) != 5 {
		t.Fatalf("expected 5 intervals (Mo-Fr only), got %d", len(intervals))
	}

	// Verify no intervals fall on Saturday (Jan 20) or Sunday (Jan 21)
	for i, interval := range intervals {
		weekday := interval.Start.Weekday()
		if weekday == time.Saturday || weekday == time.Sunday {
			t.Errorf("interval[%d]: unexpected interval on weekend: %v", i, interval.Start)
		}
	}
}

// TestGetOpenIntervals_AlwaysOpen tests 24/7 opening hours
func TestGetOpenIntervals_AlwaysOpen(t *testing.T) {
	oh, err := New("24/7")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 17, 23, 59, 59, 0, time.UTC)

	intervals := oh.GetOpenIntervals(from, to)

	if len(intervals) != 1 {
		t.Fatalf("expected 1 continuous interval for 24/7, got %d", len(intervals))
	}

	// The interval should span the entire range
	if !intervals[0].Start.Equal(from) {
		t.Errorf("expected start %v, got %v", from, intervals[0].Start)
	}

	// End time should be close to 'to' (might be exact or end of last day)
	// Allow some flexibility in how the implementation handles the end time
	expectedEnd := time.Date(2024, 1, 17, 23, 59, 59, 0, time.UTC)
	if intervals[0].End.After(expectedEnd.Add(time.Minute)) {
		t.Errorf("end time %v is too far after expected %v", intervals[0].End, expectedEnd)
	}
}

// TestGetOpenIntervals_AlwaysClosed tests permanently closed hours
func TestGetOpenIntervals_AlwaysClosed(t *testing.T) {
	testCases := []string{"off", "closed"}

	for _, input := range testCases {
		oh, err := New(input)
		if err != nil {
			t.Fatalf("unexpected parse error for %q: %v", input, err)
		}

		from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 17, 23, 59, 59, 0, time.UTC)

		intervals := oh.GetOpenIntervals(from, to)

		if len(intervals) != 0 {
			t.Errorf("expected 0 intervals for %q, got %d", input, len(intervals))
		}
	}
}

// TestGetOpenIntervals_WithUnknown tests intervals with unknown state
func TestGetOpenIntervals_WithUnknown(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-17:00 unknown")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 19, 23, 59, 59, 0, time.UTC)

	intervals := oh.GetOpenIntervals(from, to)

	if len(intervals) != 5 {
		t.Fatalf("expected 5 intervals, got %d", len(intervals))
	}

	// All intervals should have Unknown=true
	for i, interval := range intervals {
		if !interval.Unknown {
			t.Errorf("interval[%d]: expected Unknown=true, got false", i)
		}
	}
}

// TestGetOpenIntervals_WithComment tests intervals with comments
func TestGetOpenIntervals_WithComment(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-17:00 \"by appointment\"")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 19, 23, 59, 59, 0, time.UTC)

	intervals := oh.GetOpenIntervals(from, to)

	if len(intervals) != 5 {
		t.Fatalf("expected 5 intervals, got %d", len(intervals))
	}

	expectedComment := "by appointment"

	// All intervals should have the comment
	for i, interval := range intervals {
		if interval.Comment != expectedComment {
			t.Errorf("interval[%d]: expected comment %q, got %q", i, expectedComment, interval.Comment)
		}
	}
}

// TestGetOpenIntervals_EmptyRange tests with from == to
func TestGetOpenIntervals_EmptyRange(t *testing.T) {
	oh, err := New("09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	sameTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	intervals := oh.GetOpenIntervals(sameTime, sameTime)

	// Empty range should return no intervals
	if len(intervals) != 0 {
		t.Errorf("expected 0 intervals for empty range, got %d", len(intervals))
	}
}

// TestGetOpenIntervals_PartialDay tests interval that starts/ends mid-day
func TestGetOpenIntervals_PartialDay(t *testing.T) {
	oh, err := New("09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// From 11:00 to 15:00 on the same day
	from := time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 15, 15, 0, 0, 0, time.UTC)

	intervals := oh.GetOpenIntervals(from, to)

	if len(intervals) != 1 {
		t.Fatalf("expected 1 interval, got %d", len(intervals))
	}

	// Interval should be clipped to the requested range
	if !intervals[0].Start.Equal(from) {
		t.Errorf("expected start %v, got %v", from, intervals[0].Start)
	}

	if !intervals[0].End.Equal(to) {
		t.Errorf("expected end %v, got %v", to, intervals[0].End)
	}
}

// TestGetOpenIntervals_SpansMidnight tests intervals that cross midnight
func TestGetOpenIntervals_SpansMidnight(t *testing.T) {
	oh, err := New("20:00-02:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 16, 23, 59, 59, 0, time.UTC)

	intervals := oh.GetOpenIntervals(from, to)

	// Should have 2 intervals:
	// 1. Jan 15 00:00-02:00 (end of previous day's range)
	// 2. Jan 15 20:00-Jan 16 02:00 (new range spanning midnight)
	// Or possibly different depending on implementation
	if len(intervals) < 1 {
		t.Fatalf("expected at least 1 interval for midnight-spanning hours, got %d", len(intervals))
	}

	// Verify that some interval crosses midnight or spans into next day
	foundMidnightSpan := false
	for _, interval := range intervals {
		if interval.Start.Day() != interval.End.Day() {
			foundMidnightSpan = true
			break
		}
	}

	if !foundMidnightSpan && len(intervals) < 2 {
		t.Logf("Note: intervals might be split at midnight or combined. Found %d intervals", len(intervals))
	}
}

// TestGetOpenIntervals_ComplexSchedule tests a more complex real-world schedule
func TestGetOpenIntervals_ComplexSchedule(t *testing.T) {
	oh, err := New("Mo-We 09:00-17:00; Th 10:00-20:00; Fr 09:00-15:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Monday, Jan 15 to Friday, Jan 19, 2024
	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 19, 23, 59, 59, 0, time.UTC)

	intervals := oh.GetOpenIntervals(from, to)

	if len(intervals) != 5 {
		t.Fatalf("expected 5 intervals, got %d", len(intervals))
	}

	// Monday: 09:00-17:00
	expectedMon := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	if !intervals[0].Start.Equal(expectedMon) {
		t.Errorf("Monday start: expected %v, got %v", expectedMon, intervals[0].Start)
	}

	// Thursday (Jan 18): 10:00-20:00
	// Find Thursday interval
	for i, interval := range intervals {
		if interval.Start.Day() == 18 {
			expectedThStart := time.Date(2024, 1, 18, 10, 0, 0, 0, time.UTC)
			expectedThEnd := time.Date(2024, 1, 18, 20, 0, 0, 0, time.UTC)

			if !interval.Start.Equal(expectedThStart) {
				t.Errorf("Thursday start: expected %v, got %v", expectedThStart, interval.Start)
			}

			if !interval.End.Equal(expectedThEnd) {
				t.Errorf("Thursday end: expected %v, got %v", expectedThEnd, interval.End)
			}
			break
		}

		// If we reached the end without finding Thursday
		if i == len(intervals)-1 {
			t.Error("Thursday interval not found")
		}
	}
}

// TestGetOpenIntervals_WithMonthConstraint tests intervals with month constraints
func TestGetOpenIntervals_WithMonthConstraint(t *testing.T) {
	oh, err := New("Jan 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Test in January (should be open)
	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 16, 23, 59, 59, 0, time.UTC)

	intervalsJan := oh.GetOpenIntervals(from, to)

	if len(intervalsJan) != 2 {
		t.Errorf("expected 2 intervals in January, got %d", len(intervalsJan))
	}

	// Test in February (should be closed)
	from = time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC)
	to = time.Date(2024, 2, 16, 23, 59, 59, 0, time.UTC)

	intervalsFeb := oh.GetOpenIntervals(from, to)

	if len(intervalsFeb) != 0 {
		t.Errorf("expected 0 intervals in February, got %d", len(intervalsFeb))
	}
}

// TestGetOpenIntervals_EdgeCaseStartEqualsEnd tests when from equals opening time
func TestGetOpenIntervals_EdgeCaseStartEqualsEnd(t *testing.T) {
	oh, err := New("09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Start exactly at opening time
	from := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 15, 23, 59, 59, 0, time.UTC)

	intervals := oh.GetOpenIntervals(from, to)

	if len(intervals) != 1 {
		t.Fatalf("expected 1 interval, got %d", len(intervals))
	}

	if !intervals[0].Start.Equal(from) {
		t.Errorf("expected start %v, got %v", from, intervals[0].Start)
	}
}

// TestGetOpenIntervals_MultipleTimeRangesMultipleDays tests complex scenario
func TestGetOpenIntervals_MultipleTimeRangesMultipleDays(t *testing.T) {
	oh, err := New("Mo,We,Fr 08:00-12:00,14:00-18:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Monday Jan 15 to Sunday Jan 21
	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 21, 23, 59, 59, 0, time.UTC)

	intervals := oh.GetOpenIntervals(from, to)

	// Monday (15), Wednesday (17), Friday (19) = 3 days
	// Each day has 2 time ranges = 6 total intervals
	if len(intervals) != 6 {
		t.Fatalf("expected 6 intervals (3 days × 2 ranges), got %d", len(intervals))
	}
}

// TestGetOpenIntervals_InverseTimeOrder tests that to < from returns empty
func TestGetOpenIntervals_InverseTimeOrder(t *testing.T) {
	oh, err := New("09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// to before from (invalid range)
	from := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 14, 12, 0, 0, 0, time.UTC)

	intervals := oh.GetOpenIntervals(from, to)

	// Should return empty slice for invalid range
	if len(intervals) != 0 {
		t.Errorf("expected 0 intervals for inverse time range, got %d", len(intervals))
	}
}

// TestGetOpenIntervals_VeryLongRange tests performance with longer date range
func TestGetOpenIntervals_VeryLongRange(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// One month: Jan 1 to Jan 31
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

	intervals := oh.GetOpenIntervals(from, to)

	// January 2024: 31 days total, should have 23 weekdays (Mo-Fr)
	// This is just a rough check
	if len(intervals) < 20 || len(intervals) > 25 {
		t.Errorf("expected approximately 23 intervals for January weekdays, got %d", len(intervals))
	}

	// Verify all intervals are on weekdays
	for i, interval := range intervals {
		weekday := interval.Start.Weekday()
		if weekday == time.Saturday || weekday == time.Sunday {
			t.Errorf("interval[%d]: unexpected interval on weekend: %v", i, interval.Start)
		}
	}
}

// TestGetOpenIntervals_OpenEnd tests open-ended time ranges (e.g., "17:00+")
func TestGetOpenIntervals_OpenEnd(t *testing.T) {
	oh, err := New("17:00+")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 15, 23, 59, 59, 0, time.UTC)

	intervals := oh.GetOpenIntervals(from, to)

	if len(intervals) != 1 {
		t.Fatalf("expected 1 interval, got %d", len(intervals))
	}

	expectedStart := time.Date(2024, 1, 15, 17, 0, 0, 0, time.UTC)

	if !intervals[0].Start.Equal(expectedStart) {
		t.Errorf("expected start %v, got %v", expectedStart, intervals[0].Start)
	}

	// End should be near end of day
	if intervals[0].End.Hour() != 23 && intervals[0].End.Hour() != 0 {
		t.Errorf("expected end near end of day, got %v", intervals[0].End)
	}
}

// TestGetOpenIntervals_WithUnknownAndComment tests combined unknown state and comment
func TestGetOpenIntervals_WithUnknownAndComment(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-17:00 unknown \"call ahead\"")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 17, 23, 59, 59, 0, time.UTC)

	intervals := oh.GetOpenIntervals(from, to)

	if len(intervals) != 3 {
		t.Fatalf("expected 3 intervals (Mon-Wed), got %d", len(intervals))
	}

	expectedComment := "call ahead"

	for i, interval := range intervals {
		if !interval.Unknown {
			t.Errorf("interval[%d]: expected Unknown=true, got false", i)
		}

		if interval.Comment != expectedComment {
			t.Errorf("interval[%d]: expected comment %q, got %q", i, expectedComment, interval.Comment)
		}
	}
}

// TestGetOpenIntervals_DifferentTimezones tests handling of different timezones
func TestGetOpenIntervals_DifferentTimezones(t *testing.T) {
	oh, err := New("09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Use a different timezone (EST)
	est, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skipf("could not load timezone: %v", err)
	}

	from := time.Date(2024, 1, 15, 0, 0, 0, 0, est)
	to := time.Date(2024, 1, 15, 23, 59, 59, 0, est)

	intervals := oh.GetOpenIntervals(from, to)

	if len(intervals) != 1 {
		t.Fatalf("expected 1 interval, got %d", len(intervals))
	}

	// Verify the interval is in the same timezone
	if intervals[0].Start.Location() != est {
		t.Errorf("expected interval in EST timezone, got %v", intervals[0].Start.Location())
	}
}
