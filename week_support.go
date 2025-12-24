package openinghours

// This file contains week number support implementation for opening_hours.go
// To integrate, add these changes:

// 1. Add to rule struct (around line 31):
//    weekStart    int  // 0=not set, 1-53 for ISO week number
//    weekEnd      int  // 0=not set, 1-53 for ISO week number
//    weekInterval int  // 0=not set, interval for week ranges (e.g., /2 for every other week)

// 2. Add week number parsing in parseRule() after month/date parsing (around line 505):
//    s, weekStart, weekEnd, weekInterval, err := parseWeekNumber(s)
//    if err != nil {
//        return r, err
//    }
//    r.weekStart = weekStart
//    r.weekEnd = weekEnd
//    r.weekInterval = weekInterval

// 3. Add to matches() function after month/day checking (around line 429):
//    // Check week number constraints
//    if r.weekStart > 0 {
//        _, week := t.ISOWeek()
//        inRange := false
//        if r.weekStart == r.weekEnd {
//            if week == r.weekStart {
//                inRange = true
//            }
//        } else {
//            if week >= r.weekStart && week <= r.weekEnd {
//                if r.weekInterval > 0 {
//                    weekOffset := week - r.weekStart
//                    if weekOffset%r.weekInterval == 0 {
//                        inRange = true
//                    }
//                } else {
//                    inRange = true
//                }
//            }
//        }
//        if !inRange {
//            return false
//        }
//    }

import (
	"fmt"
	"strconv"
	"strings"
)

// parseWeekNumbers extracts week number information from the start of the string
// Returns: remaining string, week constraints slice, error
// Supports comma-separated week specifications like "week 01,10,20" or "week 52-53,01-02"
func parseWeekNumbers(s string) (string, []weekConstraint, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return s, nil, nil
	}

	// Check if starts with "week"
	if !strings.HasPrefix(strings.ToLower(s), "week ") {
		return s, nil, nil
	}

	// Remove "week " prefix
	s = strings.TrimSpace(s[5:])
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return s, nil, nil
	}

	// Parse the week number/range(s)
	weekPart := parts[0]

	var constraints []weekConstraint

	// Split by comma for multiple week specifications
	weekSpecs := strings.Split(weekPart, ",")

	for _, spec := range weekSpecs {
		spec = strings.TrimSpace(spec)
		if spec == "" {
			continue
		}

		// Check for interval (e.g., "01-53/2")
		var weekInterval int
		if strings.Contains(spec, "/") {
			intervalParts := strings.SplitN(spec, "/", 2)
			spec = intervalParts[0]
			interval, err := strconv.Atoi(intervalParts[1])
			if err != nil {
				return s, nil, fmt.Errorf("invalid week interval: %s", intervalParts[1])
			}
			weekInterval = interval
		}

		// Check for range (e.g., "01-10")
		if strings.Contains(spec, "-") {
			rangeParts := strings.SplitN(spec, "-", 2)
			weekStart, err := strconv.Atoi(rangeParts[0])
			if err != nil {
				return s, nil, fmt.Errorf("invalid week number: %s", rangeParts[0])
			}
			weekEnd, err := strconv.Atoi(rangeParts[1])
			if err != nil {
				return s, nil, fmt.Errorf("invalid week number: %s", rangeParts[1])
			}

			// Validate week numbers
			if weekStart < 1 || weekStart > 53 || weekEnd < 1 || weekEnd > 53 {
				return s, nil, fmt.Errorf("week numbers must be between 1 and 53")
			}

			constraints = append(constraints, weekConstraint{
				weekStart:    weekStart,
				weekEnd:      weekEnd,
				weekInterval: weekInterval,
			})
		} else {
			// Single week number
			weekNum, err := strconv.Atoi(spec)
			if err != nil {
				return s, nil, fmt.Errorf("invalid week number: %s", spec)
			}

			if weekNum < 1 || weekNum > 53 {
				return s, nil, fmt.Errorf("week number must be between 1 and 53")
			}

			constraints = append(constraints, weekConstraint{
				weekStart:    weekNum,
				weekEnd:      weekNum,
				weekInterval: 0,
			})
		}
	}

	// Remove the week part from the string
	remaining := strings.TrimSpace(s[len(parts[0]):])
	return remaining, constraints, nil
}

// parseWeekNumber extracts week number information from the start of the string
// Returns: remaining string, weekStart, weekEnd, weekInterval, error
// Returns 0 for week values if not specified
// Deprecated: Use parseWeekNumbers instead for full support of comma-separated weeks
func parseWeekNumber(s string) (string, int, int, int, error) {
	remaining, constraints, err := parseWeekNumbers(s)
	if err != nil {
		return s, 0, 0, 0, err
	}

	if len(constraints) == 0 {
		return remaining, 0, 0, 0, nil
	}

	// Return only the first constraint for backwards compatibility
	c := constraints[0]
	return remaining, c.weekStart, c.weekEnd, c.weekInterval, nil
}
