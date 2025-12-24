package openinghours

import (
	"testing"
	"time"
)

func TestConstrainedWeekday_FirstMonday(t *testing.T) {
	// Mo[1] 09:00-17:00 - first Monday of the month
	oh, err := New("Mo[1] 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// January 2024: first Monday is Jan 1
	tests := []struct {
		day  int
		hour int
		want bool
		desc string
	}{
		{1, 10, true, "Jan 1 (first Monday) at 10:00 - should be open"},
		{8, 10, false, "Jan 8 (second Monday) at 10:00 - should be closed"},
		{15, 10, false, "Jan 15 (third Monday) at 10:00 - should be closed"},
		{22, 10, false, "Jan 22 (fourth Monday) at 10:00 - should be closed"},
		{29, 10, false, "Jan 29 (fifth Monday) at 10:00 - should be closed"},
		{1, 8, false, "Jan 1 (first Monday) at 08:00 - before opening time"},
		{1, 18, false, "Jan 1 (first Monday) at 18:00 - after closing time"},
	}

	for _, tt := range tests {
		tm := time.Date(2024, 1, tt.day, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(tm)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}

func TestConstrainedWeekday_LastFriday(t *testing.T) {
	// Fr[-1] 09:00-17:00 - last Friday of the month
	oh, err := New("Fr[-1] 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// January 2024: last Friday is Jan 26
	tests := []struct {
		day  int
		hour int
		want bool
		desc string
	}{
		{5, 10, false, "Jan 5 (first Friday) at 10:00 - should be closed"},
		{12, 10, false, "Jan 12 (second Friday) at 10:00 - should be closed"},
		{19, 10, false, "Jan 19 (third Friday) at 10:00 - should be closed"},
		{26, 10, true, "Jan 26 (last Friday) at 10:00 - should be open"},
		{26, 8, false, "Jan 26 (last Friday) at 08:00 - before opening time"},
		{26, 18, false, "Jan 26 (last Friday) at 18:00 - after closing time"},
	}

	for _, tt := range tests {
		tm := time.Date(2024, 1, tt.day, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(tm)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}

func TestConstrainedWeekday_ThirdWednesday(t *testing.T) {
	// We[3] 10:00-14:00 - third Wednesday of the month
	oh, err := New("We[3] 10:00-14:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// January 2024: third Wednesday is Jan 17
	tests := []struct {
		day  int
		hour int
		want bool
		desc string
	}{
		{3, 12, false, "Jan 3 (first Wednesday) at 12:00 - should be closed"},
		{10, 12, false, "Jan 10 (second Wednesday) at 12:00 - should be closed"},
		{17, 12, true, "Jan 17 (third Wednesday) at 12:00 - should be open"},
		{24, 12, false, "Jan 24 (fourth Wednesday) at 12:00 - should be closed"},
		{31, 12, false, "Jan 31 (fifth Wednesday) at 12:00 - should be closed"},
		{17, 9, false, "Jan 17 (third Wednesday) at 09:00 - before opening time"},
		{17, 15, false, "Jan 17 (third Wednesday) at 15:00 - after closing time"},
		{17, 10, true, "Jan 17 (third Wednesday) at 10:00 - at opening time"},
		{17, 13, true, "Jan 17 (third Wednesday) at 13:00 - within opening hours"},
	}

	for _, tt := range tests {
		tm := time.Date(2024, 1, tt.day, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(tm)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}

func TestConstrainedWeekday_MultipleOccurrences(t *testing.T) {
	// Th[1],Th[-1] 09:00-17:00 - first and last Thursday of the month
	oh, err := New("Th[1],Th[-1] 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// January 2024: first Thursday is Jan 4, last is Jan 25
	tests := []struct {
		day  int
		hour int
		want bool
		desc string
	}{
		{4, 10, true, "Jan 4 (first Thursday) at 10:00 - should be open"},
		{11, 10, false, "Jan 11 (second Thursday) at 10:00 - should be closed"},
		{18, 10, false, "Jan 18 (third Thursday) at 10:00 - should be closed"},
		{25, 10, true, "Jan 25 (last Thursday) at 10:00 - should be open"},
		{4, 8, false, "Jan 4 (first Thursday) at 08:00 - before opening time"},
		{4, 18, false, "Jan 4 (first Thursday) at 18:00 - after closing time"},
		{25, 12, true, "Jan 25 (last Thursday) at 12:00 - should be open"},
	}

	for _, tt := range tests {
		tm := time.Date(2024, 1, tt.day, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(tm)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}

func TestConstrainedWeekday_Range(t *testing.T) {
	// Sa[1-2] 10:00-14:00 - first and second Saturday of the month
	oh, err := New("Sa[1-2] 10:00-14:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// January 2024: first Saturday is Jan 6, second is Jan 13
	tests := []struct {
		day  int
		hour int
		want bool
		desc string
	}{
		{6, 12, true, "Jan 6 (first Saturday) at 12:00 - should be open"},
		{13, 12, true, "Jan 13 (second Saturday) at 12:00 - should be open"},
		{20, 12, false, "Jan 20 (third Saturday) at 12:00 - should be closed"},
		{27, 12, false, "Jan 27 (fourth Saturday) at 12:00 - should be closed"},
		{6, 9, false, "Jan 6 (first Saturday) at 09:00 - before opening time"},
		{6, 15, false, "Jan 6 (first Saturday) at 15:00 - after closing time"},
		{13, 10, true, "Jan 13 (second Saturday) at 10:00 - at opening time"},
		{13, 14, false, "Jan 13 (second Saturday) at 14:00 - at closing time (exclusive)"},
	}

	for _, tt := range tests {
		tm := time.Date(2024, 1, tt.day, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(tm)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}

func TestConstrainedWeekday_WithOtherRules(t *testing.T) {
	// Mo-Fr 09:00-17:00; Th[3],Th[-1] off
	// Complex case: open Monday-Friday, but closed on 3rd and last Thursday
	oh, err := New("Mo-Fr 09:00-17:00; Th[3],Th[-1] off")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// January 2024: third Thursday is Jan 18, last Thursday is Jan 25
	tests := []struct {
		day  int
		hour int
		want bool
		desc string
	}{
		{15, 10, true, "Jan 15 (Monday) at 10:00 - should be open"},
		{16, 10, true, "Jan 16 (Tuesday) at 10:00 - should be open"},
		{17, 10, true, "Jan 17 (Wednesday) at 10:00 - should be open"},
		{4, 10, true, "Jan 4 (first Thursday) at 10:00 - should be open"},
		{11, 10, true, "Jan 11 (second Thursday) at 10:00 - should be open"},
		{18, 10, false, "Jan 18 (third Thursday) at 10:00 - should be closed (off rule)"},
		{25, 10, false, "Jan 25 (last Thursday) at 10:00 - should be closed (off rule)"},
		{19, 10, true, "Jan 19 (Friday) at 10:00 - should be open"},
		{20, 10, false, "Jan 20 (Saturday) at 10:00 - should be closed (weekend)"},
		{18, 8, false, "Jan 18 (third Thursday) at 08:00 - before normal hours and off"},
		{18, 18, false, "Jan 18 (third Thursday) at 18:00 - after normal hours and off"},
	}

	for _, tt := range tests {
		tm := time.Date(2024, 1, tt.day, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(tm)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}

func TestConstrainedWeekday_EdgeCases(t *testing.T) {
	// Test edge cases with different months

	t.Run("SecondTuesdayInFebruary2024", func(t *testing.T) {
		// Tu[2] 10:00-16:00
		oh, err := New("Tu[2] 10:00-16:00")
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}

		// February 2024: second Tuesday is Feb 13
		tests := []struct {
			day  int
			hour int
			want bool
			desc string
		}{
			{6, 12, false, "Feb 6 (first Tuesday) - should be closed"},
			{13, 12, true, "Feb 13 (second Tuesday) - should be open"},
			{20, 12, false, "Feb 20 (third Tuesday) - should be closed"},
			{27, 12, false, "Feb 27 (fourth Tuesday) - should be closed"},
		}

		for _, tt := range tests {
			tm := time.Date(2024, 2, tt.day, tt.hour, 0, 0, 0, time.UTC)
			got := oh.GetState(tm)
			if got != tt.want {
				t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
			}
		}
	})

	t.Run("FourthAndFifthMonday", func(t *testing.T) {
		// Mo[4-5] 09:00-17:00 - only months with 5 Mondays will have the 5th open
		oh, err := New("Mo[4-5] 09:00-17:00")
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}

		// January 2024 has 5 Mondays: 1, 8, 15, 22, 29
		tests := []struct {
			day  int
			hour int
			want bool
			desc string
		}{
			{1, 10, false, "Jan 1 (first Monday) - should be closed"},
			{8, 10, false, "Jan 8 (second Monday) - should be closed"},
			{15, 10, false, "Jan 15 (third Monday) - should be closed"},
			{22, 10, true, "Jan 22 (fourth Monday) - should be open"},
			{29, 10, true, "Jan 29 (fifth Monday) - should be open"},
		}

		for _, tt := range tests {
			tm := time.Date(2024, 1, tt.day, tt.hour, 0, 0, 0, time.UTC)
			got := oh.GetState(tm)
			if got != tt.want {
				t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
			}
		}
	})

	t.Run("LastMondayVariousMonths", func(t *testing.T) {
		// Mo[-1] 09:00-17:00 - last Monday varies by month
		oh, err := New("Mo[-1] 09:00-17:00")
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}

		tests := []struct {
			month time.Month
			day   int
			hour  int
			want  bool
			desc  string
		}{
			{time.January, 29, 10, true, "Jan 29 (last Monday in Jan 2024)"},
			{time.February, 26, 10, true, "Feb 26 (last Monday in Feb 2024)"},
			{time.March, 25, 10, true, "Mar 25 (last Monday in Mar 2024)"},
			{time.January, 22, 10, false, "Jan 22 (fourth Monday, not last)"},
			{time.February, 19, 10, false, "Feb 19 (third Monday, not last)"},
		}

		for _, tt := range tests {
			tm := time.Date(2024, tt.month, tt.day, tt.hour, 0, 0, 0, time.UTC)
			got := oh.GetState(tm)
			if got != tt.want {
				t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
			}
		}
	})
}

func TestConstrainedWeekday_MultipleConstrainedDaysInRule(t *testing.T) {
	// Mo[1],We[1],Fr[1] 09:00-17:00 - first Monday, Wednesday, and Friday
	oh, err := New("Mo[1],We[1],Fr[1] 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// January 2024: first Monday is Jan 1, first Wednesday is Jan 3, first Friday is Jan 5
	tests := []struct {
		day  int
		hour int
		want bool
		desc string
	}{
		{1, 10, true, "Jan 1 (first Monday) - should be open"},
		{3, 10, true, "Jan 3 (first Wednesday) - should be open"},
		{5, 10, true, "Jan 5 (first Friday) - should be open"},
		{8, 10, false, "Jan 8 (second Monday) - should be closed"},
		{10, 10, false, "Jan 10 (second Wednesday) - should be closed"},
		{12, 10, false, "Jan 12 (second Friday) - should be closed"},
	}

	for _, tt := range tests {
		tm := time.Date(2024, 1, tt.day, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(tm)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}

func TestConstrainedWeekday_NegativeIndexes(t *testing.T) {
	// Test various negative indexes

	t.Run("SecondToLastSunday", func(t *testing.T) {
		// Su[-2] 10:00-14:00 - second to last Sunday
		oh, err := New("Su[-2] 10:00-14:00")
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}

		// January 2024: Sundays are Jan 7, 14, 21, 28
		// Second to last is Jan 21
		tests := []struct {
			day  int
			hour int
			want bool
			desc string
		}{
			{7, 12, false, "Jan 7 (first Sunday) - should be closed"},
			{14, 12, false, "Jan 14 (second Sunday) - should be closed"},
			{21, 12, true, "Jan 21 (second to last Sunday) - should be open"},
			{28, 12, false, "Jan 28 (last Sunday) - should be closed"},
		}

		for _, tt := range tests {
			tm := time.Date(2024, 1, tt.day, tt.hour, 0, 0, 0, time.UTC)
			got := oh.GetState(tm)
			if got != tt.want {
				t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
			}
		}
	})
}

func TestConstrainedWeekday_CombinedWithRegularWeekdays(t *testing.T) {
	// Mo[1] 09:00-17:00; Mo 10:00-14:00
	// First Monday has special hours, other Mondays have different hours
	oh, err := New("Mo 10:00-14:00; Mo[1] 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// January 2024: first Monday is Jan 1, other Mondays are 8, 15, 22, 29
	tests := []struct {
		day  int
		hour int
		want bool
		desc string
	}{
		{1, 9, true, "Jan 1 (first Monday) at 09:00 - should be open (special hours)"},
		{1, 16, true, "Jan 1 (first Monday) at 16:00 - should be open (special hours)"},
		{1, 13, true, "Jan 1 (first Monday) at 13:00 - should be open (both rules)"},
		{8, 9, false, "Jan 8 (second Monday) at 09:00 - should be closed (regular hours start at 10:00)"},
		{8, 11, true, "Jan 8 (second Monday) at 11:00 - should be open (regular hours)"},
		{8, 16, false, "Jan 8 (second Monday) at 16:00 - should be closed (regular hours end at 14:00)"},
		{15, 11, true, "Jan 15 (third Monday) at 11:00 - should be open (regular hours)"},
	}

	for _, tt := range tests {
		tm := time.Date(2024, 1, tt.day, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(tm)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}
