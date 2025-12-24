package openinghours

import (
	"testing"
	"time"
)

// mockHolidayChecker is a simple implementation of HolidayChecker for testing
type mockHolidayChecker struct {
	holidays map[string]bool
}

func (m *mockHolidayChecker) IsHoliday(t time.Time) bool {
	return m.holidays[t.Format("2006-01-02")]
}

// TestHolidayOffset_DayAfter tests the "PH +1 day" syntax for day after a public holiday
func TestHolidayOffset_DayAfter(t *testing.T) {
	// Input: "PH +1 day 10:00-14:00"
	// This means: day after a public holiday has hours 10:00-14:00
	oh, err := New("PH +1 day 10:00-14:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Set up mock: Jan 1, 2024 is a holiday (Monday)
	hChecker := &mockHolidayChecker{
		holidays: map[string]bool{
			"2024-01-01": true, // New Year's Day (Monday)
		},
	}
	oh.SetHolidayChecker(hChecker)

	// Jan 2, 2024 (Tuesday, day after holiday) at 12:00 should be open
	dayAfterHoliday := time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(dayAfterHoliday) {
		t.Errorf("expected open on Jan 2 (day after holiday) at 12:00, got closed")
	}

	// Jan 2, 2024 at 10:00 (start time) should be open
	dayAfterHolidayStart := time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC)
	if !oh.GetState(dayAfterHolidayStart) {
		t.Errorf("expected open on Jan 2 (day after holiday) at 10:00, got closed")
	}

	// Jan 2, 2024 at 09:00 (before opening) should be closed
	dayAfterHolidayEarly := time.Date(2024, 1, 2, 9, 0, 0, 0, time.UTC)
	if oh.GetState(dayAfterHolidayEarly) {
		t.Errorf("expected closed on Jan 2 (day after holiday) at 09:00 (before opening), got open")
	}

	// Jan 2, 2024 at 14:00 (end time, exclusive) should be closed
	dayAfterHolidayEnd := time.Date(2024, 1, 2, 14, 0, 0, 0, time.UTC)
	if oh.GetState(dayAfterHolidayEnd) {
		t.Errorf("expected closed on Jan 2 (day after holiday) at 14:00 (closing time), got open")
	}

	// Jan 2, 2024 at 15:00 (after closing) should be closed
	dayAfterHolidayLate := time.Date(2024, 1, 2, 15, 0, 0, 0, time.UTC)
	if oh.GetState(dayAfterHolidayLate) {
		t.Errorf("expected closed on Jan 2 (day after holiday) at 15:00 (after closing), got open")
	}

	// Jan 1, 2024 (the holiday itself) at 12:00 should be closed
	holidayItself := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	if oh.GetState(holidayItself) {
		t.Errorf("expected closed on Jan 1 (the holiday itself) at 12:00, got open")
	}

	// Jan 3, 2024 (two days after holiday) at 12:00 should be closed
	twoDaysAfter := time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC)
	if oh.GetState(twoDaysAfter) {
		t.Errorf("expected closed on Jan 3 (two days after holiday) at 12:00, got open")
	}

	// Jan 8, 2024 (regular Monday, not related to holiday) at 12:00 should be closed
	regularDay := time.Date(2024, 1, 8, 12, 0, 0, 0, time.UTC)
	if oh.GetState(regularDay) {
		t.Errorf("expected closed on Jan 8 (regular day) at 12:00, got open")
	}
}

// TestHolidayOffset_DayBefore tests the "PH -1 day off" syntax for day before a public holiday
func TestHolidayOffset_DayBefore(t *testing.T) {
	// Input: "Mo-Fr 09:00-17:00; PH -1 day off"
	// This means: weekdays 09:00-17:00, but day before a public holiday is closed
	oh, err := New("Mo-Fr 09:00-17:00; PH -1 day off")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Set up mock: Jan 1, 2024 is a holiday (Monday)
	hChecker := &mockHolidayChecker{
		holidays: map[string]bool{
			"2024-01-01": true, // New Year's Day (Monday)
		},
	}
	oh.SetHolidayChecker(hChecker)

	// Dec 31, 2023 (Sunday, day before holiday) should be closed
	// (would be closed anyway as it's Sunday, but testing the PH -1 day rule)
	dayBeforeHoliday := time.Date(2023, 12, 31, 12, 0, 0, 0, time.UTC)
	if oh.GetState(dayBeforeHoliday) {
		t.Errorf("expected closed on Dec 31 (day before holiday, Sunday) at 12:00, got open")
	}

	// Dec 30, 2023 (Saturday) should be closed (weekend, not a weekday)
	twoDaysBefore := time.Date(2023, 12, 30, 12, 0, 0, 0, time.UTC)
	if oh.GetState(twoDaysBefore) {
		t.Errorf("expected closed on Dec 30 (Saturday) at 12:00, got open")
	}

	// Dec 29, 2023 (Friday, two days before holiday) should follow normal rules (open)
	regularFriday := time.Date(2023, 12, 29, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(regularFriday) {
		t.Errorf("expected open on Dec 29 (Friday, regular weekday) at 12:00, got closed")
	}

	// Jan 2, 2024 (Tuesday, day after holiday) should follow normal rules (open)
	dayAfterHoliday := time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(dayAfterHoliday) {
		t.Errorf("expected open on Jan 2 (Tuesday, regular weekday) at 12:00, got closed")
	}

	// Test with a holiday on a weekday where day before is also a weekday
	// Add Jan 5, 2024 (Friday) as a holiday
	hChecker.holidays["2024-01-05"] = true

	// Jan 4, 2024 (Thursday, day before Friday holiday) should be closed (PH -1 day off)
	dayBeforeWeekdayHoliday := time.Date(2024, 1, 4, 12, 0, 0, 0, time.UTC)
	if oh.GetState(dayBeforeWeekdayHoliday) {
		t.Errorf("expected closed on Jan 4 (Thursday, day before holiday) at 12:00, got open")
	}

	// Jan 3, 2024 (Wednesday, two days before Friday holiday) should follow normal rules (open)
	twoDaysBeforeWeekdayHoliday := time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(twoDaysBeforeWeekdayHoliday) {
		t.Errorf("expected open on Jan 3 (Wednesday, regular weekday) at 12:00, got closed")
	}
}

// TestHolidayOffset_CombinedWithPH tests combining regular PH rules with offset rules
func TestHolidayOffset_CombinedWithPH(t *testing.T) {
	// Input: "Mo-Fr 09:00-17:00; PH off; PH +1 day 10:00-14:00"
	// This means: weekdays 09:00-17:00, holidays are closed, day after holiday is 10:00-14:00
	oh, err := New("Mo-Fr 09:00-17:00; PH off; PH +1 day 10:00-14:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Set up mock: Jan 1, 2024 is a holiday (Monday)
	hChecker := &mockHolidayChecker{
		holidays: map[string]bool{
			"2024-01-01": true, // New Year's Day (Monday)
		},
	}
	oh.SetHolidayChecker(hChecker)

	// Jan 1, 2024 (the holiday, Monday) should be closed (PH off)
	holiday := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	if oh.GetState(holiday) {
		t.Errorf("expected closed on Jan 1 (holiday) at 12:00, got open")
	}

	// Jan 2, 2024 (Tuesday, day after holiday) at 12:00 should be open (PH +1 day hours)
	dayAfterHoliday := time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(dayAfterHoliday) {
		t.Errorf("expected open on Jan 2 (day after holiday) at 12:00, got closed")
	}

	// Jan 2, 2024 at 10:00 should be open (start of PH +1 day hours)
	dayAfterHolidayStart := time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC)
	if !oh.GetState(dayAfterHolidayStart) {
		t.Errorf("expected open on Jan 2 (day after holiday) at 10:00, got closed")
	}

	// Jan 2, 2024 at 09:00 should be closed (before PH +1 day hours, even though normal weekday hours would be open)
	dayAfterHolidayEarly := time.Date(2024, 1, 2, 9, 0, 0, 0, time.UTC)
	if oh.GetState(dayAfterHolidayEarly) {
		t.Errorf("expected closed on Jan 2 (day after holiday) at 09:00, got open")
	}

	// Jan 2, 2024 at 15:00 should be closed (after PH +1 day hours, even though normal weekday hours would be open)
	dayAfterHolidayLate := time.Date(2024, 1, 2, 15, 0, 0, 0, time.UTC)
	if oh.GetState(dayAfterHolidayLate) {
		t.Errorf("expected closed on Jan 2 (day after holiday) at 15:00, got open")
	}

	// Jan 3, 2024 (Wednesday, regular weekday) should follow normal hours (open at 12:00)
	regularDay := time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(regularDay) {
		t.Errorf("expected open on Jan 3 (regular Wednesday) at 12:00, got closed")
	}

	// Jan 3, 2024 at 09:00 should be open (normal weekday hours)
	regularDayStart := time.Date(2024, 1, 3, 9, 0, 0, 0, time.UTC)
	if !oh.GetState(regularDayStart) {
		t.Errorf("expected open on Jan 3 (regular Wednesday) at 09:00, got closed")
	}

	// Jan 3, 2024 at 16:00 should be open (normal weekday hours)
	regularDayLate := time.Date(2024, 1, 3, 16, 0, 0, 0, time.UTC)
	if !oh.GetState(regularDayLate) {
		t.Errorf("expected open on Jan 3 (regular Wednesday) at 16:00, got closed")
	}
}

// TestHolidayOffset_WithWeekday tests combining PH offset with weekday constraints
func TestHolidayOffset_WithWeekday(t *testing.T) {
	// Input: "PH +1 day Mo-Fr 09:00-17:00"
	// This means: day after holiday, only on weekdays (Mon-Fri), hours are 09:00-17:00
	oh, err := New("PH +1 day Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Set up mock holidays:
	// Jan 1, 2024 is Monday (New Year's Day)
	// Jan 6, 2024 is Saturday
	hChecker := &mockHolidayChecker{
		holidays: map[string]bool{
			"2024-01-01": true, // Monday
			"2024-01-06": true, // Saturday
		},
	}
	oh.SetHolidayChecker(hChecker)

	// Jan 2, 2024 (Tuesday, day after Monday holiday) should be open at 12:00
	// (day after holiday AND it's a weekday)
	dayAfterMondayHoliday := time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(dayAfterMondayHoliday) {
		t.Errorf("expected open on Jan 2 (Tuesday, day after Monday holiday) at 12:00, got closed")
	}

	// Jan 2, 2024 at 09:00 should be open (start time)
	dayAfterMondayHolidayStart := time.Date(2024, 1, 2, 9, 0, 0, 0, time.UTC)
	if !oh.GetState(dayAfterMondayHolidayStart) {
		t.Errorf("expected open on Jan 2 (Tuesday, day after Monday holiday) at 09:00, got closed")
	}

	// Jan 2, 2024 at 18:00 should be closed (after closing time)
	dayAfterMondayHolidayLate := time.Date(2024, 1, 2, 18, 0, 0, 0, time.UTC)
	if oh.GetState(dayAfterMondayHolidayLate) {
		t.Errorf("expected closed on Jan 2 (Tuesday, day after Monday holiday) at 18:00, got open")
	}

	// Jan 7, 2024 (Sunday, day after Saturday holiday) should be closed
	// (day after holiday but NOT a weekday, so rule doesn't apply)
	dayAfterSaturdayHoliday := time.Date(2024, 1, 7, 12, 0, 0, 0, time.UTC)
	if oh.GetState(dayAfterSaturdayHoliday) {
		t.Errorf("expected closed on Jan 7 (Sunday, day after Saturday holiday) at 12:00, got open")
	}

	// Jan 8, 2024 (Monday, not day after a holiday) should be closed
	regularMonday := time.Date(2024, 1, 8, 12, 0, 0, 0, time.UTC)
	if oh.GetState(regularMonday) {
		t.Errorf("expected closed on Jan 8 (regular Monday) at 12:00, got open")
	}

	// Jan 1, 2024 (the Monday holiday itself) should be closed
	mondayHoliday := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	if oh.GetState(mondayHoliday) {
		t.Errorf("expected closed on Jan 1 (Monday holiday) at 12:00, got open")
	}

	// Jan 6, 2024 (the Saturday holiday itself) should be closed
	saturdayHoliday := time.Date(2024, 1, 6, 12, 0, 0, 0, time.UTC)
	if oh.GetState(saturdayHoliday) {
		t.Errorf("expected closed on Jan 6 (Saturday holiday) at 12:00, got open")
	}
}

// TestHolidayOffset_MultipleOffsets tests combining multiple offset rules
func TestHolidayOffset_MultipleOffsets(t *testing.T) {
	// Input: "Mo-Fr 09:00-17:00; PH -1 day off; PH +1 day 10:00-14:00"
	// This means: normal weekday hours, day before holiday is closed, day after has special hours
	oh, err := New("Mo-Fr 09:00-17:00; PH -1 day off; PH +1 day 10:00-14:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Set up mock: Jan 3, 2024 is a holiday (Wednesday)
	hChecker := &mockHolidayChecker{
		holidays: map[string]bool{
			"2024-01-03": true, // Wednesday
		},
	}
	oh.SetHolidayChecker(hChecker)

	// Jan 2, 2024 (Tuesday, day before Wednesday holiday) should be closed (PH -1 day off)
	dayBeforeHoliday := time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)
	if oh.GetState(dayBeforeHoliday) {
		t.Errorf("expected closed on Jan 2 (Tuesday, day before holiday) at 12:00, got open")
	}

	// Jan 3, 2024 (the Wednesday holiday) should be closed (no PH rule, but also no other matching rule)
	holiday := time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC)
	if oh.GetState(holiday) {
		t.Errorf("expected closed on Jan 3 (Wednesday holiday) at 12:00, got open")
	}

	// Jan 4, 2024 (Thursday, day after Wednesday holiday) at 12:00 should be open (PH +1 day hours)
	dayAfterHoliday := time.Date(2024, 1, 4, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(dayAfterHoliday) {
		t.Errorf("expected open on Jan 4 (Thursday, day after holiday) at 12:00, got closed")
	}

	// Jan 4, 2024 at 09:00 should be closed (before PH +1 day hours)
	dayAfterHolidayEarly := time.Date(2024, 1, 4, 9, 0, 0, 0, time.UTC)
	if oh.GetState(dayAfterHolidayEarly) {
		t.Errorf("expected closed on Jan 4 (Thursday, day after holiday) at 09:00, got open")
	}

	// Jan 4, 2024 at 15:00 should be closed (after PH +1 day hours)
	dayAfterHolidayLate := time.Date(2024, 1, 4, 15, 0, 0, 0, time.UTC)
	if oh.GetState(dayAfterHolidayLate) {
		t.Errorf("expected closed on Jan 4 (Thursday, day after holiday) at 15:00, got open")
	}

	// Jan 5, 2024 (Friday, regular weekday) should follow normal hours (open at 12:00)
	regularFriday := time.Date(2024, 1, 5, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(regularFriday) {
		t.Errorf("expected open on Jan 5 (Friday, regular weekday) at 12:00, got closed")
	}

	// Jan 5, 2024 at 09:00 should be open (normal weekday hours)
	regularFridayStart := time.Date(2024, 1, 5, 9, 0, 0, 0, time.UTC)
	if !oh.GetState(regularFridayStart) {
		t.Errorf("expected open on Jan 5 (Friday, regular weekday) at 09:00, got closed")
	}

	// Jan 5, 2024 at 16:00 should be open (normal weekday hours)
	regularFridayLate := time.Date(2024, 1, 5, 16, 0, 0, 0, time.UTC)
	if !oh.GetState(regularFridayLate) {
		t.Errorf("expected open on Jan 5 (Friday, regular weekday) at 16:00, got closed")
	}
}

// TestHolidayOffset_NoChecker tests that offset rules don't match when no checker is set
func TestHolidayOffset_NoChecker(t *testing.T) {
	// When no checker is set, PH offset rules should not match (similar to PH behavior)
	oh, err := New("PH +1 day 10:00-14:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// No checker set, so any time should be closed
	testTime := time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)
	if oh.GetState(testTime) {
		t.Errorf("expected closed when no holiday checker is set, got open")
	}
}

// TestHolidayOffset_LargerOffset tests larger offset values like PH +2 day, PH -2 day
func TestHolidayOffset_LargerOffset(t *testing.T) {
	// Input: "PH +2 day 10:00-14:00"
	// This means: two days after a public holiday has hours 10:00-14:00
	oh, err := New("PH +2 day 10:00-14:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Set up mock: Jan 1, 2024 is a holiday (Monday)
	hChecker := &mockHolidayChecker{
		holidays: map[string]bool{
			"2024-01-01": true, // Monday
		},
	}
	oh.SetHolidayChecker(hChecker)

	// Jan 3, 2024 (Wednesday, two days after Monday holiday) should be open at 12:00
	twoDaysAfterHoliday := time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(twoDaysAfterHoliday) {
		t.Errorf("expected open on Jan 3 (two days after holiday) at 12:00, got closed")
	}

	// Jan 2, 2024 (Tuesday, one day after Monday holiday) should be closed
	oneDayAfterHoliday := time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)
	if oh.GetState(oneDayAfterHoliday) {
		t.Errorf("expected closed on Jan 2 (one day after holiday) at 12:00, got open")
	}

	// Jan 4, 2024 (Thursday, three days after Monday holiday) should be closed
	threeDaysAfterHoliday := time.Date(2024, 1, 4, 12, 0, 0, 0, time.UTC)
	if oh.GetState(threeDaysAfterHoliday) {
		t.Errorf("expected closed on Jan 4 (three days after holiday) at 12:00, got open")
	}
}

// TestHolidayOffset_NegativeLargerOffset tests larger negative offset values
func TestHolidayOffset_NegativeLargerOffset(t *testing.T) {
	// Input: "Mo-Fr 09:00-17:00; PH -2 day off"
	// This means: weekdays 09:00-17:00, but two days before a public holiday is closed
	oh, err := New("Mo-Fr 09:00-17:00; PH -2 day off")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Set up mock: Jan 5, 2024 is a holiday (Friday)
	hChecker := &mockHolidayChecker{
		holidays: map[string]bool{
			"2024-01-05": true, // Friday
		},
	}
	oh.SetHolidayChecker(hChecker)

	// Jan 3, 2024 (Wednesday, two days before Friday holiday) should be closed (PH -2 day off)
	twoDaysBeforeHoliday := time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC)
	if oh.GetState(twoDaysBeforeHoliday) {
		t.Errorf("expected closed on Jan 3 (Wednesday, two days before holiday) at 12:00, got open")
	}

	// Jan 4, 2024 (Thursday, one day before Friday holiday) should follow normal rules (open)
	oneDayBeforeHoliday := time.Date(2024, 1, 4, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(oneDayBeforeHoliday) {
		t.Errorf("expected open on Jan 4 (Thursday, one day before holiday) at 12:00, got closed")
	}

	// Jan 2, 2024 (Tuesday, three days before Friday holiday) should follow normal rules (open)
	threeDaysBeforeHoliday := time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(threeDaysBeforeHoliday) {
		t.Errorf("expected open on Jan 2 (Tuesday, three days before holiday) at 12:00, got closed")
	}
}
