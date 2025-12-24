package openinghours

import (
	"testing"
	"time"
)

// mockSchoolHolidayChecker is a simple implementation of SchoolHolidayChecker for testing
type mockSchoolHolidayChecker struct {
	holidays map[string]bool
}

func (m *mockSchoolHolidayChecker) IsSchoolHoliday(t time.Time) bool {
	return m.holidays[t.Format("2006-01-02")]
}

// TestSchoolHoliday_Basic tests that school holidays override normal weekday rules when using "SH off"
func TestSchoolHoliday_Basic(t *testing.T) {
	// Input: "Mo-Fr 09:00-17:00; SH off"
	// This means: open weekdays 9am-5pm, but closed during school holidays
	oh, err := New("Mo-Fr 09:00-17:00; SH off")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Set up a mock school holiday checker that returns true for Jan 1-7, 2024
	shChecker := &mockSchoolHolidayChecker{
		holidays: map[string]bool{
			"2024-01-01": true,
			"2024-01-02": true,
			"2024-01-03": true,
			"2024-01-04": true,
			"2024-01-05": true,
			"2024-01-06": true,
			"2024-01-07": true,
		},
	}

	oh.SetSchoolHolidayChecker(shChecker)

	// Jan 3, 2024 is Wednesday (normally open), but it's during school holidays
	// At 10:00 during school holidays should be closed
	duringHoliday := time.Date(2024, 1, 3, 10, 0, 0, 0, time.UTC)
	if oh.GetState(duringHoliday) {
		t.Errorf("expected closed on Jan 3 (Wed) at 10:00 during school holidays, got open")
	}

	// Jan 10, 2024 is Wednesday (not a school holiday)
	// At 10:00 (not school holiday) should be open
	notHoliday := time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC)
	if !oh.GetState(notHoliday) {
		t.Errorf("expected open on Jan 10 (Wed) at 10:00 (not school holiday), got closed")
	}
}

// TestSchoolHoliday_OpenDuring tests opening hours that apply only during school holidays
func TestSchoolHoliday_OpenDuring(t *testing.T) {
	// Input: "SH 10:00-14:00"
	// This means: open only during school holidays from 10am-2pm
	oh, err := New("SH 10:00-14:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Set up a mock school holiday checker
	shChecker := &mockSchoolHolidayChecker{
		holidays: map[string]bool{
			"2024-01-03": true,  // Wednesday is a school holiday
			"2024-01-10": false, // Wednesday is not a school holiday
		},
	}

	oh.SetSchoolHolidayChecker(shChecker)

	// During school holidays, 12:00 should be open
	duringHoliday := time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(duringHoliday) {
		t.Errorf("expected open during school holidays at 12:00, got closed")
	}

	// Outside school holidays, 12:00 should be closed
	notHoliday := time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC)
	if oh.GetState(notHoliday) {
		t.Errorf("expected closed outside school holidays at 12:00, got open")
	}

	// During school holidays but outside time range (8:00), should be closed
	duringHolidayBeforeOpen := time.Date(2024, 1, 3, 8, 0, 0, 0, time.UTC)
	if oh.GetState(duringHolidayBeforeOpen) {
		t.Errorf("expected closed during school holidays at 8:00 (before 10:00), got open")
	}

	// During school holidays but after time range (15:00), should be closed
	duringHolidayAfterClose := time.Date(2024, 1, 3, 15, 0, 0, 0, time.UTC)
	if oh.GetState(duringHolidayAfterClose) {
		t.Errorf("expected closed during school holidays at 15:00 (after 14:00), got open")
	}
}

// TestSchoolHoliday_WithWeekday tests school holidays combined with weekday constraints
func TestSchoolHoliday_WithWeekday(t *testing.T) {
	// Input: "SH Mo-Fr 09:00-17:00"
	// This means: open during school holidays, but only on weekdays (Mon-Fri) from 9am-5pm
	oh, err := New("SH Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Set up a mock school holiday checker
	// Jan 1, 2024 is Monday
	// Jan 6, 2024 is Saturday
	shChecker := &mockSchoolHolidayChecker{
		holidays: map[string]bool{
			"2024-01-01": true,  // Monday - school holiday
			"2024-01-06": true,  // Saturday - school holiday
			"2024-01-08": false, // Monday - not a school holiday
		},
	}

	oh.SetSchoolHolidayChecker(shChecker)

	// School holiday on Monday at 10:00 should be open
	holidayMonday := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	if !oh.GetState(holidayMonday) {
		t.Errorf("expected open on school holiday Monday at 10:00, got closed")
	}

	// School holiday on Saturday at 10:00 should be closed (not a weekday)
	holidaySaturday := time.Date(2024, 1, 6, 10, 0, 0, 0, time.UTC)
	if oh.GetState(holidaySaturday) {
		t.Errorf("expected closed on school holiday Saturday at 10:00 (not Mo-Fr), got open")
	}

	// Not a school holiday, even if it's Monday, should be closed
	notHolidayMonday := time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC)
	if oh.GetState(notHolidayMonday) {
		t.Errorf("expected closed on non-holiday Monday at 10:00, got open")
	}

	// School holiday on Monday but outside time range (8:00), should be closed
	holidayMondayEarly := time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC)
	if oh.GetState(holidayMondayEarly) {
		t.Errorf("expected closed on school holiday Monday at 8:00 (before 9:00), got open")
	}
}

// TestSchoolHoliday_NoChecker tests that SH rules don't match when no checker is set
func TestSchoolHoliday_NoChecker(t *testing.T) {
	// When no checker is set, SH rules should not match (similar to PH behavior)
	oh, err := New("SH 10:00-14:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// No checker set, so any time should be closed
	testTime := time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC)
	if oh.GetState(testTime) {
		t.Errorf("expected closed when no school holiday checker is set, got open")
	}
}

// TestSchoolHoliday_CombinedWithRegularRules tests SH combined with regular opening hours
func TestSchoolHoliday_CombinedWithRegularRules(t *testing.T) {
	// Regular hours with school holiday override
	oh, err := New("Mo-Fr 08:00-18:00; SH 10:00-14:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	shChecker := &mockSchoolHolidayChecker{
		holidays: map[string]bool{
			"2024-01-03": true,  // Wednesday - school holiday
			"2024-01-10": false, // Wednesday - not a school holiday
		},
	}

	oh.SetSchoolHolidayChecker(shChecker)

	// During school holiday, should use SH hours (10:00-14:00)
	// At 9:00 on school holiday Wednesday, should follow SH rule (closed, as 9:00 is before 10:00)
	duringHolidayEarly := time.Date(2024, 1, 3, 9, 0, 0, 0, time.UTC)
	if oh.GetState(duringHolidayEarly) {
		t.Errorf("expected closed at 9:00 on school holiday (SH hours are 10:00-14:00), got open")
	}

	// At 12:00 on school holiday Wednesday, should be open (within SH hours)
	duringHoliday := time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(duringHoliday) {
		t.Errorf("expected open at 12:00 on school holiday, got closed")
	}

	// At 16:00 on school holiday Wednesday, should follow SH rule (closed, as 16:00 is after 14:00)
	duringHolidayLate := time.Date(2024, 1, 3, 16, 0, 0, 0, time.UTC)
	if oh.GetState(duringHolidayLate) {
		t.Errorf("expected closed at 16:00 on school holiday (SH hours are 10:00-14:00), got open")
	}

	// Not a school holiday, should use regular hours (08:00-18:00)
	// At 9:00 on regular Wednesday, should be open
	notHolidayEarly := time.Date(2024, 1, 10, 9, 0, 0, 0, time.UTC)
	if !oh.GetState(notHolidayEarly) {
		t.Errorf("expected open at 9:00 on regular Wednesday (regular hours 08:00-18:00), got closed")
	}

	// At 16:00 on regular Wednesday, should be open
	notHolidayLate := time.Date(2024, 1, 10, 16, 0, 0, 0, time.UTC)
	if !oh.GetState(notHolidayLate) {
		t.Errorf("expected open at 16:00 on regular Wednesday (regular hours 08:00-18:00), got closed")
	}
}

// TestSchoolHoliday_AlwaysOff tests "SH off" to close during school holidays
func TestSchoolHoliday_AlwaysOff(t *testing.T) {
	oh, err := New("SH off")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	shChecker := &mockSchoolHolidayChecker{
		holidays: map[string]bool{
			"2024-01-03": true,
		},
	}

	oh.SetSchoolHolidayChecker(shChecker)

	// During school holiday, should always be closed
	duringHoliday := time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC)
	if oh.GetState(duringHoliday) {
		t.Errorf("expected closed during school holiday with 'SH off', got open")
	}

	// Not a school holiday, should also be closed (no other rules specified)
	notHoliday := time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC)
	if oh.GetState(notHoliday) {
		t.Errorf("expected closed on non-holiday with only 'SH off' rule, got open")
	}
}

// TestSchoolHoliday_MultipleTimeRanges tests SH with multiple time ranges
func TestSchoolHoliday_MultipleTimeRanges(t *testing.T) {
	oh, err := New("SH 08:00-12:00,14:00-18:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	shChecker := &mockSchoolHolidayChecker{
		holidays: map[string]bool{
			"2024-01-03": true,
		},
	}

	oh.SetSchoolHolidayChecker(shChecker)

	tests := []struct {
		hour int
		want bool
		desc string
	}{
		{7, false, "before first range"},
		{10, true, "in first range"},
		{13, false, "between ranges"},
		{16, true, "in second range"},
		{19, false, "after second range"},
	}

	for _, tt := range tests {
		tm := time.Date(2024, 1, 3, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(tm)
		if got != tt.want {
			t.Errorf("%s (hour %d): got %v, want %v", tt.desc, tt.hour, got, tt.want)
		}
	}
}

// TestSchoolHoliday_IsWeekStable tests that SH rules make IsWeekStable return false
func TestSchoolHoliday_IsWeekStable(t *testing.T) {
	// Opening hours with SH should not be week-stable because school holidays
	// can fall on any day and change from week to week
	oh, err := New("Mo-Fr 09:00-17:00; SH off")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if oh.IsWeekStable() {
		t.Errorf("expected IsWeekStable() to return false for rules with SH")
	}
}

// mockPublicHolidayChecker is a simple implementation for testing PH with SH
type mockPublicHolidayChecker struct {
	holidays map[string]bool
}

func (h *mockPublicHolidayChecker) IsHoliday(t time.Time) bool {
	return h.holidays[t.Format("2006-01-02")]
}

// TestSchoolHoliday_WithPHAndSH tests combining both public holidays and school holidays
func TestSchoolHoliday_WithPHAndSH(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-17:00; PH off; SH 10:00-14:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Set up both checkers
	phChecker := &mockPublicHolidayChecker{
		holidays: map[string]bool{
			"2024-01-01": true, // Jan 1 is a public holiday
		},
	}

	shChecker := &mockSchoolHolidayChecker{
		holidays: map[string]bool{
			"2024-01-03": true, // Jan 3 is a school holiday
		},
	}

	oh.SetHolidayChecker(phChecker)
	oh.SetSchoolHolidayChecker(shChecker)

	// Jan 1, 2024 is Monday and a public holiday - should be closed (PH off)
	publicHoliday := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	if oh.GetState(publicHoliday) {
		t.Errorf("expected closed on public holiday (PH off), got open")
	}

	// Jan 3, 2024 is Wednesday and a school holiday - should be open 10:00-14:00
	schoolHoliday := time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(schoolHoliday) {
		t.Errorf("expected open at 12:00 on school holiday (SH 10:00-14:00), got closed")
	}

	// Jan 3 at 9:00 - school holiday but before SH hours
	schoolHolidayEarly := time.Date(2024, 1, 3, 9, 0, 0, 0, time.UTC)
	if oh.GetState(schoolHolidayEarly) {
		t.Errorf("expected closed at 9:00 on school holiday (before SH hours), got open")
	}

	// Jan 10, 2024 is Wednesday, not a holiday - should use regular hours
	regularDay := time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(regularDay) {
		t.Errorf("expected open at 12:00 on regular Wednesday, got closed")
	}
}

// TestSchoolHoliday_GetMatchingRule tests that GetMatchingRule works correctly with SH rules
func TestSchoolHoliday_GetMatchingRule(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-17:00; SH 10:00-14:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	shChecker := &mockSchoolHolidayChecker{
		holidays: map[string]bool{
			"2024-01-03": true,  // Wednesday - school holiday
			"2024-01-10": false, // Wednesday - not a school holiday
		},
	}

	oh.SetSchoolHolidayChecker(shChecker)

	// During school holiday, should match SH rule (index 1)
	duringHoliday := time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC)
	ruleIndex := oh.GetMatchingRule(duringHoliday)
	if ruleIndex != 1 {
		t.Errorf("expected matching rule index 1 (SH rule) during school holiday, got %d", ruleIndex)
	}

	// Not a school holiday, should match first rule (index 0)
	notHoliday := time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC)
	ruleIndex = oh.GetMatchingRule(notHoliday)
	if ruleIndex != 0 {
		t.Errorf("expected matching rule index 0 (regular rule) on non-holiday, got %d", ruleIndex)
	}
}
