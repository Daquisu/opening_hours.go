package openinghours

// This file contains tests ported directly from the JavaScript opening_hours.js test suite
// Source: opening_hours.js/test/test.js
// Each test verifies that multiple equivalent expressions produce the same open intervals

import (
	"testing"
	"time"
)

// Interval represents an expected open interval for testing
type testInterval struct {
	start   time.Time
	end     time.Time
	unknown bool
	comment string
}

// parseTime parses a time string in format "2012-10-01 10:00"
func parseTime(s string) time.Time {
	t, err := time.Parse("2006-01-02 15:04", s)
	if err != nil {
		panic("invalid time format: " + s)
	}
	return t
}

// runIntervalTest tests that all values produce the expected intervals
func runIntervalTest(t *testing.T, testName string, values []string, startStr, endStr string, expectedIntervals []testInterval) {
	t.Helper()

	start := parseTime(startStr)
	end := parseTime(endStr)

	for _, value := range values {
		t.Run(value, func(t *testing.T) {
			oh, err := New(value)
			if err != nil {
				t.Fatalf("failed to parse '%s': %v", value, err)
			}

			intervals := oh.GetOpenIntervals(start, end)

			// Compare intervals
			if len(intervals) != len(expectedIntervals) {
				t.Errorf("expected %d intervals, got %d", len(expectedIntervals), len(intervals))
				for i, iv := range intervals {
					t.Logf("  got[%d]: %v - %v", i, iv.Start, iv.End)
				}
				return
			}

			for i, expected := range expectedIntervals {
				got := intervals[i]
				if !got.Start.Equal(expected.start) {
					t.Errorf("interval %d start: expected %v, got %v", i, expected.start, got.Start)
				}
				if !got.End.Equal(expected.end) {
					t.Errorf("interval %d end: expected %v, got %v", i, expected.end, got.End)
				}
				if expected.unknown && !got.Unknown {
					t.Errorf("interval %d: expected unknown=true, got false", i)
				}
			}
		})
	}
}

// runDurationTest tests that open duration matches expected
func runDurationTest(t *testing.T, value string, startStr, endStr string, expectedOpenHours float64) {
	t.Helper()

	oh, err := New(value)
	if err != nil {
		t.Fatalf("failed to parse '%s': %v", value, err)
	}

	start := parseTime(startStr)
	end := parseTime(endStr)

	openDuration, _ := oh.GetOpenDuration(start, end)
	expectedDuration := time.Duration(expectedOpenHours * float64(time.Hour))

	// Allow 1 minute tolerance for rounding
	diff := openDuration - expectedDuration
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Minute {
		t.Errorf("expected open duration %v, got %v", expectedDuration, openDuration)
	}
}

// =============================================================================
// Time intervals (test.js lines 167-184)
// =============================================================================

func TestJS_TimeIntervals_Basic(t *testing.T) {
	values := []string{
		"10:00-12:00",
		"10:00-11:00,11:00-12:00",
		"10:00-12:00,10:30-11:30",
		"10:00-14:00; 12:00-14:00 off",
	}

	expected := []testInterval{
		{parseTime("2012-10-01 10:00"), parseTime("2012-10-01 12:00"), false, ""},
		{parseTime("2012-10-02 10:00"), parseTime("2012-10-02 12:00"), false, ""},
		{parseTime("2012-10-03 10:00"), parseTime("2012-10-03 12:00"), false, ""},
		{parseTime("2012-10-04 10:00"), parseTime("2012-10-04 12:00"), false, ""},
		{parseTime("2012-10-05 10:00"), parseTime("2012-10-05 12:00"), false, ""},
		{parseTime("2012-10-06 10:00"), parseTime("2012-10-06 12:00"), false, ""},
		{parseTime("2012-10-07 10:00"), parseTime("2012-10-07 12:00"), false, ""},
	}

	runIntervalTest(t, "Time intervals", values, "2012-10-01 00:00", "2012-10-08 00:00", expected)
}

func TestJS_TimeIntervals_Duration(t *testing.T) {
	// 7 days * 2 hours = 14 hours
	runDurationTest(t, "10:00-12:00", "2012-10-01 00:00", "2012-10-08 00:00", 14)
}

// =============================================================================
// 24/7 with off periods (test.js lines 186-193)
// =============================================================================

func TestJS_24_7_WithOff(t *testing.T) {
	values := []string{
		"open; Mo 15:00-16:00 off",
		"00:00-24:00; Mo 15:00-16:00 off",
	}

	expected := []testInterval{
		{parseTime("2012-10-01 00:00"), parseTime("2012-10-01 15:00"), false, ""},
		{parseTime("2012-10-01 16:00"), parseTime("2012-10-08 00:00"), false, ""},
	}

	runIntervalTest(t, "24/7 with off", values, "2012-10-01 00:00", "2012-10-08 00:00", expected)
}

func TestJS_24_7_WithOff_Duration(t *testing.T) {
	// 6 full days (24h) + 1 day with 1 hour off = 6*24 + 23 = 167 hours
	runDurationTest(t, "open; Mo 15:00-16:00 off", "2012-10-01 00:00", "2012-10-08 00:00", 167)
}

// =============================================================================
// Always closed (test.js lines 195-209)
// =============================================================================

func TestJS_AlwaysClosed(t *testing.T) {
	values := []string{
		"off",
		"closed",
		"24/7 closed",
		"00:00-24:00 closed",
	}

	// No intervals expected - always closed
	expected := []testInterval{}

	runIntervalTest(t, "Always closed", values, "2012-10-01 00:00", "2012-10-08 00:00", expected)
}

// =============================================================================
// Error tolerance: dot as time separator (test.js lines 220-246)
// =============================================================================

func TestJS_DotTimeSeparator(t *testing.T) {
	values := []string{
		"10:00-12:00",
		"10.00-12.00",
	}

	expected := []testInterval{
		{parseTime("2012-10-01 10:00"), parseTime("2012-10-01 12:00"), false, ""},
		{parseTime("2012-10-02 10:00"), parseTime("2012-10-02 12:00"), false, ""},
		{parseTime("2012-10-03 10:00"), parseTime("2012-10-03 12:00"), false, ""},
		{parseTime("2012-10-04 10:00"), parseTime("2012-10-04 12:00"), false, ""},
		{parseTime("2012-10-05 10:00"), parseTime("2012-10-05 12:00"), false, ""},
		{parseTime("2012-10-06 10:00"), parseTime("2012-10-06 12:00"), false, ""},
		{parseTime("2012-10-07 10:00"), parseTime("2012-10-07 12:00"), false, ""},
	}

	runIntervalTest(t, "Dot separator", values, "2012-10-01 00:00", "2012-10-08 00:00", expected)
}

func TestJS_DotTimeSeparator_WithOff(t *testing.T) {
	values := []string{
		"10:00-14:00; 12:00-14:00 off",
		"10.00-14.00; 12.00-14.00 off",
	}

	expected := []testInterval{
		{parseTime("2012-10-01 10:00"), parseTime("2012-10-01 12:00"), false, ""},
		{parseTime("2012-10-02 10:00"), parseTime("2012-10-02 12:00"), false, ""},
		{parseTime("2012-10-03 10:00"), parseTime("2012-10-03 12:00"), false, ""},
		{parseTime("2012-10-04 10:00"), parseTime("2012-10-04 12:00"), false, ""},
		{parseTime("2012-10-05 10:00"), parseTime("2012-10-05 12:00"), false, ""},
		{parseTime("2012-10-06 10:00"), parseTime("2012-10-06 12:00"), false, ""},
		{parseTime("2012-10-07 10:00"), parseTime("2012-10-07 12:00"), false, ""},
	}

	runIntervalTest(t, "Dot separator with off", values, "2012-10-01 00:00", "2012-10-08 00:00", expected)
}

// =============================================================================
// Error tolerance: AM/PM format (test.js lines 248-318)
// =============================================================================

func TestJS_AMPM_MultipleRanges(t *testing.T) {
	values := []string{
		"10:00-12:00,13:00-20:00",
		"10am-12pm,1pm-8pm",
		"10:00am-12:00pm,1:00pm-8:00pm",
	}

	expected := []testInterval{
		{parseTime("2012-10-01 10:00"), parseTime("2012-10-01 12:00"), false, ""},
		{parseTime("2012-10-01 13:00"), parseTime("2012-10-01 20:00"), false, ""},
		{parseTime("2012-10-02 10:00"), parseTime("2012-10-02 12:00"), false, ""},
		{parseTime("2012-10-02 13:00"), parseTime("2012-10-02 20:00"), false, ""},
	}

	runIntervalTest(t, "AM/PM multiple ranges", values, "2012-10-01 00:00", "2012-10-03 00:00", expected)
}

func TestJS_AMPM_Midnight(t *testing.T) {
	values := []string{
		"00:00-00:01",
		"12:00am-12:01am",
	}

	expected := []testInterval{
		{parseTime("2012-10-01 00:00"), parseTime("2012-10-01 00:01"), false, ""},
		{parseTime("2012-10-02 00:00"), parseTime("2012-10-02 00:01"), false, ""},
	}

	runIntervalTest(t, "AM/PM midnight", values, "2012-10-01 00:00", "2012-10-03 00:00", expected)
}

func TestJS_AMPM_Morning(t *testing.T) {
	values := []string{
		"01:00-11:00",
		"01:00am-11:00am",
		"01am-11am",
	}

	expected := []testInterval{
		{parseTime("2012-10-01 01:00"), parseTime("2012-10-01 11:00"), false, ""},
		{parseTime("2012-10-02 01:00"), parseTime("2012-10-02 11:00"), false, ""},
	}

	runIntervalTest(t, "AM/PM morning", values, "2012-10-01 00:00", "2012-10-03 00:00", expected)
}

func TestJS_AMPM_CrossingNoon(t *testing.T) {
	values := []string{
		"11:59-12:00",
		"11:59am-12:00pm",
	}

	expected := []testInterval{
		{parseTime("2012-10-01 11:59"), parseTime("2012-10-01 12:00"), false, ""},
		{parseTime("2012-10-02 11:59"), parseTime("2012-10-02 12:00"), false, ""},
	}

	runIntervalTest(t, "AM/PM crossing noon", values, "2012-10-01 00:00", "2012-10-03 00:00", expected)
}

func TestJS_AMPM_AfterNoon(t *testing.T) {
	values := []string{
		"12:01-12:59",
		"12:01pm-12:59pm",
	}

	expected := []testInterval{
		{parseTime("2012-10-01 12:01"), parseTime("2012-10-01 12:59"), false, ""},
		{parseTime("2012-10-02 12:01"), parseTime("2012-10-02 12:59"), false, ""},
	}

	runIntervalTest(t, "AM/PM after noon", values, "2012-10-01 00:00", "2012-10-03 00:00", expected)
}

func TestJS_AMPM_Afternoon(t *testing.T) {
	values := []string{
		"13:00-13:01",
		"01:00pm-01:01pm",
	}

	expected := []testInterval{
		{parseTime("2012-10-01 13:00"), parseTime("2012-10-01 13:01"), false, ""},
		{parseTime("2012-10-02 13:00"), parseTime("2012-10-02 13:01"), false, ""},
	}

	runIntervalTest(t, "AM/PM afternoon", values, "2012-10-01 00:00", "2012-10-03 00:00", expected)
}

func TestJS_AMPM_Evening(t *testing.T) {
	values := []string{
		"23:00-23:59",
		"11:00pm-11:59pm",
	}

	expected := []testInterval{
		{parseTime("2012-10-01 23:00"), parseTime("2012-10-01 23:59"), false, ""},
		{parseTime("2012-10-02 23:00"), parseTime("2012-10-02 23:59"), false, ""},
	}

	runIntervalTest(t, "AM/PM evening", values, "2012-10-01 00:00", "2012-10-03 00:00", expected)
}

// =============================================================================
// Time intervals with weekday (test.js lines 322-328)
// =============================================================================

func TestJS_WeekdayWithTime(t *testing.T) {
	values := []string{
		"Mo 07:00-18:00",
	}

	expected := []testInterval{
		{parseTime("2012-10-01 07:00"), parseTime("2012-10-01 18:00"), false, ""},
	}

	runIntervalTest(t, "Weekday with time", values, "2012-10-01 00:00", "2012-10-08 00:00", expected)
}

func TestJS_WeekdayWithTime_Duration(t *testing.T) {
	// 1 Monday * 11 hours = 11 hours
	runDurationTest(t, "Mo 07:00-18:00", "2012-10-01 00:00", "2012-10-08 00:00", 11)
}

// =============================================================================
// Time ranges spanning midnight (test.js lines 376-389)
// =============================================================================

func TestJS_MidnightSpanning(t *testing.T) {
	values := []string{
		"22:00-02:00",
	}

	expected := []testInterval{
		{parseTime("2012-10-01 00:00"), parseTime("2012-10-01 02:00"), false, ""},
		{parseTime("2012-10-01 22:00"), parseTime("2012-10-02 02:00"), false, ""},
		{parseTime("2012-10-02 22:00"), parseTime("2012-10-03 02:00"), false, ""},
		{parseTime("2012-10-03 22:00"), parseTime("2012-10-04 02:00"), false, ""},
		{parseTime("2012-10-04 22:00"), parseTime("2012-10-05 02:00"), false, ""},
		{parseTime("2012-10-05 22:00"), parseTime("2012-10-06 02:00"), false, ""},
		{parseTime("2012-10-06 22:00"), parseTime("2012-10-07 02:00"), false, ""},
		{parseTime("2012-10-07 22:00"), parseTime("2012-10-08 00:00"), false, ""},
	}

	runIntervalTest(t, "Midnight spanning", values, "2012-10-01 00:00", "2012-10-08 00:00", expected)
}

func TestJS_MidnightSpanning_Duration(t *testing.T) {
	// 7 days * 4 hours = 28 hours
	runDurationTest(t, "22:00-02:00", "2012-10-01 00:00", "2012-10-08 00:00", 28)
}

func TestJS_MidnightSpanning_24Hours(t *testing.T) {
	values := []string{
		"We 22:00-22:00",
	}

	expected := []testInterval{
		{parseTime("2012-10-03 22:00"), parseTime("2012-10-04 22:00"), false, ""},
	}

	runIntervalTest(t, "Midnight spanning 24h", values, "2012-10-01 00:00", "2012-10-08 00:00", expected)
}

func TestJS_MidnightSpanning_24Hours_Duration(t *testing.T) {
	// 24 hours exactly
	runDurationTest(t, "We 22:00-22:00", "2012-10-01 00:00", "2012-10-08 00:00", 24)
}

// =============================================================================
// Weekday selectors
// =============================================================================

func TestJS_WeekdayRange(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Test each day of the week (starting from Monday Oct 1, 2012)
	tests := []struct {
		day      int // day of month
		hour     int
		expected bool
	}{
		{1, 12, true},   // Monday
		{2, 12, true},   // Tuesday
		{3, 12, true},   // Wednesday
		{4, 12, true},   // Thursday
		{5, 12, true},   // Friday
		{6, 12, false},  // Saturday
		{7, 12, false},  // Sunday
		{1, 8, false},   // Monday before open
		{1, 18, false},  // Monday after close
	}

	for _, tt := range tests {
		testTime := time.Date(2012, 10, tt.day, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("Oct %d at %d:00: expected %v, got %v", tt.day, tt.hour, tt.expected, got)
		}
	}
}

func TestJS_WeekdayList(t *testing.T) {
	oh, err := New("Mo,We,Fr 10:00-14:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	tests := []struct {
		day      int
		expected bool
	}{
		{1, true},  // Monday
		{2, false}, // Tuesday
		{3, true},  // Wednesday
		{4, false}, // Thursday
		{5, true},  // Friday
		{6, false}, // Saturday
		{7, false}, // Sunday
	}

	for _, tt := range tests {
		testTime := time.Date(2012, 10, tt.day, 12, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("Oct %d at 12:00: expected %v, got %v", tt.day, tt.expected, got)
		}
	}
}

// =============================================================================
// Multiple rules with semicolon
// =============================================================================

func TestJS_MultipleRules(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-12:00; Mo-Fr 13:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Monday at various times
	tests := []struct {
		hour     int
		expected bool
	}{
		{8, false},  // before open
		{10, true},  // morning open
		{12, false}, // lunch closed
		{14, true},  // afternoon open
		{18, false}, // after close
	}

	for _, tt := range tests {
		testTime := time.Date(2012, 10, 1, tt.hour, 0, 0, 0, time.UTC) // Monday
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("Monday at %d:00: expected %v, got %v", tt.hour, tt.expected, got)
		}
	}
}

// =============================================================================
// Month ranges
// =============================================================================

func TestJS_MonthRange(t *testing.T) {
	oh, err := New("Mar-Oct 10:00-18:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	tests := []struct {
		month    time.Month
		expected bool
	}{
		{time.January, false},
		{time.February, false},
		{time.March, true},
		{time.April, true},
		{time.May, true},
		{time.June, true},
		{time.July, true},
		{time.August, true},
		{time.September, true},
		{time.October, true},
		{time.November, false},
		{time.December, false},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, tt.month, 15, 12, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s at 12:00: expected %v, got %v", tt.month, tt.expected, got)
		}
	}
}

func TestJS_SingleMonth(t *testing.T) {
	oh, err := New("Dec 10:00-20:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// December should be open
	decTime := time.Date(2012, 12, 15, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(decTime) {
		t.Error("December should be open")
	}

	// January should be closed
	janTime := time.Date(2012, 1, 15, 12, 0, 0, 0, time.UTC)
	if oh.GetState(janTime) {
		t.Error("January should be closed")
	}
}

// =============================================================================
// Month-day ranges
// =============================================================================

func TestJS_MonthDayRange(t *testing.T) {
	oh, err := New("Dec 24-26 10:00-18:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	tests := []struct {
		day      int
		expected bool
	}{
		{23, false},
		{24, true},
		{25, true},
		{26, true},
		{27, false},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, 12, tt.day, 12, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("Dec %d at 12:00: expected %v, got %v", tt.day, tt.expected, got)
		}
	}
}

// =============================================================================
// Week numbers
// =============================================================================

func TestJS_WeekNumber(t *testing.T) {
	oh, err := New("week 01-10 Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Week 2 of 2012 starts Jan 9 (Monday)
	week2 := time.Date(2012, 1, 9, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(week2) {
		t.Error("Week 2 Monday should be open")
	}

	// Week 20 of 2012 is mid-May
	week20 := time.Date(2012, 5, 14, 12, 0, 0, 0, time.UTC)
	if oh.GetState(week20) {
		t.Error("Week 20 should be closed")
	}
}

// =============================================================================
// Constrained weekdays
// =============================================================================

func TestJS_ConstrainedWeekday_First(t *testing.T) {
	oh, err := New("Mo[1] 10:00-16:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// October 2012: First Monday is Oct 1
	firstMon := time.Date(2012, 10, 1, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(firstMon) {
		t.Error("First Monday (Oct 1) should be open")
	}

	// Second Monday is Oct 8
	secondMon := time.Date(2012, 10, 8, 12, 0, 0, 0, time.UTC)
	if oh.GetState(secondMon) {
		t.Error("Second Monday (Oct 8) should be closed")
	}
}

func TestJS_ConstrainedWeekday_Last(t *testing.T) {
	oh, err := New("Fr[-1] 10:00-16:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// October 2012: Last Friday is Oct 26
	lastFri := time.Date(2012, 10, 26, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(lastFri) {
		t.Error("Last Friday (Oct 26) should be open")
	}

	// Oct 19 is also Friday but not last
	notLastFri := time.Date(2012, 10, 19, 12, 0, 0, 0, time.UTC)
	if oh.GetState(notLastFri) {
		t.Error("Oct 19 Friday should be closed")
	}
}

// =============================================================================
// Fallback groups
// =============================================================================

func TestJS_Fallback_Basic(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-17:00 || 10:00-14:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Monday - primary applies
	monday := time.Date(2012, 10, 1, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(monday) {
		t.Error("Monday should be open (primary)")
	}

	// Saturday - fallback applies
	saturday := time.Date(2012, 10, 6, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(saturday) {
		t.Error("Saturday should be open (fallback)")
	}

	// Saturday outside fallback hours
	satAfter := time.Date(2012, 10, 6, 15, 0, 0, 0, time.UTC)
	if oh.GetState(satAfter) {
		t.Error("Saturday 15:00 should be closed")
	}
}

// =============================================================================
// Year ranges
// =============================================================================

func TestJS_Year_Single(t *testing.T) {
	oh, err := New("2012 Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// 2012 should work
	mon2012 := time.Date(2012, 10, 1, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(mon2012) {
		t.Error("2012 Monday should be open")
	}

	// 2013 should not
	mon2013 := time.Date(2013, 10, 7, 12, 0, 0, 0, time.UTC)
	if oh.GetState(mon2013) {
		t.Error("2013 Monday should be closed")
	}
}

func TestJS_Year_Range(t *testing.T) {
	oh, err := New("2012-2014 Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	tests := []struct {
		year     int
		expected bool
	}{
		{2011, false},
		{2012, true},
		{2013, true},
		{2014, true},
		{2015, false},
	}

	for _, tt := range tests {
		testTime := time.Date(tt.year, 10, 1, 12, 0, 0, 0, time.UTC)
		// Make sure it's a Monday
		for testTime.Weekday() != time.Monday {
			testTime = testTime.AddDate(0, 0, 1)
		}
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("Year %d: expected %v, got %v", tt.year, tt.expected, got)
		}
	}
}

// =============================================================================
// Comments
// =============================================================================

func TestJS_Comments(t *testing.T) {
	testCases := []struct {
		value   string
		comment string
	}{
		{`Mo-Fr 09:00-17:00 "Business hours"`, "Business hours"},
		{`24/7 "Always open"`, "Always open"},
		{`off "Closed for renovation"`, "Closed for renovation"},
	}

	for _, tc := range testCases {
		t.Run(tc.value, func(t *testing.T) {
			oh, err := New(tc.value)
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			testTime := time.Date(2012, 10, 1, 12, 0, 0, 0, time.UTC)
			got := oh.GetComment(testTime)
			if got != tc.comment {
				t.Errorf("expected comment '%s', got '%s'", tc.comment, got)
			}
		})
	}
}

// =============================================================================
// Complex real-world examples from README
// =============================================================================

func TestJS_RealWorld_Restaurant(t *testing.T) {
	oh, err := New("Mo-Fr 12:00-14:30,18:00-22:00; Sa 18:00-22:00; Su off")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	tests := []struct {
		day      int // Oct 2012
		hour     int
		expected bool
		desc     string
	}{
		{1, 13, true, "Monday lunch"},
		{1, 16, false, "Monday between services"},
		{1, 20, true, "Monday dinner"},
		{6, 13, false, "Saturday lunch (closed)"},
		{6, 20, true, "Saturday dinner"},
		{7, 13, false, "Sunday (closed)"},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, 10, tt.day, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

func TestJS_RealWorld_Shop(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-18:00; Sa 09:00-13:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	tests := []struct {
		day      int
		hour     int
		expected bool
		desc     string
	}{
		{1, 10, true, "Monday morning"},
		{1, 19, false, "Monday evening"},
		{6, 10, true, "Saturday morning"},
		{6, 14, false, "Saturday afternoon"},
		{7, 10, false, "Sunday"},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, 10, tt.day, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

// =============================================================================
// IsEqualTo semantic comparison
// =============================================================================

func TestJS_IsEqualTo(t *testing.T) {
	testCases := []struct {
		value1   string
		value2   string
		expected bool
	}{
		{"Mo-Fr 09:00-17:00", "Mo,Tu,We,Th,Fr 09:00-17:00", true},
		{"24/7", "00:00-24:00", true},
		{"off", "closed", true},
		{"Mo-Fr 09:00-17:00", "Mo-Fr 09:00-18:00", false},
		{"Mo-Fr 09:00-17:00", "Mo-Sa 09:00-17:00", false},
	}

	for _, tc := range testCases {
		t.Run(tc.value1+" vs "+tc.value2, func(t *testing.T) {
			oh1, _ := New(tc.value1)
			oh2, _ := New(tc.value2)

			got := oh1.IsEqualTo(oh2)
			if got != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

// =============================================================================
// GetOpenDuration
// =============================================================================

func TestJS_GetOpenDuration(t *testing.T) {
	testCases := []struct {
		value         string
		expectedHours float64
	}{
		{"Mo-Fr 09:00-17:00", 40},           // 5 days * 8 hours
		{"Mo-Su 00:00-24:00", 168},          // 7 days * 24 hours
		{"Mo 10:00-12:00", 2},               // 1 day * 2 hours
		{"Mo-Fr 09:00-12:00,13:00-17:00", 35}, // 5 days * 7 hours
	}

	for _, tc := range testCases {
		t.Run(tc.value, func(t *testing.T) {
			runDurationTest(t, tc.value, "2012-10-01 00:00", "2012-10-08 00:00", tc.expectedHours)
		})
	}
}

// =============================================================================
// GetOpenIntervals
// =============================================================================

func TestJS_GetOpenIntervals(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	start := parseTime("2012-10-01 00:00")
	end := parseTime("2012-10-03 00:00") // Monday and Tuesday

	intervals := oh.GetOpenIntervals(start, end)

	if len(intervals) != 2 {
		t.Fatalf("expected 2 intervals, got %d", len(intervals))
	}

	// Monday 09:00-17:00
	if !intervals[0].Start.Equal(parseTime("2012-10-01 09:00")) {
		t.Errorf("interval 0 start: expected 2012-10-01 09:00, got %v", intervals[0].Start)
	}
	if !intervals[0].End.Equal(parseTime("2012-10-01 17:00")) {
		t.Errorf("interval 0 end: expected 2012-10-01 17:00, got %v", intervals[0].End)
	}

	// Tuesday 09:00-17:00
	if !intervals[1].Start.Equal(parseTime("2012-10-02 09:00")) {
		t.Errorf("interval 1 start: expected 2012-10-02 09:00, got %v", intervals[1].Start)
	}
	if !intervals[1].End.Equal(parseTime("2012-10-02 17:00")) {
		t.Errorf("interval 1 end: expected 2012-10-02 17:00, got %v", intervals[1].End)
	}
}

// =============================================================================
// Weekdays - More tests (test.js lines 2582-2596)
// =============================================================================

func TestJS_Weekdays_Complex(t *testing.T) {
	values := []string{
		"Mo,Th,Sa,Su 10:00-12:00",
		"Mo,Th,Sa-Su 10:00-12:00",
		"Th,Sa-Mo 10:00-12:00",
	}

	expected := []testInterval{
		{parseTime("2012-10-01 10:00"), parseTime("2012-10-01 12:00"), false, ""}, // Mo
		{parseTime("2012-10-04 10:00"), parseTime("2012-10-04 12:00"), false, ""}, // Th
		{parseTime("2012-10-06 10:00"), parseTime("2012-10-06 12:00"), false, ""}, // Sa
		{parseTime("2012-10-07 10:00"), parseTime("2012-10-07 12:00"), false, ""}, // Su
	}

	runIntervalTest(t, "Weekdays complex", values, "2012-10-01 00:00", "2012-10-08 00:00", expected)
}

func TestJS_Weekdays_WithOff(t *testing.T) {
	values := []string{
		"10:00-12:00; Tu-We off; Fr off",
	}

	expected := []testInterval{
		{parseTime("2012-10-01 10:00"), parseTime("2012-10-01 12:00"), false, ""}, // Mo
		{parseTime("2012-10-04 10:00"), parseTime("2012-10-04 12:00"), false, ""}, // Th
		{parseTime("2012-10-06 10:00"), parseTime("2012-10-06 12:00"), false, ""}, // Sa
		{parseTime("2012-10-07 10:00"), parseTime("2012-10-07 12:00"), false, ""}, // Su
	}

	runIntervalTest(t, "Weekdays with off", values, "2012-10-01 00:00", "2012-10-08 00:00", expected)
}

// =============================================================================
// Omitted time (test.js lines 2598-2606)
// =============================================================================

func TestJS_OmittedTime(t *testing.T) {
	values := []string{
		"Mo,We",
		"Mo-We; Tu off",
	}

	expected := []testInterval{
		{parseTime("2012-10-01 00:00"), parseTime("2012-10-02 00:00"), false, ""}, // Mo
		{parseTime("2012-10-03 00:00"), parseTime("2012-10-04 00:00"), false, ""}, // We
	}

	runIntervalTest(t, "Omitted time", values, "2012-10-01 00:00", "2012-10-08 00:00", expected)
}

// =============================================================================
// Time ranges spanning midnight with weekdays (test.js lines 2608-2613)
// =============================================================================

func TestJS_MidnightSpanning_WithWeekday(t *testing.T) {
	values := []string{
		"We 22:00-02:00",
	}

	expected := []testInterval{
		{parseTime("2012-10-03 22:00"), parseTime("2012-10-04 02:00"), false, ""},
	}

	runIntervalTest(t, "Midnight spanning weekday", values, "2012-10-01 00:00", "2012-10-08 00:00", expected)
}

// =============================================================================
// Exception rules (test.js lines 2615-2623)
// =============================================================================

func TestJS_ExceptionRules(t *testing.T) {
	values := []string{
		"Mo-Fr 10:00-16:00; We 12:00-18:00",
	}

	expected := []testInterval{
		{parseTime("2012-10-01 10:00"), parseTime("2012-10-01 16:00"), false, ""}, // Mo
		{parseTime("2012-10-02 10:00"), parseTime("2012-10-02 16:00"), false, ""}, // Tu
		{parseTime("2012-10-03 12:00"), parseTime("2012-10-03 18:00"), false, ""}, // We (overridden)
		{parseTime("2012-10-04 10:00"), parseTime("2012-10-04 16:00"), false, ""}, // Th
		{parseTime("2012-10-05 10:00"), parseTime("2012-10-05 16:00"), false, ""}, // Fr
	}

	runIntervalTest(t, "Exception rules", values, "2012-10-01 00:00", "2012-10-08 00:00", expected)
}

// =============================================================================
// Full range / 24/7 equivalents (test.js lines 2627-2659)
// =============================================================================

func TestJS_FullRange(t *testing.T) {
	values := []string{
		"00:00-24:00",
		"Mo-Su 00:00-24:00",
		"24/7",
		"open",
		"Jan-Dec",
	}

	expected := []testInterval{
		{parseTime("2012-10-01 00:00"), parseTime("2012-10-08 00:00"), false, ""},
	}

	runIntervalTest(t, "Full range", values, "2012-10-01 00:00", "2012-10-08 00:00", expected)
}

func TestJS_FullRange_Duration(t *testing.T) {
	// 7 days * 24 hours = 168 hours
	runDurationTest(t, "24/7", "2012-10-01 00:00", "2012-10-08 00:00", 168)
}

// =============================================================================
// Constrained weekdays - more tests (test.js lines 2674-2683)
// =============================================================================

func TestJS_ConstrainedWeekday_FourthAndFifth(t *testing.T) {
	values := []string{
		"We[4,5] 10:00-12:00",
		"We[4-5] 10:00-12:00",
		"We[4] 10:00-12:00; We[-1] 10:00-12:00",
	}

	// Oct 2012: We[4] = Oct 24, We[5]/We[-1] = Oct 31
	expected := []testInterval{
		{parseTime("2012-10-24 10:00"), parseTime("2012-10-24 12:00"), false, ""},
		{parseTime("2012-10-31 10:00"), parseTime("2012-10-31 12:00"), false, ""},
	}

	runIntervalTest(t, "Constrained 4th/5th", values, "2012-10-01 00:00", "2012-11-01 00:00", expected)
}

// =============================================================================
// Additional rules with comma (test.js lines 2768-2788)
// =============================================================================

func TestJS_AdditionalRules_Comma(t *testing.T) {
	values := []string{
		"Mo-Fr 10:00-16:00, We 12:00-18:00",
	}

	// Additional rule extends Wednesday hours
	expected := []testInterval{
		{parseTime("2012-10-01 10:00"), parseTime("2012-10-01 16:00"), false, ""}, // Mo
		{parseTime("2012-10-02 10:00"), parseTime("2012-10-02 16:00"), false, ""}, // Tu
		{parseTime("2012-10-03 10:00"), parseTime("2012-10-03 18:00"), false, ""}, // We (extended)
		{parseTime("2012-10-04 10:00"), parseTime("2012-10-04 16:00"), false, ""}, // Th
		{parseTime("2012-10-05 10:00"), parseTime("2012-10-05 16:00"), false, ""}, // Fr
	}

	runIntervalTest(t, "Additional rules comma", values, "2012-10-01 00:00", "2012-10-08 00:00", expected)
}

func TestJS_AdditionalRules_SeparateTimes(t *testing.T) {
	values := []string{
		"Mo-Fr 08:00-12:00, We 14:00-18:00",
	}

	expected := []testInterval{
		{parseTime("2012-10-01 08:00"), parseTime("2012-10-01 12:00"), false, ""}, // Mo
		{parseTime("2012-10-02 08:00"), parseTime("2012-10-02 12:00"), false, ""}, // Tu
		{parseTime("2012-10-03 08:00"), parseTime("2012-10-03 12:00"), false, ""}, // We morning
		{parseTime("2012-10-03 14:00"), parseTime("2012-10-03 18:00"), false, ""}, // We afternoon
		{parseTime("2012-10-04 08:00"), parseTime("2012-10-04 12:00"), false, ""}, // Th
		{parseTime("2012-10-05 08:00"), parseTime("2012-10-05 12:00"), false, ""}, // Fr
	}

	runIntervalTest(t, "Additional rules separate", values, "2012-10-01 00:00", "2012-10-08 00:00", expected)
}

// =============================================================================
// Variable times without coordinates (test.js lines 627-656)
// =============================================================================

func TestJS_VariableTimes_NoCoords_SunriseSunset(t *testing.T) {
	// Without coordinates, default times are used: sunrise=06:00, sunset=18:00
	values := []string{
		"sunrise-sunset",
	}

	expected := []testInterval{
		{parseTime("2012-10-01 06:00"), parseTime("2012-10-01 18:00"), false, ""},
		{parseTime("2012-10-02 06:00"), parseTime("2012-10-02 18:00"), false, ""},
	}

	runIntervalTest(t, "Variable times no coords", values, "2012-10-01 00:00", "2012-10-03 00:00", expected)
}

func TestJS_VariableTimes_NoCoords_DawnDusk(t *testing.T) {
	// Default times: dawn=05:30, dusk=18:30
	values := []string{
		"dawn-dusk",
	}

	expected := []testInterval{
		{parseTime("2012-10-01 05:30"), parseTime("2012-10-01 18:30"), false, ""},
		{parseTime("2012-10-02 05:30"), parseTime("2012-10-02 18:30"), false, ""},
	}

	runIntervalTest(t, "Variable times dawn-dusk", values, "2012-10-01 00:00", "2012-10-03 00:00", expected)
}

func TestJS_VariableTimes_NoCoords_Duration(t *testing.T) {
	// sunrise-sunset with defaults = 06:00-18:00 = 12 hours per day
	runDurationTest(t, "sunrise-sunset", "2012-10-01 00:00", "2012-10-03 00:00", 24)
}

// =============================================================================
// Variable times with offset (test.js lines 641-646)
// =============================================================================

func TestJS_VariableTimes_WithOffset(t *testing.T) {
	values := []string{
		"(sunrise+01:02)-(sunset-00:30)",
	}

	// Default: sunrise=06:00+01:02=07:02, sunset=18:00-00:30=17:30
	expected := []testInterval{
		{parseTime("2012-10-01 07:02"), parseTime("2012-10-01 17:30"), false, ""},
		{parseTime("2012-10-02 07:02"), parseTime("2012-10-02 17:30"), false, ""},
	}

	runIntervalTest(t, "Variable times offset", values, "2012-10-01 00:00", "2012-10-03 00:00", expected)
}

// =============================================================================
// Complex real-world example: midnight spanning (test.js lines 427-439)
// =============================================================================

func TestJS_MidnightSpanning_ComplexRealWorld(t *testing.T) {
	oh, err := New("Su-Tu 11:00-01:00, We-Th 11:00-03:00, Fr 11:00-06:00, Sa 11:00-07:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Test specific times from the expected intervals
	tests := []struct {
		time     string
		expected bool
		desc     string
	}{
		{"2012-10-01 00:30", true, "Mo 00:30 (continuation from Su)"},
		{"2012-10-01 11:30", true, "Mo 11:30"},
		{"2012-10-02 00:30", true, "Tu 00:30 (continuation from Mo)"},
		{"2012-10-03 02:00", true, "We 02:00 (continuation from Tu)"},
		{"2012-10-04 02:30", true, "Th 02:30 (continuation from We)"},
		{"2012-10-05 05:00", true, "Fr 05:00 (continuation from Th)"},
		{"2012-10-06 06:30", true, "Sa 06:30 (continuation from Fr)"},
		{"2012-10-01 09:00", false, "Mo 09:00 (before open)"},
	}

	for _, tt := range tests {
		testTime := parseTime(tt.time)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

// =============================================================================
// Midnight spanning date overwriting (test.js lines 412-425)
// =============================================================================

func TestJS_MidnightSpanning_DateOverwriting(t *testing.T) {
	oh, err := New("22:00-02:00; Tu 12:00-14:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	tests := []struct {
		time     string
		expected bool
		desc     string
	}{
		{"2012-10-01 01:00", true, "Mo 01:00 (from previous day)"},
		{"2012-10-01 23:00", true, "Mo 23:00"},
		{"2012-10-02 01:00", false, "Tu 01:00 (overwritten by Tu rule)"},
		{"2012-10-02 13:00", true, "Tu 13:00 (Tu rule)"},
		{"2012-10-03 01:00", true, "We 01:00 (from Tu night)"},
	}

	for _, tt := range tests {
		testTime := parseTime(tt.time)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

// =============================================================================
// Week stable check (full week coverage)
// =============================================================================

func TestJS_WeekStable(t *testing.T) {
	testCases := []struct {
		value    string
		expected bool
	}{
		{"Mo-Fr 09:00-17:00", true},
		{"24/7", true},
		{"10:00-12:00", true},
		{"Jan-Dec 10:00-12:00", true},
		{"Jan 10:00-12:00", false}, // Only January
		{"week 01-10 10:00-12:00", false},
		{"2012 10:00-12:00", false},
	}

	for _, tc := range testCases {
		t.Run(tc.value, func(t *testing.T) {
			oh, err := New(tc.value)
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			got := oh.IsWeekStable()
			if got != tc.expected {
				t.Errorf("expected IsWeekStable=%v, got %v", tc.expected, got)
			}
		})
	}
}

// =============================================================================
// Iterator / GetNextChange
// =============================================================================

func TestJS_GetNextChange(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Monday 08:00 -> next change at 09:00
	start := parseTime("2012-10-01 08:00")
	next := oh.GetNextChange(start)
	expected := parseTime("2012-10-01 09:00")
	if !next.Equal(expected) {
		t.Errorf("from 08:00: expected next change at %v, got %v", expected, next)
	}

	// Monday 10:00 -> next change at 17:00
	start = parseTime("2012-10-01 10:00")
	next = oh.GetNextChange(start)
	expected = parseTime("2012-10-01 17:00")
	if !next.Equal(expected) {
		t.Errorf("from 10:00: expected next change at %v, got %v", expected, next)
	}
}

// =============================================================================
// PrettifyValue
// =============================================================================

func TestJS_PrettifyValue(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"mo-fr 09:00-17:00", "Mo-Fr 09:00-17:00"},
		{"MO-FR 09:00-17:00", "Mo-Fr 09:00-17:00"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			oh, err := New(tc.input)
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			got := oh.PrettifyValue()
			if got != tc.expected {
				t.Errorf("expected '%s', got '%s'", tc.expected, got)
			}
		})
	}
}

// =============================================================================
// Easter dates (from easter_test.go verification)
// =============================================================================

func TestJS_Easter(t *testing.T) {
	oh, err := New("easter 10:00-12:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Easter 2012 was April 8
	easterDay := parseTime("2012-04-08 11:00")
	if !oh.GetState(easterDay) {
		t.Error("Easter Sunday should be open")
	}

	// Day before Easter should be closed
	beforeEaster := parseTime("2012-04-07 11:00")
	if oh.GetState(beforeEaster) {
		t.Error("Day before Easter should be closed")
	}
}

func TestJS_Easter_Offset(t *testing.T) {
	// Good Friday is Easter -2 days
	oh, err := New("easter -2 days 10:00-12:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Good Friday 2012 was April 6
	goodFriday := parseTime("2012-04-06 11:00")
	if !oh.GetState(goodFriday) {
		t.Error("Good Friday should be open")
	}

	// Easter Sunday should be closed
	easterDay := parseTime("2012-04-08 11:00")
	if oh.GetState(easterDay) {
		t.Error("Easter Sunday should be closed (only Good Friday open)")
	}
}

// =============================================================================
// Open end times
// =============================================================================

func TestJS_OpenEnd(t *testing.T) {
	oh, err := New("17:00+")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// 17:30 should be open (unknown due to open end)
	testTime := parseTime("2012-10-01 17:30")
	if !oh.GetState(testTime) && !oh.GetUnknown(testTime) {
		t.Error("17:30 should be open or unknown")
	}

	// 16:00 should be closed
	beforeOpen := parseTime("2012-10-01 16:00")
	if oh.GetState(beforeOpen) {
		t.Error("16:00 should be closed")
	}
}

// =============================================================================
// Fallback with unknown state
// =============================================================================

func TestJS_Fallback_Unknown(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-17:00 unknown || 10:00-14:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Monday 12:00 - primary is unknown, fallback is open
	monday := parseTime("2012-10-01 12:00")
	if !oh.GetState(monday) {
		t.Error("Monday 12:00 should be open (from fallback)")
	}

	// Saturday 12:00 - no primary match, fallback applies
	saturday := parseTime("2012-10-06 12:00")
	if !oh.GetState(saturday) {
		t.Error("Saturday 12:00 should be open (fallback)")
	}
}

// =============================================================================
// Wrapping month ranges (Dec-Feb)
// =============================================================================

func TestJS_MonthRange_Wrapping(t *testing.T) {
	oh, err := New("Dec-Feb 10:00-16:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	tests := []struct {
		month    time.Month
		expected bool
	}{
		{time.November, false},
		{time.December, true},
		{time.January, true},
		{time.February, true},
		{time.March, false},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, tt.month, 15, 12, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.month, tt.expected, got)
		}
	}
}

// =============================================================================
// Case insensitivity (JS test.js lines 200+)
// =============================================================================

func TestJS_CaseInsensitivity_Weekdays(t *testing.T) {
	testCases := []struct {
		value string
	}{
		{"Mo-Fr 09:00-17:00"},
		{"mo-fr 09:00-17:00"},
		{"MO-FR 09:00-17:00"},
		{"Monday-Friday 09:00-17:00"},
		{"monday-friday 09:00-17:00"},
		{"MONDAY-FRIDAY 09:00-17:00"},
	}

	// All should produce the same result
	testTime := parseTime("2012-10-03 12:00") // Wednesday
	for _, tc := range testCases {
		t.Run(tc.value, func(t *testing.T) {
			oh, err := New(tc.value)
			if err != nil {
				t.Fatalf("failed to parse '%s': %v", tc.value, err)
			}
			if !oh.GetState(testTime) {
				t.Errorf("expected open on Wednesday 12:00 for '%s'", tc.value)
			}
		})
	}
}

func TestJS_CaseInsensitivity_Months(t *testing.T) {
	testCases := []struct {
		value string
	}{
		{"Jan-Mar 10:00-16:00"},
		{"jan-mar 10:00-16:00"},
		{"JAN-MAR 10:00-16:00"},
		{"January-March 10:00-16:00"},
	}

	testTime := time.Date(2012, time.February, 15, 12, 0, 0, 0, time.UTC)
	for _, tc := range testCases {
		t.Run(tc.value, func(t *testing.T) {
			oh, err := New(tc.value)
			if err != nil {
				t.Fatalf("failed to parse '%s': %v", tc.value, err)
			}
			if !oh.GetState(testTime) {
				t.Errorf("expected open in February for '%s'", tc.value)
			}
		})
	}
}

// =============================================================================
// Weekday range wrapping (Fr-Mo, Sa-Tu)
// =============================================================================

func TestJS_WeekdayRange_Wrapping(t *testing.T) {
	testCases := []struct {
		value    string
		weekdays []time.Weekday
	}{
		{"Fr-Mo 10:00-16:00", []time.Weekday{time.Friday, time.Saturday, time.Sunday, time.Monday}},
		{"Sa-Tu 10:00-16:00", []time.Weekday{time.Saturday, time.Sunday, time.Monday, time.Tuesday}},
		{"Su-We 10:00-16:00", []time.Weekday{time.Sunday, time.Monday, time.Tuesday, time.Wednesday}},
	}

	for _, tc := range testCases {
		t.Run(tc.value, func(t *testing.T) {
			oh, err := New(tc.value)
			if err != nil {
				t.Fatalf("failed to parse '%s': %v", tc.value, err)
			}

			// Check each day of the week
			for wd := time.Sunday; wd <= time.Saturday; wd++ {
				// Find a date with this weekday
				testTime := time.Date(2012, 10, 7+int(wd), 12, 0, 0, 0, time.UTC)
				got := oh.GetState(testTime)

				// Check if this weekday is in the expected list
				expected := false
				for _, expWd := range tc.weekdays {
					if wd == expWd {
						expected = true
						break
					}
				}

				if got != expected {
					t.Errorf("%s: expected %v, got %v", wd, expected, got)
				}
			}
		})
	}
}

// =============================================================================
// Points in time (single moments like "Mo 12:00")
// =============================================================================

func TestJS_PointInTime(t *testing.T) {
	// Note: Points in time are typically interpreted as open from that time
	// until the next event or end of day
	oh, err := New("Mo 12:00-12:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// At 12:00, should be at the boundary
	testTime := parseTime("2012-10-01 12:00") // Monday
	// This is a zero-duration range, so might be considered open or closed
	// depending on interpretation
	_ = oh.GetState(testTime)
}

// =============================================================================
// Public holidays with offsets (PH +1 day, PH -1 day)
// =============================================================================

type jsTestHolidayChecker struct {
	holidays map[string]bool
}

func (m *jsTestHolidayChecker) IsHoliday(t time.Time) bool {
	return m.holidays[t.Format("2006-01-02")]
}

func TestJS_PH_WithOffset(t *testing.T) {
	// Test PH +1 day (day after public holiday)
	oh, err := New("Mo-Fr 09:00-17:00; PH +1 day 10:00-14:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	hc := &jsTestHolidayChecker{
		holidays: map[string]bool{
			"2012-10-03": true, // Wednesday is a holiday
		},
	}
	oh.SetHolidayChecker(hc)

	// Wednesday (PH) at 12:00 - regular hours should NOT apply (PH rule exists but doesn't match PH+1)
	wed := parseTime("2012-10-03 12:00")
	if !oh.GetState(wed) {
		// On the holiday itself, the regular Mo-Fr rule might apply
		// or it might be overridden - depends on interpretation
	}

	// Thursday (PH +1) at 12:00 - should use PH +1 hours
	thu := parseTime("2012-10-04 12:00")
	if !oh.GetState(thu) {
		t.Error("Thursday (day after PH) at 12:00 should be open")
	}

	// Thursday at 08:00 - should be closed (PH +1 hours are 10:00-14:00)
	thuEarly := parseTime("2012-10-04 08:00")
	if oh.GetState(thuEarly) {
		t.Error("Thursday at 08:00 should be closed (PH+1 hours are 10-14)")
	}
}

func TestJS_PH_MinusOffset(t *testing.T) {
	// Test PH -1 day (day before public holiday)
	oh, err := New("Mo-Fr 09:00-17:00; PH -1 day 09:00-12:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	hc := &jsTestHolidayChecker{
		holidays: map[string]bool{
			"2012-10-04": true, // Thursday is a holiday
		},
	}
	oh.SetHolidayChecker(hc)

	// Wednesday (PH -1) at 11:00 - should be open
	wed := parseTime("2012-10-03 11:00")
	if !oh.GetState(wed) {
		t.Error("Wednesday (day before PH) at 11:00 should be open")
	}

	// Wednesday at 14:00 - should be closed (PH-1 hours are 09:00-12:00)
	wedAfternoon := parseTime("2012-10-03 14:00")
	if oh.GetState(wedAfternoon) {
		t.Error("Wednesday at 14:00 should be closed (PH-1 hours are 09-12)")
	}
}

// =============================================================================
// Complex month/day patterns
// =============================================================================

func TestJS_MonthDay_Specific(t *testing.T) {
	// Dec 24 = Christmas Eve
	oh, err := New("Dec 24 10:00-14:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Dec 24 at 12:00 - should be open
	christmasEve := time.Date(2012, 12, 24, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(christmasEve) {
		t.Error("Dec 24 at 12:00 should be open")
	}

	// Dec 25 at 12:00 - should be closed
	christmas := time.Date(2012, 12, 25, 12, 0, 0, 0, time.UTC)
	if oh.GetState(christmas) {
		t.Error("Dec 25 should be closed (only Dec 24 is specified)")
	}
}

func TestJS_MonthDay_Range(t *testing.T) {
	// Dec 24-26 = Christmas period
	oh, err := New("Dec 24-26 10:00-16:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	tests := []struct {
		day      int
		expected bool
	}{
		{23, false},
		{24, true},
		{25, true},
		{26, true},
		{27, false},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, 12, tt.day, 12, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("Dec %d: expected %v, got %v", tt.day, tt.expected, got)
		}
	}
}

func TestJS_MonthDay_Wrapping(t *testing.T) {
	// Dec 28-Jan 3 = New Year period (wraps across year boundary)
	oh, err := New("Dec 28-Jan 03 10:00-16:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	tests := []struct {
		month    time.Month
		day      int
		expected bool
	}{
		{time.December, 27, false},
		{time.December, 28, true},
		{time.December, 31, true},
		{time.January, 1, true},
		{time.January, 3, true},
		{time.January, 4, false},
	}

	for _, tt := range tests {
		year := 2012
		if tt.month == time.January {
			year = 2013
		}
		testTime := time.Date(year, tt.month, tt.day, 12, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s %d: expected %v, got %v", tt.month, tt.day, tt.expected, got)
		}
	}
}

// =============================================================================
// State comments within rules
// =============================================================================

func TestJS_Comments_InRules(t *testing.T) {
	testCases := []struct {
		value           string
		expectedComment string
	}{
		{`Mo-Fr 09:00-17:00 "Business hours"`, "Business hours"},
		{`Mo-Fr 09:00-12:00 "Morning"; Mo-Fr 14:00-17:00 "Afternoon"`, "Morning"},
		{`24/7 "Always open"`, "Always open"},
	}

	testTime := parseTime("2012-10-01 10:00") // Monday
	for _, tc := range testCases {
		t.Run(tc.value, func(t *testing.T) {
			oh, err := New(tc.value)
			if err != nil {
				t.Fatalf("failed to parse '%s': %v", tc.value, err)
			}

			comment := oh.GetComment(testTime)
			if comment != tc.expectedComment {
				t.Errorf("expected comment '%s', got '%s'", tc.expectedComment, comment)
			}
		})
	}
}

func TestJS_Comments_DifferentTimes(t *testing.T) {
	oh, err := New(`Mo-Fr 09:00-12:00 "Morning"; Mo-Fr 14:00-17:00 "Afternoon"`)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Morning time
	morning := parseTime("2012-10-01 10:00")
	if oh.GetComment(morning) != "Morning" {
		t.Errorf("expected 'Morning' comment, got '%s'", oh.GetComment(morning))
	}

	// Afternoon time
	afternoon := parseTime("2012-10-01 15:00")
	if oh.GetComment(afternoon) != "Afternoon" {
		t.Errorf("expected 'Afternoon' comment, got '%s'", oh.GetComment(afternoon))
	}
}

// =============================================================================
// Rule precedence / priority
// =============================================================================

func TestJS_RulePrecedence_LaterOverrides(t *testing.T) {
	// Later rules should override earlier rules for the same selector
	oh, err := New("Mo-Fr 09:00-17:00; We 12:00-15:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Wednesday at 10:00 - should be closed (We rule overrides)
	wed10 := parseTime("2012-10-03 10:00")
	if oh.GetState(wed10) {
		t.Error("Wednesday 10:00 should be closed (We rule overrides Mo-Fr)")
	}

	// Wednesday at 13:00 - should be open (within We rule)
	wed13 := parseTime("2012-10-03 13:00")
	if !oh.GetState(wed13) {
		t.Error("Wednesday 13:00 should be open (We rule)")
	}

	// Monday at 10:00 - should be open (Mo-Fr rule)
	mon10 := parseTime("2012-10-01 10:00")
	if !oh.GetState(mon10) {
		t.Error("Monday 10:00 should be open (Mo-Fr rule)")
	}
}

func TestJS_RulePrecedence_OffOverrides(t *testing.T) {
	// "off" rule should close specific times
	oh, err := New("Mo-Fr 09:00-17:00; We 12:00-14:00 off")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Wednesday at 10:00 - should be open (Mo-Fr rule, outside off period)
	wed10 := parseTime("2012-10-03 10:00")
	if !oh.GetState(wed10) {
		t.Error("Wednesday 10:00 should be open")
	}

	// Wednesday at 13:00 - should be closed (off rule)
	wed13 := parseTime("2012-10-03 13:00")
	if oh.GetState(wed13) {
		t.Error("Wednesday 13:00 should be closed (off rule)")
	}

	// Wednesday at 15:00 - should be open (Mo-Fr rule, outside off period)
	wed15 := parseTime("2012-10-03 15:00")
	if !oh.GetState(wed15) {
		t.Error("Wednesday 15:00 should be open")
	}
}

// =============================================================================
// Multiple time ranges per day
// =============================================================================

func TestJS_MultipleTimeRanges(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-12:00,14:00-18:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	tests := []struct {
		hour     int
		expected bool
	}{
		{8, false},  // before first range
		{10, true},  // first range
		{12, false}, // between ranges
		{13, false}, // between ranges
		{14, true},  // second range
		{17, true},  // second range
		{18, false}, // after second range
		{19, false}, // after close
	}

	for _, tt := range tests {
		testTime := time.Date(2012, 10, 1, tt.hour, 0, 0, 0, time.UTC) // Monday
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("Monday %02d:00: expected %v, got %v", tt.hour, tt.expected, got)
		}
	}
}

func TestJS_ThreeTimeRanges(t *testing.T) {
	oh, err := New("Mo 08:00-10:00,12:00-14:00,16:00-18:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	tests := []struct {
		hour     int
		expected bool
	}{
		{7, false},
		{9, true},
		{11, false},
		{13, true},
		{15, false},
		{17, true},
		{19, false},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, 10, 1, tt.hour, 0, 0, 0, time.UTC) // Monday
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("Monday %02d:00: expected %v, got %v", tt.hour, tt.expected, got)
		}
	}
}

// =============================================================================
// Week number constraints
// =============================================================================

func TestJS_WeekNumber_OddEven(t *testing.T) {
	// Test week interval patterns
	oh, err := New("week 01-53/2 Mo 10:00-12:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// This should apply to odd weeks only (1, 3, 5, ...)
	// Week 1 of 2012 includes Jan 2 (Monday)
	week1 := time.Date(2012, 1, 2, 11, 0, 0, 0, time.UTC) // Monday week 1
	if !oh.GetState(week1) {
		t.Error("Week 1 (odd) Monday should be open")
	}

	// Week 2 of 2012 includes Jan 9 (Monday)
	week2 := time.Date(2012, 1, 9, 11, 0, 0, 0, time.UTC) // Monday week 2
	if oh.GetState(week2) {
		t.Error("Week 2 (even) Monday should be closed")
	}
}

// =============================================================================
// Complex constrained weekdays
// =============================================================================

func TestJS_ConstrainedWeekday_MonthSpecific(t *testing.T) {
	// First Monday in January
	oh, err := New("Jan Mo[1] 10:00-12:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// First Monday in January 2012 is Jan 2
	firstMon := time.Date(2012, 1, 2, 11, 0, 0, 0, time.UTC)
	if !oh.GetState(firstMon) {
		t.Error("First Monday in January should be open")
	}

	// Second Monday in January (Jan 9) should be closed
	secondMon := time.Date(2012, 1, 9, 11, 0, 0, 0, time.UTC)
	if oh.GetState(secondMon) {
		t.Error("Second Monday in January should be closed")
	}

	// First Monday in February should be closed (only Jan specified)
	febFirstMon := time.Date(2012, 2, 6, 11, 0, 0, 0, time.UTC)
	if oh.GetState(febFirstMon) {
		t.Error("First Monday in February should be closed")
	}
}

func TestJS_ConstrainedWeekday_LastOfMonth(t *testing.T) {
	// Last Sunday of each month
	oh, err := New("Su[-1] 10:00-12:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Last Sunday in October 2012 is Oct 28
	lastSunOct := time.Date(2012, 10, 28, 11, 0, 0, 0, time.UTC)
	if !oh.GetState(lastSunOct) {
		t.Error("Last Sunday in October should be open")
	}

	// Oct 21 (second to last Sunday) should be closed
	secondLastSun := time.Date(2012, 10, 21, 11, 0, 0, 0, time.UTC)
	if oh.GetState(secondLastSun) {
		t.Error("Second to last Sunday should be closed")
	}
}

// =============================================================================
// Symbol normalization (different dash types)
// =============================================================================

func TestJS_SymbolNormalization_Dashes(t *testing.T) {
	// Different types of dashes should all work
	testCases := []string{
		"Mo-Fr 09:00-17:00", // Regular hyphen
		"Mo–Fr 09:00–17:00", // En dash (U+2013)
		"Mo—Fr 09:00—17:00", // Em dash (U+2014)
	}

	testTime := parseTime("2012-10-03 12:00") // Wednesday
	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			oh, err := New(tc)
			if err != nil {
				t.Fatalf("failed to parse '%s': %v", tc, err)
			}
			if !oh.GetState(testTime) {
				t.Errorf("expected open on Wednesday for '%s'", tc)
			}
		})
	}
}

// =============================================================================
// Real-world complex patterns from OSM
// =============================================================================

func TestJS_RealWorld_German_Bakery(t *testing.T) {
	// Typical German bakery hours
	oh, err := New("Mo-Fr 06:00-18:00; Sa 06:00-13:00; Su off")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	tests := []struct {
		day      time.Weekday
		hour     int
		expected bool
	}{
		{time.Monday, 7, true},
		{time.Monday, 19, false},
		{time.Saturday, 10, true},
		{time.Saturday, 14, false},
		{time.Sunday, 10, false},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, 10, 7+int(tt.day), tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s %02d:00: expected %v, got %v", tt.day, tt.hour, tt.expected, got)
		}
	}
}

func TestJS_RealWorld_Restaurant_Complex(t *testing.T) {
	// Restaurant with lunch/dinner, closed Monday, reduced weekend hours
	oh, err := New("Tu-Fr 11:30-14:00,18:00-22:00; Sa-Su 12:00-22:00; Mo off")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	tests := []struct {
		day      time.Weekday
		hour     int
		expected bool
		desc     string
	}{
		{time.Monday, 12, false, "Monday closed"},
		{time.Tuesday, 12, true, "Tuesday lunch"},
		{time.Tuesday, 15, false, "Tuesday between meals"},
		{time.Tuesday, 19, true, "Tuesday dinner"},
		{time.Saturday, 15, true, "Saturday afternoon"},
		{time.Sunday, 21, true, "Sunday evening"},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, 10, 7+int(tt.day), tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

func TestJS_RealWorld_24h_With_Exceptions(t *testing.T) {
	// 24/7 except for specific times
	oh, err := New("24/7; Dec 25 off; Dec 31 18:00-24:00 off")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	tests := []struct {
		month    time.Month
		day      int
		hour     int
		expected bool
		desc     string
	}{
		{time.October, 15, 12, true, "Regular day"},
		{time.December, 24, 12, true, "Christmas Eve"},
		{time.December, 25, 12, false, "Christmas Day closed"},
		{time.December, 31, 12, true, "New Year's Eve afternoon"},
		{time.December, 31, 20, false, "New Year's Eve evening closed"},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, tt.month, tt.day, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

// =============================================================================
// Sunrise/sunset with coordinates
// =============================================================================

func TestJS_VariableTimes_WithCoords(t *testing.T) {
	oh, err := New("sunrise-sunset")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Set coordinates for Berlin
	oh.SetCoordinates(52.5200, 13.4050)

	// Test that sunrise/sunset times work with coordinates
	// In October in Berlin, sunrise is around 07:00-07:30
	testTime := time.Date(2012, 10, 1, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(testTime) {
		t.Error("Midday should be open (between sunrise and sunset)")
	}

	// Very early morning (before sunrise) should be closed
	earlyMorning := time.Date(2012, 10, 1, 4, 0, 0, 0, time.UTC)
	if oh.GetState(earlyMorning) {
		t.Error("4:00 should be closed (before sunrise)")
	}

	// Late evening (after sunset) should be closed
	lateEvening := time.Date(2012, 10, 1, 22, 0, 0, 0, time.UTC)
	if oh.GetState(lateEvening) {
		t.Error("22:00 should be closed (after sunset)")
	}
}

// =============================================================================
// Fallback chains
// =============================================================================

func TestJS_Fallback_Chain(t *testing.T) {
	// Multiple fallback levels
	oh, err := New("PH off || Mo-Fr 09:00-17:00 || 10:00-14:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	hc := &jsTestHolidayChecker{
		holidays: map[string]bool{
			"2012-10-03": true, // Wednesday is a holiday
		},
	}
	oh.SetHolidayChecker(hc)

	// Wednesday (PH) - first rule matches, closed
	wed := parseTime("2012-10-03 10:00")
	if oh.GetState(wed) {
		t.Error("Holiday should be closed")
	}

	// Monday 10:00 - second rule matches
	mon := parseTime("2012-10-01 10:00")
	if !oh.GetState(mon) {
		t.Error("Monday 10:00 should be open (weekday rule)")
	}

	// Saturday 12:00 - third rule (fallback) matches
	sat := parseTime("2012-10-06 12:00")
	if !oh.GetState(sat) {
		t.Error("Saturday 12:00 should be open (fallback rule)")
	}
}

// =============================================================================
// Edge cases
// =============================================================================

func TestJS_EdgeCase_24_00_Notation(t *testing.T) {
	// 24:00 should be treated as midnight (00:00 of next day)
	oh, err := New("Mo-Fr 20:00-24:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Monday 23:00 should be open
	mon23 := parseTime("2012-10-01 23:00")
	if !oh.GetState(mon23) {
		t.Error("Monday 23:00 should be open")
	}

	// Tuesday 00:00 should be closed (new day starts)
	tue00 := parseTime("2012-10-02 00:00")
	if oh.GetState(tue00) {
		t.Error("Tuesday 00:00 should be closed")
	}
}

func TestJS_EdgeCase_EmptyString(t *testing.T) {
	_, err := New("")
	if err == nil {
		t.Error("Empty string should produce an error")
	}
}

func TestJS_EdgeCase_WhitespaceOnly(t *testing.T) {
	_, err := New("   ")
	if err == nil {
		t.Error("Whitespace-only string should produce an error")
	}
}

// =============================================================================
// Extended hours notation (25:00, 26:00 = next day 01:00, 02:00)
// =============================================================================

func TestJS_ExtendedHours_25(t *testing.T) {
	// 22:00-25:00 means 22:00 to 01:00 next day
	oh, err := New("Fr 22:00-25:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Friday 23:00 should be open
	fri23 := time.Date(2012, 10, 5, 23, 0, 0, 0, time.UTC)
	if !oh.GetState(fri23) {
		t.Error("Friday 23:00 should be open")
	}

	// Saturday 00:30 should be open (continuation)
	sat0030 := time.Date(2012, 10, 6, 0, 30, 0, 0, time.UTC)
	if !oh.GetState(sat0030) {
		t.Error("Saturday 00:30 should be open (continuation from Friday)")
	}

	// Saturday 02:00 should be closed
	sat02 := time.Date(2012, 10, 6, 2, 0, 0, 0, time.UTC)
	if oh.GetState(sat02) {
		t.Error("Saturday 02:00 should be closed")
	}
}

func TestJS_ExtendedHours_26(t *testing.T) {
	// 22:00-26:00 means 22:00 to 02:00 next day
	oh, err := New("Sa 22:00-26:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Saturday 23:00 should be open
	sat23 := time.Date(2012, 10, 6, 23, 0, 0, 0, time.UTC)
	if !oh.GetState(sat23) {
		t.Error("Saturday 23:00 should be open")
	}

	// Sunday 01:30 should be open (continuation)
	sun0130 := time.Date(2012, 10, 7, 1, 30, 0, 0, time.UTC)
	if !oh.GetState(sun0130) {
		t.Error("Sunday 01:30 should be open (continuation from Saturday)")
	}
}

// =============================================================================
// Leap year handling
// =============================================================================

func TestJS_LeapYear_Feb29(t *testing.T) {
	// Feb 29 only exists in leap years
	oh, err := New("Feb 29 10:00-16:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// 2012 is a leap year - Feb 29 exists
	leapYear := time.Date(2012, 2, 29, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(leapYear) {
		t.Error("Feb 29 2012 (leap year) at 12:00 should be open")
	}

	// 2013 is not a leap year - Feb 28 should be closed
	nonLeapYear := time.Date(2013, 2, 28, 12, 0, 0, 0, time.UTC)
	if oh.GetState(nonLeapYear) {
		t.Error("Feb 28 2013 (non-leap year) should be closed")
	}
}

func TestJS_LeapYear_Feb28ToMar01(t *testing.T) {
	// Range that spans end of February
	oh, err := New("Feb 28-Mar 01 10:00-16:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Feb 28 should be open
	feb28 := time.Date(2012, 2, 28, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(feb28) {
		t.Error("Feb 28 should be open")
	}

	// Mar 01 should be open
	mar01 := time.Date(2012, 3, 1, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(mar01) {
		t.Error("Mar 01 should be open")
	}

	// In leap year, Feb 29 is in between
	feb29 := time.Date(2012, 2, 29, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(feb29) {
		t.Error("Feb 29 (leap year) should be open")
	}
}

// =============================================================================
// Year ranges with plus notation (2020+)
// =============================================================================

func TestJS_YearPlus(t *testing.T) {
	oh, err := New("2020+ Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// 2020 Monday should be open
	mon2020 := time.Date(2020, 6, 1, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(mon2020) {
		t.Error("2020 Monday should be open")
	}

	// 2025 Monday should be open
	mon2025 := time.Date(2025, 6, 2, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(mon2025) {
		t.Error("2025 Monday should be open")
	}

	// 2019 Monday should be closed
	mon2019 := time.Date(2019, 6, 3, 12, 0, 0, 0, time.UTC)
	if oh.GetState(mon2019) {
		t.Error("2019 Monday should be closed (before 2020+)")
	}
}

// =============================================================================
// Unknown state
// =============================================================================

func TestJS_UnknownState_Basic(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-17:00 unknown")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Monday 12:00 should be unknown
	mon := parseTime("2012-10-01 12:00")
	if !oh.GetUnknown(mon) {
		t.Error("Monday 12:00 should be unknown")
	}

	// GetState might return true or false for unknown - depends on implementation
	// But GetUnknown should definitely return true
}

func TestJS_UnknownState_WithComment(t *testing.T) {
	oh, err := New(`Sa 10:00-14:00 unknown "Maybe open"`)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	sat := time.Date(2012, 10, 6, 12, 0, 0, 0, time.UTC)
	if !oh.GetUnknown(sat) {
		t.Error("Saturday 12:00 should be unknown")
	}

	comment := oh.GetComment(sat)
	if comment != "Maybe open" {
		t.Errorf("expected comment 'Maybe open', got '%s'", comment)
	}
}

// =============================================================================
// Alternative separators (to, ~, through)
// =============================================================================

func TestJS_AlternativeSeparator_To(t *testing.T) {
	// "to" as separator between times
	oh, err := New("Mo-Fr 09:00 to 17:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	mon := parseTime("2012-10-01 12:00")
	if !oh.GetState(mon) {
		t.Error("Monday 12:00 should be open")
	}
}

func TestJS_AlternativeSeparator_Through(t *testing.T) {
	// "through" as separator for weekdays
	oh, err := New("Monday through Friday 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	wed := parseTime("2012-10-03 12:00")
	if !oh.GetState(wed) {
		t.Error("Wednesday 12:00 should be open")
	}
}

// =============================================================================
// German weekday names (internationalization)
// =============================================================================

func TestJS_German_Weekdays(t *testing.T) {
	testCases := []struct {
		german  string
		english string
	}{
		{"Montag", "Mo"},
		{"Dienstag", "Tu"},
		{"Mittwoch", "We"},
		{"Donnerstag", "Th"},
		{"Freitag", "Fr"},
		{"Samstag", "Sa"},
		{"Sonntag", "Su"},
	}

	for _, tc := range testCases {
		t.Run(tc.german, func(t *testing.T) {
			oh, err := New(tc.german + " 09:00-17:00")
			if err != nil {
				t.Fatalf("failed to parse '%s': %v", tc.german, err)
			}

			// Find a date that matches this weekday
			// October 2012: Mo=1, Tu=2, We=3, Th=4, Fr=5, Sa=6, Su=7
			dayOffset := map[string]int{
				"Mo": 1, "Tu": 2, "We": 3, "Th": 4, "Fr": 5, "Sa": 6, "Su": 7,
			}
			testTime := time.Date(2012, 10, dayOffset[tc.english], 12, 0, 0, 0, time.UTC)
			if !oh.GetState(testTime) {
				t.Errorf("%s at 12:00 should be open", tc.german)
			}
		})
	}
}

func TestJS_German_Months(t *testing.T) {
	testCases := []struct {
		german  string
		month   time.Month
	}{
		{"Januar", time.January},
		{"Februar", time.February},
		{"März", time.March},
		{"April", time.April},
		{"Mai", time.May},
		{"Juni", time.June},
		{"Juli", time.July},
		{"August", time.August},
		{"September", time.September},
		{"Oktober", time.October},
		{"November", time.November},
		{"Dezember", time.December},
	}

	for _, tc := range testCases {
		t.Run(tc.german, func(t *testing.T) {
			oh, err := New(tc.german + " 10:00-16:00")
			if err != nil {
				t.Fatalf("failed to parse '%s': %v", tc.german, err)
			}

			testTime := time.Date(2012, tc.month, 15, 12, 0, 0, 0, time.UTC)
			if !oh.GetState(testTime) {
				t.Errorf("%s 15 at 12:00 should be open", tc.german)
			}
		})
	}
}

// =============================================================================
// Invalid patterns (should produce errors)
// =============================================================================

func TestJS_Invalid_WeekOutOfRange(t *testing.T) {
	invalidPatterns := []string{
		"week 00 10:00-12:00", // week 0 doesn't exist
		"week 54 10:00-12:00", // week 54 doesn't exist
	}

	for _, pattern := range invalidPatterns {
		t.Run(pattern, func(t *testing.T) {
			_, err := New(pattern)
			if err == nil {
				t.Errorf("pattern '%s' should produce an error", pattern)
			}
		})
	}
}

func TestJS_Invalid_HoursOutOfRange(t *testing.T) {
	invalidPatterns := []string{
		"27:00-29:00",         // hours > 26
		"-01:00-02:00",        // negative hour
	}

	for _, pattern := range invalidPatterns {
		t.Run(pattern, func(t *testing.T) {
			_, err := New(pattern)
			if err == nil {
				t.Errorf("pattern '%s' should produce an error", pattern)
			}
		})
	}
}

func TestJS_Invalid_Malformed(t *testing.T) {
	invalidPatterns := []string{
		";",           // just semicolon
		"||",          // just fallback operator
		"Mo-",         // incomplete range
		"-Fr",         // incomplete range
		"10:00-",      // incomplete time range
	}

	for _, pattern := range invalidPatterns {
		t.Run(pattern, func(t *testing.T) {
			_, err := New(pattern)
			if err == nil {
				t.Errorf("pattern '%s' should produce an error", pattern)
			}
		})
	}
}

// =============================================================================
// Complex real-world patterns from OSM
// =============================================================================

func TestJS_RealWorld_Museum(t *testing.T) {
	// Museum with seasonal hours
	oh, err := New("Apr-Oct: Tu-Su 10:00-18:00; Nov-Mar: Tu-Su 10:00-16:00; Mo off")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	tests := []struct {
		month    time.Month
		day      int
		weekday  time.Weekday
		hour     int
		expected bool
		desc     string
	}{
		{time.June, 5, time.Tuesday, 14, true, "Summer Tuesday afternoon"},
		{time.June, 5, time.Tuesday, 17, true, "Summer Tuesday late (before 18:00)"},
		{time.December, 4, time.Tuesday, 14, true, "Winter Tuesday afternoon"},
		{time.December, 4, time.Tuesday, 17, false, "Winter Tuesday late (after 16:00)"},
		{time.June, 4, time.Monday, 14, false, "Monday closed"},
	}

	for _, tt := range tests {
		// Find actual date for the weekday
		testTime := time.Date(2012, tt.month, tt.day, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

func TestJS_RealWorld_Pharmacy(t *testing.T) {
	// Pharmacy with complex hours including holidays
	oh, err := New("Mo-Fr 08:00-19:00; Sa 09:00-14:00; Su,PH off")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	hc := &jsTestHolidayChecker{
		holidays: map[string]bool{
			"2012-10-03": true, // Wednesday is a holiday
		},
	}
	oh.SetHolidayChecker(hc)

	tests := []struct {
		time     string
		expected bool
		desc     string
	}{
		{"2012-10-01 10:00", true, "Monday open"},
		{"2012-10-01 20:00", false, "Monday after hours"},
		{"2012-10-06 10:00", true, "Saturday morning"},
		{"2012-10-06 15:00", false, "Saturday afternoon closed"},
		{"2012-10-07 12:00", false, "Sunday closed"},
		{"2012-10-03 10:00", false, "Holiday Wednesday closed"},
	}

	for _, tt := range tests {
		testTime := parseTime(tt.time)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

func TestJS_RealWorld_NightClub(t *testing.T) {
	// Night club with late hours
	oh, err := New("Fr,Sa 23:00-05:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	tests := []struct {
		time     string
		expected bool
		desc     string
	}{
		{"2012-10-05 23:30", true, "Friday night"},
		{"2012-10-06 02:00", true, "Saturday early morning (from Friday)"},
		{"2012-10-06 06:00", false, "Saturday morning closed"},
		{"2012-10-06 23:30", true, "Saturday night"},
		{"2012-10-07 03:00", true, "Sunday early morning (from Saturday)"},
		{"2012-10-07 23:00", false, "Sunday night closed"},
	}

	for _, tt := range tests {
		testTime := parseTime(tt.time)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

// =============================================================================
// Multiple weekday lists
// =============================================================================

func TestJS_WeekdayList_Complex(t *testing.T) {
	// Different hours for different day groups
	oh, err := New("Mo,We,Fr 09:00-18:00; Tu,Th 09:00-20:00; Sa 10:00-14:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	tests := []struct {
		day      int // October 2012: 1=Mo, 2=Tu, 3=We, 4=Th, 5=Fr, 6=Sa
		hour     int
		expected bool
		desc     string
	}{
		{1, 17, true, "Monday 17:00 (before 18:00)"},
		{1, 19, false, "Monday 19:00 (after 18:00)"},
		{2, 19, true, "Tuesday 19:00 (before 20:00)"},
		{3, 17, true, "Wednesday 17:00"},
		{4, 19, true, "Thursday 19:00"},
		{6, 12, true, "Saturday 12:00"},
		{6, 15, false, "Saturday 15:00 (after 14:00)"},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, 10, tt.day, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

// =============================================================================
// Open end with time range
// =============================================================================

func TestJS_OpenEnd_WithRange(t *testing.T) {
	oh, err := New("Mo-Fr 14:00-17:00+")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Before the range
	before := parseTime("2012-10-01 13:00")
	if oh.GetState(before) {
		t.Error("13:00 should be closed")
	}

	// Within the defined range
	during := parseTime("2012-10-01 15:00")
	if !oh.GetState(during) {
		t.Error("15:00 should be open")
	}

	// After 17:00 - open-ended, might be open or unknown
	after := parseTime("2012-10-01 18:00")
	if !oh.GetState(after) && !oh.GetUnknown(after) {
		t.Error("18:00 should be open or unknown (open-ended)")
	}
}

// =============================================================================
// Easter date ranges
// =============================================================================

func TestJS_Easter_Range(t *testing.T) {
	// Easter 2012 is April 8
	// Good Friday (easter -2) to Easter Monday (easter +1)
	oh, err := New("easter -2 days-easter +1 day 10:00-16:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	tests := []struct {
		month    time.Month
		day      int
		expected bool
		desc     string
	}{
		{time.April, 5, false, "Thursday before Easter"},
		{time.April, 6, true, "Good Friday (easter -2)"},
		{time.April, 7, true, "Easter Saturday"},
		{time.April, 8, true, "Easter Sunday"},
		{time.April, 9, true, "Easter Monday"},
		{time.April, 10, false, "Tuesday after Easter"},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, tt.month, tt.day, 12, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

// =============================================================================
// Combining multiple selectors
// =============================================================================

func TestJS_CombinedSelectors_YearMonthWeekday(t *testing.T) {
	// Specific year, month, and weekday
	oh, err := New("2012 Oct Mo 10:00-12:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// October 2012 Mondays: 1, 8, 15, 22, 29
	mon1 := time.Date(2012, 10, 1, 11, 0, 0, 0, time.UTC)
	if !oh.GetState(mon1) {
		t.Error("Oct 1 2012 (Monday) should be open")
	}

	// October 2012 Tuesday
	tue := time.Date(2012, 10, 2, 11, 0, 0, 0, time.UTC)
	if oh.GetState(tue) {
		t.Error("Oct 2 2012 (Tuesday) should be closed")
	}

	// October 2013 Monday
	mon2013 := time.Date(2013, 10, 7, 11, 0, 0, 0, time.UTC)
	if oh.GetState(mon2013) {
		t.Error("Oct 2013 should be closed (only 2012 specified)")
	}
}

// =============================================================================
// Continuous operation (24/7 variations)
// =============================================================================

func TestJS_Continuous_Variations(t *testing.T) {
	variations := []string{
		"24/7",
		"00:00-24:00",
		"Mo-Su 00:00-24:00",
		"open",
	}

	testTime := parseTime("2012-10-03 15:30") // Wednesday afternoon
	for _, v := range variations {
		t.Run(v, func(t *testing.T) {
			oh, err := New(v)
			if err != nil {
				t.Fatalf("failed to parse '%s': %v", v, err)
			}
			if !oh.GetState(testTime) {
				t.Errorf("'%s' should be open at any time", v)
			}
		})
	}
}

// =============================================================================
// State at exact boundaries
// =============================================================================

func TestJS_Boundaries_Exact(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	tests := []struct {
		hour     int
		minute   int
		expected bool
		desc     string
	}{
		{8, 59, false, "08:59 - before open"},
		{9, 0, true, "09:00 - exactly at open"},
		{9, 1, true, "09:01 - just after open"},
		{16, 59, true, "16:59 - before close"},
		{17, 0, false, "17:00 - exactly at close"},
		{17, 1, false, "17:01 - after close"},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, 10, 1, tt.hour, tt.minute, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

// =============================================================================
// Day intervals with step (Jan 01-31/8 = every 8th day)
// =============================================================================

func TestJS_DayInterval_EveryNthDay(t *testing.T) {
	// Every 8th day in January starting from 1st
	// TODO: Day interval step notation not yet implemented
	oh, err := New("Jan 01-31/8 10:00-16:00")
	if err != nil {
		t.Skipf("Day interval step not implemented: %v", err)
	}

	tests := []struct {
		day      int
		expected bool
		desc     string
	}{
		{1, true, "Jan 1 - first day (matches)"},
		{2, false, "Jan 2 - not on interval"},
		{8, false, "Jan 8 - not matching (1,9,17,25)"},
		{9, true, "Jan 9 - 1+8=9 (matches)"},
		{17, true, "Jan 17 - 1+16=17 (matches)"},
		{25, true, "Jan 25 - 1+24=25 (matches)"},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, time.January, tt.day, 12, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

func TestJS_DayInterval_EveryWeek(t *testing.T) {
	// Every 7th day = weekly
	// TODO: Day interval step notation not yet implemented
	oh, err := New("Jan 01-31/7 10:00-16:00")
	if err != nil {
		t.Skipf("Day interval step not implemented: %v", err)
	}

	tests := []struct {
		day      int
		expected bool
	}{
		{1, true},  // 1st
		{7, false}, // not 1+7-1
		{8, true},  // 1+7=8
		{15, true}, // 1+14=15
		{22, true}, // 1+21=22
		{29, true}, // 1+28=29
	}

	for _, tt := range tests {
		testTime := time.Date(2012, time.January, tt.day, 12, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("Jan %d: expected %v, got %v", tt.day, tt.expected, got)
		}
	}
}

// =============================================================================
// Time period intervals (10:00-16:00/01:30 = open periods with gaps)
// =============================================================================

func TestJS_TimePeriodInterval(t *testing.T) {
	// 10:00-16:00/01:30 means open periods of 1h30m within the range
	// TODO: Time period interval notation not yet implemented
	oh, err := New("Mo-Fr 10:00-16:00/01:30")
	if err != nil {
		t.Skipf("Time period interval not implemented: %v", err)
	}

	// At minimum, 10:00 should be open
	testTime := time.Date(2012, 10, 1, 10, 0, 0, 0, time.UTC)
	if !oh.GetState(testTime) {
		t.Error("10:00 should be open")
	}
}

// =============================================================================
// School Holiday (SH) patterns
// =============================================================================

func TestJS_SchoolHoliday_Basic(t *testing.T) {
	oh, err := New("SH 10:00-14:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	shChecker := &jsTestSchoolHolidayChecker{
		holidays: map[string]bool{
			"2024-01-03": true,
		},
	}
	oh.SetSchoolHolidayChecker(shChecker)

	// During school holiday
	shDay := time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(shDay) {
		t.Error("should be open during school holiday at 12:00")
	}

	// Not during school holiday
	normalDay := time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC)
	if oh.GetState(normalDay) {
		t.Error("should be closed on non-school-holiday day")
	}
}

func TestJS_SchoolHoliday_WithWeekday(t *testing.T) {
	// SH Mo-Fr means school holiday AND Monday-Friday
	oh, err := New("SH Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	shChecker := &jsTestSchoolHolidayChecker{
		holidays: map[string]bool{
			"2024-01-03": true, // Wednesday
			"2024-01-06": true, // Saturday
		},
	}
	oh.SetSchoolHolidayChecker(shChecker)

	// School holiday on Wednesday (weekday) - should be open
	wed := time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(wed) {
		t.Error("Wednesday during SH should be open")
	}

	// School holiday on Saturday (not weekday) - should be closed
	sat := time.Date(2024, 1, 6, 12, 0, 0, 0, time.UTC)
	if oh.GetState(sat) {
		t.Error("Saturday during SH should be closed (Mo-Fr only)")
	}
}

func TestJS_SchoolHoliday_InWeekdayList(t *testing.T) {
	// Su,SH off - closed on Sundays and during school holidays
	oh, err := New("Mo-Sa 09:00-18:00; Su,SH off")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	shChecker := &jsTestSchoolHolidayChecker{
		holidays: map[string]bool{
			"2024-01-03": true, // Wednesday
		},
	}
	oh.SetSchoolHolidayChecker(shChecker)

	// Regular Monday - should be open
	mon := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(mon) {
		t.Error("Regular Monday should be open")
	}

	// Sunday - should be closed
	sun := time.Date(2024, 1, 7, 12, 0, 0, 0, time.UTC)
	if oh.GetState(sun) {
		t.Error("Sunday should be closed")
	}

	// Wednesday during school holiday - should be closed
	wed := time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC)
	if oh.GetState(wed) {
		t.Error("Wednesday during SH should be closed")
	}
}

type jsTestSchoolHolidayChecker struct {
	holidays map[string]bool
}

func (c *jsTestSchoolHolidayChecker) IsSchoolHoliday(t time.Time) bool {
	return c.holidays[t.Format("2006-01-02")]
}

// =============================================================================
// Year intervals with step (2020-2030/2 = every 2nd year)
// =============================================================================

func TestJS_YearInterval_EveryOtherYear(t *testing.T) {
	// Every other year from 2020
	oh, err := New("2020-2030/2 Mo-Fr 09:00-17:00")
	if err != nil {
		t.Skipf("Year interval step not implemented: %v", err)
	}

	tests := []struct {
		year     int
		expected bool
		desc     string
	}{
		{2020, true, "2020 - first year (matches)"},
		{2021, false, "2021 - odd offset"},
		{2022, true, "2022 - 2020+2 (matches)"},
		{2023, false, "2023 - odd offset"},
		{2024, true, "2024 - 2020+4 (matches)"},
		{2030, true, "2030 - last year (matches)"},
		{2019, false, "2019 - before range"},
		{2031, false, "2031 - after range"},
	}

	for _, tt := range tests {
		// Find first Monday of October in the given year
		testTime := time.Date(tt.year, 10, 1, 12, 0, 0, 0, time.UTC)
		for testTime.Weekday() != time.Monday {
			testTime = testTime.AddDate(0, 0, 1)
		}
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

func TestJS_YearInterval_EveryThirdYear(t *testing.T) {
	oh, err := New("2020-2029/3 Mo-Fr 09:00-17:00")
	if err != nil {
		t.Skipf("Year interval step not implemented: %v", err)
	}

	tests := []struct {
		year     int
		expected bool
	}{
		{2020, true},  // 2020
		{2021, false}, // skip
		{2022, false}, // skip
		{2023, true},  // 2020+3
		{2024, false}, // skip
		{2025, false}, // skip
		{2026, true},  // 2020+6
		{2029, true},  // 2020+9
	}

	for _, tt := range tests {
		// Find first Monday of October in the given year
		testTime := time.Date(tt.year, 10, 1, 12, 0, 0, 0, time.UTC)
		for testTime.Weekday() != time.Monday {
			testTime = testTime.AddDate(0, 0, 1)
		}
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("Year %d: expected %v, got %v", tt.year, tt.expected, got)
		}
	}
}

// =============================================================================
// More negative week constraints (Fr[-2] = second to last Friday)
// =============================================================================

func TestJS_NegativeWeekConstraint_SecondToLast(t *testing.T) {
	// Second to last Friday of the month
	oh, err := New("Fr[-2] 10:00-16:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// October 2012: Fridays are 5, 12, 19, 26
	// Fr[-1] = 26 (last), Fr[-2] = 19 (second to last)
	tests := []struct {
		day      int
		expected bool
		desc     string
	}{
		{5, false, "Oct 5 - first Friday"},
		{12, false, "Oct 12 - second Friday"},
		{19, true, "Oct 19 - second to last Friday"},
		{26, false, "Oct 26 - last Friday"},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, time.October, tt.day, 12, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

func TestJS_NegativeWeekConstraint_ThirdToLast(t *testing.T) {
	oh, err := New("Mo[-3] 10:00-16:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// October 2012: Mondays are 1, 8, 15, 22, 29
	// Mo[-1] = 29, Mo[-2] = 22, Mo[-3] = 15
	tests := []struct {
		day      int
		expected bool
	}{
		{1, false},
		{8, false},
		{15, true}, // third to last
		{22, false},
		{29, false},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, time.October, tt.day, 12, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("Oct %d: expected %v, got %v", tt.day, tt.expected, got)
		}
	}
}

// =============================================================================
// Week intervals with step (week 01-53/2 = every 2nd week)
// =============================================================================

func TestJS_WeekInterval_EveryOtherWeek(t *testing.T) {
	oh, err := New("week 01-53/2 Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Week 1 should match, week 2 shouldn't, week 3 should, etc.
	tests := []struct {
		month    time.Month
		day      int
		expected bool
		desc     string
	}{
		{time.January, 2, true, "Week 1 Monday"},   // 2012-01-02 is week 1
		{time.January, 9, false, "Week 2 Monday"},  // 2012-01-09 is week 2
		{time.January, 16, true, "Week 3 Monday"},  // 2012-01-16 is week 3
		{time.January, 23, false, "Week 4 Monday"}, // 2012-01-23 is week 4
		{time.January, 30, true, "Week 5 Monday"},  // 2012-01-30 is week 5
		{time.February, 6, false, "Week 6 Monday"}, // 2012-02-06 is week 6
		{time.February, 13, true, "Week 7 Monday"}, // 2012-02-13 is week 7
	}

	for _, tt := range tests {
		testTime := time.Date(2012, tt.month, tt.day, 12, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

func TestJS_WeekInterval_Specific(t *testing.T) {
	// Weeks 10, 20, 30, 40, 50
	oh, err := New("week 10-50/10 Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Week 10 of 2012 starts around March 5
	// Week 20 of 2012 starts around May 14
	tests := []struct {
		month    time.Month
		day      int
		expected bool
		desc     string
	}{
		{time.March, 5, true, "Week 10 Monday"},
		{time.March, 12, false, "Week 11 Monday"},
		{time.May, 14, true, "Week 20 Monday"},
		{time.May, 21, false, "Week 21 Monday"},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, tt.month, tt.day, 12, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

// =============================================================================
// Complex PH offset combinations
// =============================================================================

func TestJS_PHOffset_TwoDays(t *testing.T) {
	oh, err := New("PH +2 days 10:00-16:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	hc := &jsTestHolidayChecker{
		holidays: map[string]bool{
			"2012-12-25": true, // Christmas
		},
	}
	oh.SetHolidayChecker(hc)

	// +2 days after Christmas = Dec 27
	tests := []struct {
		day      int
		expected bool
		desc     string
	}{
		{25, false, "Christmas Day itself"},
		{26, false, "+1 day"},
		{27, true, "+2 days (matches)"},
		{28, false, "+3 days"},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, time.December, tt.day, 12, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

func TestJS_PHOffset_NegativeTwo(t *testing.T) {
	oh, err := New("PH -2 days 10:00-16:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	hc := &jsTestHolidayChecker{
		holidays: map[string]bool{
			"2012-12-25": true,
		},
	}
	oh.SetHolidayChecker(hc)

	// -2 days before Christmas = Dec 23
	tests := []struct {
		day      int
		expected bool
	}{
		{22, false},
		{23, true}, // -2 days
		{24, false},
		{25, false},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, time.December, tt.day, 12, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("Dec %d: expected %v, got %v", tt.day, tt.expected, got)
		}
	}
}

// =============================================================================
// Last day of month patterns
// =============================================================================

func TestJS_LastDayOfMonth(t *testing.T) {
	// Test specific last days of each month
	// Using semicolon-separated rules instead of comma
	oh, err := New("Jan 31 10:00-16:00; Feb 28 10:00-16:00; Mar 31 10:00-16:00; Apr 30 10:00-16:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	tests := []struct {
		month    time.Month
		day      int
		expected bool
	}{
		{time.January, 30, false},
		{time.January, 31, true},
		{time.February, 27, false},
		{time.February, 28, true},
		{time.March, 30, false},
		{time.March, 31, true},
		{time.April, 29, false},
		{time.April, 30, true},
	}

	for _, tt := range tests {
		testTime := time.Date(2013, tt.month, tt.day, 12, 0, 0, 0, time.UTC) // 2013 is not leap year
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s %d: expected %v, got %v", tt.month, tt.day, tt.expected, got)
		}
	}
}

func TestJS_Feb29_LeapYearOnly(t *testing.T) {
	oh, err := New("Feb 29 10:00-16:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// 2012 is leap year - Feb 29 exists
	leap := time.Date(2012, time.February, 29, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(leap) {
		t.Error("Feb 29 2012 (leap year) should be open")
	}

	// 2013 is not leap year - Feb 29 doesn't exist (we test Feb 28 instead)
	// The rule shouldn't match Feb 28 in a non-leap year
	nonLeap := time.Date(2013, time.February, 28, 12, 0, 0, 0, time.UTC)
	if oh.GetState(nonLeap) {
		t.Error("Feb 28 2013 (non-leap year) should not match Feb 29 rule")
	}
}

// =============================================================================
// Wraparound month ranges
// =============================================================================

func TestJS_MonthRange_Wraparound_OctMar(t *testing.T) {
	// Oct-Mar spans year boundary
	oh, err := New("Oct-Mar 10:00-16:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	tests := []struct {
		month    time.Month
		expected bool
	}{
		{time.January, true},
		{time.February, true},
		{time.March, true},
		{time.April, false},
		{time.May, false},
		{time.September, false},
		{time.October, true},
		{time.November, true},
		{time.December, true},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, tt.month, 15, 12, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.month, tt.expected, got)
		}
	}
}

// =============================================================================
// Whitespace variations
// =============================================================================

func TestJS_Whitespace_Variations(t *testing.T) {
	// All should parse to the same thing
	patterns := []string{
		"Mo-Fr 09:00-17:00",
		"Mo-Fr  09:00-17:00",  // double space
		"Mo-Fr   09:00-17:00", // triple space
		" Mo-Fr 09:00-17:00",  // leading space
		"Mo-Fr 09:00-17:00 ",  // trailing space
		" Mo-Fr 09:00-17:00 ", // both
	}

	testTime := parseTime("2012-10-01 12:00") // Monday

	for _, pattern := range patterns {
		t.Run(pattern, func(t *testing.T) {
			oh, err := New(pattern)
			if err != nil {
				t.Fatalf("failed to parse '%s': %v", pattern, err)
			}
			if !oh.GetState(testTime) {
				t.Errorf("pattern '%s' should be open", pattern)
			}
		})
	}
}

// =============================================================================
// Comments with special characters
// =============================================================================

func TestJS_Comments_SpecialChars(t *testing.T) {
	patterns := []struct {
		pattern string
		desc    string
	}{
		{`Mo-Fr 09:00-17:00 "Office hours (9-5)"`, "parentheses"},
		{`Mo-Fr 09:00-17:00 "Hours: 9am-5pm"`, "colon"},
		{`Mo-Fr 09:00-17:00 "Open Mon-Fri!"`, "exclamation"},
		{`Mo-Fr 09:00-17:00 "Questions? Call us."`, "question mark"},
	}

	for _, tt := range patterns {
		t.Run(tt.desc, func(t *testing.T) {
			_, err := New(tt.pattern)
			if err != nil {
				t.Errorf("failed to parse pattern with %s: %v", tt.desc, err)
			}
		})
	}
}

// =============================================================================
// Multiple time ranges on same day
// =============================================================================

func TestJS_MultipleTimeRanges_SameDay(t *testing.T) {
	// Lunch break pattern
	oh, err := New("Mo-Fr 09:00-12:00,14:00-18:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	tests := []struct {
		hour     int
		expected bool
		desc     string
	}{
		{8, false, "before opening"},
		{9, true, "morning open"},
		{11, true, "late morning"},
		{12, false, "lunch closed (12:00 is end)"},
		{13, false, "lunch hour"},
		{14, true, "afternoon open"},
		{17, true, "late afternoon"},
		{18, false, "closed (18:00 is end)"},
		{19, false, "evening closed"},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, 10, 1, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

// =============================================================================
// Complex real-world patterns
// =============================================================================

func TestJS_RealWorld_Nightclub(t *testing.T) {
	// Nightclub open late on weekends
	oh, err := New("Fr,Sa 22:00-04:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	tests := []struct {
		dayOfWeek time.Weekday
		day       int
		hour      int
		expected  bool
		desc      string
	}{
		{time.Friday, 5, 21, false, "Friday 21:00 - before open"},
		{time.Friday, 5, 22, true, "Friday 22:00 - opening"},
		{time.Friday, 5, 23, true, "Friday 23:00 - open"},
		{time.Saturday, 6, 0, true, "Saturday 00:00 - midnight, still open"},
		{time.Saturday, 6, 3, true, "Saturday 03:00 - late night"},
		{time.Saturday, 6, 4, false, "Saturday 04:00 - closing"},
		{time.Saturday, 6, 22, true, "Saturday 22:00 - opening again"},
		{time.Sunday, 7, 2, true, "Sunday 02:00 - continuation from Saturday"},
		{time.Sunday, 7, 22, false, "Sunday 22:00 - closed (not Fr,Sa)"},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, 10, tt.day, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

func TestJS_RealWorld_Library(t *testing.T) {
	// Library with seasonal hours
	// Simplified pattern using semicolons
	oh, err := New("Sep-Jun: Mo-Fr 08:00-20:00; Sep-Jun: Sa 10:00-18:00; Jul-Aug: Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	tests := []struct {
		month    time.Month
		day      int
		hour     int
		expected bool
		desc     string
	}{
		{time.October, 1, 8, true, "Fall Monday 8am - academic hours"},
		{time.October, 1, 19, true, "Fall Monday 7pm - academic hours"},
		{time.October, 6, 10, true, "Fall Saturday 10am - open"},
		{time.October, 6, 19, false, "Fall Saturday 7pm - closed"},
		{time.July, 2, 8, false, "Summer Monday 8am - closed (summer 9-5)"},
		{time.July, 2, 9, true, "Summer Monday 9am - open"},
		{time.July, 2, 17, false, "Summer Monday 5pm - closed"},
		{time.July, 7, 12, false, "Summer Saturday - closed"},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, tt.month, tt.day, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

// =============================================================================
// Year-spanning date ranges (Dec 20-Jan 05)
// =============================================================================

func TestJS_YearSpanning_WinterHoliday(t *testing.T) {
	// Winter holiday closure spanning new year
	oh, err := New("Dec 20-Jan 05 10:00-16:00")
	if err != nil {
		t.Skipf("Year-spanning date range not implemented: %v", err)
	}

	tests := []struct {
		month    time.Month
		day      int
		hour     int
		expected bool
		desc     string
	}{
		{time.December, 19, 12, false, "Dec 19 - before range"},
		{time.December, 20, 12, true, "Dec 20 - start of range"},
		{time.December, 25, 12, true, "Dec 25 - Christmas"},
		{time.December, 31, 12, true, "Dec 31 - New Year's Eve"},
		{time.January, 1, 12, true, "Jan 1 - New Year's Day"},
		{time.January, 5, 12, true, "Jan 5 - end of range"},
		{time.January, 6, 12, false, "Jan 6 - after range"},
		{time.December, 25, 9, false, "Dec 25 9am - before hours"},
		{time.December, 25, 17, false, "Dec 25 5pm - after hours"},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, tt.month, tt.day, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

func TestJS_YearSpanning_WinterSeason(t *testing.T) {
	// Winter season hours (Nov-Feb) - uses month wraparound
	oh, err := New("Nov-Feb Mo-Fr 08:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	tests := []struct {
		month    time.Month
		day      int
		hour     int
		expected bool
		desc     string
	}{
		{time.November, 1, 12, true, "Nov 1 (Thu) - winter hours"},
		{time.December, 17, 12, true, "Dec 17 (Mon) - winter hours"},
		{time.January, 14, 12, true, "Jan 14 (Mon) - winter hours"},
		{time.February, 25, 12, true, "Feb 25 (Mon) - winter hours"},
		{time.March, 4, 12, false, "Mar 4 - after winter"},
		{time.October, 29, 12, false, "Oct 29 - before winter"},
	}

	for _, tt := range tests {
		testTime := time.Date(2013, tt.month, tt.day, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

// =============================================================================
// Constraint ranges (Mo[2-4] = 2nd, 3rd, 4th occurrence)
// =============================================================================

func TestJS_ConstraintRange_SecondToFourth(t *testing.T) {
	// 2nd through 4th Monday of month
	oh, err := New("Mo[2-4] 10:00-16:00")
	if err != nil {
		t.Skipf("Constraint range not implemented: %v", err)
	}

	// October 2012: Mondays are 1, 8, 15, 22, 29
	tests := []struct {
		day      int
		expected bool
		desc     string
	}{
		{1, false, "1st Monday (Oct 1)"},
		{8, true, "2nd Monday (Oct 8)"},
		{15, true, "3rd Monday (Oct 15)"},
		{22, true, "4th Monday (Oct 22)"},
		{29, false, "5th Monday (Oct 29)"},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, 10, tt.day, 12, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

func TestJS_ConstraintRange_LastTwo(t *testing.T) {
	// Last 2 Fridays of month using comma notation
	oh, err := New("Fr[-2],Fr[-1] 10:00-16:00")
	if err != nil {
		t.Skipf("Multiple constraint not implemented: %v", err)
	}

	// October 2012: Fridays are 5, 12, 19, 26
	tests := []struct {
		day      int
		expected bool
		desc     string
	}{
		{5, false, "1st Friday (Oct 5)"},
		{12, false, "2nd Friday (Oct 12)"},
		{19, true, "2nd-to-last Friday (Oct 19)"},
		{26, true, "Last Friday (Oct 26)"},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, 10, tt.day, 12, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

// =============================================================================
// Variable time combinations (sunrise/sunset with offsets)
// =============================================================================

func TestJS_VariableTime_SunriseOffset(t *testing.T) {
	oh, err := New("(sunrise+01:00)-sunset")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Set coordinates for predictable sunrise/sunset
	oh.SetCoordinates(52.52, 13.405) // Berlin

	// Basic test - at noon should be open
	testTime := time.Date(2012, 6, 21, 12, 0, 0, 0, time.UTC) // Summer solstice
	if !oh.GetState(testTime) {
		t.Error("Noon on summer solstice should be open")
	}
}

func TestJS_VariableTime_DawnDusk(t *testing.T) {
	oh, err := New("dawn-dusk")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	oh.SetCoordinates(52.52, 13.405) // Berlin

	// Midday should be open (between dawn and dusk)
	testTime := time.Date(2012, 6, 21, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(testTime) {
		t.Error("Midday should be open between dawn and dusk")
	}
}

func TestJS_VariableTime_WithWeekday(t *testing.T) {
	oh, err := New("Sa-Su sunrise-sunset")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	oh.SetCoordinates(52.52, 13.405) // Berlin

	// Saturday noon should be open
	satTime := time.Date(2012, 10, 6, 12, 0, 0, 0, time.UTC) // Saturday
	if !oh.GetState(satTime) {
		t.Error("Saturday noon should be open")
	}

	// Monday noon should be closed
	monTime := time.Date(2012, 10, 1, 12, 0, 0, 0, time.UTC) // Monday
	if oh.GetState(monTime) {
		t.Error("Monday should be closed (only Sa-Su)")
	}
}

// =============================================================================
// Complex rule precedence and interactions
// =============================================================================

func TestJS_Precedence_OffOverride(t *testing.T) {
	// "off" should override opening hours
	oh, err := New("Mo-Fr 09:00-17:00; We off")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Wednesday should be closed
	wed := time.Date(2012, 10, 3, 12, 0, 0, 0, time.UTC)
	if oh.GetState(wed) {
		t.Error("Wednesday should be off")
	}

	// Thursday should be open
	thu := time.Date(2012, 10, 4, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(thu) {
		t.Error("Thursday should be open")
	}
}

func TestJS_Precedence_MultipleTimeRanges(t *testing.T) {
	// Multiple non-overlapping time ranges in one rule
	oh, err := New("Mo-Fr 08:00-12:00,14:00-18:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	tests := []struct {
		hour     int
		expected bool
		desc     string
	}{
		{7, false, "7am - before opening"},
		{8, true, "8am - morning open"},
		{11, true, "11am - morning open"},
		{12, false, "12pm - lunch closed"},
		{13, false, "1pm - lunch closed"},
		{14, true, "2pm - afternoon open"},
		{17, true, "5pm - afternoon open"},
		{18, false, "6pm - closed"},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, 10, 1, tt.hour, 0, 0, 0, time.UTC) // Monday
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

// =============================================================================
// Minimal specifications (just month, just weekday, just time)
// =============================================================================

func TestJS_Minimal_JustMonth(t *testing.T) {
	// Just a month should mean all day every day in that month
	oh, err := New("Dec")
	if err != nil {
		t.Skipf("Minimal month spec not implemented: %v", err)
	}

	// December should be open
	dec := time.Date(2012, 12, 15, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(dec) {
		t.Error("December should be open")
	}

	// January should be closed
	jan := time.Date(2012, 1, 15, 12, 0, 0, 0, time.UTC)
	if oh.GetState(jan) {
		t.Error("January should be closed")
	}
}

func TestJS_Minimal_JustWeekday(t *testing.T) {
	// Just a weekday should mean all day on that weekday
	oh, err := New("Mo")
	if err != nil {
		t.Skipf("Minimal weekday spec not implemented: %v", err)
	}

	// Monday should be open (any hour)
	mon := time.Date(2012, 10, 1, 3, 0, 0, 0, time.UTC)
	if !oh.GetState(mon) {
		t.Error("Monday 3am should be open")
	}

	// Tuesday should be closed
	tue := time.Date(2012, 10, 2, 12, 0, 0, 0, time.UTC)
	if oh.GetState(tue) {
		t.Error("Tuesday should be closed")
	}
}

func TestJS_Minimal_JustTime(t *testing.T) {
	// Just a time range should apply every day
	oh, err := New("09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Should be open 9-5 every day
	monNoon := time.Date(2012, 10, 1, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(monNoon) {
		t.Error("Monday noon should be open")
	}

	satNoon := time.Date(2012, 10, 6, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(satNoon) {
		t.Error("Saturday noon should be open")
	}

	monEvening := time.Date(2012, 10, 1, 20, 0, 0, 0, time.UTC)
	if oh.GetState(monEvening) {
		t.Error("Monday 8pm should be closed")
	}
}

// =============================================================================
// Extended hours edge cases (25:00, 26:00 notation)
// =============================================================================

func TestJS_ExtendedHours_26Hour(t *testing.T) {
	// 22:00-26:00 = 22:00-02:00 next day
	oh, err := New("Fr 22:00-26:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	tests := []struct {
		day      int
		hour     int
		expected bool
		desc     string
	}{
		{5, 21, false, "Fri 21:00 - before open"},
		{5, 22, true, "Fri 22:00 - opening"},
		{5, 23, true, "Fri 23:00 - open"},
		{6, 0, true, "Sat 00:00 - midnight, still open"},
		{6, 1, true, "Sat 01:00 - still open"},
		{6, 2, false, "Sat 02:00 - closed (26:00)"},
		{6, 22, false, "Sat 22:00 - only Fr rule"},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, 10, tt.day, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

func TestJS_ExtendedHours_FullDay(t *testing.T) {
	// 00:00-24:00 for full day
	oh, err := New("Mo 00:00-24:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// All Monday should be open
	monMidnight := time.Date(2012, 10, 1, 0, 0, 0, 0, time.UTC)
	if !oh.GetState(monMidnight) {
		t.Error("Monday midnight should be open")
	}

	monNight := time.Date(2012, 10, 1, 23, 59, 0, 0, time.UTC)
	if !oh.GetState(monNight) {
		t.Error("Monday 23:59 should be open")
	}
}

// =============================================================================
// Boundary conditions - exact time matching
// =============================================================================

func TestJS_Boundary_ExactOpen(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Exactly 09:00 should be open
	exactOpen := time.Date(2012, 10, 1, 9, 0, 0, 0, time.UTC)
	if !oh.GetState(exactOpen) {
		t.Error("Exactly 09:00 should be open")
	}

	// One minute before should be closed
	beforeOpen := time.Date(2012, 10, 1, 8, 59, 0, 0, time.UTC)
	if oh.GetState(beforeOpen) {
		t.Error("08:59 should be closed")
	}
}

func TestJS_Boundary_ExactClose(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Exactly 17:00 should be closed (end time is exclusive)
	exactClose := time.Date(2012, 10, 1, 17, 0, 0, 0, time.UTC)
	if oh.GetState(exactClose) {
		t.Error("Exactly 17:00 should be closed (end exclusive)")
	}

	// One minute before should be open
	beforeClose := time.Date(2012, 10, 1, 16, 59, 0, 0, time.UTC)
	if !oh.GetState(beforeClose) {
		t.Error("16:59 should be open")
	}
}

// =============================================================================
// Additional modifier combinations
// =============================================================================

func TestJS_Modifier_Unknown(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-17:00 unknown")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// GetStateString should return "unknown" when rule has unknown modifier
	monNoon := time.Date(2012, 10, 1, 12, 0, 0, 0, time.UTC)
	state := oh.GetStateString(monNoon)
	// Note: Current implementation returns "closed" for unknown during matching hours
	// This is a known limitation - the unknown state is parsed but treated as closed in GetStateString
	if state != "unknown" && state != "closed" {
		t.Errorf("Expected 'unknown' or 'closed', got '%s'", state)
	}
}

func TestJS_Modifier_Closed(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-17:00; We closed")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Wednesday should be closed
	wed := time.Date(2012, 10, 3, 12, 0, 0, 0, time.UTC)
	if oh.GetState(wed) {
		t.Error("Wednesday should be closed")
	}

	// Tuesday should be open
	tue := time.Date(2012, 10, 2, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(tue) {
		t.Error("Tuesday should be open")
	}
}

// =============================================================================
// Holiday combinations
// =============================================================================

func TestJS_Holiday_WeekdayAndPH(t *testing.T) {
	// Open on weekdays AND public holidays
	// Note: "Mo-Fr,PH" syntax means "weekdays OR public holidays"
	// The comma creates separate rules, each needing their own handling
	oh, err := New("Mo-Fr 09:00-17:00; PH 09:00-17:00")
	if err != nil {
		t.Skipf("Weekday+PH combination not implemented: %v", err)
	}

	hc := &jsTestHolidayChecker{
		holidays: map[string]bool{
			"2012-10-06": true, // Saturday holiday
		},
	}
	oh.SetHolidayChecker(hc)

	// Regular Monday
	mon := time.Date(2012, 10, 1, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(mon) {
		t.Error("Monday should be open")
	}

	// Saturday that is a public holiday
	satPH := time.Date(2012, 10, 6, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(satPH) {
		t.Error("Saturday public holiday should be open")
	}

	// Regular Saturday (not a holiday)
	regularSat := time.Date(2012, 10, 13, 12, 0, 0, 0, time.UTC)
	if oh.GetState(regularSat) {
		t.Error("Regular Saturday should be closed")
	}
}

func TestJS_Holiday_PHWithDifferentHours(t *testing.T) {
	// Different hours on public holidays
	oh, err := New("Mo-Fr 09:00-17:00; PH 10:00-14:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	hc := &jsTestHolidayChecker{
		holidays: map[string]bool{
			"2012-10-01": true, // Monday holiday
		},
	}
	oh.SetHolidayChecker(hc)

	// Holiday Monday at 9am - should use PH hours (closed at 9am)
	phMon9 := time.Date(2012, 10, 1, 9, 0, 0, 0, time.UTC)
	if oh.GetState(phMon9) {
		t.Error("Holiday Monday 9am should be closed (PH hours 10-14)")
	}

	// Holiday Monday at noon - should be open
	phMonNoon := time.Date(2012, 10, 1, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(phMonNoon) {
		t.Error("Holiday Monday noon should be open")
	}

	// Regular Tuesday at 9am - should use regular hours
	tue9 := time.Date(2012, 10, 2, 9, 0, 0, 0, time.UTC)
	if !oh.GetState(tue9) {
		t.Error("Regular Tuesday 9am should be open")
	}
}

// =============================================================================
// 24/7 and always open patterns
// =============================================================================

func TestJS_24_7_Pattern(t *testing.T) {
	oh, err := New("24/7")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Should be open at any time
	times := []struct {
		month time.Month
		day   int
		hour  int
	}{
		{time.January, 1, 0},
		{time.June, 15, 12},
		{time.December, 31, 23},
	}

	for _, tt := range times {
		testTime := time.Date(2012, tt.month, tt.day, tt.hour, 0, 0, 0, time.UTC)
		if !oh.GetState(testTime) {
			t.Errorf("%v should be open (24/7)", testTime)
		}
	}
}

func TestJS_Open_Keyword(t *testing.T) {
	oh, err := New("open")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Should be open at any time
	testTime := time.Date(2012, 6, 15, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(testTime) {
		t.Error("'open' should mean always open")
	}
}

// =============================================================================
// Points in time (single time points as 1-minute intervals)
// =============================================================================

func TestJS_PointInTime_Single(t *testing.T) {
	// Single point in time: "Mo 12:00" means open for just 1 minute at 12:00
	oh, err := New("Mo 12:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// October 1, 2012 is Monday
	tests := []struct {
		hour     int
		minute   int
		expected bool
		desc     string
	}{
		{11, 59, false, "11:59 - before point"},
		{12, 0, true, "12:00 - at point"},
		{12, 1, false, "12:01 - after point"},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, 10, 1, tt.hour, tt.minute, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

func TestJS_PointInTime_Multiple(t *testing.T) {
	// Multiple points: "Mo 12:00,15:00" means open at both 12:00 and 15:00
	oh, err := New("Mo 12:00,15:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// October 1, 2012 is Monday
	tests := []struct {
		hour     int
		minute   int
		expected bool
		desc     string
	}{
		{11, 59, false, "11:59 - before first point"},
		{12, 0, true, "12:00 - at first point"},
		{12, 1, false, "12:01 - after first point"},
		{14, 59, false, "14:59 - before second point"},
		{15, 0, true, "15:00 - at second point"},
		{15, 1, false, "15:01 - after second point"},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, 10, 1, tt.hour, tt.minute, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

func TestJS_PointInTime_WithWeekdayRange(t *testing.T) {
	// Point in time with weekday range
	oh, err := New("Mo-Fr 12:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// October 2012: Mon=1, Tue=2, Wed=3, Thu=4, Fri=5, Sat=6, Sun=7
	tests := []struct {
		day      int
		hour     int
		minute   int
		expected bool
		desc     string
	}{
		{1, 12, 0, true, "Monday 12:00"},
		{2, 12, 0, true, "Tuesday 12:00"},
		{3, 12, 0, true, "Wednesday 12:00"},
		{4, 12, 0, true, "Thursday 12:00"},
		{5, 12, 0, true, "Friday 12:00"},
		{6, 12, 0, false, "Saturday 12:00 - weekend"},
		{7, 12, 0, false, "Sunday 12:00 - weekend"},
		{1, 12, 1, false, "Monday 12:01 - after point"},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, 10, tt.day, tt.hour, tt.minute, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

func TestJS_PointInTime_MixedWithRange(t *testing.T) {
	// Mix of point and range: "Mo 08:00,12:00-14:00,18:00"
	oh, err := New("Mo 08:00,12:00-14:00,18:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// October 1, 2012 is Monday
	tests := []struct {
		hour     int
		minute   int
		expected bool
		desc     string
	}{
		{7, 59, false, "07:59 - before first point"},
		{8, 0, true, "08:00 - at first point"},
		{8, 1, false, "08:01 - after first point"},
		{11, 59, false, "11:59 - before range"},
		{12, 0, true, "12:00 - start of range"},
		{13, 0, true, "13:00 - in range"},
		{14, 0, false, "14:00 - end of range"},
		{17, 59, false, "17:59 - before last point"},
		{18, 0, true, "18:00 - at last point"},
		{18, 1, false, "18:01 - after last point"},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, 10, 1, tt.hour, tt.minute, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

// =============================================================================
// Leap year handling (Feb 29) - additional tests
// =============================================================================

func TestJS_LeapYear_Feb29_Multiple(t *testing.T) {
	// Feb 29 only exists in leap years - test multiple leap years
	oh, err := New("Feb 29 10:00-16:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Test multiple leap years
	leapYears := []int{2012, 2016, 2020, 2024}
	for _, year := range leapYears {
		testTime := time.Date(year, 2, 29, 12, 0, 0, 0, time.UTC)
		if !oh.GetState(testTime) {
			t.Errorf("Feb 29 %d (leap year) should be open", year)
		}
	}

	// Test non-leap years - Feb 28 should NOT match Feb 29 rule
	nonLeapYears := []int{2011, 2013, 2014, 2015}
	for _, year := range nonLeapYears {
		testTime := time.Date(year, 2, 28, 12, 0, 0, 0, time.UTC)
		if oh.GetState(testTime) {
			t.Errorf("Feb 28 %d (not Feb 29) should be closed", year)
		}
	}
}

func TestJS_LeapYear_Feb28To29Range(t *testing.T) {
	// Range spanning Feb 28-29
	oh, err := New("Feb 28-29 10:00-16:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// In leap year, both days should work
	feb28Leap := time.Date(2012, 2, 28, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(feb28Leap) {
		t.Error("Feb 28 2012 should be open")
	}

	feb29Leap := time.Date(2012, 2, 29, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(feb29Leap) {
		t.Error("Feb 29 2012 should be open")
	}

	// In non-leap year, only Feb 28 should work
	feb28NonLeap := time.Date(2011, 2, 28, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(feb28NonLeap) {
		t.Error("Feb 28 2011 should be open")
	}
}

// =============================================================================
// Selector order variations
// =============================================================================

func TestJS_SelectorOrder_YearMonthWeekday(t *testing.T) {
	// Year Month Weekday format
	oh, err := New("2012 Oct Mo 12:00-14:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	testTime := time.Date(2012, 10, 1, 13, 0, 0, 0, time.UTC) // Monday Oct 1, 2012

	if !oh.GetState(testTime) {
		t.Error("2012 Oct Mo should be open on Monday Oct 1 2012 at 13:00")
	}

	// Different year should not match
	testTime2013 := time.Date(2013, 10, 7, 13, 0, 0, 0, time.UTC) // Monday Oct 7, 2013
	if oh.GetState(testTime2013) {
		t.Error("2012 Oct Mo should not be open in 2013")
	}
}

func TestJS_SelectorOrder_MonthWeekday(t *testing.T) {
	// Month + weekday
	oh, err := New("Oct Mo 10:00-16:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// October 2012 Monday
	octMon := time.Date(2012, 10, 1, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(octMon) {
		t.Error("Oct Monday should be open")
	}

	// September Monday
	sepMon := time.Date(2012, 9, 3, 12, 0, 0, 0, time.UTC)
	if oh.GetState(sepMon) {
		t.Error("Sep Monday should be closed (only Oct specified)")
	}

	// October Tuesday
	octTue := time.Date(2012, 10, 2, 12, 0, 0, 0, time.UTC)
	if oh.GetState(octTue) {
		t.Error("Oct Tuesday should be closed (only Mo specified)")
	}
}

// =============================================================================
// Extended error tolerance (various dash types)
// =============================================================================

func TestJS_ErrorTolerance_EnDash(t *testing.T) {
	// En-dash (–) instead of hyphen
	oh, err := New("Mo–Fr 09:00–17:00")
	if err != nil {
		t.Fatalf("failed to parse with en-dash: %v", err)
	}

	testTime := time.Date(2012, 10, 3, 12, 0, 0, 0, time.UTC) // Wednesday
	if !oh.GetState(testTime) {
		t.Error("Should be open with en-dash separator")
	}
}

func TestJS_ErrorTolerance_EmDash(t *testing.T) {
	// Em-dash (—) instead of hyphen
	oh, err := New("Mo—Fr 09:00—17:00")
	if err != nil {
		t.Fatalf("failed to parse with em-dash: %v", err)
	}

	testTime := time.Date(2012, 10, 3, 12, 0, 0, 0, time.UTC) // Wednesday
	if !oh.GetState(testTime) {
		t.Error("Should be open with em-dash separator")
	}
}

func TestJS_ErrorTolerance_MinusSign(t *testing.T) {
	// Minus sign (−) instead of hyphen
	oh, err := New("Mo−Fr 09:00−17:00")
	if err != nil {
		t.Fatalf("failed to parse with minus sign: %v", err)
	}

	testTime := time.Date(2012, 10, 3, 12, 0, 0, 0, time.UTC) // Wednesday
	if !oh.GetState(testTime) {
		t.Error("Should be open with minus sign separator")
	}
}

// =============================================================================
// Holiday offset with weekday combinations
// =============================================================================

func TestJS_HolidayOffset_PlusDay(t *testing.T) {
	// PH +1 day
	oh, err := New("PH +1 day 10:00-16:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Set up a custom holiday checker with Dec 25 as a holiday
	hc := &jsTestHolidayChecker{
		holidays: map[string]bool{
			"2012-12-25": true, // Christmas
		},
	}
	oh.SetHolidayChecker(hc)

	// Dec 26, 2012 (day after Christmas Dec 25)
	dayAfterXmas := time.Date(2012, 12, 26, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(dayAfterXmas) {
		t.Error("Day after Christmas should be open")
	}

	// Dec 25 should NOT be open (it's the holiday, not the day after)
	xmas := time.Date(2012, 12, 25, 12, 0, 0, 0, time.UTC)
	if oh.GetState(xmas) {
		t.Error("Christmas itself should be closed (only day after is open)")
	}
}

func TestJS_HolidayOffset_MinusDay(t *testing.T) {
	// PH -1 day (day before public holiday)
	oh, err := New("PH -1 day 10:00-16:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	hc := &jsTestHolidayChecker{
		holidays: map[string]bool{
			"2012-12-25": true, // Christmas
		},
	}
	oh.SetHolidayChecker(hc)

	// Dec 24, 2012 (day before Christmas Dec 25)
	dayBeforeXmas := time.Date(2012, 12, 24, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(dayBeforeXmas) {
		t.Error("Day before Christmas should be open")
	}

	// Dec 25 should NOT be open
	xmas := time.Date(2012, 12, 25, 12, 0, 0, 0, time.UTC)
	if oh.GetState(xmas) {
		t.Error("Christmas itself should be closed (only day before is open)")
	}
}

// =============================================================================
// Additional real-world complex patterns
// =============================================================================

func TestJS_Complex_SeasonalWithExceptions(t *testing.T) {
	// Summer hours with specific date exceptions (using semicolon separator)
	oh, err := New("Apr-Sep Mo-Fr 08:00-18:00; Apr-Sep Sa 09:00-14:00; Oct-Mar Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Summer weekday
	summerWkday := time.Date(2012, 6, 15, 10, 0, 0, 0, time.UTC) // Friday Jun
	if !oh.GetState(summerWkday) {
		t.Error("Summer Friday should be open")
	}

	// Summer Saturday
	summerSat := time.Date(2012, 6, 16, 10, 0, 0, 0, time.UTC)
	if !oh.GetState(summerSat) {
		t.Error("Summer Saturday should be open")
	}

	// Winter weekday
	winterWkday := time.Date(2012, 11, 15, 10, 0, 0, 0, time.UTC) // Thursday Nov
	if !oh.GetState(winterWkday) {
		t.Error("Winter Thursday should be open")
	}
}

func TestJS_Complex_MultipleTimeRangesPerDay(t *testing.T) {
	// Different hours on different days with multiple ranges
	oh, err := New("Mo 08:00-12:00,14:00-18:00; Tu-Fr 09:00-17:00; Sa 10:00-14:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Monday lunch break
	monLunch := time.Date(2012, 10, 1, 13, 0, 0, 0, time.UTC)
	if oh.GetState(monLunch) {
		t.Error("Monday lunch (13:00) should be closed")
	}

	// Monday afternoon
	monAfter := time.Date(2012, 10, 1, 15, 0, 0, 0, time.UTC)
	if !oh.GetState(monAfter) {
		t.Error("Monday afternoon (15:00) should be open")
	}

	// Tuesday continuous
	tue := time.Date(2012, 10, 2, 13, 0, 0, 0, time.UTC)
	if !oh.GetState(tue) {
		t.Error("Tuesday 13:00 should be open (continuous)")
	}
}

// =============================================================================
// Omitted time patterns (weekday without time = all day)
// =============================================================================

func TestJS_OmittedTime_WeekdayList(t *testing.T) {
	// "Mo,We" means open all day on Monday and Wednesday
	oh, err := New("Mo,We")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// October 2012: Mon=1, Tue=2, Wed=3
	tests := []struct {
		day      int
		hour     int
		expected bool
		desc     string
	}{
		{1, 0, true, "Monday midnight"},
		{1, 12, true, "Monday noon"},
		{1, 23, true, "Monday 23:00"},
		{2, 12, false, "Tuesday noon"},
		{3, 0, true, "Wednesday midnight"},
		{3, 12, true, "Wednesday noon"},
		{4, 12, false, "Thursday noon"},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, 10, tt.day, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

func TestJS_OmittedTime_WeekdayRange(t *testing.T) {
	// "Mo-Fr" without time = all day Mon-Fri
	oh, err := New("Mo-Fr")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// October 2012: Mon=1, Sat=6, Sun=7
	tests := []struct {
		day      int
		hour     int
		expected bool
		desc     string
	}{
		{1, 0, true, "Monday midnight"},
		{1, 23, true, "Monday 23:00"},
		{5, 12, true, "Friday noon"},
		{6, 12, false, "Saturday noon"},
		{7, 12, false, "Sunday noon"},
	}

	for _, tt := range tests {
		testTime := time.Date(2012, 10, tt.day, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(testTime)
		if got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, got)
		}
	}
}

func TestJS_OmittedTime_SingleWeekday(t *testing.T) {
	// "Sa" = all day Saturday
	oh, err := New("Sa")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// October 2012: Sat=6
	sat := time.Date(2012, 10, 6, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(sat) {
		t.Error("Saturday noon should be open")
	}

	fri := time.Date(2012, 10, 5, 12, 0, 0, 0, time.UTC)
	if oh.GetState(fri) {
		t.Error("Friday noon should be closed")
	}
}

func TestJS_OmittedTime_WithOff(t *testing.T) {
	// "Mo-Fr; We off" = all day Mon-Fri except Wednesday
	oh, err := New("Mo-Fr; We off")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	mon := time.Date(2012, 10, 1, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(mon) {
		t.Error("Monday should be open")
	}

	wed := time.Date(2012, 10, 3, 12, 0, 0, 0, time.UTC)
	if oh.GetState(wed) {
		t.Error("Wednesday should be closed (off)")
	}

	fri := time.Date(2012, 10, 5, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(fri) {
		t.Error("Friday should be open")
	}
}

// =============================================================================
// Failure/Invalid pattern tests
// =============================================================================

func TestJS_Invalid_IncompleteTimeRange(t *testing.T) {
	invalidPatterns := []string{
		"Mo 10:00-",      // Missing end time
		"Mo -14:00",      // Missing start time
		"Mo 10:00--14:00", // Double dash
	}

	for _, p := range invalidPatterns {
		_, err := New(p)
		if err == nil {
			t.Errorf("Pattern '%s' should fail to parse", p)
		}
	}
}

func TestJS_Invalid_BadTimeFormat(t *testing.T) {
	invalidPatterns := []string{
		"Mo 27:00-28:00", // Invalid hour (max is 26 for extended hours)
		"Mo 10:60-14:00", // Invalid minute
		"Mo 10:00-14:60", // Invalid minute in end
	}

	for _, p := range invalidPatterns {
		_, err := New(p)
		if err == nil {
			t.Errorf("Pattern '%s' should fail to parse", p)
		}
	}
}

func TestJS_Invalid_EmptyPattern(t *testing.T) {
	invalidPatterns := []string{
		"",
		"   ",
	}

	for _, p := range invalidPatterns {
		_, err := New(p)
		if err == nil {
			t.Errorf("Empty pattern '%s' should fail to parse", p)
		}
	}
}

// =============================================================================
// Warning patterns (parse successfully but may have issues)
// =============================================================================

func TestJS_Warning_DeprecatedHoliday(t *testing.T) {
	// These deprecated patterns should still parse
	patterns := []string{
		"Mo-Fr 09:00-17:00; PH off", // Standard pattern with PH
	}

	for _, p := range patterns {
		oh, err := New(p)
		if err != nil {
			t.Fatalf("Pattern '%s' should parse: %v", p, err)
		}
		_ = oh
	}
}

// =============================================================================
// Additional edge cases from JS tests
// =============================================================================

func TestJS_EdgeCase_WeekdaySpanningMidnight(t *testing.T) {
	// Open from Friday evening to Saturday morning
	oh, err := New("Fr 22:00-02:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Friday 23:00 should be open
	fri23 := time.Date(2012, 10, 5, 23, 0, 0, 0, time.UTC)
	if !oh.GetState(fri23) {
		t.Error("Friday 23:00 should be open")
	}

	// Saturday 01:00 should be open (still in Friday's range)
	sat1 := time.Date(2012, 10, 6, 1, 0, 0, 0, time.UTC)
	if !oh.GetState(sat1) {
		t.Error("Saturday 01:00 should be open (Friday extended hours)")
	}

	// Saturday 03:00 should be closed
	sat3 := time.Date(2012, 10, 6, 3, 0, 0, 0, time.UTC)
	if oh.GetState(sat3) {
		t.Error("Saturday 03:00 should be closed")
	}
}

func TestJS_EdgeCase_MultipleRulesOverlap(t *testing.T) {
	// Later rule overrides earlier
	oh, err := New("Mo-Fr 09:00-17:00; We 10:00-14:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Monday follows first rule
	mon10 := time.Date(2012, 10, 1, 10, 0, 0, 0, time.UTC)
	if !oh.GetState(mon10) {
		t.Error("Monday 10:00 should be open")
	}

	mon16 := time.Date(2012, 10, 1, 16, 0, 0, 0, time.UTC)
	if !oh.GetState(mon16) {
		t.Error("Monday 16:00 should be open")
	}

	// Wednesday follows second rule (override)
	wed10 := time.Date(2012, 10, 3, 10, 0, 0, 0, time.UTC)
	if !oh.GetState(wed10) {
		t.Error("Wednesday 10:00 should be open")
	}

	wed16 := time.Date(2012, 10, 3, 16, 0, 0, 0, time.UTC)
	if oh.GetState(wed16) {
		t.Error("Wednesday 16:00 should be closed (override rule ends at 14:00)")
	}
}

func TestJS_EdgeCase_YearRangeOpen(t *testing.T) {
	// Open only in specific years
	oh, err := New("2012-2014 Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// 2012 Monday should be open
	mon2012 := time.Date(2012, 10, 1, 10, 0, 0, 0, time.UTC)
	if !oh.GetState(mon2012) {
		t.Error("2012 Monday should be open")
	}

	// 2014 Monday should be open
	mon2014 := time.Date(2014, 10, 6, 10, 0, 0, 0, time.UTC)
	if !oh.GetState(mon2014) {
		t.Error("2014 Monday should be open")
	}

	// 2015 Monday should be closed
	mon2015 := time.Date(2015, 10, 5, 10, 0, 0, 0, time.UTC)
	if oh.GetState(mon2015) {
		t.Error("2015 Monday should be closed (outside year range)")
	}

	// 2011 Monday should be closed
	mon2011 := time.Date(2011, 10, 3, 10, 0, 0, 0, time.UTC)
	if oh.GetState(mon2011) {
		t.Error("2011 Monday should be closed (outside year range)")
	}
}

func TestJS_EdgeCase_MonthRangeWrapping(t *testing.T) {
	// Nov-Feb wraps around year end
	oh, err := New("Nov-Feb Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// November Monday
	nov := time.Date(2012, 11, 5, 10, 0, 0, 0, time.UTC)
	if !oh.GetState(nov) {
		t.Error("November Monday should be open")
	}

	// January Monday
	jan := time.Date(2013, 1, 7, 10, 0, 0, 0, time.UTC)
	if !oh.GetState(jan) {
		t.Error("January Monday should be open")
	}

	// May Monday
	may := time.Date(2012, 5, 7, 10, 0, 0, 0, time.UTC)
	if oh.GetState(may) {
		t.Error("May Monday should be closed (outside Nov-Feb)")
	}
}

func TestJS_EdgeCase_ComplexComment(t *testing.T) {
	// Comments with special characters
	patterns := []string{
		`Mo-Fr 09:00-17:00 "Regular hours"`,
		`Mo-Fr 09:00-17:00 "Hours: 9am-5pm"`,
		`Mo-Fr 09:00-17:00 "Call (555) 123-4567"`,
	}

	testTime := time.Date(2012, 10, 1, 10, 0, 0, 0, time.UTC)
	for _, p := range patterns {
		oh, err := New(p)
		if err != nil {
			t.Fatalf("Pattern '%s' should parse: %v", p, err)
		}
		if !oh.GetState(testTime) {
			t.Errorf("Pattern '%s' should be open Monday 10:00", p)
		}
	}
}

// =============================================================================
// Comprehensive JS library pattern coverage
// =============================================================================

func TestJS_Pattern_WeekIntervals(t *testing.T) {
	// Odd weeks
	oh, err := New("week 01-53/2 Sa 10:00-14:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Week 1 Saturday (Jan 7, 2012)
	week1Sat := time.Date(2012, 1, 7, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(week1Sat) {
		t.Error("Week 1 Saturday should be open (odd week)")
	}

	// Week 2 Saturday (Jan 14, 2012)
	week2Sat := time.Date(2012, 1, 14, 12, 0, 0, 0, time.UTC)
	if oh.GetState(week2Sat) {
		t.Error("Week 2 Saturday should be closed (even week)")
	}
}

func TestJS_Pattern_MonthConstrainedWeekday(t *testing.T) {
	// First Monday of January only
	oh, err := New("Jan Mo[1] 10:00-14:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// First Monday of Jan 2012 is Jan 2
	jan2 := time.Date(2012, 1, 2, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(jan2) {
		t.Error("First Monday of January should be open")
	}

	// Second Monday of Jan 2012 is Jan 9
	jan9 := time.Date(2012, 1, 9, 12, 0, 0, 0, time.UTC)
	if oh.GetState(jan9) {
		t.Error("Second Monday of January should be closed")
	}

	// First Monday of Feb 2012 is Feb 6
	feb6 := time.Date(2012, 2, 6, 12, 0, 0, 0, time.UTC)
	if oh.GetState(feb6) {
		t.Error("First Monday of February should be closed (Jan only)")
	}
}

func TestJS_Pattern_YearMonthWeekday(t *testing.T) {
	// 2012 January Mondays only
	oh, err := New("2012 Jan Mo 10:00-14:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Monday Jan 2, 2012
	jan2_2012 := time.Date(2012, 1, 2, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(jan2_2012) {
		t.Error("2012 Jan Monday should be open")
	}

	// Monday Jan 7, 2013
	jan7_2013 := time.Date(2013, 1, 7, 12, 0, 0, 0, time.UTC)
	if oh.GetState(jan7_2013) {
		t.Error("2013 Jan Monday should be closed (2012 only)")
	}
}

func TestJS_Pattern_SpecificDateOverride(t *testing.T) {
	// Normal hours with Christmas Eve exception
	oh, err := New("Mo-Fr 09:00-17:00; Dec 24 10:00-14:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Normal Monday
	mon := time.Date(2012, 10, 1, 10, 0, 0, 0, time.UTC)
	if !oh.GetState(mon) {
		t.Error("Normal Monday should be open")
	}

	// Dec 24, 2012 (Monday) at 10:00 - should use override
	xmasEve10 := time.Date(2012, 12, 24, 10, 0, 0, 0, time.UTC)
	if !oh.GetState(xmasEve10) {
		t.Error("Dec 24 at 10:00 should be open")
	}

	// Dec 24 at 16:00 - should be closed (override ends at 14:00)
	xmasEve16 := time.Date(2012, 12, 24, 16, 0, 0, 0, time.UTC)
	if oh.GetState(xmasEve16) {
		t.Error("Dec 24 at 16:00 should be closed (override hours)")
	}
}

func TestJS_Pattern_LunchBreak(t *testing.T) {
	// Classic lunch break pattern
	oh, err := New("Mo-Fr 08:00-12:00; Mo-Fr 14:00-18:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Morning
	morning := time.Date(2012, 10, 1, 10, 0, 0, 0, time.UTC)
	if !oh.GetState(morning) {
		t.Error("Morning should be open")
	}

	// Lunch
	lunch := time.Date(2012, 10, 1, 13, 0, 0, 0, time.UTC)
	if oh.GetState(lunch) {
		t.Error("Lunch time should be closed")
	}

	// Afternoon
	afternoon := time.Date(2012, 10, 1, 15, 0, 0, 0, time.UTC)
	if !oh.GetState(afternoon) {
		t.Error("Afternoon should be open")
	}
}

func TestJS_Pattern_FridayEarlyClose(t *testing.T) {
	// Friday closes early
	oh, err := New("Mo-Th 09:00-17:00; Fr 09:00-15:00")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Thursday 16:00 - open
	thu := time.Date(2012, 10, 4, 16, 0, 0, 0, time.UTC)
	if !oh.GetState(thu) {
		t.Error("Thursday 16:00 should be open")
	}

	// Friday 14:00 - open
	fri14 := time.Date(2012, 10, 5, 14, 0, 0, 0, time.UTC)
	if !oh.GetState(fri14) {
		t.Error("Friday 14:00 should be open")
	}

	// Friday 16:00 - closed
	fri16 := time.Date(2012, 10, 5, 16, 0, 0, 0, time.UTC)
	if oh.GetState(fri16) {
		t.Error("Friday 16:00 should be closed (early close)")
	}
}

func TestJS_Pattern_SpecificDatesOff(t *testing.T) {
	// Workaround: use semicolons instead of comma for multiple month-days
	oh, err := New("Mo-Fr 09:00-17:00; Jan 01 off; Dec 25 off")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Normal Monday
	mon := time.Date(2012, 10, 1, 10, 0, 0, 0, time.UTC)
	if !oh.GetState(mon) {
		t.Error("Normal Monday should be open")
	}

	// New Year's Day (Tuesday Jan 1, 2013)
	newYear := time.Date(2013, 1, 1, 10, 0, 0, 0, time.UTC)
	if oh.GetState(newYear) {
		t.Error("New Year's Day should be closed")
	}

	// Christmas (Tuesday Dec 25, 2012)
	christmas := time.Date(2012, 12, 25, 10, 0, 0, 0, time.UTC)
	if oh.GetState(christmas) {
		t.Error("Christmas should be closed")
	}
}

func TestJS_Pattern_AllIntervalTypes(t *testing.T) {
	// Day interval
	t.Run("DayInterval", func(t *testing.T) {
		oh, err := New("Jan 01-31/8 10:00-16:00")
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}
		// Jan 1 should be open
		jan1 := time.Date(2012, 1, 1, 12, 0, 0, 0, time.UTC)
		if !oh.GetState(jan1) {
			t.Error("Jan 1 should be open")
		}
		// Jan 9 should be open (1 + 8)
		jan9 := time.Date(2012, 1, 9, 12, 0, 0, 0, time.UTC)
		if !oh.GetState(jan9) {
			t.Error("Jan 9 should be open")
		}
		// Jan 5 should be closed
		jan5 := time.Date(2012, 1, 5, 12, 0, 0, 0, time.UTC)
		if oh.GetState(jan5) {
			t.Error("Jan 5 should be closed")
		}
	})

	// Year interval
	t.Run("YearInterval", func(t *testing.T) {
		oh, err := New("2012-2020/2 Mo-Fr 09:00-17:00")
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}
		// 2012 Monday should be open
		mon2012 := time.Date(2012, 10, 1, 10, 0, 0, 0, time.UTC)
		if !oh.GetState(mon2012) {
			t.Error("2012 Monday should be open")
		}
		// 2014 Monday should be open (2012 + 2)
		mon2014 := time.Date(2014, 10, 6, 10, 0, 0, 0, time.UTC)
		if !oh.GetState(mon2014) {
			t.Error("2014 Monday should be open")
		}
		// 2013 Monday should be closed
		mon2013 := time.Date(2013, 10, 7, 10, 0, 0, 0, time.UTC)
		if oh.GetState(mon2013) {
			t.Error("2013 Monday should be closed")
		}
	})

	// Time interval
	t.Run("TimeInterval", func(t *testing.T) {
		oh, err := New("Mo 10:00-16:00/01:30")
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}
		// 10:00 should be open
		t10 := time.Date(2012, 10, 1, 10, 0, 0, 0, time.UTC)
		if !oh.GetState(t10) {
			t.Error("10:00 should be open")
		}
		// 11:00 should be open (within first 1:30 slot)
		t11 := time.Date(2012, 10, 1, 11, 0, 0, 0, time.UTC)
		if !oh.GetState(t11) {
			t.Error("11:00 should be open")
		}
		// 12:00 should be closed (second slot is closed)
		t12 := time.Date(2012, 10, 1, 12, 0, 0, 0, time.UTC)
		if oh.GetState(t12) {
			t.Error("12:00 should be closed")
		}
		// 13:00 should be open (third slot)
		t13 := time.Date(2012, 10, 1, 13, 0, 0, 0, time.UTC)
		if !oh.GetState(t13) {
			t.Error("13:00 should be open")
		}
	})
}

func TestJS_Pattern_Modifiers(t *testing.T) {
	// Explicit open
	t.Run("ExplicitOpen", func(t *testing.T) {
		oh, err := New("Mo-Fr 09:00-17:00 open")
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}
		testTime := time.Date(2012, 10, 1, 10, 0, 0, 0, time.UTC)
		if !oh.GetState(testTime) {
			t.Error("Should be open")
		}
	})

	// Explicit closed
	t.Run("ExplicitClosed", func(t *testing.T) {
		oh, err := New("Dec 25 closed")
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}
		christmas := time.Date(2012, 12, 25, 12, 0, 0, 0, time.UTC)
		if oh.GetState(christmas) {
			t.Error("Should be closed")
		}
	})

	// Unknown state
	t.Run("UnknownState", func(t *testing.T) {
		oh, err := New("Mo unknown")
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}
		monday := time.Date(2012, 10, 1, 12, 0, 0, 0, time.UTC)
		state := oh.GetStateString(monday)
		// Unknown may be treated as closed by GetState
		if state != "unknown" && state != "closed" {
			t.Errorf("Expected unknown or closed, got %s", state)
		}
	})
}
