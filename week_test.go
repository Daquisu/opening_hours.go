package openinghours

import (
	"testing"
	"time"
)

func TestWeekNumber_SingleWeek(t *testing.T) {
	// week 02 Mo-Fr 09:00-17:00
	// January 2024: Week 2 is Jan 8-14 (ISO week numbering)
	oh, err := New("week 02 Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		date time.Time
		want bool
		desc string
	}{
		// Week 2, Monday Jan 8, 2024 at 10:00 - should be open
		{time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC), true, "Week 2 (Jan 8, Monday) at 10:00 - should be open"},
		// Week 2, Tuesday Jan 9, 2024 at 10:00 - should be open
		{time.Date(2024, 1, 9, 10, 0, 0, 0, time.UTC), true, "Week 2 (Jan 9, Tuesday) at 10:00 - should be open"},
		// Week 2, Friday Jan 12, 2024 at 10:00 - should be open
		{time.Date(2024, 1, 12, 10, 0, 0, 0, time.UTC), true, "Week 2 (Jan 12, Friday) at 10:00 - should be open"},
		// Week 2, Saturday Jan 13, 2024 at 10:00 - should be closed (weekend)
		{time.Date(2024, 1, 13, 10, 0, 0, 0, time.UTC), false, "Week 2 (Jan 13, Saturday) at 10:00 - should be closed (weekend)"},
		// Week 3, Monday Jan 15, 2024 at 10:00 - should be closed (different week)
		{time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC), false, "Week 3 (Jan 15, Monday) at 10:00 - should be closed (week 3)"},
		// Week 1, Monday Jan 1, 2024 at 10:00 - should be closed (different week)
		{time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), false, "Week 1 (Jan 1, Monday) at 10:00 - should be closed (week 1)"},
		// Week 2, Monday Jan 8 at 08:00 - should be closed (before opening time)
		{time.Date(2024, 1, 8, 8, 0, 0, 0, time.UTC), false, "Week 2 (Jan 8, Monday) at 08:00 - should be closed (before opening time)"},
		// Week 2, Monday Jan 8 at 18:00 - should be closed (after closing time)
		{time.Date(2024, 1, 8, 18, 0, 0, 0, time.UTC), false, "Week 2 (Jan 8, Monday) at 18:00 - should be closed (after closing time)"},
	}

	for _, tt := range tests {
		got := oh.GetState(tt.date)
		if got != tt.want {
			year, week := tt.date.ISOWeek()
			t.Errorf("%s: got %v, want %v (ISO week: %d-%02d)", tt.desc, got, tt.want, year, week)
		}
	}
}

func TestWeekNumber_WeekRange(t *testing.T) {
	// week 01-10 Mo-Fr 09:00-17:00
	// Tests that weeks 1-10 are open on weekdays
	oh, err := New("week 01-10 Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		date time.Time
		want bool
		desc string
	}{
		// Week 1, Monday Jan 1, 2024 at 10:00 - should be open
		{time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), true, "Week 1 (Jan 1, Monday) at 10:00 - should be open"},
		// Week 2, Wednesday Jan 10, 2024 at 10:00 - should be open
		{time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC), true, "Week 2 (Jan 10, Wednesday) at 10:00 - should be open"},
		// Week 5, Monday Jan 29, 2024 at 10:00 - should be open
		{time.Date(2024, 1, 29, 10, 0, 0, 0, time.UTC), true, "Week 5 (Jan 29, Monday) at 10:00 - should be open"},
		// Week 10, Friday Mar 8, 2024 at 10:00 - should be open
		{time.Date(2024, 3, 8, 10, 0, 0, 0, time.UTC), true, "Week 10 (Mar 8, Friday) at 10:00 - should be open"},
		// Week 11, Monday Mar 11, 2024 at 10:00 - should be closed (week 11, outside range)
		{time.Date(2024, 3, 11, 10, 0, 0, 0, time.UTC), false, "Week 11 (Mar 11, Monday) at 10:00 - should be closed (week 11)"},
		// Week 15, Monday Apr 8, 2024 at 10:00 - should be closed (week 15, outside range)
		{time.Date(2024, 4, 8, 10, 0, 0, 0, time.UTC), false, "Week 15 (Apr 8, Monday) at 10:00 - should be closed (week 15)"},
		// Week 1, Saturday Jan 6, 2024 at 10:00 - should be closed (weekend)
		{time.Date(2024, 1, 6, 10, 0, 0, 0, time.UTC), false, "Week 1 (Jan 6, Saturday) at 10:00 - should be closed (weekend)"},
		// Week 5, Sunday Feb 4, 2024 at 10:00 - should be closed (weekend)
		{time.Date(2024, 2, 4, 10, 0, 0, 0, time.UTC), false, "Week 5 (Feb 4, Sunday) at 10:00 - should be closed (weekend)"},
	}

	for _, tt := range tests {
		got := oh.GetState(tt.date)
		if got != tt.want {
			year, week := tt.date.ISOWeek()
			t.Errorf("%s: got %v, want %v (ISO week: %d-%02d)", tt.desc, got, tt.want, year, week)
		}
	}
}

func TestWeekNumber_WeekInterval(t *testing.T) {
	// week 01-53/2 Sa 10:00-14:00
	// Tests odd weeks (1, 3, 5, ..., 53) are open on Saturday
	oh, err := New("week 01-53/2 Sa 10:00-14:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		date time.Time
		want bool
		desc string
	}{
		// Week 1, Saturday Jan 6, 2024 at 12:00 - should be open (odd week)
		{time.Date(2024, 1, 6, 12, 0, 0, 0, time.UTC), true, "Week 1 (Jan 6, Saturday) at 12:00 - should be open (odd week)"},
		// Week 3, Saturday Jan 20, 2024 at 12:00 - should be open (odd week)
		{time.Date(2024, 1, 20, 12, 0, 0, 0, time.UTC), true, "Week 3 (Jan 20, Saturday) at 12:00 - should be open (odd week)"},
		// Week 5, Saturday Feb 3, 2024 at 12:00 - should be open (odd week)
		{time.Date(2024, 2, 3, 12, 0, 0, 0, time.UTC), true, "Week 5 (Feb 3, Saturday) at 12:00 - should be open (odd week)"},
		// Week 7, Saturday Feb 17, 2024 at 12:00 - should be open (odd week)
		{time.Date(2024, 2, 17, 12, 0, 0, 0, time.UTC), true, "Week 7 (Feb 17, Saturday) at 12:00 - should be open (odd week)"},
		// Week 2, Saturday Jan 13, 2024 at 12:00 - should be closed (even week)
		{time.Date(2024, 1, 13, 12, 0, 0, 0, time.UTC), false, "Week 2 (Jan 13, Saturday) at 12:00 - should be closed (even week)"},
		// Week 4, Saturday Jan 27, 2024 at 12:00 - should be closed (even week)
		{time.Date(2024, 1, 27, 12, 0, 0, 0, time.UTC), false, "Week 4 (Jan 27, Saturday) at 12:00 - should be closed (even week)"},
		// Week 6, Saturday Feb 10, 2024 at 12:00 - should be closed (even week)
		{time.Date(2024, 2, 10, 12, 0, 0, 0, time.UTC), false, "Week 6 (Feb 10, Saturday) at 12:00 - should be closed (even week)"},
		// Week 8, Saturday Feb 24, 2024 at 12:00 - should be closed (even week)
		{time.Date(2024, 2, 24, 12, 0, 0, 0, time.UTC), false, "Week 8 (Feb 24, Saturday) at 12:00 - should be closed (even week)"},
		// Week 1, Monday Jan 1, 2024 at 12:00 - should be closed (not Saturday)
		{time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), false, "Week 1 (Jan 1, Monday) at 12:00 - should be closed (not Saturday)"},
		// Week 1, Saturday Jan 6 at 09:00 - should be closed (before opening time)
		{time.Date(2024, 1, 6, 9, 0, 0, 0, time.UTC), false, "Week 1 (Jan 6, Saturday) at 09:00 - should be closed (before opening time)"},
		// Week 1, Saturday Jan 6 at 15:00 - should be closed (after closing time)
		{time.Date(2024, 1, 6, 15, 0, 0, 0, time.UTC), false, "Week 1 (Jan 6, Saturday) at 15:00 - should be closed (after closing time)"},
	}

	for _, tt := range tests {
		got := oh.GetState(tt.date)
		if got != tt.want {
			year, week := tt.date.ISOWeek()
			t.Errorf("%s: got %v, want %v (ISO week: %d-%02d)", tt.desc, got, tt.want, year, week)
		}
	}
}

func TestWeekNumber_WithOtherRules(t *testing.T) {
	// Mo-Fr 09:00-17:00; week 01 off
	// Normal weekday hours, but week 1 is closed (override)
	oh, err := New("Mo-Fr 09:00-17:00; week 01 off")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		date time.Time
		want bool
		desc string
	}{
		// Week 1, Monday Jan 1, 2024 at 10:00 - should be closed (week 1 off)
		{time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), false, "Week 1 (Jan 1, Monday) at 10:00 - should be closed (week 1 off)"},
		// Week 1, Tuesday Jan 2, 2024 at 10:00 - should be closed (week 1 off)
		{time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC), false, "Week 1 (Jan 2, Tuesday) at 10:00 - should be closed (week 1 off)"},
		// Week 1, Friday Jan 5, 2024 at 10:00 - should be closed (week 1 off)
		{time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC), false, "Week 1 (Jan 5, Friday) at 10:00 - should be closed (week 1 off)"},
		// Week 2, Monday Jan 8, 2024 at 10:00 - should be open (normal hours)
		{time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC), true, "Week 2 (Jan 8, Monday) at 10:00 - should be open (normal hours)"},
		// Week 3, Wednesday Jan 17, 2024 at 10:00 - should be open (normal hours)
		{time.Date(2024, 1, 17, 10, 0, 0, 0, time.UTC), true, "Week 3 (Jan 17, Wednesday) at 10:00 - should be open (normal hours)"},
		// Week 10, Friday Mar 8, 2024 at 10:00 - should be open (normal hours)
		{time.Date(2024, 3, 8, 10, 0, 0, 0, time.UTC), true, "Week 10 (Mar 8, Friday) at 10:00 - should be open (normal hours)"},
		// Week 2, Saturday Jan 13, 2024 at 10:00 - should be closed (weekend)
		{time.Date(2024, 1, 13, 10, 0, 0, 0, time.UTC), false, "Week 2 (Jan 13, Saturday) at 10:00 - should be closed (weekend)"},
	}

	for _, tt := range tests {
		got := oh.GetState(tt.date)
		if got != tt.want {
			year, week := tt.date.ISOWeek()
			t.Errorf("%s: got %v, want %v (ISO week: %d-%02d)", tt.desc, got, tt.want, year, week)
		}
	}
}

func TestWeekNumber_YearBoundary(t *testing.T) {
	// week 52-53,01-02 Sa-Su 10:00-18:00
	// Tests week behavior at year boundaries
	// ISO week 1 of 2024 starts on Jan 1, 2024 (Monday)
	// ISO week 52 of 2023 is Dec 25-31, 2023
	oh, err := New("week 52-53,01-02 Sa-Su 10:00-18:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		date time.Time
		want bool
		desc string
	}{
		// Week 52 of 2023, Saturday Dec 30, 2023 at 12:00 - should be open
		{time.Date(2023, 12, 30, 12, 0, 0, 0, time.UTC), true, "Week 52 (Dec 30, 2023, Saturday) at 12:00 - should be open"},
		// Week 52 of 2023, Sunday Dec 31, 2023 at 12:00 - should be open
		{time.Date(2023, 12, 31, 12, 0, 0, 0, time.UTC), true, "Week 52 (Dec 31, 2023, Sunday) at 12:00 - should be open"},
		// Week 1 of 2024, Saturday Jan 6, 2024 at 12:00 - should be open
		{time.Date(2024, 1, 6, 12, 0, 0, 0, time.UTC), true, "Week 1 (Jan 6, 2024, Saturday) at 12:00 - should be open"},
		// Week 1 of 2024, Sunday Jan 7, 2024 at 12:00 - should be open
		{time.Date(2024, 1, 7, 12, 0, 0, 0, time.UTC), true, "Week 1 (Jan 7, 2024, Sunday) at 12:00 - should be open"},
		// Week 2 of 2024, Saturday Jan 13, 2024 at 12:00 - should be open
		{time.Date(2024, 1, 13, 12, 0, 0, 0, time.UTC), true, "Week 2 (Jan 13, 2024, Saturday) at 12:00 - should be open"},
		// Week 3 of 2024, Saturday Jan 20, 2024 at 12:00 - should be closed (week 3 not in range)
		{time.Date(2024, 1, 20, 12, 0, 0, 0, time.UTC), false, "Week 3 (Jan 20, 2024, Saturday) at 12:00 - should be closed (week 3)"},
		// Week 51 of 2023, Saturday Dec 23, 2023 at 12:00 - should be closed (week 51 not in range)
		{time.Date(2023, 12, 23, 12, 0, 0, 0, time.UTC), false, "Week 51 (Dec 23, 2023, Saturday) at 12:00 - should be closed (week 51)"},
		// Week 1 of 2024, Monday Jan 1, 2024 at 12:00 - should be closed (not weekend)
		{time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), false, "Week 1 (Jan 1, 2024, Monday) at 12:00 - should be closed (not weekend)"},
	}

	for _, tt := range tests {
		got := oh.GetState(tt.date)
		if got != tt.want {
			year, week := tt.date.ISOWeek()
			t.Errorf("%s: got %v, want %v (ISO week: %d-%02d)", tt.desc, got, tt.want, year, week)
		}
	}
}

func TestWeekNumber_MultipleWeeks(t *testing.T) {
	// week 01,10,20,30 Mo 09:00-17:00
	// Tests specific weeks (not a range)
	oh, err := New("week 01,10,20,30 Mo 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		date time.Time
		want bool
		desc string
	}{
		// Week 1, Monday Jan 1, 2024 at 10:00 - should be open
		{time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), true, "Week 1 (Jan 1, Monday) at 10:00 - should be open"},
		// Week 10, Monday Mar 4, 2024 at 10:00 - should be open
		{time.Date(2024, 3, 4, 10, 0, 0, 0, time.UTC), true, "Week 10 (Mar 4, Monday) at 10:00 - should be open"},
		// Week 20, Monday May 13, 2024 at 10:00 - should be open
		{time.Date(2024, 5, 13, 10, 0, 0, 0, time.UTC), true, "Week 20 (May 13, Monday) at 10:00 - should be open"},
		// Week 30, Monday Jul 22, 2024 at 10:00 - should be open
		{time.Date(2024, 7, 22, 10, 0, 0, 0, time.UTC), true, "Week 30 (Jul 22, Monday) at 10:00 - should be open"},
		// Week 2, Monday Jan 8, 2024 at 10:00 - should be closed (not in list)
		{time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC), false, "Week 2 (Jan 8, Monday) at 10:00 - should be closed (not in list)"},
		// Week 15, Monday Apr 8, 2024 at 10:00 - should be closed (not in list)
		{time.Date(2024, 4, 8, 10, 0, 0, 0, time.UTC), false, "Week 15 (Apr 8, Monday) at 10:00 - should be closed (not in list)"},
		// Week 1, Tuesday Jan 2, 2024 at 10:00 - should be closed (not Monday)
		{time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC), false, "Week 1 (Jan 2, Tuesday) at 10:00 - should be closed (not Monday)"},
	}

	for _, tt := range tests {
		got := oh.GetState(tt.date)
		if got != tt.want {
			year, week := tt.date.ISOWeek()
			t.Errorf("%s: got %v, want %v (ISO week: %d-%02d)", tt.desc, got, tt.want, year, week)
		}
	}
}

func TestWeekNumber_EvenWeeks(t *testing.T) {
	// week 02-52/2 Mo-Fr 09:00-17:00
	// Tests even weeks (2, 4, 6, ..., 52)
	oh, err := New("week 02-52/2 Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		date time.Time
		want bool
		desc string
	}{
		// Week 2, Monday Jan 8, 2024 at 10:00 - should be open (even week)
		{time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC), true, "Week 2 (Jan 8, Monday) at 10:00 - should be open (even week)"},
		// Week 4, Wednesday Jan 24, 2024 at 10:00 - should be open (even week)
		{time.Date(2024, 1, 24, 10, 0, 0, 0, time.UTC), true, "Week 4 (Jan 24, Wednesday) at 10:00 - should be open (even week)"},
		// Week 6, Tuesday Feb 6, 2024 at 10:00 - should be open (even week)
		{time.Date(2024, 2, 6, 10, 0, 0, 0, time.UTC), true, "Week 6 (Feb 6, Tuesday) at 10:00 - should be open (even week)"},
		// Week 1, Monday Jan 1, 2024 at 10:00 - should be closed (odd week)
		{time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), false, "Week 1 (Jan 1, Monday) at 10:00 - should be closed (odd week)"},
		// Week 3, Tuesday Jan 16, 2024 at 10:00 - should be closed (odd week)
		{time.Date(2024, 1, 16, 10, 0, 0, 0, time.UTC), false, "Week 3 (Jan 16, Tuesday) at 10:00 - should be closed (odd week)"},
		// Week 5, Friday Feb 2, 2024 at 10:00 - should be closed (odd week)
		{time.Date(2024, 2, 2, 10, 0, 0, 0, time.UTC), false, "Week 5 (Feb 2, Friday) at 10:00 - should be closed (odd week)"},
		// Week 2, Saturday Jan 13, 2024 at 10:00 - should be closed (weekend)
		{time.Date(2024, 1, 13, 10, 0, 0, 0, time.UTC), false, "Week 2 (Jan 13, Saturday) at 10:00 - should be closed (weekend)"},
	}

	for _, tt := range tests {
		got := oh.GetState(tt.date)
		if got != tt.want {
			year, week := tt.date.ISOWeek()
			t.Errorf("%s: got %v, want %v (ISO week: %d-%02d)", tt.desc, got, tt.want, year, week)
		}
	}
}

func TestWeekNumber_NoTimeRange(t *testing.T) {
	// week 15 off
	// Week 15 is completely closed (no time ranges)
	oh, err := New("week 15 off")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		date time.Time
		want bool
		desc string
	}{
		// Week 15, Monday Apr 8, 2024 at 00:00 - should be closed
		{time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC), false, "Week 15 (Apr 8, Monday) at 00:00 - should be closed"},
		// Week 15, Wednesday Apr 10, 2024 at 12:00 - should be closed
		{time.Date(2024, 4, 10, 12, 0, 0, 0, time.UTC), false, "Week 15 (Apr 10, Wednesday) at 12:00 - should be closed"},
		// Week 15, Sunday Apr 14, 2024 at 23:59 - should be closed
		{time.Date(2024, 4, 14, 23, 59, 0, 0, time.UTC), false, "Week 15 (Apr 14, Sunday) at 23:59 - should be closed"},
		// Week 14, Monday Apr 1, 2024 at 12:00 - should be closed (different week, no general rule)
		{time.Date(2024, 4, 1, 12, 0, 0, 0, time.UTC), false, "Week 14 (Apr 1, Monday) at 12:00 - should be closed (no general rule)"},
		// Week 16, Monday Apr 15, 2024 at 12:00 - should be closed (different week, no general rule)
		{time.Date(2024, 4, 15, 12, 0, 0, 0, time.UTC), false, "Week 16 (Apr 15, Monday) at 12:00 - should be closed (no general rule)"},
	}

	for _, tt := range tests {
		got := oh.GetState(tt.date)
		if got != tt.want {
			year, week := tt.date.ISOWeek()
			t.Errorf("%s: got %v, want %v (ISO week: %d-%02d)", tt.desc, got, tt.want, year, week)
		}
	}
}

func TestWeekNumber_ComplexInterval(t *testing.T) {
	// week 01-52/3 Sa 10:00-14:00
	// Tests weeks divisible by 3 (every third week: 3, 6, 9, 12, ...)
	// Note: 01-52/3 means start at week 1, go to week 52, every 3rd week
	// So this should match weeks: 1, 4, 7, 10, 13, 16, 19, 22, 25, 28, 31, 34, 37, 40, 43, 46, 49, 52
	oh, err := New("week 01-52/3 Sa 10:00-14:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		date time.Time
		want bool
		desc string
	}{
		// Week 1, Saturday Jan 6, 2024 at 12:00 - should be open (week 1 = 1 + 0*3)
		{time.Date(2024, 1, 6, 12, 0, 0, 0, time.UTC), true, "Week 1 (Jan 6, Saturday) at 12:00 - should be open"},
		// Week 4, Saturday Jan 27, 2024 at 12:00 - should be open (week 4 = 1 + 1*3)
		{time.Date(2024, 1, 27, 12, 0, 0, 0, time.UTC), true, "Week 4 (Jan 27, Saturday) at 12:00 - should be open"},
		// Week 7, Saturday Feb 17, 2024 at 12:00 - should be open (week 7 = 1 + 2*3)
		{time.Date(2024, 2, 17, 12, 0, 0, 0, time.UTC), true, "Week 7 (Feb 17, Saturday) at 12:00 - should be open"},
		// Week 10, Saturday Mar 9, 2024 at 12:00 - should be open (week 10 = 1 + 3*3)
		{time.Date(2024, 3, 9, 12, 0, 0, 0, time.UTC), true, "Week 10 (Mar 9, Saturday) at 12:00 - should be open"},
		// Week 2, Saturday Jan 13, 2024 at 12:00 - should be closed (week 2 not in pattern)
		{time.Date(2024, 1, 13, 12, 0, 0, 0, time.UTC), false, "Week 2 (Jan 13, Saturday) at 12:00 - should be closed"},
		// Week 3, Saturday Jan 20, 2024 at 12:00 - should be closed (week 3 not in pattern)
		{time.Date(2024, 1, 20, 12, 0, 0, 0, time.UTC), false, "Week 3 (Jan 20, Saturday) at 12:00 - should be closed"},
		// Week 5, Saturday Feb 3, 2024 at 12:00 - should be closed (week 5 not in pattern)
		{time.Date(2024, 2, 3, 12, 0, 0, 0, time.UTC), false, "Week 5 (Feb 3, Saturday) at 12:00 - should be closed"},
		// Week 6, Saturday Feb 10, 2024 at 12:00 - should be closed (week 6 not in pattern)
		{time.Date(2024, 2, 10, 12, 0, 0, 0, time.UTC), false, "Week 6 (Feb 10, Saturday) at 12:00 - should be closed"},
	}

	for _, tt := range tests {
		got := oh.GetState(tt.date)
		if got != tt.want {
			year, week := tt.date.ISOWeek()
			t.Errorf("%s: got %v, want %v (ISO week: %d-%02d)", tt.desc, got, tt.want, year, week)
		}
	}
}
