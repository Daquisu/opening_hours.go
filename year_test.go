package openinghours

import (
	"testing"
	"time"
)

func TestYear_SingleYear(t *testing.T) {
	// 2024 Mo-Fr 09:00-17:00
	// Only applies in 2024, not in other years
	oh, err := New("2024 Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		date time.Time
		want bool
		desc string
	}{
		// Jan 15, 2024 is Monday - should be open in 2024
		{time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC), true, "Jan 15, 2024 (Monday) at 10:00 - should be open"},
		// Jan 15, 2025 is Wednesday - should be closed (different year)
		{time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), false, "Jan 15, 2025 (Wednesday) at 10:00 - should be closed (different year)"},
		// Jan 15, 2023 is Sunday - should be closed (different year)
		{time.Date(2023, 1, 15, 10, 0, 0, 0, time.UTC), false, "Jan 15, 2023 (Sunday) at 10:00 - should be closed (different year)"},
		// Feb 20, 2024 is Tuesday - should be open in 2024
		{time.Date(2024, 2, 20, 10, 0, 0, 0, time.UTC), true, "Feb 20, 2024 (Tuesday) at 10:00 - should be open"},
		// Feb 20, 2025 is Thursday - should be closed (different year)
		{time.Date(2025, 2, 20, 10, 0, 0, 0, time.UTC), false, "Feb 20, 2025 (Thursday) at 10:00 - should be closed (different year)"},
		// Jun 14, 2024 is Friday - should be open in 2024
		{time.Date(2024, 6, 14, 10, 0, 0, 0, time.UTC), true, "Jun 14, 2024 (Friday) at 10:00 - should be open"},
		// Jun 14, 2023 is Wednesday - should be closed (different year)
		{time.Date(2023, 6, 14, 10, 0, 0, 0, time.UTC), false, "Jun 14, 2023 (Wednesday) at 10:00 - should be closed (different year)"},
		// Saturday in 2024 at 10:00 - should be closed (weekend)
		{time.Date(2024, 1, 20, 10, 0, 0, 0, time.UTC), false, "Jan 20, 2024 (Saturday) at 10:00 - should be closed (weekend)"},
		// Monday in 2024 before opening time - should be closed
		{time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC), false, "Jan 15, 2024 (Monday) at 08:00 - should be closed (before opening)"},
		// Monday in 2024 after closing time - should be closed
		{time.Date(2024, 1, 15, 18, 0, 0, 0, time.UTC), false, "Jan 15, 2024 (Monday) at 18:00 - should be closed (after closing)"},
	}

	for _, tt := range tests {
		got := oh.GetState(tt.date)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}

func TestYear_YearRange(t *testing.T) {
	// 2024-2026 Dec 25 off
	// Applies to years 2024, 2025, and 2026
	oh, err := New("2024-2026 Dec 25 off")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		date time.Time
		want bool
		desc string
	}{
		// Dec 25, 2024 - should be closed (within range)
		{time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC), false, "Dec 25, 2024 at 00:00 - should be closed"},
		{time.Date(2024, 12, 25, 12, 0, 0, 0, time.UTC), false, "Dec 25, 2024 at 12:00 - should be closed"},
		{time.Date(2024, 12, 25, 23, 59, 0, 0, time.UTC), false, "Dec 25, 2024 at 23:59 - should be closed"},
		// Dec 25, 2025 - should be closed (within range)
		{time.Date(2025, 12, 25, 12, 0, 0, 0, time.UTC), false, "Dec 25, 2025 at 12:00 - should be closed"},
		// Dec 25, 2026 - should be closed (within range)
		{time.Date(2026, 12, 25, 12, 0, 0, 0, time.UTC), false, "Dec 25, 2026 at 12:00 - should be closed"},
		// Dec 25, 2027 - should be closed (outside range, no general rule)
		{time.Date(2027, 12, 25, 12, 0, 0, 0, time.UTC), false, "Dec 25, 2027 at 12:00 - should be closed (no matching rule)"},
		// Dec 25, 2023 - should be closed (outside range, no general rule)
		{time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC), false, "Dec 25, 2023 at 12:00 - should be closed (no matching rule)"},
		// Dec 24, 2024 - should be closed (different day, no general rule)
		{time.Date(2024, 12, 24, 12, 0, 0, 0, time.UTC), false, "Dec 24, 2024 at 12:00 - should be closed (no matching rule)"},
		// Dec 26, 2025 - should be closed (different day, no general rule)
		{time.Date(2025, 12, 26, 12, 0, 0, 0, time.UTC), false, "Dec 26, 2025 at 12:00 - should be closed (no matching rule)"},
	}

	for _, tt := range tests {
		got := oh.GetState(tt.date)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}

func TestYear_WithMonthDay(t *testing.T) {
	// 2024 Dec 24-26 10:00-22:00
	// Applies only to Dec 24-26 in 2024
	oh, err := New("2024 Dec 24-26 10:00-22:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		date time.Time
		want bool
		desc string
	}{
		// Dec 24, 2024 during hours - should be open
		{time.Date(2024, 12, 24, 10, 0, 0, 0, time.UTC), true, "Dec 24, 2024 at 10:00 - should be open"},
		{time.Date(2024, 12, 24, 15, 0, 0, 0, time.UTC), true, "Dec 24, 2024 at 15:00 - should be open"},
		{time.Date(2024, 12, 24, 21, 59, 0, 0, time.UTC), true, "Dec 24, 2024 at 21:59 - should be open"},
		// Dec 25, 2024 during hours - should be open
		{time.Date(2024, 12, 25, 12, 0, 0, 0, time.UTC), true, "Dec 25, 2024 at 12:00 - should be open"},
		// Dec 26, 2024 during hours - should be open
		{time.Date(2024, 12, 26, 12, 0, 0, 0, time.UTC), true, "Dec 26, 2024 at 12:00 - should be open"},
		// Dec 24, 2024 before opening time - should be closed
		{time.Date(2024, 12, 24, 9, 0, 0, 0, time.UTC), false, "Dec 24, 2024 at 09:00 - should be closed (before opening)"},
		// Dec 24, 2024 after closing time - should be closed
		{time.Date(2024, 12, 24, 22, 0, 0, 0, time.UTC), false, "Dec 24, 2024 at 22:00 - should be closed (after closing)"},
		// Dec 24, 2025 during hours - should be closed (different year)
		{time.Date(2025, 12, 24, 15, 0, 0, 0, time.UTC), false, "Dec 24, 2025 at 15:00 - should be closed (different year)"},
		// Dec 25, 2025 during hours - should be closed (different year)
		{time.Date(2025, 12, 25, 15, 0, 0, 0, time.UTC), false, "Dec 25, 2025 at 15:00 - should be closed (different year)"},
		// Dec 26, 2023 during hours - should be closed (different year)
		{time.Date(2023, 12, 26, 15, 0, 0, 0, time.UTC), false, "Dec 26, 2023 at 15:00 - should be closed (different year)"},
		// Dec 23, 2024 during hours - should be closed (different day)
		{time.Date(2024, 12, 23, 15, 0, 0, 0, time.UTC), false, "Dec 23, 2024 at 15:00 - should be closed (different day)"},
		// Dec 27, 2024 during hours - should be closed (different day)
		{time.Date(2024, 12, 27, 15, 0, 0, 0, time.UTC), false, "Dec 27, 2024 at 15:00 - should be closed (different day)"},
	}

	for _, tt := range tests {
		got := oh.GetState(tt.date)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}

func TestYear_MultipleRules(t *testing.T) {
	// Mo-Fr 09:00-17:00; 2024 Dec 25 off
	// Regular weekday hours, but Dec 25, 2024 is closed
	oh, err := New("Mo-Fr 09:00-17:00; 2024 Dec 25 off")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		date time.Time
		want bool
		desc string
	}{
		// Regular weekday in 2024 - should be open
		{time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC), true, "Jan 15, 2024 (Monday) at 10:00 - should be open"},
		// Regular weekday in 2025 - should be open
		{time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), true, "Jan 15, 2025 (Wednesday) at 10:00 - should be open"},
		// Dec 25, 2024 - should be closed (specific override)
		{time.Date(2024, 12, 25, 10, 0, 0, 0, time.UTC), false, "Dec 25, 2024 (Wednesday) at 10:00 - should be closed (override)"},
		// Dec 25, 2025 is Thursday - should be open (only 2024 override)
		{time.Date(2025, 12, 25, 10, 0, 0, 0, time.UTC), true, "Dec 25, 2025 (Thursday) at 10:00 - should be open (regular weekday)"},
		// Dec 25, 2023 is Monday - should be open (only 2024 override)
		{time.Date(2023, 12, 25, 10, 0, 0, 0, time.UTC), true, "Dec 25, 2023 (Monday) at 10:00 - should be open (regular weekday)"},
		// Weekend in 2024 - should be closed
		{time.Date(2024, 1, 20, 10, 0, 0, 0, time.UTC), false, "Jan 20, 2024 (Saturday) at 10:00 - should be closed (weekend)"},
		// Weekend in 2025 - should be closed
		{time.Date(2025, 1, 18, 10, 0, 0, 0, time.UTC), false, "Jan 18, 2025 (Saturday) at 10:00 - should be closed (weekend)"},
	}

	for _, tt := range tests {
		got := oh.GetState(tt.date)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}

func TestYear_FullDateRange(t *testing.T) {
	// 2024 Jan 01-2024 Jun 30 Mo-Fr 09:00-17:00
	// Applies only to Jan-Jun 2024 on weekdays
	oh, err := New("2024 Jan 01-2024 Jun 30 Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		date time.Time
		want bool
		desc string
	}{
		// Jan 2024 weekday - should be open
		{time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC), true, "Jan 15, 2024 (Monday) at 10:00 - should be open"},
		// Feb 2024 weekday - should be open
		{time.Date(2024, 2, 20, 10, 0, 0, 0, time.UTC), true, "Feb 20, 2024 (Tuesday) at 10:00 - should be open"},
		// Mar 2024 weekday - should be open
		{time.Date(2024, 3, 13, 10, 0, 0, 0, time.UTC), true, "Mar 13, 2024 (Wednesday) at 10:00 - should be open"},
		// Apr 2024 weekday - should be open
		{time.Date(2024, 4, 18, 10, 0, 0, 0, time.UTC), true, "Apr 18, 2024 (Thursday) at 10:00 - should be open"},
		// May 2024 weekday - should be open
		{time.Date(2024, 5, 17, 10, 0, 0, 0, time.UTC), true, "May 17, 2024 (Friday) at 10:00 - should be open"},
		// Jun 2024 weekday - should be open
		{time.Date(2024, 6, 14, 10, 0, 0, 0, time.UTC), true, "Jun 14, 2024 (Friday) at 10:00 - should be open"},
		// Jun 30, 2024 is Sunday - should be closed (weekend)
		{time.Date(2024, 6, 30, 10, 0, 0, 0, time.UTC), false, "Jun 30, 2024 (Sunday) at 10:00 - should be closed (weekend)"},
		// Jul 2024 weekday - should be closed (outside date range)
		{time.Date(2024, 7, 1, 10, 0, 0, 0, time.UTC), false, "Jul 1, 2024 (Monday) at 10:00 - should be closed (outside range)"},
		// Aug 2024 weekday - should be closed (outside date range)
		{time.Date(2024, 8, 15, 10, 0, 0, 0, time.UTC), false, "Aug 15, 2024 (Thursday) at 10:00 - should be closed (outside range)"},
		// Dec 2024 weekday - should be closed (outside date range)
		{time.Date(2024, 12, 16, 10, 0, 0, 0, time.UTC), false, "Dec 16, 2024 (Monday) at 10:00 - should be closed (outside range)"},
		// Jan 2025 weekday - should be closed (different year)
		{time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), false, "Jan 15, 2025 (Wednesday) at 10:00 - should be closed (different year)"},
		// Jan 2023 weekday - should be closed (different year)
		{time.Date(2023, 1, 16, 10, 0, 0, 0, time.UTC), false, "Jan 16, 2023 (Monday) at 10:00 - should be closed (different year)"},
		// Weekend in Jan-Jun 2024 - should be closed (weekend)
		{time.Date(2024, 3, 16, 10, 0, 0, 0, time.UTC), false, "Mar 16, 2024 (Saturday) at 10:00 - should be closed (weekend)"},
	}

	for _, tt := range tests {
		got := oh.GetState(tt.date)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}

func TestYear_ComplexWithYearBoundaries(t *testing.T) {
	// Test year boundaries with multiple years
	// 2023-2025 Jan 01 off
	oh, err := New("2023-2025 Jan 01 off")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		date time.Time
		want bool
		desc string
	}{
		// Jan 1, 2023 - should be closed
		{time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), false, "Jan 1, 2023 at 12:00 - should be closed"},
		// Jan 1, 2024 - should be closed
		{time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), false, "Jan 1, 2024 at 12:00 - should be closed"},
		// Jan 1, 2025 - should be closed
		{time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC), false, "Jan 1, 2025 at 12:00 - should be closed"},
		// Jan 1, 2022 - should be closed (outside range, no general rule)
		{time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC), false, "Jan 1, 2022 at 12:00 - should be closed (no matching rule)"},
		// Jan 1, 2026 - should be closed (outside range, no general rule)
		{time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC), false, "Jan 1, 2026 at 12:00 - should be closed (no matching rule)"},
		// Jan 2, 2024 - should be closed (different day, no general rule)
		{time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC), false, "Jan 2, 2024 at 12:00 - should be closed (no matching rule)"},
	}

	for _, tt := range tests {
		got := oh.GetState(tt.date)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}

func TestYear_YearWithWeekNumber(t *testing.T) {
	// 2024 week 01 Mo-Fr 09:00-17:00
	// Only week 1 of 2024
	oh, err := New("2024 week 01 Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		date time.Time
		want bool
		desc string
	}{
		// Week 1 of 2024, Monday Jan 1 at 10:00 - should be open
		{time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), true, "Week 1, Jan 1, 2024 (Monday) at 10:00 - should be open"},
		// Week 1 of 2024, Friday Jan 5 at 10:00 - should be open
		{time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC), true, "Week 1, Jan 5, 2024 (Friday) at 10:00 - should be open"},
		// Week 2 of 2024, Monday Jan 8 at 10:00 - should be closed (different week)
		{time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC), false, "Week 2, Jan 8, 2024 (Monday) at 10:00 - should be closed (week 2)"},
		// Week 1 of 2025, Monday Jan 6 at 10:00 - should be closed (different year)
		{time.Date(2025, 1, 6, 10, 0, 0, 0, time.UTC), false, "Week 1, Jan 6, 2025 (Monday) at 10:00 - should be closed (different year)"},
		// Week 1 of 2023, Monday Jan 2 at 10:00 - should be closed (different year)
		{time.Date(2023, 1, 2, 10, 0, 0, 0, time.UTC), false, "Week 1, Jan 2, 2023 (Monday) at 10:00 - should be closed (different year)"},
		// Week 1 of 2024, Saturday Jan 6 at 10:00 - should be closed (weekend)
		{time.Date(2024, 1, 6, 10, 0, 0, 0, time.UTC), false, "Week 1, Jan 6, 2024 (Saturday) at 10:00 - should be closed (weekend)"},
	}

	for _, tt := range tests {
		got := oh.GetState(tt.date)
		if got != tt.want {
			year, week := tt.date.ISOWeek()
			t.Errorf("%s: got %v, want %v (ISO week: %d-%02d)", tt.desc, got, tt.want, year, week)
		}
	}
}

func TestYear_MultipleYears(t *testing.T) {
	// 2024,2026,2028 Dec 25 off
	// Specific years (not a range)
	oh, err := New("2024,2026,2028 Dec 25 off")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		date time.Time
		want bool
		desc string
	}{
		// Dec 25, 2024 - should be closed
		{time.Date(2024, 12, 25, 12, 0, 0, 0, time.UTC), false, "Dec 25, 2024 at 12:00 - should be closed"},
		// Dec 25, 2026 - should be closed
		{time.Date(2026, 12, 25, 12, 0, 0, 0, time.UTC), false, "Dec 25, 2026 at 12:00 - should be closed"},
		// Dec 25, 2028 - should be closed
		{time.Date(2028, 12, 25, 12, 0, 0, 0, time.UTC), false, "Dec 25, 2028 at 12:00 - should be closed"},
		// Dec 25, 2025 - should be closed (not in list, no general rule)
		{time.Date(2025, 12, 25, 12, 0, 0, 0, time.UTC), false, "Dec 25, 2025 at 12:00 - should be closed (not in year list)"},
		// Dec 25, 2027 - should be closed (not in list, no general rule)
		{time.Date(2027, 12, 25, 12, 0, 0, 0, time.UTC), false, "Dec 25, 2027 at 12:00 - should be closed (not in year list)"},
		// Dec 25, 2023 - should be closed (not in list, no general rule)
		{time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC), false, "Dec 25, 2023 at 12:00 - should be closed (not in year list)"},
	}

	for _, tt := range tests {
		got := oh.GetState(tt.date)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}
