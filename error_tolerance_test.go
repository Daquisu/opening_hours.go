package openinghours

import (
	"testing"
	"time"
)

func TestErrorTolerance_AMPMFormat_SimpleTime(t *testing.T) {
	// Test case: "10am-12pm" should parse as 10:00-12:00
	oh, err := New("10am-12pm")
	if err != nil {
		t.Fatalf("unexpected parse error for '10am-12pm': %v", err)
	}

	tests := []struct {
		hour int
		want bool
		desc string
	}{
		{9, false, "09:00 (before range)"},
		{10, true, "10:00 (start of range)"},
		{11, true, "11:00 (within range)"},
		{12, false, "12:00 (end of range, exclusive)"},
		{13, false, "13:00 (after range)"},
	}

	for _, tt := range tests {
		tm := time.Date(2024, 1, 15, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(tm)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}

func TestErrorTolerance_AMPMFormat_TimeWithColon(t *testing.T) {
	// Test case: "10:00am-12:00pm" should parse as 10:00-12:00
	oh, err := New("10:00am-12:00pm")
	if err != nil {
		t.Fatalf("unexpected parse error for '10:00am-12:00pm': %v", err)
	}

	tests := []struct {
		hour int
		want bool
		desc string
	}{
		{9, false, "09:00 (before range)"},
		{10, true, "10:00 (start of range)"},
		{11, true, "11:00 (within range)"},
		{12, false, "12:00 (end of range, exclusive)"},
		{13, false, "13:00 (after range)"},
	}

	for _, tt := range tests {
		tm := time.Date(2024, 1, 15, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(tm)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}

func TestErrorTolerance_AMPMFormat_EarlyMorning(t *testing.T) {
	// Test case: "01am-11am" should parse as 01:00-11:00
	oh, err := New("01am-11am")
	if err != nil {
		t.Fatalf("unexpected parse error for '01am-11am': %v", err)
	}

	tests := []struct {
		hour int
		want bool
		desc string
	}{
		{0, false, "00:00 (before range)"},
		{1, true, "01:00 (start of range)"},
		{6, true, "06:00 (within range)"},
		{11, false, "11:00 (end of range, exclusive)"},
		{12, false, "12:00 (after range)"},
	}

	for _, tt := range tests {
		tm := time.Date(2024, 1, 15, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(tm)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}

func TestErrorTolerance_AMPMFormat_NoonSpecialCase(t *testing.T) {
	// Test case: "12:01pm-12:59pm" should parse as 12:01-12:59
	oh, err := New("12:01pm-12:59pm")
	if err != nil {
		t.Fatalf("unexpected parse error for '12:01pm-12:59pm': %v", err)
	}

	tests := []struct {
		hour   int
		minute int
		want   bool
		desc   string
	}{
		{12, 0, false, "12:00 (before range)"},
		{12, 1, true, "12:01 (start of range)"},
		{12, 30, true, "12:30 (within range)"},
		{12, 59, false, "12:59 (end of range, exclusive)"},
		{13, 0, false, "13:00 (after range)"},
	}

	for _, tt := range tests {
		tm := time.Date(2024, 1, 15, tt.hour, tt.minute, 0, 0, time.UTC)
		got := oh.GetState(tm)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}

func TestErrorTolerance_AMPMFormat_CrossingNoon(t *testing.T) {
	// Test case: "11:59am-12:00pm" should parse as 11:59-12:00
	oh, err := New("11:59am-12:00pm")
	if err != nil {
		t.Fatalf("unexpected parse error for '11:59am-12:00pm': %v", err)
	}

	tests := []struct {
		hour   int
		minute int
		want   bool
		desc   string
	}{
		{11, 58, false, "11:58 (before range)"},
		{11, 59, true, "11:59 (start of range)"},
		{12, 0, false, "12:00 (end of range, exclusive)"},
		{12, 1, false, "12:01 (after range)"},
	}

	for _, tt := range tests {
		tm := time.Date(2024, 1, 15, tt.hour, tt.minute, 0, 0, time.UTC)
		got := oh.GetState(tm)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}

func TestErrorTolerance_AMPMFormat_WithSpacesAndDots(t *testing.T) {
	// Test case: "10:00 a.m. - 12:00 p.m." should work with spaces and dots
	oh, err := New("10:00 a.m. - 12:00 p.m.")
	if err != nil {
		t.Fatalf("unexpected parse error for '10:00 a.m. - 12:00 p.m.': %v", err)
	}

	tests := []struct {
		hour int
		want bool
		desc string
	}{
		{9, false, "09:00 (before range)"},
		{10, true, "10:00 (start of range)"},
		{11, true, "11:00 (within range)"},
		{12, false, "12:00 (end of range, exclusive)"},
		{13, false, "13:00 (after range)"},
	}

	for _, tt := range tests {
		tm := time.Date(2024, 1, 15, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(tm)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}

func TestErrorTolerance_AMPMFormat_WithWeekdays(t *testing.T) {
	// Test case: "Mo 9am-5pm" should work with weekdays
	oh, err := New("Mo 9am-5pm")
	if err != nil {
		t.Fatalf("unexpected parse error for 'Mo 9am-5pm': %v", err)
	}

	// Jan 15, 2024 is Monday
	// Jan 16, 2024 is Tuesday

	tests := []struct {
		day  int
		hour int
		want bool
		desc string
	}{
		{15, 8, false, "Monday 08:00 (before range)"},
		{15, 9, true, "Monday 09:00 (start of range)"},
		{15, 12, true, "Monday 12:00 (within range)"},
		{15, 17, false, "Monday 17:00 (end of range, exclusive)"},
		{15, 18, false, "Monday 18:00 (after range)"},
		{16, 12, false, "Tuesday 12:00 (not Monday)"},
	}

	for _, tt := range tests {
		tm := time.Date(2024, 1, tt.day, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(tm)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}

func TestErrorTolerance_DotsInTime_Basic(t *testing.T) {
	// Test case: "10.00-14.00" should parse as 10:00-14:00
	oh, err := New("10.00-14.00")
	if err != nil {
		t.Fatalf("unexpected parse error for '10.00-14.00': %v", err)
	}

	tests := []struct {
		hour int
		want bool
		desc string
	}{
		{9, false, "09:00 (before range)"},
		{10, true, "10:00 (start of range)"},
		{12, true, "12:00 (within range)"},
		{14, false, "14:00 (end of range, exclusive)"},
		{15, false, "15:00 (after range)"},
	}

	for _, tt := range tests {
		tm := time.Date(2024, 1, 15, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(tm)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}

func TestErrorTolerance_DotsInTime_WithWeekdays(t *testing.T) {
	// Test case: "Mo 8.30-12.30" should work with weekdays
	oh, err := New("Mo 8.30-12.30")
	if err != nil {
		t.Fatalf("unexpected parse error for 'Mo 8.30-12.30': %v", err)
	}

	// Jan 15, 2024 is Monday
	// Jan 16, 2024 is Tuesday

	tests := []struct {
		day    int
		hour   int
		minute int
		want   bool
		desc   string
	}{
		{15, 8, 29, false, "Monday 08:29 (before range)"},
		{15, 8, 30, true, "Monday 08:30 (start of range)"},
		{15, 10, 0, true, "Monday 10:00 (within range)"},
		{15, 12, 30, false, "Monday 12:30 (end of range, exclusive)"},
		{15, 13, 0, false, "Monday 13:00 (after range)"},
		{16, 10, 0, false, "Tuesday 10:00 (not Monday)"},
	}

	for _, tt := range tests {
		tm := time.Date(2024, 1, tt.day, tt.hour, tt.minute, 0, 0, time.UTC)
		got := oh.GetState(tm)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}

func TestErrorTolerance_DotsInTime_MultipleRanges(t *testing.T) {
	// Test case: "9.00-12.00,14.00-18.00" should work with multiple ranges
	oh, err := New("9.00-12.00,14.00-18.00")
	if err != nil {
		t.Fatalf("unexpected parse error for '9.00-12.00,14.00-18.00': %v", err)
	}

	tests := []struct {
		hour int
		want bool
		desc string
	}{
		{8, false, "08:00 (before first range)"},
		{9, true, "09:00 (start of first range)"},
		{10, true, "10:00 (within first range)"},
		{12, false, "12:00 (end of first range, exclusive)"},
		{13, false, "13:00 (between ranges)"},
		{14, true, "14:00 (start of second range)"},
		{16, true, "16:00 (within second range)"},
		{18, false, "18:00 (end of second range, exclusive)"},
		{19, false, "19:00 (after second range)"},
	}

	for _, tt := range tests {
		tm := time.Date(2024, 1, 15, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(tm)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}
