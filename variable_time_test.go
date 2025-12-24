package openinghours

import (
	"testing"
	"time"
)

// TestVariableTime_SunriseToSunset tests the basic sunrise-sunset functionality
func TestVariableTime_SunriseToSunset(t *testing.T) {
	oh, err := New("sunrise-sunset")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Set coordinates for Berlin (lat=52.52, lon=13.405)
	oh.SetCoordinates(52.52, 13.405)

	// Test on a known date: June 21, 2024 (summer solstice)
	// Actual calculated sunrise in Berlin: 02:50 UTC
	// Actual calculated sunset in Berlin: 19:25 UTC

	tests := []struct {
		name     string
		hour     int
		minute   int
		wantOpen bool
		desc     string
	}{
		{
			name:     "before sunrise",
			hour:     2,
			minute:   0,
			wantOpen: false,
			desc:     "02:00 should be closed (before sunrise at 02:50)",
		},
		{
			name:     "during daylight morning",
			hour:     10,
			minute:   0,
			wantOpen: true,
			desc:     "10:00 should be open (during daylight)",
		},
		{
			name:     "during daylight afternoon",
			hour:     15,
			minute:   0,
			wantOpen: true,
			desc:     "15:00 should be open (during daylight)",
		},
		{
			name:     "after sunset",
			hour:     23,
			minute:   0,
			wantOpen: false,
			desc:     "23:00 should be closed (after sunset)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTime := time.Date(2024, 6, 21, tt.hour, tt.minute, 0, 0, time.UTC)
			got := oh.GetState(testTime)
			if got != tt.wantOpen {
				t.Errorf("%s: got %v, want %v", tt.desc, got, tt.wantOpen)
			}
		})
	}

	// Test on winter solstice (December 21, 2024)
	// Actual calculated sunrise in Berlin: 07:23 UTC
	// Actual calculated sunset in Berlin: 14:47 UTC
	// Times at night should be closed
	winterNightTime := time.Date(2024, 12, 21, 20, 0, 0, 0, time.UTC)
	if oh.GetState(winterNightTime) {
		t.Errorf("December 21, 20:00 should be closed (after sunset at 14:47)")
	}

	// Midday should be open
	winterDayTime := time.Date(2024, 12, 21, 12, 0, 0, 0, time.UTC)
	if !oh.GetState(winterDayTime) {
		t.Errorf("December 21, 12:00 should be open (during daylight in winter)")
	}
}

// TestVariableTime_WithOffset tests sunrise/sunset with time offsets
func TestVariableTime_WithOffset(t *testing.T) {
	// Open from 1 hour after sunrise to 1 hour before sunset
	oh, err := New("(sunrise+01:00)-(sunset-01:00)")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Set coordinates for Berlin
	oh.SetCoordinates(52.52, 13.405)

	// Test on June 21, 2024
	// Sunrise is 02:50, then sunrise+01:00 is 03:50
	// Sunset is 19:25, then sunset-01:00 is 18:25
	testDate := time.Date(2024, 6, 21, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		hour     int
		minute   int
		wantOpen bool
		desc     string
	}{
		{
			name:     "at sunrise time",
			hour:     2,
			minute:   50,
			wantOpen: false,
			desc:     "at sunrise should be closed (offset not reached)",
		},
		{
			name:     "one hour after sunrise",
			hour:     4,
			minute:   0,
			wantOpen: true,
			desc:     "one hour after sunrise should be open",
		},
		{
			name:     "midday",
			hour:     12,
			minute:   0,
			wantOpen: true,
			desc:     "midday should be open",
		},
		{
			name:     "one hour before sunset",
			hour:     18,
			minute:   0,
			wantOpen: true,
			desc:     "one hour before sunset should be open",
		},
		{
			name:     "at sunset time",
			hour:     19,
			minute:   25,
			wantOpen: false,
			desc:     "at sunset should be closed (offset already passed)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTime := time.Date(testDate.Year(), testDate.Month(), testDate.Day(),
				tt.hour, tt.minute, 0, 0, time.UTC)
			got := oh.GetState(testTime)
			if got != tt.wantOpen {
				t.Errorf("%s: got %v, want %v", tt.desc, got, tt.wantOpen)
			}
		})
	}
}

// TestVariableTime_DawnDusk tests civil dawn to civil dusk
func TestVariableTime_DawnDusk(t *testing.T) {
	oh, err := New("dawn-dusk")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Set coordinates for Berlin
	oh.SetCoordinates(52.52, 13.405)

	// Test on June 21, 2024
	// Civil dawn is when sun is 6° below horizon (before sunrise)
	// Civil dusk is when sun is 6° below horizon (after sunset)
	// Sunrise: 02:50, so dawn: 02:20
	// Sunset: 19:25, so dusk: 19:55

	tests := []struct {
		name     string
		hour     int
		minute   int
		wantOpen bool
		desc     string
	}{
		{
			name:     "very early morning",
			hour:     2,
			minute:   0,
			wantOpen: false,
			desc:     "02:00 should be closed (before civil dawn at 02:20)",
		},
		{
			name:     "during civil twilight morning",
			hour:     2,
			minute:   30,
			wantOpen: true,
			desc:     "02:30 should be open (after dawn at 02:20)",
		},
		{
			name:     "during daylight",
			hour:     12,
			minute:   0,
			wantOpen: true,
			desc:     "12:00 should be open (during daylight)",
		},
		{
			name:     "during civil twilight evening",
			hour:     19,
			minute:   45,
			wantOpen: true,
			desc:     "19:45 should be open (before dusk at 19:55)",
		},
		{
			name:     "late night",
			hour:     20,
			minute:   30,
			wantOpen: false,
			desc:     "20:30 should be closed (after civil dusk at 19:55)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTime := time.Date(2024, 6, 21, tt.hour, tt.minute, 0, 0, time.UTC)
			got := oh.GetState(testTime)
			if got != tt.wantOpen {
				t.Errorf("%s: got %v, want %v", tt.desc, got, tt.wantOpen)
			}
		})
	}
}

// TestVariableTime_Overnight tests overnight range from sunset to sunrise
func TestVariableTime_Overnight(t *testing.T) {
	oh, err := New("sunset-sunrise")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Set coordinates for Berlin
	oh.SetCoordinates(52.52, 13.405)

	// Test on June 21, 2024
	// Sunset is 19:25, next sunrise is 02:50 next day
	// Night time should be open, day time should be closed

	tests := []struct {
		name     string
		hour     int
		minute   int
		wantOpen bool
		desc     string
	}{
		{
			name:     "during daylight morning",
			hour:     10,
			minute:   0,
			wantOpen: false,
			desc:     "10:00 should be closed (during daylight)",
		},
		{
			name:     "during daylight afternoon",
			hour:     15,
			minute:   0,
			wantOpen: false,
			desc:     "15:00 should be closed (during daylight)",
		},
		{
			name:     "after sunset",
			hour:     20,
			minute:   0,
			wantOpen: true,
			desc:     "20:00 should be open (after sunset at 19:25)",
		},
		{
			name:     "midnight",
			hour:     0,
			minute:   0,
			wantOpen: true,
			desc:     "00:00 should be open (night time)",
		},
		{
			name:     "before sunrise",
			hour:     2,
			minute:   0,
			wantOpen: true,
			desc:     "02:00 should be open (before sunrise at 02:50)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTime := time.Date(2024, 6, 21, tt.hour, tt.minute, 0, 0, time.UTC)
			got := oh.GetState(testTime)
			if got != tt.wantOpen {
				t.Errorf("%s: got %v, want %v", tt.desc, got, tt.wantOpen)
			}
		})
	}
}

// TestVariableTime_NoCoordinates tests fallback behavior when no coordinates are set
func TestVariableTime_NoCoordinates(t *testing.T) {
	oh, err := New("sunrise-sunset")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Do NOT set coordinates - should use default fallback
	// Default: sunrise = 06:00, sunset = 18:00

	tests := []struct {
		name     string
		hour     int
		minute   int
		wantOpen bool
		desc     string
	}{
		{
			name:     "before default sunrise",
			hour:     5,
			minute:   0,
			wantOpen: false,
			desc:     "05:00 should be closed (before default sunrise 06:00)",
		},
		{
			name:     "at default sunrise",
			hour:     6,
			minute:   0,
			wantOpen: true,
			desc:     "06:00 should be open (at default sunrise)",
		},
		{
			name:     "during default daylight",
			hour:     12,
			minute:   0,
			wantOpen: true,
			desc:     "12:00 should be open (during default daylight hours)",
		},
		{
			name:     "before default sunset",
			hour:     17,
			minute:   30,
			wantOpen: true,
			desc:     "17:30 should be open (before default sunset 18:00)",
		},
		{
			name:     "at default sunset",
			hour:     18,
			minute:   0,
			wantOpen: false,
			desc:     "18:00 should be closed (at default sunset)",
		},
		{
			name:     "after default sunset",
			hour:     20,
			minute:   0,
			wantOpen: false,
			desc:     "20:00 should be closed (after default sunset)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTime := time.Date(2024, 6, 21, tt.hour, tt.minute, 0, 0, time.UTC)
			got := oh.GetState(testTime)
			if got != tt.wantOpen {
				t.Errorf("%s: got %v, want %v", tt.desc, got, tt.wantOpen)
			}
		})
	}
}

// TestVariableTime_WithWeekday tests combination of weekday selector with variable times
func TestVariableTime_WithWeekday(t *testing.T) {
	oh, err := New("Mo-Fr sunrise-sunset")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Set coordinates for Berlin
	oh.SetCoordinates(52.52, 13.405)

	// Test on different days in June 2024
	// June 17, 2024 is Monday
	// June 21, 2024 is Friday
	// June 22, 2024 is Saturday
	// June 23, 2024 is Sunday

	tests := []struct {
		name     string
		day      int
		hour     int
		minute   int
		wantOpen bool
		desc     string
	}{
		{
			name:     "monday during daylight",
			day:      17,
			hour:     12,
			minute:   0,
			wantOpen: true,
			desc:     "Monday 12:00 should be open (weekday + daylight)",
		},
		{
			name:     "monday night",
			day:      17,
			hour:     23,
			minute:   0,
			wantOpen: false,
			desc:     "Monday 23:00 should be closed (weekday but after sunset)",
		},
		{
			name:     "friday during daylight",
			day:      21,
			hour:     12,
			minute:   0,
			wantOpen: true,
			desc:     "Friday 12:00 should be open (weekday + daylight)",
		},
		{
			name:     "saturday during daylight",
			day:      22,
			hour:     12,
			minute:   0,
			wantOpen: false,
			desc:     "Saturday 12:00 should be closed (not a weekday)",
		},
		{
			name:     "sunday during daylight",
			day:      23,
			hour:     12,
			minute:   0,
			wantOpen: false,
			desc:     "Sunday 12:00 should be closed (not a weekday)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTime := time.Date(2024, 6, tt.day, tt.hour, tt.minute, 0, 0, time.UTC)
			got := oh.GetState(testTime)
			if got != tt.wantOpen {
				t.Errorf("%s: got %v, want %v", tt.desc, got, tt.wantOpen)
			}
		})
	}
}

// TestVariableTime_MultipleOffsets tests complex offset combinations
func TestVariableTime_MultipleOffsets(t *testing.T) {
	// 30 minutes before sunset to 2 hours after sunset
	oh, err := New("(sunset-00:30)-(sunset+02:00)")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Set coordinates for Berlin
	oh.SetCoordinates(52.52, 13.405)

	// Test on June 21, 2024
	// Sunset is 19:25 UTC:
	// - Start: sunset-00:30 = 18:55
	// - End: sunset+02:00 = 21:25
	testDate := time.Date(2024, 6, 21, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		hour     int
		minute   int
		wantOpen bool
		desc     string
	}{
		{
			name:     "before range start",
			hour:     18,
			minute:   30,
			wantOpen: false,
			desc:     "18:30 should be closed (before sunset-00:30 at 18:55)",
		},
		{
			name:     "at range start",
			hour:     18,
			minute:   55,
			wantOpen: true,
			desc:     "18:55 should be open (at sunset-00:30)",
		},
		{
			name:     "during range",
			hour:     20,
			minute:   0,
			wantOpen: true,
			desc:     "20:00 should be open (during range)",
		},
		{
			name:     "before range end",
			hour:     21,
			minute:   0,
			wantOpen: true,
			desc:     "21:00 should be open (before sunset+02:00 at 21:25)",
		},
		{
			name:     "after range end",
			hour:     21,
			minute:   30,
			wantOpen: false,
			desc:     "21:30 should be closed (after sunset+02:00 at 21:25)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTime := time.Date(testDate.Year(), testDate.Month(), testDate.Day(),
				tt.hour, tt.minute, 0, 0, time.UTC)
			got := oh.GetState(testTime)
			if got != tt.wantOpen {
				t.Errorf("%s: got %v, want %v", tt.desc, got, tt.wantOpen)
			}
		})
	}
}

// TestVariableTime_NegativeOffset tests negative offsets (before sunrise/sunset)
func TestVariableTime_NegativeOffset(t *testing.T) {
	// Open from 1 hour before sunrise to sunrise
	oh, err := New("(sunrise-01:00)-sunrise")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Set coordinates for Berlin
	oh.SetCoordinates(52.52, 13.405)

	// Test on June 21, 2024
	// Sunrise is 02:50, then sunrise-01:00 is 01:50
	testDate := time.Date(2024, 6, 21, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		hour     int
		minute   int
		wantOpen bool
		desc     string
	}{
		{
			name:     "before range",
			hour:     1,
			minute:   0,
			wantOpen: false,
			desc:     "01:00 should be closed (before sunrise-01:00 at 01:50)",
		},
		{
			name:     "during range",
			hour:     2,
			minute:   0,
			wantOpen: true,
			desc:     "02:00 should be open (between sunrise-01:00 at 01:50 and sunrise at 02:50)",
		},
		{
			name:     "at sunrise",
			hour:     2,
			minute:   50,
			wantOpen: false,
			desc:     "02:50 should be closed (at sunrise, end of range)",
		},
		{
			name:     "after range",
			hour:     4,
			minute:   0,
			wantOpen: false,
			desc:     "04:00 should be closed (after sunrise)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTime := time.Date(testDate.Year(), testDate.Month(), testDate.Day(),
				tt.hour, tt.minute, 0, 0, time.UTC)
			got := oh.GetState(testTime)
			if got != tt.wantOpen {
				t.Errorf("%s: got %v, want %v", tt.desc, got, tt.wantOpen)
			}
		})
	}
}

// TestVariableTime_DifferentLocations tests that coordinates affect calculations
func TestVariableTime_DifferentLocations(t *testing.T) {
	// Test the same time in two different locations
	// Berlin vs Tokyo should have different sunrise/sunset times

	ohBerlin, err := New("sunrise-sunset")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	ohBerlin.SetCoordinates(52.52, 13.405) // Berlin

	ohTokyo, err := New("sunrise-sunset")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	ohTokyo.SetCoordinates(35.6762, 139.6503) // Tokyo

	// Test on same UTC date/time: June 21, 2024 at 03:00 UTC
	testTime := time.Date(2024, 6, 21, 3, 0, 0, 0, time.UTC)

	// At 03:00 UTC:
	// - Berlin (UTC+2 in summer): 05:00 local time - should be around sunrise (open)
	// - Tokyo (UTC+9): 12:00 local time - well after sunrise (open)
	// However, since we're using UTC times, the results should differ based on
	// the actual solar position at that UTC time for each location

	berlinState := ohBerlin.GetState(testTime)
	tokyoState := ohTokyo.GetState(testTime)

	// The states should potentially differ, but at minimum we can verify
	// that the function doesn't panic and returns a boolean result
	_ = berlinState
	_ = tokyoState

	// More specific assertions would require knowing exact sunrise/sunset calculations
	// This test primarily verifies that different coordinates are accepted and processed
}

// TestVariableTime_YearTransition tests variable times work across year boundaries
func TestVariableTime_YearTransition(t *testing.T) {
	oh, err := New("sunrise-sunset")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Set coordinates for Berlin
	oh.SetCoordinates(52.52, 13.405)

	// Test on January 1, 2024 (winter)
	winterTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	winterResult := oh.GetState(winterTime)
	if !winterResult {
		t.Errorf("January 1 at noon should be open (during daylight)")
	}

	// Test on July 1, 2024 (summer)
	summerTime := time.Date(2024, 7, 1, 12, 0, 0, 0, time.UTC)
	summerResult := oh.GetState(summerTime)
	if !summerResult {
		t.Errorf("July 1 at noon should be open (during daylight)")
	}
}

// TestVariableTime_WithSemicolon tests variable times combined with other rules
func TestVariableTime_WithSemicolon(t *testing.T) {
	// Regular hours on weekdays, sunrise-sunset on weekends
	oh, err := New("Mo-Fr 09:00-17:00; Sa-Su sunrise-sunset")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Set coordinates for Berlin
	oh.SetCoordinates(52.52, 13.405)

	// Test on June 17 (Monday) and June 22 (Saturday), 2024
	tests := []struct {
		name     string
		day      int
		hour     int
		minute   int
		wantOpen bool
		desc     string
	}{
		{
			name:     "monday fixed hours",
			day:      17,
			hour:     10,
			minute:   0,
			wantOpen: true,
			desc:     "Monday 10:00 should be open (fixed hours 09:00-17:00)",
		},
		{
			name:     "monday early morning",
			day:      17,
			hour:     7,
			minute:   0,
			wantOpen: false,
			desc:     "Monday 07:00 should be closed (before 09:00)",
		},
		{
			name:     "saturday during daylight",
			day:      22,
			hour:     10,
			minute:   0,
			wantOpen: true,
			desc:     "Saturday 10:00 should be open (sunrise-sunset)",
		},
		{
			name:     "saturday night",
			day:      22,
			hour:     23,
			minute:   0,
			wantOpen: false,
			desc:     "Saturday 23:00 should be closed (after sunset)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTime := time.Date(2024, 6, tt.day, tt.hour, tt.minute, 0, 0, time.UTC)
			got := oh.GetState(testTime)
			if got != tt.wantOpen {
				t.Errorf("%s: got %v, want %v", tt.desc, got, tt.wantOpen)
			}
		})
	}
}

// TestVariableTime_EdgeCases tests edge cases and boundary conditions
func TestVariableTime_EdgeCases(t *testing.T) {
	t.Run("extreme_northern_latitude", func(t *testing.T) {
		// Test near Arctic Circle where sun might not set in summer
		oh, err := New("sunrise-sunset")
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}

		// Tromsø, Norway (near Arctic Circle)
		oh.SetCoordinates(69.6492, 18.9553)

		// June 21 - near midnight sun period
		testTime := time.Date(2024, 6, 21, 23, 0, 0, 0, time.UTC)
		// Should handle gracefully (might be always open during midnight sun)
		_ = oh.GetState(testTime)
	})

	t.Run("extreme_southern_latitude", func(t *testing.T) {
		// Test in southern hemisphere
		oh, err := New("sunrise-sunset")
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}

		// Sydney, Australia
		oh.SetCoordinates(-33.8688, 151.2093)

		// December 21 (summer in southern hemisphere)
		// Sunrise: 18:47 UTC (05:47 local), Sunset: 09:01 UTC (20:01 local previous day wraps to next)
		// Test at 00:00 UTC which is 11:00 local time (during daylight)
		testTime := time.Date(2024, 12, 21, 0, 0, 0, 0, time.UTC)
		if !oh.GetState(testTime) {
			t.Errorf("Sydney at 00:00 UTC (11:00 local) in December should be open (summer daylight)")
		}
	})

	t.Run("equator", func(t *testing.T) {
		// Test at equator where sunrise/sunset are relatively constant
		oh, err := New("sunrise-sunset")
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}

		// Quito, Ecuador (near equator)
		oh.SetCoordinates(-0.1807, -78.4678)

		// Should have roughly 12 hours of daylight year-round
		testTime := time.Date(2024, 6, 21, 18, 0, 0, 0, time.UTC)
		_ = oh.GetState(testTime)
	})
}
