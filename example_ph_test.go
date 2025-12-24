package openinghours

import (
	"fmt"
	"time"
)

// Example of using public holiday support
func ExampleOpeningHours_SetHolidayChecker() {
	// Create a simple holiday checker that marks Jan 1, 2024 as a holiday
	type simpleHolidayChecker struct {
		holidays map[string]bool
	}

	hc := &simpleHolidayChecker{
		holidays: map[string]bool{
			"2024-01-01": true, // New Year's Day
		},
	}

	// Implement the IsHoliday method
	isHoliday := func(t time.Time) bool {
		key := t.Format("2006-01-02")
		return hc.holidays[key]
	}

	// Parse opening hours with PH support
	oh, _ := New("Mo-Fr 09:00-17:00; PH off")

	// Set the holiday checker (using anonymous struct that implements interface)
	oh.SetHolidayChecker(HolidayCheckerFunc(isHoliday))

	// Jan 1, 2024 is Monday (but also a holiday)
	holiday := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	fmt.Printf("Open on New Year's Day (Monday, PH): %v\n", oh.GetState(holiday))

	// Jan 2, 2024 is Tuesday (regular weekday, not a holiday)
	regularDay := time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC)
	fmt.Printf("Open on regular Tuesday: %v\n", oh.GetState(regularDay))

	// Output:
	// Open on New Year's Day (Monday, PH): false
	// Open on regular Tuesday: true
}

// HolidayCheckerFunc is a function type that implements HolidayChecker
type HolidayCheckerFunc func(time.Time) bool

func (f HolidayCheckerFunc) IsHoliday(t time.Time) bool {
	return f(t)
}
