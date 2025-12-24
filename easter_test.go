package openinghours

import (
	"testing"
	"time"
)

// TestEaster_EasterSunday tests the basic "easter HH:MM-HH:MM" syntax for Easter Sunday
func TestEaster_EasterSunday(t *testing.T) {
	// Input: "easter 10:00-14:00"
	// This means: Easter Sunday has hours 10:00-14:00
	oh, err := New("easter 10:00-14:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// March 31, 2024 at 12:00 should be open (Easter Sunday 2024)
	easter2024 := time.Date(2024, 3, 31, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(easter2024) {
		t.Errorf("expected open on March 31, 2024 at 12:00 (Easter Sunday 2024), got closed")
	}

	// April 9, 2023 at 12:00 should be open (Easter Sunday 2023)
	easter2023 := time.Date(2023, 4, 9, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(easter2023) {
		t.Errorf("expected open on April 9, 2023 at 12:00 (Easter Sunday 2023), got closed")
	}

	// April 20, 2025 at 12:00 should be open (Easter Sunday 2025)
	easter2025 := time.Date(2025, 4, 20, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(easter2025) {
		t.Errorf("expected open on April 20, 2025 at 12:00 (Easter Sunday 2025), got closed")
	}

	// March 30, 2024 at 12:00 should be closed (not Easter)
	notEaster := time.Date(2024, 3, 30, 12, 0, 0, 0, time.UTC)
	if oh.GetState(notEaster) {
		t.Errorf("expected closed on March 30, 2024 at 12:00 (not Easter), got open")
	}

	// March 31, 2024 at 10:00 should be open (start time)
	easterStart := time.Date(2024, 3, 31, 10, 0, 0, 0, time.UTC)
	if !oh.GetState(easterStart) {
		t.Errorf("expected open on March 31, 2024 at 10:00 (Easter start time), got closed")
	}

	// March 31, 2024 at 09:00 should be closed (before opening)
	easterBeforeOpen := time.Date(2024, 3, 31, 9, 0, 0, 0, time.UTC)
	if oh.GetState(easterBeforeOpen) {
		t.Errorf("expected closed on March 31, 2024 at 09:00 (before opening), got open")
	}

	// March 31, 2024 at 14:00 should be closed (end time, exclusive)
	easterEnd := time.Date(2024, 3, 31, 14, 0, 0, 0, time.UTC)
	if oh.GetState(easterEnd) {
		t.Errorf("expected closed on March 31, 2024 at 14:00 (end time), got open")
	}

	// March 31, 2024 at 15:00 should be closed (after closing)
	easterAfterClose := time.Date(2024, 3, 31, 15, 0, 0, 0, time.UTC)
	if oh.GetState(easterAfterClose) {
		t.Errorf("expected closed on March 31, 2024 at 15:00 (after closing), got open")
	}
}

// TestEaster_EasterMonday tests the "easter +1 day HH:MM-HH:MM" syntax for Easter Monday
func TestEaster_EasterMonday(t *testing.T) {
	// Input: "easter +1 day 09:00-17:00"
	// This means: Easter Monday (day after Easter Sunday) has hours 09:00-17:00
	oh, err := New("easter +1 day 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// April 1, 2024 at 10:00 should be open (Easter Monday 2024)
	// Easter 2024 is March 31, so April 1 is Easter Monday
	easterMonday2024 := time.Date(2024, 4, 1, 10, 0, 0, 0, time.UTC)
	if !oh.GetState(easterMonday2024) {
		t.Errorf("expected open on April 1, 2024 at 10:00 (Easter Monday 2024), got closed")
	}

	// April 10, 2023 at 10:00 should be open (Easter Monday 2023)
	// Easter 2023 is April 9, so April 10 is Easter Monday
	easterMonday2023 := time.Date(2023, 4, 10, 10, 0, 0, 0, time.UTC)
	if !oh.GetState(easterMonday2023) {
		t.Errorf("expected open on April 10, 2023 at 10:00 (Easter Monday 2023), got closed")
	}

	// March 31, 2024 at 10:00 should be closed (Easter Sunday, not +1 day)
	easterSunday := time.Date(2024, 3, 31, 10, 0, 0, 0, time.UTC)
	if oh.GetState(easterSunday) {
		t.Errorf("expected closed on March 31, 2024 at 10:00 (Easter Sunday, not Monday), got open")
	}

	// April 2, 2024 at 10:00 should be closed (two days after Easter)
	twoDaysAfter := time.Date(2024, 4, 2, 10, 0, 0, 0, time.UTC)
	if oh.GetState(twoDaysAfter) {
		t.Errorf("expected closed on April 2, 2024 at 10:00 (two days after Easter), got open")
	}

	// April 1, 2024 at 09:00 should be open (start time)
	easterMondayStart := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)
	if !oh.GetState(easterMondayStart) {
		t.Errorf("expected open on April 1, 2024 at 09:00 (Easter Monday start time), got closed")
	}

	// April 1, 2024 at 08:00 should be closed (before opening)
	easterMondayEarly := time.Date(2024, 4, 1, 8, 0, 0, 0, time.UTC)
	if oh.GetState(easterMondayEarly) {
		t.Errorf("expected closed on April 1, 2024 at 08:00 (before opening), got open")
	}

	// April 1, 2024 at 17:00 should be closed (end time, exclusive)
	easterMondayEnd := time.Date(2024, 4, 1, 17, 0, 0, 0, time.UTC)
	if oh.GetState(easterMondayEnd) {
		t.Errorf("expected closed on April 1, 2024 at 17:00 (end time), got open")
	}
}

// TestEaster_GoodFriday tests the "easter -2 days off" syntax for Good Friday
func TestEaster_GoodFriday(t *testing.T) {
	// Input: "easter -2 days off"
	// This means: Good Friday (2 days before Easter Sunday) is closed
	oh, err := New("easter -2 days off")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// March 29, 2024 at 10:00 should be closed (Good Friday 2024)
	// Easter 2024 is March 31, so March 29 is Good Friday
	goodFriday2024 := time.Date(2024, 3, 29, 10, 0, 0, 0, time.UTC)
	if oh.GetState(goodFriday2024) {
		t.Errorf("expected closed on March 29, 2024 at 10:00 (Good Friday 2024), got open")
	}

	// April 7, 2023 at 10:00 should be closed (Good Friday 2023)
	// Easter 2023 is April 9, so April 7 is Good Friday
	goodFriday2023 := time.Date(2023, 4, 7, 10, 0, 0, 0, time.UTC)
	if oh.GetState(goodFriday2023) {
		t.Errorf("expected closed on April 7, 2023 at 10:00 (Good Friday 2023), got open")
	}

	// March 31, 2024 at 10:00 should be open (Easter Sunday, not Good Friday)
	easterSunday := time.Date(2024, 3, 31, 10, 0, 0, 0, time.UTC)
	if oh.GetState(easterSunday) {
		t.Errorf("expected closed on March 31, 2024 at 10:00 (Easter Sunday, not Good Friday), got open")
	}

	// March 28, 2024 at 10:00 should be open (three days before Easter)
	threeDaysBefore := time.Date(2024, 3, 28, 10, 0, 0, 0, time.UTC)
	if oh.GetState(threeDaysBefore) {
		t.Errorf("expected closed on March 28, 2024 at 10:00 (three days before Easter), got open")
	}

	// March 30, 2024 at 10:00 should be open (one day before Easter, not Good Friday)
	oneDayBefore := time.Date(2024, 3, 30, 10, 0, 0, 0, time.UTC)
	if oh.GetState(oneDayBefore) {
		t.Errorf("expected closed on March 30, 2024 at 10:00 (one day before Easter), got open")
	}
}

// TestEaster_WithWeekdayRule tests combining Easter rules with weekday rules
func TestEaster_WithWeekdayRule(t *testing.T) {
	// Input: "Mo-Fr 09:00-17:00; easter off; easter +1 day off"
	// This means: weekdays 09:00-17:00, but Easter Sunday and Easter Monday are closed
	oh, err := New("Mo-Fr 09:00-17:00; easter off; easter +1 day off")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// March 31, 2024 is Easter Sunday 2024 (Sunday)
	// Should be closed (easter off rule, even though it's Sunday which wouldn't match Mo-Fr anyway)
	easterSunday := time.Date(2024, 3, 31, 12, 0, 0, 0, time.UTC)
	if oh.GetState(easterSunday) {
		t.Errorf("expected closed on March 31, 2024 at 12:00 (Easter Sunday, off rule), got open")
	}

	// April 1, 2024 is Easter Monday 2024 (Monday)
	// Should be closed (easter +1 day off rule, overrides Mo-Fr rule)
	easterMonday := time.Date(2024, 4, 1, 12, 0, 0, 0, time.UTC)
	if oh.GetState(easterMonday) {
		t.Errorf("expected closed on April 1, 2024 at 12:00 (Easter Monday, off rule), got open")
	}

	// April 2, 2024 (Tuesday, regular weekday after Easter)
	// Should be open (regular Mo-Fr rule applies)
	regularTuesday := time.Date(2024, 4, 2, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(regularTuesday) {
		t.Errorf("expected open on April 2, 2024 at 12:00 (regular Tuesday), got closed")
	}

	// March 28, 2024 (Thursday, before Easter)
	// Should be open (regular Mo-Fr rule applies)
	regularThursday := time.Date(2024, 3, 28, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(regularThursday) {
		t.Errorf("expected open on March 28, 2024 at 12:00 (regular Thursday), got closed")
	}

	// April 6, 2024 (Saturday, weekend)
	// Should be closed (not a weekday)
	saturday := time.Date(2024, 4, 6, 12, 0, 0, 0, time.UTC)
	if oh.GetState(saturday) {
		t.Errorf("expected closed on April 6, 2024 at 12:00 (Saturday), got open")
	}
}

// TestEaster_DateRange tests the "easter -2 days-easter +1 day off" syntax for date ranges
func TestEaster_DateRange(t *testing.T) {
	// Input: "easter -2 days-easter +1 day off"
	// This means: from Good Friday through Easter Monday (4 days total) is closed
	oh, err := New("easter -2 days-easter +1 day off")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Easter 2024 is March 31
	// Good Friday (easter -2 days) is March 29
	// Easter Monday (easter +1 day) is April 1

	// March 29, 2024 (Good Friday) should be closed
	goodFriday := time.Date(2024, 3, 29, 12, 0, 0, 0, time.UTC)
	if oh.GetState(goodFriday) {
		t.Errorf("expected closed on March 29, 2024 at 12:00 (Good Friday), got open")
	}

	// March 30, 2024 (Saturday, day before Easter) should be closed
	saturdayBeforeEaster := time.Date(2024, 3, 30, 12, 0, 0, 0, time.UTC)
	if oh.GetState(saturdayBeforeEaster) {
		t.Errorf("expected closed on March 30, 2024 at 12:00 (Saturday before Easter), got open")
	}

	// March 31, 2024 (Easter Sunday) should be closed
	easterSunday := time.Date(2024, 3, 31, 12, 0, 0, 0, time.UTC)
	if oh.GetState(easterSunday) {
		t.Errorf("expected closed on March 31, 2024 at 12:00 (Easter Sunday), got open")
	}

	// April 1, 2024 (Easter Monday) should be closed
	easterMonday := time.Date(2024, 4, 1, 12, 0, 0, 0, time.UTC)
	if oh.GetState(easterMonday) {
		t.Errorf("expected closed on April 1, 2024 at 12:00 (Easter Monday), got open")
	}

	// March 28, 2024 (Thursday, before Good Friday) should be open
	beforeRange := time.Date(2024, 3, 28, 12, 0, 0, 0, time.UTC)
	if oh.GetState(beforeRange) {
		t.Errorf("expected closed on March 28, 2024 at 12:00 (before Easter range), got open")
	}

	// April 2, 2024 (Tuesday, after Easter Monday) should be open
	afterRange := time.Date(2024, 4, 2, 12, 0, 0, 0, time.UTC)
	if oh.GetState(afterRange) {
		t.Errorf("expected closed on April 2, 2024 at 12:00 (after Easter range), got open")
	}
}

// TestEaster_AshWednesday tests larger negative offset for Ash Wednesday
func TestEaster_AshWednesday(t *testing.T) {
	// Input: "easter -49 days off"
	// This means: Ash Wednesday (49 days before Easter) is closed
	oh, err := New("easter -49 days off")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Easter 2024 is March 31
	// Ash Wednesday is 49 days before = February 11, 2024
	ashWednesday2024 := time.Date(2024, 2, 11, 12, 0, 0, 0, time.UTC)
	if oh.GetState(ashWednesday2024) {
		t.Errorf("expected closed on February 11, 2024 at 12:00 (Ash Wednesday 2024), got open")
	}

	// Easter 2023 is April 9
	// Ash Wednesday is 49 days before = February 19, 2023
	ashWednesday2023 := time.Date(2023, 2, 19, 12, 0, 0, 0, time.UTC)
	if oh.GetState(ashWednesday2023) {
		t.Errorf("expected closed on February 19, 2023 at 12:00 (Ash Wednesday 2023), got open")
	}

	// March 31, 2024 (Easter Sunday) should be open (not Ash Wednesday)
	easterSunday := time.Date(2024, 3, 31, 12, 0, 0, 0, time.UTC)
	if oh.GetState(easterSunday) {
		t.Errorf("expected closed on March 31, 2024 at 12:00 (Easter Sunday, not Ash Wednesday), got open")
	}

	// February 12, 2024 (day after Ash Wednesday) should be open
	dayAfterAsh := time.Date(2024, 2, 12, 12, 0, 0, 0, time.UTC)
	if oh.GetState(dayAfterAsh) {
		t.Errorf("expected closed on February 12, 2024 at 12:00 (day after Ash Wednesday), got open")
	}
}

// TestEaster_MultipleYears tests that Easter calculations work correctly across multiple years
func TestEaster_MultipleYears(t *testing.T) {
	// Input: "easter 10:00-14:00"
	oh, err := New("easter 10:00-14:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Test Easter dates for various years
	testCases := []struct {
		year  int
		month time.Month
		day   int
		desc  string
	}{
		{2020, time.April, 12, "Easter 2020"},
		{2021, time.April, 4, "Easter 2021"},
		{2022, time.April, 17, "Easter 2022"},
		{2023, time.April, 9, "Easter 2023"},
		{2024, time.March, 31, "Easter 2024"},
		{2025, time.April, 20, "Easter 2025"},
		{2026, time.April, 5, "Easter 2026"},
	}

	for _, tc := range testCases {
		easterDate := time.Date(tc.year, tc.month, tc.day, 12, 0, 0, 0, time.UTC)
		if !oh.GetState(easterDate) {
			t.Errorf("expected open on %s (%v at 12:00), got closed", tc.desc, easterDate.Format("2006-01-02"))
		}

		// Day before should be closed
		dayBefore := easterDate.AddDate(0, 0, -1)
		if oh.GetState(dayBefore) {
			t.Errorf("expected closed on day before %s (%v at 12:00), got open", tc.desc, dayBefore.Format("2006-01-02"))
		}

		// Day after should be closed
		dayAfter := easterDate.AddDate(0, 0, 1)
		if oh.GetState(dayAfter) {
			t.Errorf("expected closed on day after %s (%v at 12:00), got open", tc.desc, dayAfter.Format("2006-01-02"))
		}
	}
}

// TestEaster_CombinedWithOtherRules tests Easter rules combined with various other rule types
func TestEaster_CombinedWithOtherRules(t *testing.T) {
	// Input: "Mo-Fr 09:00-17:00; Sa 10:00-14:00; easter -2 days-easter +1 day off"
	// This means: weekday and Saturday hours, but Good Friday through Easter Monday is closed
	oh, err := New("Mo-Fr 09:00-17:00; Sa 10:00-14:00; easter -2 days-easter +1 day off")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Easter 2024 is March 31 (Sunday)
	// Good Friday is March 29 (Friday)
	// Easter Monday is April 1 (Monday)

	// March 29, 2024 (Good Friday) should be closed (Easter range overrides Mo-Fr rule)
	goodFriday := time.Date(2024, 3, 29, 12, 0, 0, 0, time.UTC)
	if oh.GetState(goodFriday) {
		t.Errorf("expected closed on March 29, 2024 at 12:00 (Good Friday, Easter override), got open")
	}

	// March 30, 2024 (Saturday before Easter) should be closed (Easter range overrides Sa rule)
	saturdayBeforeEaster := time.Date(2024, 3, 30, 12, 0, 0, 0, time.UTC)
	if oh.GetState(saturdayBeforeEaster) {
		t.Errorf("expected closed on March 30, 2024 at 12:00 (Saturday, Easter override), got open")
	}

	// April 1, 2024 (Easter Monday) should be closed (Easter range overrides Mo-Fr rule)
	easterMonday := time.Date(2024, 4, 1, 12, 0, 0, 0, time.UTC)
	if oh.GetState(easterMonday) {
		t.Errorf("expected closed on April 1, 2024 at 12:00 (Easter Monday, Easter override), got open")
	}

	// March 28, 2024 (Thursday before Good Friday) should be open (normal Mo-Fr rule)
	thursday := time.Date(2024, 3, 28, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(thursday) {
		t.Errorf("expected open on March 28, 2024 at 12:00 (Thursday, normal hours), got closed")
	}

	// April 2, 2024 (Tuesday after Easter Monday) should be open (normal Mo-Fr rule)
	tuesday := time.Date(2024, 4, 2, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(tuesday) {
		t.Errorf("expected open on April 2, 2024 at 12:00 (Tuesday, normal hours), got closed")
	}

	// April 6, 2024 (Saturday after Easter) should be open with Saturday hours
	saturday := time.Date(2024, 4, 6, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(saturday) {
		t.Errorf("expected open on April 6, 2024 at 12:00 (Saturday, normal hours), got closed")
	}

	// April 6, 2024 at 15:00 should be closed (after Saturday closing time)
	saturdayLate := time.Date(2024, 4, 6, 15, 0, 0, 0, time.UTC)
	if oh.GetState(saturdayLate) {
		t.Errorf("expected closed on April 6, 2024 at 15:00 (after Saturday hours), got open")
	}
}

// TestEaster_PluralDays tests both "day" and "days" syntax
func TestEaster_PluralDays(t *testing.T) {
	// Both "easter +1 day" and "easter +1 days" should work
	testCases := []string{
		"easter +1 day 09:00-17:00",
		"easter +1 days 09:00-17:00",
	}

	for _, input := range testCases {
		oh, err := New(input)
		if err != nil {
			t.Fatalf("unexpected parse error for %q: %v", input, err)
		}

		// April 1, 2024 at 10:00 should be open (Easter Monday 2024)
		easterMonday := time.Date(2024, 4, 1, 10, 0, 0, 0, time.UTC)
		if !oh.GetState(easterMonday) {
			t.Errorf("for input %q: expected open on April 1, 2024 at 10:00 (Easter Monday 2024), got closed", input)
		}
	}
}

// TestEaster_EdgeCases tests edge cases and boundary conditions
func TestEaster_EdgeCases(t *testing.T) {
	// Test easter +0 day (same as just "easter")
	oh, err := New("easter +0 day 10:00-14:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	easterSunday := time.Date(2024, 3, 31, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(easterSunday) {
		t.Errorf("expected open on March 31, 2024 at 12:00 (Easter +0 day), got closed")
	}

	// Test negative offset without explicit minus sign in alternate format
	oh2, err := New("easter -1 day off")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// March 30, 2024 is one day before Easter 2024
	dayBeforeEaster := time.Date(2024, 3, 30, 12, 0, 0, 0, time.UTC)
	if oh2.GetState(dayBeforeEaster) {
		t.Errorf("expected closed on March 30, 2024 at 12:00 (Easter -1 day), got open")
	}
}
