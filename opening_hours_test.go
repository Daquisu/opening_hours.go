package openinghours

import (
	"testing"
	"time"
)

func TestBasicTimeRange(t *testing.T) {
	oh, err := New("10:00-12:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Test time within range - should be open
	openTime := time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)
	if !oh.GetState(openTime) {
		t.Errorf("expected open at %v, got closed", openTime)
	}

	// Test time before range - should be closed
	beforeTime := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	if oh.GetState(beforeTime) {
		t.Errorf("expected closed at %v, got open", beforeTime)
	}

	// Test time after range - should be closed
	afterTime := time.Date(2024, 1, 15, 13, 0, 0, 0, time.UTC)
	if oh.GetState(afterTime) {
		t.Errorf("expected closed at %v, got open", afterTime)
	}

	// Test exact start time - should be open
	startTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	if !oh.GetState(startTime) {
		t.Errorf("expected open at %v, got closed", startTime)
	}

	// Test exact end time - should be closed (exclusive)
	endTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	if oh.GetState(endTime) {
		t.Errorf("expected closed at %v, got open", endTime)
	}
}

func TestAlwaysOpen(t *testing.T) {
	oh, err := New("24/7")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Any time should be open
	times := []time.Time{
		time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 15, 23, 59, 0, 0, time.UTC),
		time.Date(2024, 6, 15, 3, 30, 0, 0, time.UTC),
	}

	for _, tm := range times {
		if !oh.GetState(tm) {
			t.Errorf("expected open at %v for 24/7", tm)
		}
	}
}

func TestAlwaysClosed(t *testing.T) {
	testCases := []string{"off", "closed"}

	for _, input := range testCases {
		oh, err := New(input)
		if err != nil {
			t.Fatalf("unexpected parse error for %q: %v", input, err)
		}

		testTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
		if oh.GetState(testTime) {
			t.Errorf("expected closed for %q at %v", input, testTime)
		}
	}
}

func TestMultipleTimeRangesComma(t *testing.T) {
	// Multiple time ranges with comma separator
	oh, err := New("08:00-12:00,14:00-18:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		hour   int
		minute int
		want   bool
	}{
		{7, 0, false},   // before first range
		{9, 0, true},    // in first range
		{12, 30, false}, // between ranges
		{15, 0, true},   // in second range
		{19, 0, false},  // after second range
	}

	for _, tt := range tests {
		tm := time.Date(2024, 1, 15, tt.hour, tt.minute, 0, 0, time.UTC)
		got := oh.GetState(tm)
		if got != tt.want {
			t.Errorf("at %02d:%02d got %v, want %v", tt.hour, tt.minute, got, tt.want)
		}
	}
}

func TestMultipleTimeRangesSemicolon(t *testing.T) {
	// Multiple rules with semicolon separator
	oh, err := New("08:00-12:00; 14:00-18:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		hour   int
		minute int
		want   bool
	}{
		{9, 0, true},    // in first rule
		{12, 30, false}, // between rules
		{15, 0, true},   // in second rule
	}

	for _, tt := range tests {
		tm := time.Date(2024, 1, 15, tt.hour, tt.minute, 0, 0, time.UTC)
		got := oh.GetState(tm)
		if got != tt.want {
			t.Errorf("at %02d:%02d got %v, want %v", tt.hour, tt.minute, got, tt.want)
		}
	}
}

func TestTimeRangeWithOff(t *testing.T) {
	// Time range with off for specific time
	oh, err := New("08:00-18:00; 12:00-14:00 off")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		hour int
		want bool
	}{
		{9, true},   // in main range, before off
		{13, false}, // in off range
		{15, true},  // in main range, after off
	}

	for _, tt := range tests {
		tm := time.Date(2024, 1, 15, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(tm)
		if got != tt.want {
			t.Errorf("at %02d:00 got %v, want %v", tt.hour, got, tt.want)
		}
	}
}

func TestWeekdayRange(t *testing.T) {
	// Mo-Fr 09:00-17:00
	oh, err := New("Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Jan 15, 2024 is Monday
	// Jan 16, 2024 is Tuesday
	// Jan 20, 2024 is Saturday
	// Jan 21, 2024 is Sunday

	tests := []struct {
		day  int
		hour int
		want bool
		desc string
	}{
		{15, 10, true, "Monday 10:00"},
		{16, 10, true, "Tuesday 10:00"},
		{17, 10, true, "Wednesday 10:00"},
		{18, 10, true, "Thursday 10:00"},
		{19, 10, true, "Friday 10:00"},
		{20, 10, false, "Saturday 10:00"},
		{21, 10, false, "Sunday 10:00"},
		{15, 8, false, "Monday 08:00 (before open)"},
		{15, 18, false, "Monday 18:00 (after close)"},
	}

	for _, tt := range tests {
		tm := time.Date(2024, 1, tt.day, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(tm)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}

func TestWeekdayList(t *testing.T) {
	// Mo,We,Fr 09:00-17:00
	oh, err := New("Mo,We,Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		day  int
		want bool
		desc string
	}{
		{15, true, "Monday"},
		{16, false, "Tuesday"},
		{17, true, "Wednesday"},
		{18, false, "Thursday"},
		{19, true, "Friday"},
		{20, false, "Saturday"},
	}

	for _, tt := range tests {
		tm := time.Date(2024, 1, tt.day, 12, 0, 0, 0, time.UTC)
		got := oh.GetState(tm)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}

func TestWeekdayWrap(t *testing.T) {
	// Sa-Mo 10:00-14:00 (wraps around the week)
	oh, err := New("Sa-Mo 10:00-14:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		day  int
		want bool
		desc string
	}{
		{15, true, "Monday"},
		{16, false, "Tuesday"},
		{19, false, "Friday"},
		{20, true, "Saturday"},
		{21, true, "Sunday"},
	}

	for _, tt := range tests {
		tm := time.Date(2024, 1, tt.day, 12, 0, 0, 0, time.UTC)
		got := oh.GetState(tm)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}

func TestComplexWeekdayRule(t *testing.T) {
	// Mo-Fr 08:00-18:00; Sa 10:00-14:00
	oh, err := New("Mo-Fr 08:00-18:00; Sa 10:00-14:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		day  int
		hour int
		want bool
		desc string
	}{
		{15, 10, true, "Monday 10:00"},
		{20, 12, true, "Saturday 12:00"},
		{20, 16, false, "Saturday 16:00 (outside Sa hours)"},
		{21, 12, false, "Sunday 12:00"},
	}

	for _, tt := range tests {
		tm := time.Date(2024, 1, tt.day, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(tm)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}

func TestUnknownState(t *testing.T) {
	// Test basic unknown state
	oh, err := New("Mo-Fr 10:00-18:00 unknown")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	testTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC) // Monday 12:00

	// GetState should return false for unknown
	if oh.GetState(testTime) {
		t.Errorf("GetState should return false for unknown state")
	}

	// GetUnknown should return true for unknown
	if !oh.GetUnknown(testTime) {
		t.Errorf("GetUnknown should return true for unknown state at %v", testTime)
	}

	// Outside the time range, GetUnknown should return false
	outsideTime := time.Date(2024, 1, 15, 20, 0, 0, 0, time.UTC) // Monday 20:00
	if oh.GetUnknown(outsideTime) {
		t.Errorf("GetUnknown should return false outside unknown time range at %v", outsideTime)
	}
}

func TestComments(t *testing.T) {
	// Test basic comment
	oh, err := New("Mo-Fr 10:00-18:00 \"by appointment\"")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	testTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC) // Monday 12:00

	// GetComment should return the comment
	comment := oh.GetComment(testTime)
	if comment != "by appointment" {
		t.Errorf("expected comment 'by appointment', got '%s'", comment)
	}

	// Outside the time range, GetComment should return empty string
	outsideTime := time.Date(2024, 1, 15, 20, 0, 0, 0, time.UTC) // Monday 20:00
	comment = oh.GetComment(outsideTime)
	if comment != "" {
		t.Errorf("expected empty comment outside range, got '%s'", comment)
	}
}

func TestCommentsWithUnknownState(t *testing.T) {
	// Test comment with unknown state
	oh, err := New("Mo-Fr 10:00-18:00 unknown \"call ahead\"")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	testTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC) // Monday 12:00

	// Should be in unknown state
	if !oh.GetUnknown(testTime) {
		t.Errorf("expected unknown state at %v", testTime)
	}

	// Should have the comment
	comment := oh.GetComment(testTime)
	if comment != "call ahead" {
		t.Errorf("expected comment 'call ahead', got '%s'", comment)
	}
}

func TestGetNextChange_CurrentlyOpen(t *testing.T) {
	// Currently open, should return closing time
	oh, err := New("09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// At 10:00, should return 17:00
	testTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	expected := time.Date(2024, 1, 15, 17, 0, 0, 0, time.UTC)

	result := oh.GetNextChange(testTime)
	if !result.Equal(expected) {
		t.Errorf("GetNextChange at 10:00 with hours 09:00-17:00: got %v, want %v", result, expected)
	}
}

func TestGetNextChange_CurrentlyClosed(t *testing.T) {
	// Currently closed, should return opening time
	oh, err := New("09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// At 08:00, should return 09:00
	testTime := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC)
	expected := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)

	result := oh.GetNextChange(testTime)
	if !result.Equal(expected) {
		t.Errorf("GetNextChange at 08:00 with hours 09:00-17:00: got %v, want %v", result, expected)
	}
}

func TestGetNextChange_WeekdayBoundary(t *testing.T) {
	// Mo-Fr 09:00-17:00
	oh, err := New("Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Friday 18:00 (closed), should return Monday 09:00
	// Jan 19, 2024 is Friday
	// Jan 22, 2024 is Monday
	testTime := time.Date(2024, 1, 19, 18, 0, 0, 0, time.UTC)
	expected := time.Date(2024, 1, 22, 9, 0, 0, 0, time.UTC)

	result := oh.GetNextChange(testTime)
	if !result.Equal(expected) {
		t.Errorf("GetNextChange Friday 18:00 with Mo-Fr 09:00-17:00: got %v, want %v", result, expected)
	}
}

func TestGetNextChange_AlwaysOpen(t *testing.T) {
	oh, err := New("24/7")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// For always open, there's no next change - should return zero time
	testTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	result := oh.GetNextChange(testTime)

	if !result.IsZero() {
		t.Errorf("GetNextChange for 24/7: expected zero time, got %v", result)
	}
}

func TestGetNextChange_AlwaysClosed(t *testing.T) {
	oh, err := New("off")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// For always closed, there's no next change - should return zero time
	testTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	result := oh.GetNextChange(testTime)

	if !result.IsZero() {
		t.Errorf("GetNextChange for off: expected zero time, got %v", result)
	}
}

func TestMidnightSpanning(t *testing.T) {
	// Test time range spanning midnight (22:00-02:00)
	oh, err := New("22:00-02:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		hour int
		want bool
		desc string
	}{
		{21, false, "21:00 (before range)"},
		{22, true, "22:00 (start of range)"},
		{23, true, "23:00 (within range, before midnight)"},
		{0, true, "00:00 (within range, after midnight)"},
		{1, true, "01:00 (within range, after midnight)"},
		{2, false, "02:00 (end of range, exclusive)"},
		{3, false, "03:00 (after range)"},
	}

	for _, tt := range tests {
		// Test on same day
		tm := time.Date(2024, 1, 15, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(tm)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}

func TestMidnightSpanningWithWeekdays(t *testing.T) {
	// Test midnight spanning with weekday constraints
	// Fr 22:00-02:00 means Friday 22:00 to Saturday 02:00
	oh, err := New("Fr 22:00-02:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Jan 19, 2024 is Friday
	// Jan 20, 2024 is Saturday

	tests := []struct {
		day  int
		hour int
		want bool
		desc string
	}{
		{19, 21, false, "Friday 21:00 (before range)"},
		{19, 22, true, "Friday 22:00 (start of range)"},
		{19, 23, true, "Friday 23:00 (within range)"},
		{20, 0, true, "Saturday 00:00 (within range, next day)"},
		{20, 1, true, "Saturday 01:00 (within range, next day)"},
		{20, 2, false, "Saturday 02:00 (end of range)"},
		{20, 22, false, "Saturday 22:00 (not Friday start)"},
	}

	for _, tt := range tests {
		tm := time.Date(2024, 1, tt.day, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetState(tm)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}

func TestGetMatchingRule_SingleRule(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		day  int
		hour int
		want int
		desc string
	}{
		{15, 10, 0, "Monday 10:00 (matches first rule)"},
		{20, 10, -1, "Saturday 10:00 (no match)"},
	}

	// Jan 15, 2024 is Monday
	// Jan 20, 2024 is Saturday
	for _, tt := range tests {
		tm := time.Date(2024, 1, tt.day, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetMatchingRule(tm)
		if got != tt.want {
			t.Errorf("%s: got %d, want %d", tt.desc, got, tt.want)
		}
	}
}

func TestGetMatchingRule_MultipleRules(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-17:00; Sa 10:00-14:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		day  int
		hour int
		want int
		desc string
	}{
		{15, 10, 0, "Monday 10:00 (matches first rule)"},
		{20, 12, 1, "Saturday 12:00 (matches second rule)"},
		{21, 10, -1, "Sunday 10:00 (no match)"},
	}

	// Jan 15, 2024 is Monday
	// Jan 20, 2024 is Saturday
	// Jan 21, 2024 is Sunday
	for _, tt := range tests {
		tm := time.Date(2024, 1, tt.day, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetMatchingRule(tm)
		if got != tt.want {
			t.Errorf("%s: got %d, want %d", tt.desc, got, tt.want)
		}
	}
}

func TestGetMatchingRule_OverrideRule(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-17:00; We off")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		day  int
		hour int
		want int
		desc string
	}{
		{15, 10, 0, "Monday 10:00 (matches first rule)"},
		{17, 10, 1, "Wednesday 10:00 (later rule matches)"},
	}

	// Jan 15, 2024 is Monday
	// Jan 17, 2024 is Wednesday
	for _, tt := range tests {
		tm := time.Date(2024, 1, tt.day, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetMatchingRule(tm)
		if got != tt.want {
			t.Errorf("%s: got %d, want %d", tt.desc, got, tt.want)
		}
	}
}

func TestGetMatchingRule_AlwaysOpen(t *testing.T) {
	oh, err := New("24/7")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		day  int
		hour int
		want int
		desc string
	}{
		{15, 10, 0, "Monday 10:00 (any time matches)"},
		{20, 3, 0, "Saturday 03:00 (any time matches)"},
		{21, 23, 0, "Sunday 23:00 (any time matches)"},
	}

	for _, tt := range tests {
		tm := time.Date(2024, 1, tt.day, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetMatchingRule(tm)
		if got != tt.want {
			t.Errorf("%s: got %d, want %d", tt.desc, got, tt.want)
		}
	}
}

func TestGetMatchingRule_Off(t *testing.T) {
	oh, err := New("off")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		day  int
		hour int
		want int
		desc string
	}{
		{15, 10, 0, "Monday 10:00 (any time matches off rule)"},
		{20, 3, 0, "Saturday 03:00 (any time matches off rule)"},
		{21, 23, 0, "Sunday 23:00 (any time matches off rule)"},
	}

	for _, tt := range tests {
		tm := time.Date(2024, 1, tt.day, tt.hour, 0, 0, 0, time.UTC)
		got := oh.GetMatchingRule(tm)
		if got != tt.want {
			t.Errorf("%s: got %d, want %d", tt.desc, got, tt.want)
		}
	}
}

func TestIsWeekStable_BasicWeekdays(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if !oh.IsWeekStable() {
		t.Errorf("expected IsWeekStable() to return true for 'Mo-Fr 09:00-17:00'")
	}
}

func TestIsWeekStable_AlwaysOpen(t *testing.T) {
	oh, err := New("24/7")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if !oh.IsWeekStable() {
		t.Errorf("expected IsWeekStable() to return true for '24/7'")
	}
}

func TestIsWeekStable_Off(t *testing.T) {
	oh, err := New("off")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if !oh.IsWeekStable() {
		t.Errorf("expected IsWeekStable() to return true for 'off'")
	}
}

func TestIsWeekStable_WithMonth(t *testing.T) {
	oh, err := New("Dec Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if oh.IsWeekStable() {
		t.Errorf("expected IsWeekStable() to return false for 'Dec Mo-Fr 09:00-17:00'")
	}
}

func TestIsWeekStable_WithYear(t *testing.T) {
	oh, err := New("2024 Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if oh.IsWeekStable() {
		t.Errorf("expected IsWeekStable() to return false for '2024 Mo-Fr 09:00-17:00'")
	}
}

func TestIsWeekStable_WithWeekNumber(t *testing.T) {
	oh, err := New("week 01-10 Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if oh.IsWeekStable() {
		t.Errorf("expected IsWeekStable() to return false for 'week 01-10 Mo-Fr 09:00-17:00'")
	}
}

func TestIsWeekStable_WithConstrainedWeekday(t *testing.T) {
	oh, err := New("Mo[1] 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if oh.IsWeekStable() {
		t.Errorf("expected IsWeekStable() to return false for 'Mo[1] 09:00-17:00'")
	}
}

func TestIsWeekStable_WithPH(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-17:00; PH off")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if oh.IsWeekStable() {
		t.Errorf("expected IsWeekStable() to return false for 'Mo-Fr 09:00-17:00; PH off'")
	}
}

func TestIsWeekStable_ComplexStable(t *testing.T) {
	oh, err := New("Mo-Fr 09:00-17:00; Sa 10:00-14:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if !oh.IsWeekStable() {
		t.Errorf("expected IsWeekStable() to return true for 'Mo-Fr 09:00-17:00; Sa 10:00-14:00'")
	}
}

