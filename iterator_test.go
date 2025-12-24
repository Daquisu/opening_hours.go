package openinghours

import (
	"testing"
	"time"
)

// Iterator API Tests

func TestIterator_CreateAndCheckInitialState(t *testing.T) {
	// Test creating iterator and checking initial state
	oh, err := New("Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Monday 10:00 - should be open
	startTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	it := oh.GetIterator(startTime)

	if it == nil {
		t.Fatal("GetIterator returned nil")
	}

	// Check initial date
	if !it.GetDate().Equal(startTime) {
		t.Errorf("GetDate: got %v, want %v", it.GetDate(), startTime)
	}

	// Check initial state - should be open
	if !it.GetState() {
		t.Error("GetState: expected true (open), got false")
	}

	// Check initial state string
	if it.GetStateString() != "open" {
		t.Errorf("GetStateString: got %q, want %q", it.GetStateString(), "open")
	}

	// Check initial comment (should be empty)
	if it.GetComment() != "" {
		t.Errorf("GetComment: got %q, want empty string", it.GetComment())
	}
}

func TestIterator_AdvanceToNextStateChange(t *testing.T) {
	// Test advancing iterator to next state change
	oh, err := New("09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Start at 10:00 (open)
	startTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	it := oh.GetIterator(startTime)

	// Advance should move to 17:00 (closing time)
	newTime := it.Advance()
	expectedTime := time.Date(2024, 1, 15, 17, 0, 0, 0, time.UTC)

	if !newTime.Equal(expectedTime) {
		t.Errorf("Advance returned %v, want %v", newTime, expectedTime)
	}

	// Iterator's current time should be updated
	if !it.GetDate().Equal(expectedTime) {
		t.Errorf("GetDate after Advance: got %v, want %v", it.GetDate(), expectedTime)
	}

	// State should now be closed
	if it.GetState() {
		t.Error("GetState after Advance: expected false (closed), got true")
	}
}

func TestIterator_MultipleStateChangesInDay(t *testing.T) {
	// Test iterating through multiple state changes in a day
	oh, err := New("08:00-12:00,14:00-18:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Start before opening (07:00)
	startTime := time.Date(2024, 1, 15, 7, 0, 0, 0, time.UTC)
	it := oh.GetIterator(startTime)

	// Should be closed initially
	if it.GetState() {
		t.Error("Initial state: expected closed, got open")
	}

	// First advance: 07:00 -> 08:00 (opening)
	newTime := it.Advance()
	if newTime.Hour() != 8 || newTime.Minute() != 0 {
		t.Errorf("First advance: got %v, want 08:00", newTime)
	}
	if !it.GetState() {
		t.Error("After first advance: expected open, got closed")
	}

	// Second advance: 08:00 -> 12:00 (closing)
	newTime = it.Advance()
	if newTime.Hour() != 12 || newTime.Minute() != 0 {
		t.Errorf("Second advance: got %v, want 12:00", newTime)
	}
	if it.GetState() {
		t.Error("After second advance: expected closed, got open")
	}

	// Third advance: 12:00 -> 14:00 (opening)
	newTime = it.Advance()
	if newTime.Hour() != 14 || newTime.Minute() != 0 {
		t.Errorf("Third advance: got %v, want 14:00", newTime)
	}
	if !it.GetState() {
		t.Error("After third advance: expected open, got closed")
	}

	// Fourth advance: 14:00 -> 18:00 (closing)
	newTime = it.Advance()
	if newTime.Hour() != 18 || newTime.Minute() != 0 {
		t.Errorf("Fourth advance: got %v, want 18:00", newTime)
	}
	if it.GetState() {
		t.Error("After fourth advance: expected closed, got open")
	}
}

func TestIterator_SetDate(t *testing.T) {
	// Test SetDate to jump to a specific time
	oh, err := New("Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Start on Monday at 10:00
	startTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	it := oh.GetIterator(startTime)

	// Should be open
	if !it.GetState() {
		t.Error("Initial state: expected open, got closed")
	}

	// Jump to Saturday at 10:00 (should be closed)
	saturdayTime := time.Date(2024, 1, 20, 10, 0, 0, 0, time.UTC)
	it.SetDate(saturdayTime)

	if !it.GetDate().Equal(saturdayTime) {
		t.Errorf("GetDate after SetDate: got %v, want %v", it.GetDate(), saturdayTime)
	}

	if it.GetState() {
		t.Error("After SetDate to Saturday: expected closed, got open")
	}

	// Jump to Tuesday at 16:00 (should be open)
	tuesdayTime := time.Date(2024, 1, 16, 16, 0, 0, 0, time.UTC)
	it.SetDate(tuesdayTime)

	if !it.GetDate().Equal(tuesdayTime) {
		t.Errorf("GetDate after second SetDate: got %v, want %v", it.GetDate(), tuesdayTime)
	}

	if !it.GetState() {
		t.Error("After SetDate to Tuesday 16:00: expected open, got closed")
	}
}

func TestIterator_AdvanceAcrossDays(t *testing.T) {
	// Test advancing across day boundaries with weekday constraints
	oh, err := New("Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Start Friday at 18:00 (closed, after business hours)
	startTime := time.Date(2024, 1, 19, 18, 0, 0, 0, time.UTC)
	it := oh.GetIterator(startTime)

	// Should be closed
	if it.GetState() {
		t.Error("Initial state: expected closed, got open")
	}

	// Advance should skip weekend and go to Monday 09:00
	newTime := it.Advance()
	expectedTime := time.Date(2024, 1, 22, 9, 0, 0, 0, time.UTC)

	if !newTime.Equal(expectedTime) {
		t.Errorf("Advance from Friday 18:00: got %v, want %v", newTime, expectedTime)
	}

	if !it.GetState() {
		t.Error("After advance to Monday: expected open, got closed")
	}
}

func TestIterator_WithComments(t *testing.T) {
	// Test iterator with comments
	oh, err := New("Mo-Fr 10:00-18:00 \"by appointment\"")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// Monday 12:00 - should have comment
	startTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	it := oh.GetIterator(startTime)

	if it.GetComment() != "by appointment" {
		t.Errorf("GetComment: got %q, want %q", it.GetComment(), "by appointment")
	}

	// Advance to 18:00 (closed)
	it.Advance()

	// Outside hours, comment should be empty
	if it.GetComment() != "" {
		t.Errorf("GetComment after closing: got %q, want empty string", it.GetComment())
	}
}

func TestIterator_AlwaysOpen(t *testing.T) {
	// Test iterator with 24/7 hours
	oh, err := New("24/7")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	startTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	it := oh.GetIterator(startTime)

	// Should be open
	if !it.GetState() {
		t.Error("Expected open for 24/7")
	}

	// Advance should return zero time (no state changes)
	newTime := it.Advance()
	if !newTime.IsZero() {
		t.Errorf("Advance for 24/7: expected zero time, got %v", newTime)
	}

	// Current time should not change
	if !it.GetDate().Equal(startTime) {
		t.Errorf("GetDate after Advance: got %v, want %v", it.GetDate(), startTime)
	}
}
