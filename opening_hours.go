package openinghours

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// HolidayChecker is an interface that users can implement to provide public holiday information
type HolidayChecker interface {
	IsHoliday(t time.Time) bool
}

// SchoolHolidayChecker is an interface for checking school holidays
type SchoolHolidayChecker interface {
	IsSchoolHoliday(t time.Time) bool
}

// OpeningHours represents parsed opening hours
type OpeningHours struct {
	rules                []rule   // Primary group of rules (before ||)
	fallbackGroups       [][]rule // Fallback groups (after ||), each group separated by ||
	holidayChecker       HolidayChecker
	schoolHolidayChecker SchoolHolidayChecker
	latitude             float64 // Latitude for sunrise/sunset calculations
	longitude            float64 // Longitude for sunrise/sunset calculations
	hasCoordinates       bool    // Whether coordinates have been set
	warnings             []string // Warnings collected during parsing
}

type weekConstraint struct {
	weekStart    int // 1-53 for ISO week number
	weekEnd      int // 1-53 for ISO week number
	weekInterval int // 0=not set, interval for week ranges (e.g., /2 for every other week)
}

type rule struct {
	weekdays           []bool              // 7 bools, index 0=Sunday, 1=Monday, ..., 6=Saturday
	weekdayConstraints []weekdayConstraint // constraints like [1], [-1], [2-4]
	weekConstraints    []weekConstraint    // week number constraints
	timeRanges         []timeRange
	state              State
	comment            string
	yearStart          int  // 0=not set, otherwise the year (e.g., 2024)
	yearEnd            int  // 0=not set, otherwise the end year (e.g., 2026)
	yearInterval       int  // 0=not set, interval for year ranges (e.g., /2 for every other year)
	monthStart         int  // 0=not set, 1-12 for Jan-Dec
	monthEnd           int  // 0=not set, 1-12 for Jan-Dec
	dayStart           int  // 0=not set, 1-31 for day of month
	dayEnd             int  // 0=not set, 1-31 for day of month
	dayInterval        int  // 0=not set, interval for day ranges (e.g., /8 for every 8th day)
	isPH               bool // true if this rule applies to public holidays
	isSH               bool // true if this rule applies to school holidays
	phOffset           int  // days offset from public holiday (-1 = day before, +1 = day after, 0 = no offset/actual PH)
	isEaster           bool // true if this rule uses Easter
	easterOffset       int  // days offset from Easter (-2 = Good Friday, +1 = Easter Monday)
	isEasterRange      bool // true if this is an Easter date range
	easterOffsetEnd    int  // end offset for Easter ranges (e.g., "easter -2 days-easter +1 day")
	ruleGroup          int  // rules from same comma-separated expression share a group; 0 = no group
}

type weekdayConstraint struct {
	weekday int // 0-6 for Su-Sa
	nthFrom int // positive: nth occurrence (1 = first), negative: from end (-1 = last)
	nthTo   int // for ranges like [1-2], 0 if single value
}

type timeRange struct {
	start       int    // minutes from midnight (or -1 for variable)
	end         int    // minutes from midnight (or -1 for variable)
	openEnd     bool   // true if this is an open-ended range (e.g., 17:00+)
	startVar    string // "sunrise", "sunset", "dawn", "dusk" (empty if fixed time)
	endVar      string // "sunrise", "sunset", "dawn", "dusk" (empty if fixed time)
	startOffset int    // offset in minutes (+60 means +01:00)
	endOffset   int    // offset in minutes (+60 means +01:00)
	interval    int    // 0=not set, interval in minutes for periodic opening (e.g., 90 for 01:30)
}

// State represents the opening state
type State int

const (
	StateOpen State = iota
	StateClosed
	StateUnknown
)

// Interval represents a time interval when the business is open
type Interval struct {
	Start   time.Time
	End     time.Time
	Unknown bool   // true if this interval is "unknown" state
	Comment string // comment for this interval
}

var weekdayNames = map[string]int{
	"su": 0, "mo": 1, "tu": 2, "we": 3, "th": 4, "fr": 5, "sa": 6,
	"sun": 0, "mon": 1, "tue": 2, "wed": 3, "thu": 4, "fri": 5, "sat": 6,
	"sunday": 0, "monday": 1, "tuesday": 2, "wednesday": 3, "thursday": 4, "friday": 5, "saturday": 6,
	// German weekday names
	"sonntag": 0, "montag": 1, "dienstag": 2, "mittwoch": 3, "donnerstag": 4, "freitag": 5, "samstag": 6,
	"so": 0, "di": 2, "mi": 3, "do": 4,
}

var monthNames = map[string]int{
	"jan": 1, "feb": 2, "mar": 3, "apr": 4, "may": 5, "jun": 6,
	"jul": 7, "aug": 8, "sep": 9, "oct": 10, "nov": 11, "dec": 12,
	"january": 1, "february": 2, "march": 3, "april": 4, "june": 6,
	"july": 7, "august": 8, "september": 9, "october": 10, "november": 11, "december": 12,
	// German month names
	"januar": 1, "februar": 2, "märz": 3, "maerz": 3, "mai": 5, "juni": 6,
	"juli": 7, "oktober": 10, "dezember": 12,
}

var timeRangePattern = regexp.MustCompile(`^(\d{1,2}):(\d{2})\s*-\s*(\d{1,2}):(\d{2})(?:/(\d{2}):(\d{2}))?$`)
var singleTimePattern = regexp.MustCompile(`^(\d{1,2}):(\d{2})$`)
var openEndPattern = regexp.MustCompile(`^(\d{1,2}):(\d{2})\+$`)
var openEndRangePattern = regexp.MustCompile(`^(\d{1,2}):(\d{2})\s*-\s*(\d{1,2}):(\d{2})\+$`)
var variableTimePattern = regexp.MustCompile(`^\(?(sunrise|sunset|dawn|dusk)([+-]\d{2}:\d{2})?\)?$`)
var dotTimePattern = regexp.MustCompile(`\b(\d{1,2})\.(\d{2})\b`)
var ampmPattern = regexp.MustCompile(`(?i)(\d{1,2})(?::(\d{2}))?\s*([ap]\.?m\.?)`)
var phOffsetPattern = regexp.MustCompile(`^\s*([+-]?\d+)\s*days?\s*`)
var easterPattern = regexp.MustCompile(`^easter\s*([+-]?\d+\s*days?)?`)
var easterRangePattern = regexp.MustCompile(`^easter\s*([+-]?\d+)\s*days?\s*-\s*easter\s*([+-]?\d+)\s*days?\s*`)

// normalizeTimeString converts various time formats to standard HH:MM-HH:MM format
func normalizeTimeString(s string) string {
	// 0. Normalize different dash types to standard hyphen
	// En dash (U+2013), Em dash (U+2014), minus sign (U+2212) -> hyphen-minus (U+002D)
	s = strings.ReplaceAll(s, "–", "-") // En dash
	s = strings.ReplaceAll(s, "—", "-") // Em dash
	s = strings.ReplaceAll(s, "−", "-") // Minus sign

	// 0.5. Normalize alternative separators to hyphen
	// "to" and "through" can be used instead of "-" in time and weekday ranges
	// Use word boundaries to avoid replacing inside words
	toPattern := regexp.MustCompile(`(?i)\s+to\s+`)
	throughPattern := regexp.MustCompile(`(?i)\s+through\s+`)
	s = toPattern.ReplaceAllString(s, "-")
	s = throughPattern.ReplaceAllString(s, "-")

	// 1. Convert dots to colons in time: 10.00 -> 10:00
	// This must be done FIRST before short time format conversion
	// Pattern: \b(\d{1,2})\.(\d{2})\b
	s = dotTimePattern.ReplaceAllString(s, "$1:$2")

	// 2. Convert short time format to standard format: 10-12 -> 10:00-12:00
	// Must not match things like "week 1-10", "Mo-Fr", or "Jan 01-15" (day ranges)
	// Split by spaces and check each word to avoid converting week numbers or day ranges
	words := strings.Fields(s)
	for i, word := range words {
		// Only convert if the previous word isn't "week" (case insensitive)
		prevIsWeek := i > 0 && strings.ToLower(words[i-1]) == "week"
		// Also don't convert if previous word is a month name (it's a day range like "Jan 01-15")
		prevIsMonth := false
		if i > 0 {
			_, prevIsMonth = monthNames[strings.ToLower(words[i-1])]
		}
		if !prevIsWeek && !prevIsMonth {
			shortPattern := regexp.MustCompile(`^(\d{1,2})-(\d{1,2})$`)
			if match := shortPattern.FindStringSubmatch(word); match != nil {
				start, err1 := strconv.Atoi(match[1])
				end, err2 := strconv.Atoi(match[2])
				if err1 == nil && err2 == nil && start >= 0 && start <= 24 && end >= 0 && end <= 24 {
					words[i] = fmt.Sprintf("%d:00-%d:00", start, end)
				}
			}
		}
	}
	s = strings.Join(words, " ")

	// 3. Convert AM/PM to 24-hour format
	// Pattern for time with am/pm: (\d{1,2})(?::(\d{2}))?\s*(a\.?m\.?|p\.?m\.?)
	s = ampmPattern.ReplaceAllStringFunc(s, func(match string) string {
		// Parse the match
		parts := ampmPattern.FindStringSubmatch(match)
		if parts == nil {
			return match
		}

		hour, _ := strconv.Atoi(parts[1])
		minute := 0
		if parts[2] != "" {
			minute, _ = strconv.Atoi(parts[2])
		}
		ampm := strings.ToLower(parts[3])

		// Normalize ampm (remove dots and spaces)
		ampm = strings.ReplaceAll(ampm, ".", "")
		ampm = strings.TrimSpace(ampm)

		// Convert to 24-hour format
		if strings.HasPrefix(ampm, "p") && hour != 12 {
			// PM: add 12 hours (except for 12pm which stays 12)
			hour += 12
		} else if strings.HasPrefix(ampm, "a") && hour == 12 {
			// 12am is midnight (00:00)
			hour = 0
		}

		return fmt.Sprintf("%d:%02d", hour, minute)
	})

	return s
}

// New parses an opening hours string and returns an OpeningHours instance
func New(value string) (*OpeningHours, error) {
	oh := &OpeningHours{}
	if err := oh.parse(value); err != nil {
		return nil, err
	}
	return oh, nil
}

// SetHolidayChecker sets the holiday checker for this OpeningHours instance
func (oh *OpeningHours) SetHolidayChecker(hc HolidayChecker) {
	oh.holidayChecker = hc
}

// SetSchoolHolidayChecker sets the school holiday checker for this OpeningHours instance
func (oh *OpeningHours) SetSchoolHolidayChecker(shc SchoolHolidayChecker) {
	oh.schoolHolidayChecker = shc
}

// SetCoordinates sets the geographic coordinates for sunrise/sunset calculations
func (oh *OpeningHours) SetCoordinates(latitude, longitude float64) {
	oh.latitude = latitude
	oh.longitude = longitude
	oh.hasCoordinates = true
}

// GetWarnings returns any warnings that were collected during parsing
func (oh *OpeningHours) GetWarnings() []string {
	return oh.warnings
}

// addWarning adds a warning message to the warnings list
func (oh *OpeningHours) addWarning(msg string) {
	oh.warnings = append(oh.warnings, msg)
}

// resolveVariableTime resolves a variable time (sunrise, sunset, dawn, dusk) to minutes from midnight
func (oh *OpeningHours) resolveVariableTime(t time.Time, varType string, offset int) int {
	var baseTime int

	if oh.hasCoordinates {
		// Use calculated times based on coordinates
		switch varType {
		case "sunrise":
			baseTime = calculateSunrise(t, oh.latitude, oh.longitude)
		case "sunset":
			baseTime = calculateSunset(t, oh.latitude, oh.longitude)
		case "dawn":
			baseTime = calculateDawn(t, oh.latitude, oh.longitude)
		case "dusk":
			baseTime = calculateDusk(t, oh.latitude, oh.longitude)
		default:
			// Fallback to default
			baseTime = defaultSunrise
		}
	} else {
		// Use default times when no coordinates are set
		switch varType {
		case "sunrise":
			baseTime = defaultSunrise
		case "sunset":
			baseTime = defaultSunset
		case "dawn":
			baseTime = defaultDawn
		case "dusk":
			baseTime = defaultDusk
		default:
			baseTime = defaultSunrise
		}
	}

	// Apply offset
	result := baseTime + offset

	// Normalize to 0-1440 range
	for result < 0 {
		result += 1440
	}
	for result >= 1440 {
		result -= 1440
	}

	return result
}

// GetState returns true if open at the given time
func (oh *OpeningHours) GetState(t time.Time) bool {
	// Check for extended midnight continuation in comma-separated rule groups
	// This handles cases like "Su-Tu 11:00-01:00, We-Th 11:00-03:00" where
	// Tuesday's opening should extend to Wednesday 03:00 (using We's end time)
	if oh.checkExtendedMidnightContinuation(t) {
		return true
	}

	// Track ruleGroups that had a selector match but no time match
	// These groups "own" the time but none of their rules matched
	selectorMatchedGroups := make(map[int]bool)

	// Track if we've seen a selector match that should override
	var overridingRule *rule

	for i := len(oh.rules) - 1; i >= 0; i-- {
		r := oh.rules[i]
		if r.matchesWithOH(t, oh.holidayChecker, oh) {
			if r.state == StateUnknown {
				// Primary is unknown, check fallback groups
				return oh.getStateFromFallback(t)
			}
			// If an earlier rule matches, check if an overriding rule already claimed this day
			if overridingRule != nil {
				// Only apply override if the overriding rule has a MORE SPECIFIC selector
				// (i.e., doesn't match the same selector as this rule)
				if !oh.hasSameSelector(overridingRule, &r, t) {
					return false
				}
			}
			return r.state == StateOpen
		}
		// If the rule's selector (day/date) matches but time doesn't,
		// and the rule has state=Open (not Off/Closed), this rule may override
		// earlier rules. This implements the OSM spec where
		// "We 12:00-18:00" overrides "Mo-Fr 10:00-16:00" for Wednesday entirely.
		// But "off" rules like "Mo 15:00-16:00 off" should only apply during their time.
		if r.state == StateOpen && len(r.timeRanges) > 0 &&
			r.matchesSelectorWithOH(t, oh.holidayChecker, oh) {
			// This "open" rule owns this day but we're outside its time ranges
			// For comma-separated rules (same ruleGroup), don't override immediately -
			// check if other rules in the group might match
			if r.ruleGroup > 0 {
				selectorMatchedGroups[r.ruleGroup] = true
				continue
			}
			// Remember this rule as potentially overriding, but check if earlier rules match first
			if overridingRule == nil {
				overridingRule = &r
			}
		}
	}

	// If we had an overriding rule and no matching rule was found, return closed
	if overridingRule != nil {
		return false
	}

	// If any comma-separated group had selector matches but no full match, return closed
	if len(selectorMatchedGroups) > 0 {
		return false
	}

	// No match in primary, check fallback groups
	if len(oh.fallbackGroups) > 0 {
		return oh.getStateFromFallback(t)
	}
	return false
}

// hasSameSelector checks if two rules have the same selector (weekdays, dates, etc.)
// This is used to determine if a later rule should override an earlier rule.
// Rules with the same selector should not override each other - they union their times.
func (oh *OpeningHours) hasSameSelector(r1, r2 *rule, t time.Time) bool {
	// Check weekdays - if both have weekday constraints, compare them
	if r1.weekdays != nil && r2.weekdays != nil {
		// If all weekdays are the same, the selectors are the same
		same := true
		for i := 0; i < 7; i++ {
			if r1.weekdays[i] != r2.weekdays[i] {
				same = false
				break
			}
		}
		if same {
			// Also check month constraints
			if r1.monthStart == r2.monthStart && r1.monthEnd == r2.monthEnd {
				return true
			}
		}
	}

	// If neither has weekday constraints, check date constraints
	if r1.weekdays == nil && r2.weekdays == nil {
		if r1.monthStart == r2.monthStart && r1.monthEnd == r2.monthEnd &&
			r1.dayStart == r2.dayStart && r1.dayEnd == r2.dayEnd {
			return true
		}
	}

	return false
}

// checkExtendedMidnightContinuation checks if the current time falls within
// an extended midnight continuation period for comma-separated rule groups.
// For example, with "Su-Tu 11:00-01:00, We-Th 11:00-03:00", if it's Wednesday 02:00,
// this checks if Tuesday opened (using Su-Tu rule) and should extend to Wednesday 03:00
// (using We-Th's end time because Wednesday is covered by We-Th).
func (oh *OpeningHours) checkExtendedMidnightContinuation(t time.Time) bool {
	minuteOfDay := t.Hour()*60 + t.Minute()
	weekday := int(t.Weekday())
	prevWeekday := (weekday + 6) % 7

	// Group rules by ruleGroup
	rulesByGroup := make(map[int][]rule)
	for _, r := range oh.rules {
		if r.ruleGroup > 0 {
			rulesByGroup[r.ruleGroup] = append(rulesByGroup[r.ruleGroup], r)
		}
	}

	// For each group, check for extended midnight continuation
	for _, rules := range rulesByGroup {
		// Find if there's a rule where:
		// 1. Previous day matches the rule's weekdays
		// 2. The rule has midnight-spanning time (end <= start)
		var prevDayRule *rule
		var prevDayEndTime int = -1

		for i := range rules {
			r := &rules[i]
			if r.weekdays != nil && r.weekdays[prevWeekday] && len(r.timeRanges) > 0 {
				tr := r.timeRanges[0] // Use first time range
				if tr.end <= tr.start { // Midnight spanning
					prevDayRule = r
					prevDayEndTime = tr.end
					break
				}
			}
		}

		if prevDayRule == nil {
			continue
		}

		// Find if there's a rule in the same group for the current day
		// with a later end time
		for i := range rules {
			r := &rules[i]
			if r.weekdays != nil && r.weekdays[weekday] && len(r.timeRanges) > 0 {
				tr := r.timeRanges[0]
				extendedEnd := tr.end
				if tr.end <= tr.start { // Also midnight spanning
					extendedEnd = tr.end
				}
				// If current day's rule has a later end time and we're before it
				if extendedEnd > prevDayEndTime && minuteOfDay < extendedEnd {
					return true
				}
			}
		}
	}

	return false
}

// GetUnknown returns true if state is unknown at the given time
func (oh *OpeningHours) GetUnknown(t time.Time) bool {
	for i := len(oh.rules) - 1; i >= 0; i-- {
		r := oh.rules[i]
		if r.matchesWithOH(t, oh.holidayChecker, oh) {
			if r.state == StateUnknown {
				// Primary is unknown
				// If there are no fallback groups, return true (it's unknown)
				if len(oh.fallbackGroups) == 0 {
					return true
				}
				// If there are fallback groups, check if they can resolve it
				return oh.getUnknownFromFallback(t)
			}
			return false
		}
	}
	// No match in primary, check fallback groups
	if len(oh.fallbackGroups) > 0 {
		return oh.getUnknownFromFallback(t)
	}
	return false
}

// GetComment returns the comment for the given time, or empty string if no comment
func (oh *OpeningHours) GetComment(t time.Time) string {
	for i := len(oh.rules) - 1; i >= 0; i-- {
		r := oh.rules[i]
		if r.matchesWithOH(t, oh.holidayChecker, oh) {
			return r.comment
		}
	}
	// No match in primary, check fallback groups
	if len(oh.fallbackGroups) > 0 {
		return oh.getCommentFromFallback(t)
	}
	return ""
}

// getCommentFromFallback returns comment from matching fallback rule
func (oh *OpeningHours) getCommentFromFallback(t time.Time) string {
	for _, fallbackGroup := range oh.fallbackGroups {
		for i := len(fallbackGroup) - 1; i >= 0; i-- {
			if fallbackGroup[i].matchesWithOH(t, oh.holidayChecker, oh) {
				return fallbackGroup[i].comment
			}
		}
	}
	return ""
}

// GetMatchingRule returns the index of the rule that matches for the given time
// Returns -1 if no rule matches
func (oh *OpeningHours) GetMatchingRule(t time.Time) int {
	// Iterate through rules in reverse order (later rules have higher priority)
	for i := len(oh.rules) - 1; i >= 0; i-- {
		if oh.rules[i].matchesWithOH(t, oh.holidayChecker, oh) {
			return i
		}
	}
	return -1
}

// GetOpenDuration returns total open and unknown duration between from and to
func (oh *OpeningHours) GetOpenDuration(from, to time.Time) (openDuration, unknownDuration time.Duration) {
	// Iterate through time in 1-minute increments and sum up open/unknown minutes
	current := from

	for current.Before(to) {
		if oh.GetState(current) {
			openDuration += time.Minute
		} else if oh.GetUnknown(current) {
			unknownDuration += time.Minute
		}

		current = current.Add(time.Minute)
	}

	return openDuration, unknownDuration
}

// GetOpenIntervals returns all open/unknown intervals between from and to
func (oh *OpeningHours) GetOpenIntervals(from, to time.Time) []Interval {
	if from.After(to) || from.Equal(to) {
		return nil
	}

	var intervals []Interval

	// Helper function to find the next state change
	// This handles unknown states correctly, unlike GetNextChange
	findNextChange := func(t time.Time) time.Time {
		currentOpen := oh.GetState(t)
		currentUnknown := oh.GetUnknown(t)
		currentComment := oh.GetComment(t)

		// Check if always open or always closed (no weekdays, no time ranges)
		if len(oh.rules) == 1 && oh.rules[0].weekdays == nil && len(oh.rules[0].timeRanges) == 0 {
			// No next change for 24/7, always closed, or always unknown
			return time.Time{}
		}

		// Try GetNextChange first for better performance
		// It works well for open/closed states, but may not handle unknown states correctly
		nextChange := oh.GetNextChange(t)

		// If we have a next change, verify it actually represents a state change
		// This handles unknown states correctly
		if !nextChange.IsZero() {
			nextOpen := oh.GetState(nextChange)
			nextUnknown := oh.GetUnknown(nextChange)
			nextComment := oh.GetComment(nextChange)

			// If state actually changed, use this time
			if nextOpen != currentOpen || nextUnknown != currentUnknown || nextComment != currentComment {
				return nextChange
			}
		}

		// Fallback: Search minute by minute for state changes
		// This is slower but handles all cases including unknown states
		// Search up to 35 days for constrained weekdays like "4th Wednesday"
		checkTime := t.Add(time.Minute)
		endTime := t.Add(35 * 24 * time.Hour)

		for checkTime.Before(endTime) {
			nextOpen := oh.GetState(checkTime)
			nextUnknown := oh.GetUnknown(checkTime)
			nextComment := oh.GetComment(checkTime)

			// State changed if any of these changed
			if nextOpen != currentOpen || nextUnknown != currentUnknown || nextComment != currentComment {
				return checkTime
			}

			checkTime = checkTime.Add(time.Minute)
		}

		// No change found within 7 days
		return time.Time{}
	}

	current := from

	for current.Before(to) {
		isOpen := oh.GetState(current)
		isUnknown := oh.GetUnknown(current)
		comment := oh.GetComment(current)

		if isOpen || isUnknown {
			// Start of an open/unknown interval
			intervalStart := current

			// Find when this interval ends
			nextChange := findNextChange(current)

			var intervalEnd time.Time
			if nextChange.IsZero() {
				// No next change (e.g., 24/7) - interval goes to 'to'
				intervalEnd = to
			} else if nextChange.After(to) {
				// Next change is beyond our range
				intervalEnd = to
			} else {
				intervalEnd = nextChange
			}

			intervals = append(intervals, Interval{
				Start:   intervalStart,
				End:     intervalEnd,
				Unknown: isUnknown,
				Comment: comment,
			})

			current = intervalEnd
		} else {
			// Currently closed, find next opening
			nextChange := findNextChange(current)

			if nextChange.IsZero() || nextChange.After(to) {
				// No more changes or beyond our range
				break
			}

			current = nextChange
		}
	}

	return intervals
}

// IsEqualTo compares two OpeningHours objects for semantic equality.
// Two OpeningHours are considered equal if they produce the same state
// at all times over a representative period.
func (oh *OpeningHours) IsEqualTo(other *OpeningHours) bool {
	if other == nil {
		return false
	}

	// We sample states over 1 week at 15-minute intervals
	// to check for semantic equality. This covers most patterns.
	// For week-stable values, 1 week is sufficient.
	// For non-week-stable values (holidays, specific dates), we need longer periods.

	// Start from a reference date that ensures we cover all weekdays
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) // Monday

	// Sample for 1 week at 15-minute intervals for week-stable comparison
	sampleDuration := 7 * 24 * time.Hour
	sampleInterval := 15 * time.Minute

	current := start
	end := start.Add(sampleDuration)

	for current.Before(end) {
		state1 := oh.GetState(current)
		state2 := other.GetState(current)
		if state1 != state2 {
			return false
		}

		unknown1 := oh.GetUnknown(current)
		unknown2 := other.GetUnknown(current)
		if unknown1 != unknown2 {
			return false
		}

		comment1 := oh.GetComment(current)
		comment2 := other.GetComment(current)
		if comment1 != comment2 {
			return false
		}

		current = current.Add(sampleInterval)
	}

	return true
}

// GetStateString returns "open", "closed", or "unknown" for the given time
func (oh *OpeningHours) GetStateString(t time.Time) string {
	for i := len(oh.rules) - 1; i >= 0; i-- {
		r := oh.rules[i]
		if r.matchesWithOH(t, oh.holidayChecker, oh) {
			if r.state == StateUnknown {
				// Primary is unknown, check fallback groups
				return oh.getStateStringFromFallback(t)
			}
			switch r.state {
			case StateOpen:
				return "open"
			case StateClosed:
				return "closed"
			}
		}
	}
	// No match in primary, return closed (don't check fallback)
	return "closed"
}

// IsWeekStable returns true if the opening hours follow a stable weekly pattern
// (same hours repeat every week without variations like months, years, dates, holidays, or week numbers)
func (oh *OpeningHours) IsWeekStable() bool {
	// Check all rules (including fallback groups)
	allRules := oh.rules
	for _, fg := range oh.fallbackGroups {
		allRules = append(allRules, fg...)
	}

	for _, r := range allRules {
		// If rule has month constraints (except full year Jan-Dec), not stable
		if r.monthStart > 0 {
			// Jan-Dec (1-12) covers full year, so it's still week stable
			if !(r.monthStart == 1 && r.monthEnd == 12) {
				return false
			}
		}

		// If rule has year constraints, not stable
		if r.yearStart > 0 {
			return false
		}

		// If rule has week number constraints, not stable
		if len(r.weekConstraints) > 0 {
			return false
		}

		// If rule has constrained weekdays (Mo[1], Fr[-1]), not stable
		if len(r.weekdayConstraints) > 0 {
			return false
		}

		// If rule applies to public holidays, not stable
		if r.isPH {
			return false
		}

		// If rule applies to school holidays, not stable
		if r.isSH {
			return false
		}

		// If rule uses Easter, not stable
		if r.isEaster {
			return false
		}
	}

	return true
}

// GetNextChange returns the next time the opening state changes
func (oh *OpeningHours) GetNextChange(t time.Time) time.Time {
	currentState := oh.GetState(t)

	// Check if always open or always closed (no weekdays, no time ranges)
	if len(oh.rules) == 1 && oh.rules[0].weekdays == nil && len(oh.rules[0].timeRanges) == 0 {
		// No next change for 24/7 or always closed
		return time.Time{}
	}

	// Search for next change up to 35 days ahead
	// (needed for constrained weekdays like "4th Wednesday" which may be ~30 days away)
	searchTime := t
	for day := 0; day < 36; day++ {
		// For the first day, start from current time; for other days, start from midnight
		var startMinute int
		if day == 0 {
			startMinute = searchTime.Hour()*60 + searchTime.Minute()
		} else {
			startMinute = 0
			searchTime = time.Date(searchTime.Year(), searchTime.Month(), searchTime.Day()+1, 0, 0, 0, 0, searchTime.Location())
		}

		// Collect all transition times for this day
		transitions := make(map[int]bool) // minute -> has transition
		weekday := int(searchTime.Weekday())
		prevWeekday := (weekday + 6) % 7

		for _, r := range oh.rules {
			// Check if rule applies to this weekday for start times
			if r.weekdays != nil && r.weekdays[weekday] {
				// Add all time range boundaries for this day
				for _, tr := range r.timeRanges {
					// Resolve variable times for this specific day
					trStart := tr.start
					trEnd := tr.end
					if tr.startVar != "" {
						trStart = oh.resolveVariableTime(searchTime, tr.startVar, tr.startOffset)
					}
					if tr.endVar != "" {
						trEnd = oh.resolveVariableTime(searchTime, tr.endVar, tr.endOffset)
					}

					if trStart > startMinute || day > 0 {
						transitions[trStart] = true
					}
					// For midnight-spanning (end <= start), don't add end on same day
					if trEnd > trStart {
						if trEnd > startMinute || day > 0 {
							transitions[trEnd] = true
						}
					}
				}
			}

			// Check if PREVIOUS day had a midnight-spanning rule that ends today
			if r.weekdays != nil && r.weekdays[prevWeekday] {
				for _, tr := range r.timeRanges {
					trEnd := tr.end
					if tr.endVar != "" {
						trEnd = oh.resolveVariableTime(searchTime, tr.endVar, tr.endOffset)
					}
					// If end <= start, it spans midnight and ends on TODAY
					if trEnd <= tr.start {
						if trEnd > startMinute || day > 0 {
							transitions[trEnd] = true
						}
					}
				}
			}

			// Handle rules without weekday constraints
			if r.weekdays == nil {
				for _, tr := range r.timeRanges {
					trStart := tr.start
					trEnd := tr.end
					if tr.startVar != "" {
						trStart = oh.resolveVariableTime(searchTime, tr.startVar, tr.startOffset)
					}
					if tr.endVar != "" {
						trEnd = oh.resolveVariableTime(searchTime, tr.endVar, tr.endOffset)
					}

					if trStart > startMinute || day > 0 {
						transitions[trStart] = true
					}
					if trEnd > startMinute || day > 0 {
						transitions[trEnd] = true
					}
				}
			}
		}

		// Sort transitions and find first one where state changes
		sortedTimes := make([]int, 0, len(transitions))
		for minute := range transitions {
			sortedTimes = append(sortedTimes, minute)
		}

		// Simple bubble sort for small arrays
		for i := 0; i < len(sortedTimes); i++ {
			for j := i + 1; j < len(sortedTimes); j++ {
				if sortedTimes[i] > sortedTimes[j] {
					sortedTimes[i], sortedTimes[j] = sortedTimes[j], sortedTimes[i]
				}
			}
		}

		for _, minute := range sortedTimes {
			checkTime := time.Date(searchTime.Year(), searchTime.Month(), searchTime.Day(),
				minute/60, minute%60, 0, 0, searchTime.Location())

			// Check if state is different at this time
			if oh.GetState(checkTime) != currentState {
				return checkTime
			}
		}
	}

	// No change found within 35 days
	return time.Time{}
}

// GetNextChangeWithMaxDate returns the next time the opening state changes,
// but only if it occurs before or at maxdate.
// If no change is found before maxdate, returns zero time.
func (oh *OpeningHours) GetNextChangeWithMaxDate(t time.Time, maxdate time.Time) time.Time {
	currentState := oh.GetState(t)

	// Check if always open or always closed (no weekdays, no time ranges)
	if len(oh.rules) == 1 && oh.rules[0].weekdays == nil && len(oh.rules[0].timeRanges) == 0 {
		// No next change for 24/7 or always closed
		return time.Time{}
	}

	// Calculate days to search (limited by maxdate)
	maxDays := int(maxdate.Sub(t).Hours()/24) + 1
	if maxDays > 365 {
		maxDays = 365 // Safety limit
	}

	searchTime := t
	for day := 0; day < maxDays+1; day++ {
		// For the first day, start from current time; for other days, start from midnight
		var startMinute int
		if day == 0 {
			startMinute = searchTime.Hour()*60 + searchTime.Minute()
		} else {
			startMinute = 0
			searchTime = time.Date(searchTime.Year(), searchTime.Month(), searchTime.Day()+1, 0, 0, 0, 0, searchTime.Location())
		}

		// Stop if we've passed maxdate
		if searchTime.After(maxdate) && day > 0 {
			break
		}

		// Collect all transition times for this day
		transitions := make(map[int]bool) // minute -> has transition
		weekday := int(searchTime.Weekday())
		prevWeekday := (weekday + 6) % 7

		for _, r := range oh.rules {
			// Check if rule applies to this weekday for start times
			if r.weekdays != nil && r.weekdays[weekday] {
				// Add all time range boundaries for this day
				for _, tr := range r.timeRanges {
					// Resolve variable times for this specific day
					trStart := tr.start
					trEnd := tr.end
					if tr.startVar != "" {
						trStart = oh.resolveVariableTime(searchTime, tr.startVar, tr.startOffset)
					}
					if tr.endVar != "" {
						trEnd = oh.resolveVariableTime(searchTime, tr.endVar, tr.endOffset)
					}

					if trStart > startMinute || day > 0 {
						transitions[trStart] = true
					}
					// For midnight-spanning (end <= start), don't add end on same day
					if trEnd > trStart {
						if trEnd > startMinute || day > 0 {
							transitions[trEnd] = true
						}
					}
				}
			}

			// Check if PREVIOUS day had a midnight-spanning rule that ends today
			if r.weekdays != nil && r.weekdays[prevWeekday] {
				for _, tr := range r.timeRanges {
					trEnd := tr.end
					if tr.endVar != "" {
						trEnd = oh.resolveVariableTime(searchTime, tr.endVar, tr.endOffset)
					}
					// If end <= start, it spans midnight and ends on TODAY
					if trEnd <= tr.start {
						if trEnd > startMinute || day > 0 {
							transitions[trEnd] = true
						}
					}
				}
			}

			// Handle rules without weekday constraints
			if r.weekdays == nil {
				for _, tr := range r.timeRanges {
					trStart := tr.start
					trEnd := tr.end
					if tr.startVar != "" {
						trStart = oh.resolveVariableTime(searchTime, tr.startVar, tr.startOffset)
					}
					if tr.endVar != "" {
						trEnd = oh.resolveVariableTime(searchTime, tr.endVar, tr.endOffset)
					}

					if trStart > startMinute || day > 0 {
						transitions[trStart] = true
					}
					if trEnd > startMinute || day > 0 {
						transitions[trEnd] = true
					}
				}
			}
		}

		// Sort transitions and find first one where state changes
		sortedTimes := make([]int, 0, len(transitions))
		for minute := range transitions {
			sortedTimes = append(sortedTimes, minute)
		}

		// Simple bubble sort for small arrays
		for i := 0; i < len(sortedTimes); i++ {
			for j := i + 1; j < len(sortedTimes); j++ {
				if sortedTimes[i] > sortedTimes[j] {
					sortedTimes[i], sortedTimes[j] = sortedTimes[j], sortedTimes[i]
				}
			}
		}

		for _, minute := range sortedTimes {
			checkTime := time.Date(searchTime.Year(), searchTime.Month(), searchTime.Day(),
				minute/60, minute%60, 0, 0, searchTime.Location())

			// Check if we've exceeded maxdate
			if checkTime.After(maxdate) {
				return time.Time{}
			}

			// Check if state is different at this time
			if oh.GetState(checkTime) != currentState {
				return checkTime
			}
		}
	}

	// No change found within maxdate range
	return time.Time{}
}

// getStateFromFallback checks fallback groups and returns the state
// Returns the state from the first fallback group that doesn't return unknown
func (oh *OpeningHours) getStateFromFallback(t time.Time) bool {
	for _, fallbackGroup := range oh.fallbackGroups {
		for i := len(fallbackGroup) - 1; i >= 0; i-- {
			if fallbackGroup[i].matchesWithOH(t, oh.holidayChecker, oh) {
				if fallbackGroup[i].state == StateUnknown {
					// This fallback is also unknown, try next fallback group
					break
				}
				return fallbackGroup[i].state == StateOpen
			}
		}
	}
	return false
}

// getUnknownFromFallback checks fallback groups and returns if state is unknown
// Returns true only if all checked fallback groups also return unknown
func (oh *OpeningHours) getUnknownFromFallback(t time.Time) bool {
	for _, fallbackGroup := range oh.fallbackGroups {
		for i := len(fallbackGroup) - 1; i >= 0; i-- {
			if fallbackGroup[i].matchesWithOH(t, oh.holidayChecker, oh) {
				if fallbackGroup[i].state == StateUnknown {
					// This fallback is also unknown, try next fallback group
					break
				}
				// Found a definite state (open or closed), so not unknown
				return false
			}
		}
	}
	// All fallback groups returned unknown or no match
	return false
}

// getStateStringFromFallback checks fallback groups and returns the state string
func (oh *OpeningHours) getStateStringFromFallback(t time.Time) string {
	for _, fallbackGroup := range oh.fallbackGroups {
		for i := len(fallbackGroup) - 1; i >= 0; i-- {
			if fallbackGroup[i].matchesWithOH(t, oh.holidayChecker, oh) {
				if fallbackGroup[i].state == StateUnknown {
					// This fallback is also unknown, try next fallback group
					break
				}
				switch fallbackGroup[i].state {
				case StateOpen:
					return "open"
				case StateClosed:
					return "closed"
				}
			}
		}
	}
	return "closed"
}

// nthWeekdayOfMonth returns which occurrence (1-indexed) of the weekday this date is in its month
// e.g., if t is the 3rd Monday of the month, returns 3
func nthWeekdayOfMonth(t time.Time) int {
	day := t.Day()
	return (day-1)/7 + 1
}

// nthWeekdayFromEnd returns which occurrence from the end (1-indexed) the date is
// e.g., if t is the last Friday of the month, returns 1
func nthWeekdayFromEnd(t time.Time) int {
	// Get the last day of the month
	nextMonth := time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, t.Location())
	lastDay := nextMonth.Add(-24 * time.Hour).Day()
	return (lastDay-t.Day())/7 + 1
}

func (r *rule) matches(t time.Time, hc HolidayChecker) bool {
	return r.matchesWithOH(t, hc, nil)
}

// isOffsetHolidayDay checks if the given day is N days away from a holiday
// This is used to prevent regular rules from matching on days that should be handled by PH offset rules
func isOffsetHolidayDay(t time.Time, hc HolidayChecker, rules []rule) bool {
	if hc == nil {
		return false
	}

	// Check if any PH offset rule would apply to this day
	for _, r := range rules {
		if r.isPH && r.phOffset != 0 {
			// Check if this day is r.phOffset days after a holiday
			checkDate := t.AddDate(0, 0, -r.phOffset)
			if hc.IsHoliday(checkDate) && !hc.IsHoliday(t) {
				return true
			}
		}
	}
	return false
}

// matchesSelectorWithOH checks if the rule's selector (weekday, date, holiday, etc.)
// matches the given time, WITHOUT checking time ranges.
// This is used to determine if a later rule "owns" a day even if outside its time ranges.
func (r *rule) matchesSelectorWithOH(t time.Time, hc HolidayChecker, oh *OpeningHours) bool {
	// Rules without weekday constraints (time-only rules like "10:00-18:00") don't own any day
	if r.weekdays == nil && !r.isPH && !r.isSH && !r.isEaster &&
		r.monthStart == 0 && r.yearStart == 0 && len(r.weekConstraints) == 0 &&
		len(r.weekdayConstraints) == 0 {
		return false
	}

	// Check year constraints
	if r.yearStart > 0 {
		year := t.Year()
		if year < r.yearStart || year > r.yearEnd {
			return false
		}
		// Check year interval if specified (e.g., 2020-2030/2 means every other year)
		if r.yearInterval > 0 {
			yearOffset := year - r.yearStart
			if yearOffset%r.yearInterval != 0 {
				return false
			}
		}
	}

	// Check Easter rules
	if r.isEaster {
		easterDate := calculateEaster(t.Year())
		if r.isEasterRange {
			startDate := easterDate.AddDate(0, 0, r.easterOffset)
			endDate := easterDate.AddDate(0, 0, r.easterOffsetEnd)
			tDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
			startDay := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
			endDay := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, time.UTC)
			if tDay.Before(startDay) || tDay.After(endDay) {
				return false
			}
		} else {
			targetDate := easterDate.AddDate(0, 0, r.easterOffset)
			if t.Year() != targetDate.Year() || t.Month() != targetDate.Month() || t.Day() != targetDate.Day() {
				return false
			}
		}
		return true // Easter selector matches
	}

	// Check school holidays
	if r.isSH {
		// Need to check if it's actually a school holiday
		if oh != nil && oh.schoolHolidayChecker != nil && oh.schoolHolidayChecker.IsSchoolHoliday(t) {
			return true
		}
		return false
	}

	// Check public holidays
	if r.isPH {
		if hc != nil && hc.IsHoliday(t) {
			return true
		}
		return false
	}

	// Check month/day constraints
	if r.monthStart > 0 {
		month := int(t.Month())
		day := t.Day()

		if r.dayStart > 0 {
			inRange := false
			if r.monthStart == r.monthEnd {
				if month == r.monthStart && day >= r.dayStart && day <= r.dayEnd {
					// Check day interval if specified (e.g., Jan 01-31/8 means every 8th day)
					if r.dayInterval > 0 {
						dayOffset := day - r.dayStart
						if dayOffset%r.dayInterval == 0 {
							inRange = true
						}
					} else {
						inRange = true
					}
				}
			} else if r.monthStart < r.monthEnd {
				if month > r.monthStart && month < r.monthEnd {
					inRange = true
				} else if month == r.monthStart && day >= r.dayStart {
					inRange = true
				} else if month == r.monthEnd && day <= r.dayEnd {
					inRange = true
				}
			} else {
				if month > r.monthStart || month < r.monthEnd {
					inRange = true
				} else if month == r.monthStart && day >= r.dayStart {
					inRange = true
				} else if month == r.monthEnd && day <= r.dayEnd {
					inRange = true
				}
			}
			if !inRange {
				return false
			}
		} else {
			inRange := false
			if r.monthStart == r.monthEnd {
				if month == r.monthStart {
					inRange = true
				}
			} else if r.monthStart < r.monthEnd {
				if month >= r.monthStart && month <= r.monthEnd {
					inRange = true
				}
			} else {
				if month >= r.monthStart || month <= r.monthEnd {
					inRange = true
				}
			}
			if !inRange {
				return false
			}
		}
	}

	// Check week number constraints
	if len(r.weekConstraints) > 0 {
		_, week := t.ISOWeek()
		inRange := false
		for _, wc := range r.weekConstraints {
			if wc.weekStart == wc.weekEnd {
				if week == wc.weekStart {
					inRange = true
					break
				}
			} else {
				if week >= wc.weekStart && week <= wc.weekEnd {
					if wc.weekInterval > 0 {
						weekOffset := week - wc.weekStart
						if weekOffset%wc.weekInterval == 0 {
							inRange = true
							break
						}
					} else {
						inRange = true
						break
					}
				}
			}
		}
		if !inRange {
			return false
		}
	}

	// Check weekday constraints
	if len(r.weekdayConstraints) > 0 {
		weekday := int(t.Weekday())
		constraintMatched := false
		for _, constraint := range r.weekdayConstraints {
			if constraint.weekday != weekday {
				continue
			}
			nthFromStart := nthWeekdayOfMonth(t)
			nthFromEnd := nthWeekdayFromEnd(t)
			if constraint.nthFrom > 0 {
				if constraint.nthTo == 0 {
					if nthFromStart == constraint.nthFrom {
						constraintMatched = true
						break
					}
				} else {
					if nthFromStart >= constraint.nthFrom && nthFromStart <= constraint.nthTo {
						constraintMatched = true
						break
					}
				}
			} else {
				if constraint.nthTo == 0 {
					if nthFromEnd == -constraint.nthFrom {
						constraintMatched = true
						break
					}
				} else {
					if nthFromEnd >= -constraint.nthTo && nthFromEnd <= -constraint.nthFrom {
						constraintMatched = true
						break
					}
				}
			}
		}
		return constraintMatched
	}

	// Check regular weekday
	if r.weekdays != nil {
		weekday := int(t.Weekday())
		return r.weekdays[weekday]
	}

	// No selector - matches everything (time-only rule)
	return true
}

func (r *rule) matchesWithOH(t time.Time, hc HolidayChecker, oh *OpeningHours) bool {
	// Check year constraints first
	if r.yearStart > 0 {
		year := t.Year()
		if year < r.yearStart || year > r.yearEnd {
			return false
		}
		// Check year interval if specified (e.g., 2020-2030/2 means every other year)
		if r.yearInterval > 0 {
			yearOffset := year - r.yearStart
			if yearOffset%r.yearInterval != 0 {
				return false
			}
		}
	}

	// Check Easter rules
	if r.isEaster {
		easterDate := calculateEaster(t.Year())

		if r.isEasterRange {
			// Check if we're in the Easter date range
			startDate := easterDate.AddDate(0, 0, r.easterOffset)
			endDate := easterDate.AddDate(0, 0, r.easterOffsetEnd)

			// Normalize dates to start of day for comparison
			tDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
			startDay := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
			endDay := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, time.UTC)

			// Check if current date is within the range (inclusive)
			if tDay.Before(startDay) || tDay.After(endDay) {
				return false
			}

			// We're in the date range, check time ranges if any
			if len(r.timeRanges) == 0 {
				return true
			}
			// Continue to time range checks below
		} else {
			// Single Easter date
			targetDate := easterDate.AddDate(0, 0, r.easterOffset)

			// Check if current date matches the Easter target date
			if t.Year() != targetDate.Year() || t.Month() != targetDate.Month() || t.Day() != targetDate.Day() {
				return false
			}
			// Continue with time range checks below
		}
	}

	// Check if this is a SH (school holiday) rule
	if r.isSH {
		// This rule only applies on school holidays
		if oh == nil || oh.schoolHolidayChecker == nil || !oh.schoolHolidayChecker.IsSchoolHoliday(t) {
			return false
		}
		// If it's an SH rule and today is a school holiday, continue checking time ranges below
	} else {
		// This is a regular rule (not SH)
		// If today is a school holiday and we have a school holiday checker, don't match regular rules
		// This allows SH rules to override regular weekday rules
		if oh != nil && oh.schoolHolidayChecker != nil && oh.schoolHolidayChecker.IsSchoolHoliday(t) {
			return false
		}
	}

	// Check if this is a PH (public holiday) rule
	if r.isPH {
		// This rule only applies on public holidays (or offset from holidays)
		if hc == nil {
			return false
		}

		if r.phOffset == 0 {
			// Actual holiday (no offset)
			if !hc.IsHoliday(t) {
				return false
			}
		} else {
			// Check if offset days ago/ahead is a holiday
			// To check if today is N days after a holiday, we check if (today - N days) is a holiday
			checkDate := t.AddDate(0, 0, -r.phOffset)
			if !hc.IsHoliday(checkDate) {
				return false
			}
			// Also make sure TODAY is not a holiday (for offset rules)
			// This ensures offset rules don't match on the holiday itself
			if hc.IsHoliday(t) {
				return false
			}
		}
		// If it's a PH rule and conditions are met, continue checking time ranges below
	} else {
		// This is a regular rule (not PH)
		// If today is a public holiday and we have a holiday checker, don't match regular rules
		// This allows PH rules to override regular weekday rules
		if hc != nil && hc.IsHoliday(t) {
			return false
		}
		// Also check if today is a "PH offset day" (e.g., day after/before a holiday)
		// This allows PH offset rules to override regular weekday rules
		if oh != nil && isOffsetHolidayDay(t, hc, oh.rules) {
			return false
		}
	}

	// Check month/day constraints first
	if r.monthStart > 0 {
		month := int(t.Month())
		day := t.Day()

		// Check if we're in a month-day range
		if r.dayStart > 0 {
			// We have specific day constraints
			// Check if we're in the range (which might span months)
			inRange := false

			if r.monthStart == r.monthEnd {
				// Same month range
				if month == r.monthStart && day >= r.dayStart && day <= r.dayEnd {
					// Check day interval if specified (e.g., Jan 01-31/8 means every 8th day)
					if r.dayInterval > 0 {
						dayOffset := day - r.dayStart
						if dayOffset%r.dayInterval == 0 {
							inRange = true
						}
					} else {
						inRange = true
					}
				}
			} else if r.monthStart < r.monthEnd {
				// Normal month range (e.g., Mar-May)
				if month > r.monthStart && month < r.monthEnd {
					inRange = true
				} else if month == r.monthStart && day >= r.dayStart {
					inRange = true
				} else if month == r.monthEnd && day <= r.dayEnd {
					inRange = true
				}
			} else {
				// Wrapping month range (e.g., Dec-Jan)
				if month > r.monthStart || month < r.monthEnd {
					inRange = true
				} else if month == r.monthStart && day >= r.dayStart {
					inRange = true
				} else if month == r.monthEnd && day <= r.dayEnd {
					inRange = true
				}
			}

			if !inRange {
				return false
			}
		} else {
			// Just month constraint, no specific days
			inRange := false
			if r.monthStart == r.monthEnd {
				// Single month
				if month == r.monthStart {
					inRange = true
				}
			} else if r.monthStart < r.monthEnd {
				// Normal month range
				if month >= r.monthStart && month <= r.monthEnd {
					inRange = true
				}
			} else {
				// Wrapping month range
				if month >= r.monthStart || month <= r.monthEnd {
					inRange = true
				}
			}

			if !inRange {
				return false
			}
		}
	}

	// Check week number constraints
	if len(r.weekConstraints) > 0 {
		_, week := t.ISOWeek()
		inRange := false
		for _, wc := range r.weekConstraints {
			if wc.weekStart == wc.weekEnd {
				// Single week
				if week == wc.weekStart {
					inRange = true
					break
				}
			} else {
				// Week range
				if week >= wc.weekStart && week <= wc.weekEnd {
					if wc.weekInterval > 0 {
						weekOffset := week - wc.weekStart
						if weekOffset%wc.weekInterval == 0 {
							inRange = true
							break
						}
					} else {
						inRange = true
						break
					}
				}
			}
		}
		if !inRange {
			return false
		}
	}

	// Check weekday constraints if present
	constraintMatched := false
	if len(r.weekdayConstraints) > 0 {
		weekday := int(t.Weekday())

		for _, constraint := range r.weekdayConstraints {
			if constraint.weekday != weekday {
				continue
			}

			// Calculate which occurrence this is
			nthFromStart := nthWeekdayOfMonth(t)
			nthFromEnd := nthWeekdayFromEnd(t)

			// Check if this matches the constraint
			if constraint.nthFrom > 0 {
				// Positive index (from start)
				if constraint.nthTo == 0 {
					// Single value like [1]
					if nthFromStart == constraint.nthFrom {
						constraintMatched = true
						break
					}
				} else {
					// Range like [1-2]
					if nthFromStart >= constraint.nthFrom && nthFromStart <= constraint.nthTo {
						constraintMatched = true
						break
					}
				}
			} else {
				// Negative index (from end)
				if constraint.nthTo == 0 {
					// Single value like [-1]
					if nthFromEnd == -constraint.nthFrom {
						constraintMatched = true
						break
					}
				} else {
					// Range like [-2--1] (not typical but handle it)
					if nthFromEnd >= -constraint.nthTo && nthFromEnd <= -constraint.nthFrom {
						constraintMatched = true
						break
					}
				}
			}
		}

		if !constraintMatched {
			return false
		}

		// If matched and no time ranges, return true
		if len(r.timeRanges) == 0 {
			return true
		}

		// Continue to check time ranges below
	}

	// If no time ranges, matches all times for the day (if weekday matches)
	if len(r.timeRanges) == 0 {
		if r.weekdays != nil {
			weekday := int(t.Weekday())
			return r.weekdays[weekday]
		}
		return true
	}

	minuteOfDay := t.Hour()*60 + t.Minute()

	for _, tr := range r.timeRanges {
		// Resolve variable times if present
		trStart := tr.start
		trEnd := tr.end

		if tr.startVar != "" && oh != nil {
			trStart = oh.resolveVariableTime(t, tr.startVar, tr.startOffset)
		}
		if tr.endVar != "" && oh != nil {
			trEnd = oh.resolveVariableTime(t, tr.endVar, tr.endOffset)
		}

		// Check if this is a midnight-spanning range or extended hours (25:00, 26:00 etc.)
		// Extended hours means end time > 24:00 (1440 minutes), which wraps to next day
		spansMidnight := trEnd <= trStart
		extendedHours := trEnd > 1440

		// For extended hours, normalize the end time to the next day equivalent
		if extendedHours {
			trEnd = trEnd - 1440 // 25:00 becomes 01:00, 26:00 becomes 02:00
			spansMidnight = true
		}

		if spansMidnight {
			// For midnight-spanning ranges, we need special handling
			if r.weekdays != nil && !constraintMatched {
				// With weekday constraints (but not constrained weekdays):
				// We match if we're on a valid start day and >= start time
				// OR if we're on the next day after a valid start day and < end time
				weekday := int(t.Weekday())
				prevWeekday := (weekday + 6) % 7 // Previous day

				// Case 1: Current day is a valid start day and time >= start
				if r.weekdays[weekday] && minuteOfDay >= trStart {
					return true
				}

				// Case 2: Previous day was a valid start day and current time < end
				if r.weekdays[prevWeekday] && minuteOfDay < trEnd {
					return true
				}
			} else {
				// Without weekday constraints or with constrained weekdays: match if time >= start OR time < end
				if minuteOfDay >= trStart || minuteOfDay < trEnd {
					return true
				}
			}
		} else {
			// Normal non-spanning range
			if r.weekdays != nil && !constraintMatched {
				weekday := int(t.Weekday())
				if !r.weekdays[weekday] {
					continue
				}
			}

			if minuteOfDay >= trStart && minuteOfDay < trEnd {
				// Check time interval if specified (e.g., 10:00-16:00/01:30)
				// Interval means alternating open/closed periods within the range
				if tr.interval > 0 {
					offset := minuteOfDay - trStart
					slotIndex := offset / tr.interval
					// Even slots are open, odd slots are closed
					if slotIndex%2 == 0 {
						return true
					}
					// In a closed slot, don't match
					continue
				}
				return true
			}
		}
	}

	return false
}

func (oh *OpeningHours) parse(value string) error {
	value = strings.TrimSpace(value)

	// Check for short time format BEFORE normalization
	// Pattern: number-number that is NOT preceded or followed by a colon or another digit
	shortTimePattern := regexp.MustCompile(`(?:^|[^\d:])(\d{1,2})-(\d{1,2})(?:[^\d:]|$)`)
	if shortTimePattern.MatchString(value) {
		// Additional check: make sure the matched portion doesn't have colons
		// by checking the entire value doesn't have time format around it
		match := shortTimePattern.FindStringSubmatch(value)
		if len(match) >= 3 {
			start, err1 := strconv.Atoi(match[1])
			end, err2 := strconv.Atoi(match[2])
			// Only warn if both are valid hour values (0-24)
			if err1 == nil && err2 == nil && start >= 0 && start <= 24 && end >= 0 && end <= 24 {
				oh.addWarning("Abbreviated time format: use HH:MM instead of H")
			}
		}
	}

	value = normalizeTimeString(value)

	// Handle special cases
	lower := strings.ToLower(value)
	if lower == "24/7" || lower == "open" {
		oh.rules = []rule{{state: StateOpen}}
		return nil
	}

	if lower == "off" || lower == "closed" {
		oh.rules = []rule{{state: StateClosed}}
		return nil
	}

	// Handle "24/7 closed", "24/7 off", "open closed", etc.
	if strings.HasPrefix(lower, "24/7 closed") || strings.HasPrefix(lower, "24/7 off") ||
		strings.HasPrefix(lower, "open closed") || strings.HasPrefix(lower, "open off") ||
		strings.HasPrefix(lower, "00:00-24:00 closed") || strings.HasPrefix(lower, "00:00-24:00 off") {
		// Extract possible comment
		comment := ""
		if idx := strings.Index(value, "\""); idx != -1 {
			endIdx := strings.LastIndex(value, "\"")
			if endIdx > idx {
				comment = value[idx+1 : endIdx]
			}
		}
		oh.rules = []rule{{state: StateClosed, comment: comment}}
		return nil
	}

	// Split by || for fallback groups
	groups := strings.Split(value, "||")

	// Parse primary group (first group before ||)
	if err := oh.parseRuleGroup(groups[0], &oh.rules); err != nil {
		return err
	}

	// Parse fallback groups (groups after ||)
	for i := 1; i < len(groups); i++ {
		var fallbackRules []rule
		if err := oh.parseRuleGroup(groups[i], &fallbackRules); err != nil {
			return err
		}
		oh.fallbackGroups = append(oh.fallbackGroups, fallbackRules)
	}

	if len(oh.rules) == 0 {
		return fmt.Errorf("unable to parse: %s", value)
	}

	// Check for redundant 24/7: if first rule is 24/7 and there are more rules
	if len(oh.rules) > 1 {
		firstRule := oh.rules[0]
		// Check if first rule is 24/7 (open all the time with no constraints)
		if firstRule.state == StateOpen && firstRule.weekdays == nil &&
		   len(firstRule.timeRanges) == 0 && firstRule.monthStart == 0 &&
		   firstRule.yearStart == 0 && len(firstRule.weekConstraints) == 0 &&
		   len(firstRule.weekdayConstraints) == 0 && !firstRule.isPH &&
		   !firstRule.isSH && !firstRule.isEaster {
			oh.addWarning("Redundant 24/7: additional rules override parts of 24/7")
		}
	}

	return nil
}

// parseRuleGroup parses a group of rules separated by semicolons
func (oh *OpeningHours) parseRuleGroup(groupStr string, rules *[]rule) error {
	groupStr = strings.TrimSpace(groupStr)
	if groupStr == "" {
		return nil
	}

	// Handle special cases within a group
	lower := strings.ToLower(groupStr)
	if lower == "24/7" || lower == "open" {
		*rules = append(*rules, rule{state: StateOpen})
		return nil
	}

	if lower == "off" || lower == "closed" {
		*rules = append(*rules, rule{state: StateClosed})
		return nil
	}

	// Counter for ruleGroup IDs (used for comma-separated rules)
	ruleGroupCounter := 1

	// Split by semicolon for multiple rules
	ruleParts := strings.Split(groupStr, ";")
	for _, rulePart := range ruleParts {
		rulePart = strings.TrimSpace(rulePart)
		if rulePart == "" {
			continue
		}

		// Check if this rule has comma-separated years
		// We need to expand it into multiple rules
		_, _, _, _, years, err := parseYearWithList(rulePart)
		if err != nil {
			return err
		}

		if len(years) > 0 {
			// Extract the part after the years
			match := yearPattern.FindString(rulePart)
			if match == "" {
				return fmt.Errorf("failed to parse years from: %s", rulePart)
			}
			remainingPart := strings.TrimSpace(rulePart[len(match):])

			// Create a rule for each year
			for _, year := range years {
				yearRule := fmt.Sprintf("%d %s", year, remainingPart)
				r, err := parseRule(yearRule, oh)
				if err != nil {
					return err
				}
				*rules = append(*rules, r)
			}
		} else {
			// First, expand any month lists (e.g., "Jun-Aug,Dec Mo 10:00-12:00")
			monthExpandedRules := expandMonthList(rulePart)

			for _, monthRule := range monthExpandedRules {
				// Check if this rule has comma-separated weekday+time combinations
				// e.g., "Mo-Fr 10:00-16:00, We 12:00-18:00" should be split into two rules
				subRules := splitByCommaOutsideBracketsAndTime(monthRule)

				// If multiple sub-rules, they share a ruleGroup (comma-separated = merge, not override)
				groupID := 0
				if len(subRules) > 1 {
					groupID = ruleGroupCounter
					ruleGroupCounter++
				}

				for _, subRule := range subRules {
					r, err := parseRule(subRule, oh)
					if err != nil {
						return err
					}
					r.ruleGroup = groupID
					*rules = append(*rules, r)
				}
			}
		}
	}

	return nil
}

// splitByCommaOutsideBracketsAndTime splits a rule by comma, but only when
// both parts are complete weekday+time combinations
// e.g., "Mo-Fr 10:00-16:00, We 12:00-18:00" -> ["Mo-Fr 10:00-16:00", "We 12:00-18:00"]
// but "Mo,Th 10:00-12:00" stays as one part (Mo,Th share the time)
// and "10:00-12:00, 14:00-18:00" stays together (just time ranges)
func splitByCommaOutsideBracketsAndTime(s string) []string {
	var parts []string
	var current strings.Builder
	bracketDepth := 0

	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		ch := runes[i]
		if ch == '[' {
			bracketDepth++
			current.WriteRune(ch)
		} else if ch == ']' {
			bracketDepth--
			current.WriteRune(ch)
		} else if ch == ',' && bracketDepth == 0 {
			// Check if BOTH current part and rest are complete weekday+time combinations
			currentPart := strings.TrimSpace(current.String())
			rest := strings.TrimSpace(string(runes[i+1:]))

			// Only split if current part also has a time, and rest starts with weekday+time
			if hasWeekdayAndTime(currentPart) && startsWithWeekday(rest) {
				// Split here - both parts are complete weekday+time combinations
				parts = append(parts, currentPart)
				current.Reset()
			} else {
				// Keep the comma - this is either a weekday list or time range separator
				current.WriteRune(ch)
			}
		} else {
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, strings.TrimSpace(current.String()))
	}

	if len(parts) == 0 {
		return []string{s}
	}
	return parts
}

// hasWeekdayAndTime checks if a string contains both weekday and time components
func hasWeekdayAndTime(s string) bool {
	s = strings.TrimSpace(s)

	// Must have a space separating weekday from time
	spaceIdx := strings.Index(s, " ")
	if spaceIdx < 0 {
		return false
	}

	weekdayPart := s[:spaceIdx]
	timePart := strings.TrimSpace(s[spaceIdx+1:])

	// Check weekday part contains weekday
	if !containsWeekday(weekdayPart) {
		return false
	}

	// Check time part contains a time (starts with digit or variable time)
	if len(timePart) == 0 {
		return false
	}

	// Time should start with digit
	if timePart[0] >= '0' && timePart[0] <= '9' {
		return true
	}

	// Could be a variable time
	variableTimes := []string{"sunrise", "sunset", "dawn", "dusk"}
	lowerTime := strings.ToLower(timePart)
	for _, vt := range variableTimes {
		if strings.HasPrefix(lowerTime, vt) {
			return true
		}
	}

	return false
}

// startsWithWeekdayAndTime checks if a string starts with a weekday+time combination
// like "We 12:00-18:00" or "Mo-Fr 10:00-16:00", not just a weekday like "Mo,Th"
func startsWithWeekday(s string) bool {
	// Pattern: weekday (with optional range/list) followed by space and time
	// We need to find a weekday selector followed by a space and a time
	s = strings.TrimSpace(s)
	if len(s) < 7 { // Minimum: "Mo 0:00"
		return false
	}

	// Find the first space
	spaceIdx := strings.Index(s, " ")
	if spaceIdx < 0 {
		return false // No space, can't be weekday+time
	}

	// The part before space should be a weekday selector
	weekdayPart := s[:spaceIdx]
	timePart := strings.TrimSpace(s[spaceIdx+1:])

	// Check if weekday part looks like weekday selectors (Mo, Mo-Fr, Mo,Tu,We, etc.)
	// It should contain at least one weekday abbreviation
	if !containsWeekday(weekdayPart) {
		return false
	}

	// Check if the time part starts with a time (digits followed by : or .)
	if len(timePart) == 0 {
		return false
	}

	// Time should start with digit
	if timePart[0] < '0' || timePart[0] > '9' {
		// Could also be a variable time like "sunrise"
		variableTimes := []string{"sunrise", "sunset", "dawn", "dusk"}
		lowerTime := strings.ToLower(timePart)
		for _, vt := range variableTimes {
			if strings.HasPrefix(lowerTime, vt) {
				return true
			}
		}
		return false
	}

	return true
}

// containsWeekday checks if a string contains a weekday abbreviation
func containsWeekday(s string) bool {
	lower := strings.ToLower(s)
	weekdays := []string{"mo", "tu", "we", "th", "fr", "sa", "su"}
	for _, wd := range weekdays {
		if strings.Contains(lower, wd) {
			return true
		}
	}
	return false
}

func parseRule(s string, oh *OpeningHours) (rule, error) {
	r := rule{state: StateOpen}

	// Extract comment if present (quoted string at the end)
	s, comment := extractComment(s, oh)
	r.comment = comment

	// Handle special cases
	lower := strings.ToLower(s)
	if lower == "24/7" || lower == "open" {
		return rule{state: StateOpen, comment: comment}, nil
	}
	if lower == "off" || lower == "closed" {
		return rule{state: StateClosed, comment: comment}, nil
	}

	// Check for state at the end (off, closed, open, unknown)
	lower = strings.ToLower(s)
	if strings.HasSuffix(lower, " off") {
		r.state = StateClosed
		s = strings.TrimSpace(s[:len(s)-len(" off")])
	} else if strings.HasSuffix(lower, " closed") {
		r.state = StateClosed
		s = strings.TrimSpace(s[:len(s)-len(" closed")])
	} else if strings.HasSuffix(lower, " open") {
		s = strings.TrimSpace(s[:len(s)-len(" open")])
	} else if strings.HasSuffix(lower, " unknown") {
		r.state = StateUnknown
		s = strings.TrimSpace(s[:len(s)-len(" unknown")])
	}

	// Try to extract year first
	s, yearStart, yearEnd, yearInterval, years, err := parseYearWithList(s)
	if err != nil {
		return r, err
	}

	// If we have multiple years (comma-separated), this is handled in parse() not here
	// We should never get here with multiple years
	if len(years) > 0 {
		return r, fmt.Errorf("internal error: multiple years should be handled in parse()")
	}

	r.yearStart = yearStart
	r.yearEnd = yearEnd
	r.yearInterval = yearInterval

	// Parse week number constraints first (e.g., "week 01 Jan Mo" - need to extract week before month)
	s, weekConstraints, err := parseWeekNumbers(s)
	if err != nil {
		return r, err
	}
	r.weekConstraints = weekConstraints

	// Try to extract month/date ranges
	s, monthStart, monthEnd, dayStart, dayEnd, dayInterval, err := parseMonthDate(s)
	if err != nil {
		return r, err
	}
	r.monthStart = monthStart
	r.monthEnd = monthEnd
	r.dayStart = dayStart
	r.dayEnd = dayEnd
	r.dayInterval = dayInterval

	// Check for Easter patterns
	lower = strings.ToLower(s)
	if strings.HasPrefix(lower, "easter") {
		// First check for Easter date range (e.g., "easter -2 days-easter +1 day")
		if match := easterRangePattern.FindStringSubmatch(s); match != nil {
			r.isEaster = true
			r.isEasterRange = true

			// Parse start offset
			offsetStr := strings.TrimSpace(match[1])
			r.easterOffset, _ = strconv.Atoi(offsetStr)

			// Parse end offset
			offsetEndStr := strings.TrimSpace(match[2])
			r.easterOffsetEnd, _ = strconv.Atoi(offsetEndStr)

			// Remove the matched part and continue parsing
			s = strings.TrimSpace(s[len(match[0]):])

			// Parse any remaining time ranges or state
			if s != "" {
				timeRanges, err := parseTimeRanges(s, oh)
				if err != nil {
					return r, err
				}
				r.timeRanges = timeRanges
			}
			return r, nil
		}

		// Single Easter date (e.g., "easter" or "easter +1 day")
		if match := easterPattern.FindStringSubmatch(s); match != nil {
			r.isEaster = true

			if match[1] != "" {
				// Parse offset like "+1 day" or "-2 days"
				offsetStr := strings.TrimSpace(match[1])
				offsetStr = strings.TrimSuffix(offsetStr, "days")
				offsetStr = strings.TrimSuffix(offsetStr, "day")
				offsetStr = strings.TrimSpace(offsetStr)
				r.easterOffset, _ = strconv.Atoi(offsetStr)
			}

			// Remove the matched part and continue parsing time ranges
			s = strings.TrimSpace(s[len(match[0]):])

			if s != "" {
				timeRanges, err := parseTimeRanges(s, oh)
				if err != nil {
					return r, err
				}
				r.timeRanges = timeRanges
			}
			return r, nil
		}
	}

	// Check if this is a PH (public holiday) rule
	if strings.HasPrefix(strings.ToUpper(s), "PH") {
		r.isPH = true
		// Remove "PH" prefix
		s = strings.TrimSpace(s[2:])

		// Check for offset like "+1 day" or "-1 day"
		// Pattern: optional space, then +N or -N, then "day" or "days"
		if match := phOffsetPattern.FindStringSubmatch(s); match != nil {
			offset, err := strconv.Atoi(match[1])
			if err != nil {
				return r, fmt.Errorf("invalid PH offset: %s", match[1])
			}
			r.phOffset = offset
			// Remove the offset part from the string
			s = strings.TrimSpace(s[len(match[0]):])
		}

		// Parse the rest (may contain weekdays and/or time ranges)
		if s != "" {
			// Try to parse weekdays and time ranges
			weekdays, constraints, timeStr, err := parseWeekdaysAndTimeWithConstraints(s)
			if err != nil {
				return r, err
			}
			r.weekdays = weekdays
			r.weekdayConstraints = constraints

			if timeStr != "" {
				timeRanges, err := parseTimeRanges(timeStr, oh)
				if err != nil {
					return r, err
				}
				r.timeRanges = timeRanges
			}
		}
		return r, nil
	}

	// Check if this is a SH (school holiday) rule
	if strings.HasPrefix(strings.ToUpper(s), "SH") {
		r.isSH = true
		// Remove "SH" prefix and parse the rest (may contain weekdays and/or time ranges)
		s = strings.TrimSpace(s[2:])
		if s != "" {
			// Try to parse weekdays and time ranges
			weekdays, constraints, timeStr, err := parseWeekdaysAndTimeWithConstraints(s)
			if err != nil {
				return r, err
			}
			r.weekdays = weekdays
			r.weekdayConstraints = constraints

			if timeStr != "" {
				timeRanges, err := parseTimeRanges(timeStr, oh)
				if err != nil {
					return r, err
				}
				r.timeRanges = timeRanges
			}
		}
		return r, nil
	}

	// Try to extract weekdays and time ranges
	weekdays, constraints, timeStr, hasPH, hasSH, err := parseWeekdaysAndTimeWithConstraintsAndHolidays(s)
	if err != nil {
		return r, err
	}

	r.weekdays = weekdays
	r.weekdayConstraints = constraints

	// Set holiday flags if PH/SH were found in the weekday list (e.g., "Su,PH off")
	if hasPH {
		r.isPH = true
	}
	if hasSH {
		r.isSH = true
	}

	if timeStr != "" {
		timeRanges, err := parseTimeRanges(timeStr, oh)
		if err != nil {
			return r, err
		}
		r.timeRanges = timeRanges
	}

	return r, nil
}

var yearPattern = regexp.MustCompile(`^(\d{4}(?:,\d{4})*)(?:-(\d{4})(/\d+)?|\+)?\s+`)

func parseYearWithList(s string) (string, int, int, int, []int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return s, 0, 0, 0, nil, nil
	}

	// Try to match year pattern at the start
	match := yearPattern.FindStringSubmatch(s)
	if match == nil {
		// No year found
		return s, 0, 0, 0, nil, nil
	}

	// Check if we have comma-separated years
	if strings.Contains(match[1], ",") {
		// Parse comma-separated years
		yearStrs := strings.Split(match[1], ",")
		years := make([]int, len(yearStrs))
		for i, yearStr := range yearStrs {
			year, err := strconv.Atoi(yearStr)
			if err != nil {
				return s, 0, 0, 0, nil, fmt.Errorf("invalid year: %s", yearStr)
			}
			years[i] = year
		}
		// Remove the matched year part from the string
		remaining := strings.TrimSpace(s[len(match[0]):])
		return remaining, 0, 0, 0, years, nil
	}

	// Extract year values
	yearStart, err := strconv.Atoi(match[1])
	if err != nil {
		return s, 0, 0, 0, nil, fmt.Errorf("invalid year: %s", match[1])
	}

	yearEnd := yearStart // Default to same year for single year
	yearInterval := 0

	// Check if we have a plus notation (e.g., "2020+")
	fullMatch := match[0]
	if strings.Contains(fullMatch, "+") {
		// Year plus means "from this year onwards"
		yearEnd = 9999
	} else if match[2] != "" {
		// We have a year range
		yearEnd, err = strconv.Atoi(match[2])
		if err != nil {
			return s, 0, 0, 0, nil, fmt.Errorf("invalid year: %s", match[2])
		}
		// Check for interval notation (e.g., /2 for every other year)
		if len(match) > 3 && match[3] != "" {
			intervalStr := strings.TrimPrefix(match[3], "/")
			yearInterval, err = strconv.Atoi(intervalStr)
			if err != nil {
				return s, 0, 0, 0, nil, fmt.Errorf("invalid year interval: %s", match[3])
			}
		}
	}

	// Remove the matched year part from the string
	remaining := strings.TrimSpace(s[len(match[0]):])
	return remaining, yearStart, yearEnd, yearInterval, nil, nil
}

// expandMonthList expands comma-separated month lists in a rule string
// e.g., "Jun-Aug,Dec Mo 10:00-12:00" -> ["Jun-Aug Mo 10:00-12:00", "Dec Mo 10:00-12:00"]
// Returns a list of expanded rule strings, or the original string if no expansion needed
func expandMonthList(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return []string{s}
	}

	parts := strings.Fields(s)
	if len(parts) == 0 {
		return []string{s}
	}

	// Check if first part contains comma-separated months
	firstPart := strings.ToLower(parts[0])
	if !strings.Contains(firstPart, ",") {
		return []string{s}
	}

	// Check if this looks like a month list (not a weekday list like "Mo,Tu")
	monthParts := strings.Split(firstPart, ",")
	for _, mp := range monthParts {
		mp = strings.TrimSpace(mp)
		// Each part should be a month or month range
		if strings.Contains(mp, "-") {
			// Month range like "jun-aug"
			rangeParts := strings.SplitN(mp, "-", 2)
			_, isMonth1 := monthNames[rangeParts[0]]
			_, isMonth2 := monthNames[rangeParts[1]]
			if !isMonth1 || !isMonth2 {
				return []string{s}
			}
		} else {
			// Single month like "dec"
			_, isMonth := monthNames[mp]
			if !isMonth {
				return []string{s}
			}
		}
	}

	// It's a month list, expand it
	remaining := strings.TrimSpace(s[len(parts[0]):])
	var result []string
	for _, mp := range monthParts {
		mp = strings.TrimSpace(mp)
		// Capitalize first letter for consistency
		if len(mp) >= 3 {
			mp = strings.ToUpper(string(mp[0])) + strings.ToLower(mp[1:])
		}
		if remaining != "" {
			result = append(result, mp+" "+remaining)
		} else {
			result = append(result, mp)
		}
	}
	return result
}

func parseMonthDate(s string) (string, int, int, int, int, int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return s, 0, 0, 0, 0, 0, nil
	}

	parts := strings.Fields(s)
	if len(parts) == 0 {
		return s, 0, 0, 0, 0, 0, nil
	}

	// Check if first part is a month range like "Jan-Mar" or "Jan-Mar:"
	firstPart := strings.ToLower(parts[0])
	// Strip trailing colon if present (used as separator in some patterns)
	firstPartClean := strings.TrimSuffix(firstPart, ":")
	if strings.Contains(firstPartClean, "-") {
		rangeParts := strings.SplitN(firstPartClean, "-", 2)
		month1, isMonth1 := monthNames[rangeParts[0]]
		month2, isMonth2 := monthNames[rangeParts[1]]
		if isMonth1 && isMonth2 {
			// It's a month range like "Jan-Mar" or "Jan-Mar:"
			remaining := strings.TrimSpace(s[len(parts[0]):])
			return remaining, month1, month2, 0, 0, 0, nil
		}
	}

	// Check if first part looks like a month
	month1, isMonth := monthNames[firstPart]
	if !isMonth {
		// Not a month, return unchanged
		return s, 0, 0, 0, 0, 0, nil
	}

	// We have at least one month
	// Now check what follows: could be another month, a day, a range, or nothing

	if len(parts) == 1 {
		// Just "Dec" - single month
		remaining := strings.TrimSpace(s[len(parts[0]):])
		return remaining, month1, month1, 0, 0, 0, nil
	}

	// Check if second part is a day number or a range or another selector
	secondPart := parts[1]

	// Check for month-day range like "Dec 24-Jan 02" or "Jan 01-2024 Jun 30" or "Jan 01-31/8"
	if strings.Contains(secondPart, "-") && !strings.Contains(secondPart, ":") {
		// This could be a day range or month-day range
		rangeParts := strings.SplitN(secondPart, "-", 2)

		// Check if first part of range is a number (day)
		if day1, err := strconv.Atoi(rangeParts[0]); err == nil {
			// It's a day number, check the second part
			// Could be "Dec 24-Jan", "Dec 24-26", "Jan 01-2024", or "Jan 01-31/8"

			// Check if second part starts with a 4-digit year (format: "Jan 01-2024 Jun 30")
			if len(rangeParts[1]) == 4 {
				if _, err := strconv.Atoi(rangeParts[1]); err == nil {
					// This is a year, so we have format like "Jan 01-2024 Jun 30"
					// We need to look ahead to get the month and day
					if len(parts) > 3 {
						month2, isMonth2 := monthNames[strings.ToLower(parts[2])]
						if isMonth2 {
							if day2, err := strconv.Atoi(parts[3]); err == nil {
								// "Jan 01-2024 Jun 30" -> month1=1, day1=1, month2=6, day2=30
								// Skip 4 parts: "Jan", "01-2024", "Jun", "30"
								consumed := len(parts[0]) + 1 + len(parts[1]) + 1 + len(parts[2]) + 1 + len(parts[3])
								remaining := strings.TrimSpace(s[consumed:])
								return remaining, month1, month2, day1, day2, 0, nil
							}
						}
					}
				}
			}

			// Check if second part of range is a month name
			month2, isMonth2 := monthNames[strings.ToLower(rangeParts[1])]
			if isMonth2 {
				// Format: "Dec 24-Jan" followed by day number
				if len(parts) > 2 {
					if day2, err := strconv.Atoi(parts[2]); err == nil {
						// "Dec 24-Jan 02"
						remaining := strings.TrimSpace(s[len(parts[0])+1+len(parts[1])+1+len(parts[2]):])
						return remaining, month1, month2, day1, day2, 0, nil
					}
				}
			}

			// Check if it's a day range with optional interval like "Dec 24-26" or "Jan 01-31/8"
			dayRangePart := rangeParts[1]
			dayInterval := 0

			// Check for interval notation like "31/8"
			if strings.Contains(dayRangePart, "/") {
				intervalParts := strings.SplitN(dayRangePart, "/", 2)
				if interval, err := strconv.Atoi(intervalParts[1]); err == nil {
					dayInterval = interval
					dayRangePart = intervalParts[0]
				}
			}

			if day2, err := strconv.Atoi(dayRangePart); err == nil {
				// Same month, day range (with optional interval)
				remaining := strings.TrimSpace(s[len(parts[0])+1+len(parts[1]):])
				return remaining, month1, month1, day1, day2, dayInterval, nil
			}
		}
	}

	// Check if second part is just a day number
	if day1, err := strconv.Atoi(secondPart); err == nil {
		// Single month-day like "Dec 25"
		remaining := strings.TrimSpace(s[len(parts[0])+1+len(parts[1]):])
		return remaining, month1, month1, day1, day1, 0, nil
	}

	// Second part is not a day, so just return the month
	remaining := strings.TrimSpace(s[len(parts[0]):])
	return remaining, month1, month1, 0, 0, 0, nil
}

func extractComment(s string, oh *OpeningHours) (string, string) {
	// Look for quoted string at the end
	s = strings.TrimSpace(s)
	if !strings.HasSuffix(s, "\"") {
		return s, ""
	}

	// Find the starting quote
	lastQuote := len(s) - 1
	startQuote := -1
	for i := lastQuote - 1; i >= 0; i-- {
		if s[i] == '"' && (i == 0 || s[i-1] != '\\') {
			startQuote = i
			break
		}
	}

	if startQuote == -1 {
		return s, ""
	}

	comment := s[startQuote+1 : lastQuote]

	// Warn if comment is empty
	if comment == "" && oh != nil {
		oh.addWarning("Empty comment")
	}

	remaining := strings.TrimSpace(s[:startQuote])
	return remaining, comment
}

var startsWithTimePattern = regexp.MustCompile(`^(\d{1,2}):(\d{2})`)
var startsWithVariableTimePattern = regexp.MustCompile(`^\(?(sunrise|sunset|dawn|dusk)`)
var startsWithShortTimePattern = regexp.MustCompile(`^\d{1,2}-\d{1,2}$`)

func parseWeekdaysAndTimeWithConstraints(s string) ([]bool, []weekdayConstraint, string, error) {
	weekdays, constraints, timeStr, _, _, err := parseWeekdaysAndTimeWithConstraintsAndHolidays(s)
	return weekdays, constraints, timeStr, err
}

// parseWeekdaysAndTimeWithConstraintsAndHolidays is like parseWeekdaysAndTimeWithConstraints but also returns PH/SH flags
func parseWeekdaysAndTimeWithConstraintsAndHolidays(s string) ([]bool, []weekdayConstraint, string, bool, bool, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil, "", false, false, nil
	}

	// Check if this starts with "PH" (public holiday)
	if strings.HasPrefix(strings.ToUpper(s), "PH") {
		// PH is handled separately, not as a weekday
		return nil, nil, s, false, false, nil
	}

	// Check if this starts with "SH" (school holiday)
	if strings.HasPrefix(strings.ToUpper(s), "SH") {
		// SH is handled separately, not as a weekday
		return nil, nil, s, false, false, nil
	}

	// Check if this starts with time range (no weekday prefix)
	if startsWithTimePattern.MatchString(s) {
		return nil, nil, s, false, false, nil
	}

	// Check if this is a short time format like "10-12" (entire string is this pattern)
	if startsWithShortTimePattern.MatchString(s) {
		return nil, nil, s, false, false, nil
	}

	// Check if this starts with a variable time (sunrise, sunset, dawn, dusk)
	if startsWithVariableTimePattern.MatchString(s) {
		return nil, nil, s, false, false, nil
	}

	// Look for weekday prefix followed by space and time
	// Try to find where weekdays end and time begins
	parts := strings.SplitN(s, " ", 2)
	if len(parts) == 1 {
		// Only weekdays, no time (shouldn't happen in valid input, but handle gracefully)
		weekdays, constraints, hasPH, hasSH, err := parseWeekdaysWithHolidays(parts[0])
		return weekdays, constraints, "", hasPH, hasSH, err
	}

	weekdays, constraints, hasPH, hasSH, err := parseWeekdaysWithHolidays(parts[0])
	if err != nil {
		// Maybe it's all time ranges?
		return nil, nil, s, false, false, nil
	}

	return weekdays, constraints, parts[1], hasPH, hasSH, nil
}

func parseWeekdaysAndTime(s string) ([]bool, string, error) {
	weekdays, _, timeStr, err := parseWeekdaysAndTimeWithConstraints(s)
	return weekdays, timeStr, err
}

func parseWeekdays(s string) ([]bool, []weekdayConstraint, error) {
	weekdays, constraints, _, _, err := parseWeekdaysWithHolidays(s)
	return weekdays, constraints, err
}

// parseWeekdaysWithHolidays parses weekdays and also returns flags for PH and SH if found
func parseWeekdaysWithHolidays(s string) ([]bool, []weekdayConstraint, bool, bool, error) {
	weekdays := make([]bool, 7)
	var constraints []weekdayConstraint
	hasPH := false
	hasSH := false

	// Split by comma for multiple weekday selectors, but not inside brackets
	// e.g., "Mo,Tu" -> ["Mo", "Tu"] but "We[4,5]" should stay as one part
	var parts []string
	var current strings.Builder
	bracketDepth := 0
	for _, ch := range s {
		if ch == '[' {
			bracketDepth++
			current.WriteRune(ch)
		} else if ch == ']' {
			bracketDepth--
			current.WriteRune(ch)
		} else if ch == ',' && bracketDepth == 0 {
			parts = append(parts, strings.TrimSpace(current.String()))
			current.Reset()
		} else {
			current.WriteRune(ch)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, strings.TrimSpace(current.String()))
	}

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check for PH (public holiday) or SH (school holiday)
		upperPart := strings.ToUpper(part)
		if upperPart == "PH" {
			hasPH = true
			continue
		}
		if upperPart == "SH" {
			hasSH = true
			continue
		}

		parsedConstraints, err := parseWeekdaySelectorWithConstraint(part, weekdays)
		if err != nil {
			return nil, nil, false, false, err
		}

		constraints = append(constraints, parsedConstraints...)
	}

	return weekdays, constraints, hasPH, hasSH, nil
}

// Matches weekday with optional constraint: We, We[1], We[1-3], We[1,3], We[1,3,5]
var weekdayConstraintPattern = regexp.MustCompile(`^([A-Za-z]{2,3})(\[([^\]]+)\])?$`)

func parseWeekdaySelectorWithConstraint(s string, weekdays []bool) ([]weekdayConstraint, error) {
	// Try to match the constraint pattern
	matches := weekdayConstraintPattern.FindStringSubmatch(s)

	if matches == nil {
		// No constraint pattern, try to parse as regular weekday
		if err := parseWeekdaySelector(s, weekdays); err != nil {
			return nil, err
		}
		return nil, nil
	}

	// Extract parts
	weekdayName := matches[1]
	hasConstraint := matches[2] != ""
	constraintStr := matches[3]

	// Parse weekday name
	weekday, ok := weekdayNames[strings.ToLower(weekdayName)]
	if !ok {
		return nil, fmt.Errorf("invalid weekday: %s", weekdayName)
	}

	if !hasConstraint {
		// No constraint, just a simple weekday
		weekdays[weekday] = true
		return nil, nil
	}

	// Parse constraint - could be single value, range, or comma-separated list
	var constraints []weekdayConstraint

	// Check if it's comma-separated
	if strings.Contains(constraintStr, ",") {
		// Comma-separated list like [4,5]
		parts := strings.Split(constraintStr, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			nth, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid constraint number: %s", part)
			}
			constraints = append(constraints, weekdayConstraint{
				weekday: weekday,
				nthFrom: nth,
				nthTo:   0,
			})
		}
	} else {
		// Check if it's a range by looking for hyphen not at the start
		// [-1] is a negative number, [1-3] is a range
		trimmed := strings.TrimSpace(constraintStr)
		hyphenIdx := strings.Index(trimmed[1:], "-") // Skip first char to handle negative numbers
		if hyphenIdx >= 0 {
			// It's a range like [1-3] (hyphen found after first character)
			hyphenIdx++ // Adjust for the skipped first character
			nthFrom, err := strconv.Atoi(strings.TrimSpace(trimmed[:hyphenIdx]))
			if err != nil {
				return nil, fmt.Errorf("invalid constraint range start: %s", trimmed[:hyphenIdx])
			}
			nthTo, err := strconv.Atoi(strings.TrimSpace(trimmed[hyphenIdx+1:]))
			if err != nil {
				return nil, fmt.Errorf("invalid constraint range end: %s", trimmed[hyphenIdx+1:])
			}
			constraints = append(constraints, weekdayConstraint{
				weekday: weekday,
				nthFrom: nthFrom,
				nthTo:   nthTo,
			})
		} else {
			// Single value like [4] or [-1]
			nth, err := strconv.Atoi(trimmed)
			if err != nil {
				return nil, fmt.Errorf("invalid constraint number: %s", constraintStr)
			}
			constraints = append(constraints, weekdayConstraint{
				weekday: weekday,
				nthFrom: nth,
				nthTo:   0,
			})
		}
	}

	return constraints, nil
}

func parseWeekdaySelector(s string, weekdays []bool) error {
	// Check for range (e.g., Mo-Fr)
	if strings.Contains(s, "-") {
		parts := strings.SplitN(s, "-", 2)
		startDay, ok := weekdayNames[strings.ToLower(parts[0])]
		if !ok {
			return fmt.Errorf("invalid weekday: %s", parts[0])
		}
		endDay, ok := weekdayNames[strings.ToLower(parts[1])]
		if !ok {
			return fmt.Errorf("invalid weekday: %s", parts[1])
		}

		// Handle wraparound (e.g., Sa-Mo)
		if startDay <= endDay {
			for i := startDay; i <= endDay; i++ {
				weekdays[i] = true
			}
		} else {
			// Wrap around (e.g., Sa-Mo includes Sa, Su, Mo)
			for i := startDay; i < 7; i++ {
				weekdays[i] = true
			}
			for i := 0; i <= endDay; i++ {
				weekdays[i] = true
			}
		}
	} else {
		// Single weekday
		day, ok := weekdayNames[strings.ToLower(s)]
		if !ok {
			return fmt.Errorf("invalid weekday: %s", s)
		}
		weekdays[day] = true
	}

	return nil
}

func parseTimeRanges(s string, oh *OpeningHours) ([]timeRange, error) {
	var ranges []timeRange

	parts := strings.Split(s, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		tr, err := parseTimeRange(part, oh, s)
		if err != nil {
			return nil, err
		}
		ranges = append(ranges, tr)
	}

	// Check for overlapping time ranges (only for fixed time ranges)
	if oh != nil && len(ranges) > 1 {
		for i := 0; i < len(ranges)-1; i++ {
			for j := i + 1; j < len(ranges); j++ {
				// Only check fixed time ranges (not variable times)
				if ranges[i].start >= 0 && ranges[i].end >= 0 &&
				   ranges[j].start >= 0 && ranges[j].end >= 0 {
					// Check if ranges overlap (not just touch)
					// Range i: [start_i, end_i), Range j: [start_j, end_j)
					// They overlap if start_i < end_j AND start_j < end_i
					if ranges[i].start < ranges[j].end && ranges[j].start < ranges[i].end {
						oh.addWarning("Overlapping time ranges detected")
						// Only warn once
						goto done
					}
				}
			}
		}
	}
done:

	return ranges, nil
}

func parseTimeRange(s string, oh *OpeningHours, originalInput string) (timeRange, error) {
	s = normalizeTimeString(s)

	// Check for open-end syntax (e.g., "17:00+")
	if match := openEndPattern.FindStringSubmatch(s); match != nil {
		startHour, _ := strconv.Atoi(match[1])
		startMin, _ := strconv.Atoi(match[2])
		return timeRange{
			start:   startHour*60 + startMin,
			end:     24 * 60, // End of day
			openEnd: true,
		}, nil
	}

	// Check for open-end range syntax (e.g., "14:00-17:00+")
	// This means "open from start time, with uncertain end time (at least until end time)"
	// We treat this as open until end of day since the closing time is uncertain
	if match := openEndRangePattern.FindStringSubmatch(s); match != nil {
		startHour, _ := strconv.Atoi(match[1])
		startMin, _ := strconv.Atoi(match[2])
		// endHour and endMin are the "at least until" times, but we extend to end of day
		return timeRange{
			start:   startHour*60 + startMin,
			end:     24 * 60, // Extend to end of day since close time is uncertain
			openEnd: true,
		}, nil
	}

	// Check for variable time range (e.g., "sunrise-sunset", "(sunrise+01:00)-(sunset-01:00)")
	// We need to be careful splitting on "-" since offsets can contain "-"
	// Strategy: find the main separator between start and end times

	// Look for patterns like "sunrise-sunset", "(sunrise+01:00)-(sunset-01:00)", etc.
	// The main separator is the "-" that's not inside parentheses and not part of an offset
	var startPart, endPart string

	// Check if we have parenthesized variable times
	if strings.HasPrefix(s, "(") {
		// Find the matching closing paren for the first part
		depth := 0
		splitIdx := -1
		for i, ch := range s {
			if ch == '(' {
				depth++
			} else if ch == ')' {
				depth--
				if depth == 0 && i+1 < len(s) && s[i+1] == '-' {
					splitIdx = i + 1
					break
				}
			}
		}
		if splitIdx > 0 {
			startPart = strings.TrimSpace(s[:splitIdx])
			endPart = strings.TrimSpace(s[splitIdx+1:])
		}
	} else {
		// Simple case: split on first "-" that's not part of an offset
		// For patterns like "sunrise-sunset" or "sunset-sunrise"
		parts := strings.Split(s, "-")
		if len(parts) >= 2 {
			// Check if first part is a variable time keyword
			if variableTimePattern.MatchString(parts[0]) {
				startPart = parts[0]
				endPart = strings.Join(parts[1:], "-")
			}
		}
	}

	if startPart != "" && endPart != "" {
		startMatch := variableTimePattern.FindStringSubmatch(startPart)
		endMatch := variableTimePattern.FindStringSubmatch(endPart)

		// If we have at least one variable time, this is a variable time range
		if startMatch != nil || endMatch != nil {
			tr := timeRange{
				start: -1, // Will be resolved at runtime
				end:   -1, // Will be resolved at runtime
			}

			// Parse start time
			if startMatch != nil {
				tr.startVar = strings.ToLower(startMatch[1])
				if startMatch[2] != "" {
					// Parse offset like "+01:00" or "-00:30"
					offset, err := parseTimeOffset(startMatch[2])
					if err != nil {
						return timeRange{}, err
					}
					tr.startOffset = offset
				}
			} else {
				// Start is a fixed time
				fixedTime, err := parseFixedTime(startPart)
				if err != nil {
					return timeRange{}, err
				}
				tr.start = fixedTime
			}

			// Parse end time
			if endMatch != nil {
				tr.endVar = strings.ToLower(endMatch[1])
				if endMatch[2] != "" {
					// Parse offset like "+01:00" or "-00:30"
					offset, err := parseTimeOffset(endMatch[2])
					if err != nil {
						return timeRange{}, err
					}
					tr.endOffset = offset
				}
			} else {
				// End is a fixed time
				fixedTime, err := parseFixedTime(endPart)
				if err != nil {
					return timeRange{}, err
				}
				tr.end = fixedTime
			}

			return tr, nil
		}
	}

	// Try single time point format (e.g., "12:00" represents a 1-minute interval)
	if match := singleTimePattern.FindStringSubmatch(s); match != nil {
		hour, _ := strconv.Atoi(match[1])
		min, _ := strconv.Atoi(match[2])

		// Validate hours and minutes
		if hour > 24 {
			return timeRange{}, fmt.Errorf("invalid time: hours cannot exceed 24")
		}
		if min > 59 {
			return timeRange{}, fmt.Errorf("invalid time: minutes cannot exceed 59")
		}

		startMinutes := hour*60 + min
		endMinutes := startMinutes + 1 // 1-minute interval

		return timeRange{
			start: startMinutes,
			end:   endMinutes,
		}, nil
	}

	// Try standard format
	match := timeRangePattern.FindStringSubmatch(s)
	if match != nil {
		startHour, _ := strconv.Atoi(match[1])
		startMin, _ := strconv.Atoi(match[2])
		endHour, _ := strconv.Atoi(match[3])
		endMin, _ := strconv.Atoi(match[4])

		// Validate hours - max 26 for extended hours notation (26:00 = 02:00 next day)
		if startHour > 26 || endHour > 26 {
			return timeRange{}, fmt.Errorf("invalid time: hours cannot exceed 26")
		}
		// Validate minutes
		if startMin > 59 || endMin > 59 {
			return timeRange{}, fmt.Errorf("invalid time: minutes cannot exceed 59")
		}

		// Check for time interval notation (e.g., 10:00-16:00/01:30)
		interval := 0
		if len(match) > 6 && match[5] != "" && match[6] != "" {
			intervalHour, _ := strconv.Atoi(match[5])
			intervalMin, _ := strconv.Atoi(match[6])
			interval = intervalHour*60 + intervalMin
		}

		return timeRange{
			start:    startHour*60 + startMin,
			end:      endHour*60 + endMin,
			interval: interval,
		}, nil
	}

	return timeRange{}, fmt.Errorf("invalid time range: %s", s)
}

// parseTimeOffset parses a time offset like "+01:00" or "-00:30" and returns minutes
func parseTimeOffset(s string) (int, error) {
	if len(s) < 6 {
		return 0, fmt.Errorf("invalid time offset: %s", s)
	}

	sign := 1
	if s[0] == '-' {
		sign = -1
	}

	hourStr := s[1:3]
	minStr := s[4:6]

	hour, err := strconv.Atoi(hourStr)
	if err != nil {
		return 0, fmt.Errorf("invalid hour in offset: %s", s)
	}

	min, err := strconv.Atoi(minStr)
	if err != nil {
		return 0, fmt.Errorf("invalid minute in offset: %s", s)
	}

	return sign * (hour*60 + min), nil
}

// parseFixedTime parses a fixed time like "06:00" and returns minutes from midnight
func parseFixedTime(s string) (int, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid time format: %s", s)
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid hour: %s", s)
	}

	min, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid minute: %s", s)
	}

	return hour*60 + min, nil
}

// Iterator provides efficient traversal of state changes
type Iterator struct {
	oh      *OpeningHours
	current time.Time
}

// GetIterator creates an iterator starting at the given time
func (oh *OpeningHours) GetIterator(start time.Time) *Iterator {
	return &Iterator{
		oh:      oh,
		current: start,
	}
}

// GetDate returns the current time of the iterator
func (it *Iterator) GetDate() time.Time {
	return it.current
}

// SetDate sets the iterator to a specific time
func (it *Iterator) SetDate(t time.Time) {
	it.current = t
}

// GetState returns the opening state at the current iterator time
func (it *Iterator) GetState() bool {
	return it.oh.GetState(it.current)
}

// GetStateString returns the state as a string ("open", "closed", or "unknown")
func (it *Iterator) GetStateString() string {
	return it.oh.GetStateString(it.current)
}

// GetComment returns any comment associated with the current state
func (it *Iterator) GetComment() string {
	return it.oh.GetComment(it.current)
}

// Advance moves the iterator to the next state change and returns the new time
// Returns zero time if there are no more state changes (e.g., for 24/7 or always closed)
func (it *Iterator) Advance() time.Time {
	nextChange := it.oh.GetNextChange(it.current)

	// If there's a next change, update current time
	if !nextChange.IsZero() {
		it.current = nextChange
	}

	return nextChange
}

// PrettifyValue returns a normalized/canonicalized version of the opening hours string
func (oh *OpeningHours) PrettifyValue() string {
	var parts []string

	// Handle special cases first
	if len(oh.rules) == 1 && oh.rules[0].weekdays == nil && len(oh.rules[0].timeRanges) == 0 {
		if oh.rules[0].state == StateOpen {
			return "24/7"
		} else if oh.rules[0].state == StateClosed {
			return "off"
		}
	}

	// Check if single rule with 00:00-24:00 (equivalent to 24/7)
	if len(oh.rules) == 1 && oh.rules[0].weekdays == nil && len(oh.rules[0].timeRanges) == 1 {
		tr := oh.rules[0].timeRanges[0]
		if tr.start == 0 && tr.end == 1440 && !tr.openEnd && tr.startVar == "" && tr.endVar == "" {
			return "24/7"
		}
	}

	for _, r := range oh.rules {
		part := prettifyRule(r)
		if part != "" {
			parts = append(parts, part)
		}
	}

	return strings.Join(parts, "; ")
}

func prettifyRule(r rule) string {
	var result strings.Builder

	// Add year if specified
	if r.yearStart > 0 {
		if r.yearStart == r.yearEnd {
			result.WriteString(fmt.Sprintf("%d ", r.yearStart))
		} else {
			result.WriteString(fmt.Sprintf("%d-%d ", r.yearStart, r.yearEnd))
		}
	}

	// Add month if specified
	if r.monthStart > 0 {
		result.WriteString(monthName(r.monthStart))
		if r.dayStart > 0 {
			result.WriteString(fmt.Sprintf(" %02d", r.dayStart))
			if r.dayEnd > r.dayStart {
				result.WriteString(fmt.Sprintf("-%02d", r.dayEnd))
			}
		}
		if r.monthEnd != r.monthStart {
			result.WriteString("-")
			result.WriteString(monthName(r.monthEnd))
			if r.dayEnd > 0 {
				result.WriteString(fmt.Sprintf(" %02d", r.dayEnd))
			}
		}
		result.WriteString(" ")
	}

	// Add weekdays
	if r.weekdays != nil {
		result.WriteString(prettifyWeekdays(r.weekdays, r.weekdayConstraints))
	}

	// Add PH/SH
	if r.isPH {
		if result.Len() > 0 {
			result.WriteString(" ")
		}
		result.WriteString("PH")
		if r.phOffset != 0 {
			if r.phOffset > 0 {
				result.WriteString(fmt.Sprintf(" +%d day", r.phOffset))
			} else {
				result.WriteString(fmt.Sprintf(" %d day", r.phOffset))
			}
		}
	}
	if r.isSH {
		if result.Len() > 0 {
			result.WriteString(" ")
		}
		result.WriteString("SH")
	}

	// Add time ranges
	if len(r.timeRanges) > 0 {
		if result.Len() > 0 {
			result.WriteString(" ")
		}
		timeStrs := make([]string, len(r.timeRanges))
		for i, tr := range r.timeRanges {
			timeStrs[i] = prettifyTimeRange(tr)
		}
		result.WriteString(strings.Join(timeStrs, ","))
	}

	// Add state
	switch r.state {
	case StateClosed:
		if result.Len() > 0 {
			result.WriteString(" ")
		}
		result.WriteString("off")
	case StateUnknown:
		if result.Len() > 0 {
			result.WriteString(" ")
		}
		result.WriteString("unknown")
	}

	// Add comment
	if r.comment != "" {
		result.WriteString(fmt.Sprintf(" \"%s\"", r.comment))
	}

	return strings.TrimSpace(result.String())
}

func prettifyWeekdays(weekdays []bool, constraints []weekdayConstraint) string {
	// Convert bool array to weekday range string
	names := []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	var parts []string

	// Handle constraints first
	if len(constraints) > 0 {
		for _, c := range constraints {
			name := names[c.weekday]
			if c.nthTo > 0 {
				parts = append(parts, fmt.Sprintf("%s[%d-%d]", name, c.nthFrom, c.nthTo))
			} else {
				parts = append(parts, fmt.Sprintf("%s[%d]", name, c.nthFrom))
			}
		}
		return strings.Join(parts, ",")
	}

	// Find ranges in weekdays, starting from Monday (index 1) instead of Sunday (index 0)
	// This gives more natural output like "Mo-Fr" instead of "Su,Mo-Fr"
	// However, we need to handle wraparound cases like "Sa-Su" correctly

	// First, try starting from Monday (1) and wrap around if needed
	startIdx := 1 // Monday
	for j := 0; j < 7; j++ {
		i := (startIdx + j) % 7
		if weekdays[i] {
			start := i
			// Find the end of this consecutive range
			count := 0
			for count < 7 && weekdays[(start+count)%7] {
				count++
			}
			end := (start + count - 1) % 7

			if count == 1 {
				// Single day
				parts = append(parts, names[start])
			} else if count == 3 {
				// Exactly 3 consecutive days: list individually
				for k := 0; k < count; k++ {
					parts = append(parts, names[(start+k)%7])
				}
			} else {
				// 2 days or 4+ days: use range
				parts = append(parts, fmt.Sprintf("%s-%s", names[start], names[end]))
			}

			// Skip the days we've already processed
			j += count - 1
		}
	}

	return strings.Join(parts, ",")
}

func prettifyTimeRange(tr timeRange) string {
	startH, startM := tr.start/60, tr.start%60
	endH, endM := tr.end/60, tr.end%60

	if tr.openEnd {
		return fmt.Sprintf("%02d:%02d+", startH, startM)
	}
	return fmt.Sprintf("%02d:%02d-%02d:%02d", startH, startM, endH, endM)
}

func monthName(m int) string {
	names := []string{"", "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	if m >= 1 && m <= 12 {
		return names[m]
	}
	return ""
}
