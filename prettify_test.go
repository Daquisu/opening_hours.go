package openinghours

import (
	"testing"
)

func TestPrettify_TimeFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "zero-pad single digit hours",
			input:    "9:00-17:00",
			expected: "09:00-17:00",
		},
		{
			name:     "normalize spaces around time range",
			input:    "09:00 - 17:00",
			expected: "09:00-17:00",
		},
		{
			name:     "zero-pad both start and end times",
			input:    "8:30-9:45",
			expected: "08:30-09:45",
		},
		{
			name:     "already properly formatted",
			input:    "09:00-17:00",
			expected: "09:00-17:00",
		},
		{
			name:     "multiple spaces around dash",
			input:    "09:00  -  17:00",
			expected: "09:00-17:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oh, err := New(tt.input)
			if err != nil {
				t.Fatalf("failed to parse input %q: %v", tt.input, err)
			}
			result := oh.PrettifyValue()
			if result != tt.expected {
				t.Errorf("PrettifyValue() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestPrettify_WeekdayNames(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Mon to Mo",
			input:    "Mon-Fri 09:00-17:00",
			expected: "Mo-Fr 09:00-17:00",
		},
		{
			name:     "Monday to Mo (full name)",
			input:    "Monday-Friday 09:00-17:00",
			expected: "Mo-Fr 09:00-17:00",
		},
		{
			name:     "single weekday Mon to Mo",
			input:    "Mon 09:00-17:00",
			expected: "Mo 09:00-17:00",
		},
		{
			name:     "Saturday to Sa",
			input:    "Saturday 10:00-14:00",
			expected: "Sa 10:00-14:00",
		},
		{
			name:     "Sunday to Su",
			input:    "Sunday 10:00-14:00",
			expected: "Su 10:00-14:00",
		},
		{
			name:     "all weekdays with full names",
			input:    "Monday,Tuesday,Wednesday 09:00-17:00",
			expected: "Mo,Tu,We 09:00-17:00",
		},
		{
			name:     "already standardized weekdays",
			input:    "Mo-Fr 09:00-17:00",
			expected: "Mo-Fr 09:00-17:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oh, err := New(tt.input)
			if err != nil {
				t.Fatalf("failed to parse input %q: %v", tt.input, err)
			}
			result := oh.PrettifyValue()
			if result != tt.expected {
				t.Errorf("PrettifyValue() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestPrettify_MultipleRules(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "add space after semicolon",
			input:    "Mo 09:00-17:00;Tu 09:00-17:00",
			expected: "Mo 09:00-17:00; Tu 09:00-17:00",
		},
		{
			name:     "normalize multiple rules",
			input:    "Mo 9:00-17:00;Tu 10:00-18:00",
			expected: "Mo 09:00-17:00; Tu 10:00-18:00",
		},
		{
			name:     "preserve existing space after semicolon",
			input:    "Mo 09:00-17:00; Tu 09:00-17:00",
			expected: "Mo 09:00-17:00; Tu 09:00-17:00",
		},
		{
			name:     "multiple rules with three entries",
			input:    "Mo 09:00-17:00;Tu 10:00-18:00;We 09:00-17:00",
			expected: "Mo 09:00-17:00; Tu 10:00-18:00; We 09:00-17:00",
		},
		{
			name:     "remove extra spaces before semicolon",
			input:    "Mo 09:00-17:00 ; Tu 09:00-17:00",
			expected: "Mo 09:00-17:00; Tu 09:00-17:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oh, err := New(tt.input)
			if err != nil {
				t.Fatalf("failed to parse input %q: %v", tt.input, err)
			}
			result := oh.PrettifyValue()
			if result != tt.expected {
				t.Errorf("PrettifyValue() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestPrettify_24_7(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "24/7 unchanged",
			input:    "24/7",
			expected: "24/7",
		},
		{
			name:     "simplify 00:00-24:00 to 24/7",
			input:    "00:00-24:00",
			expected: "24/7",
		},
		{
			name:     "24/7 case variations",
			input:    "24/7",
			expected: "24/7",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oh, err := New(tt.input)
			if err != nil {
				t.Fatalf("failed to parse input %q: %v", tt.input, err)
			}
			result := oh.PrettifyValue()
			if result != tt.expected {
				t.Errorf("PrettifyValue() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestPrettify_Off(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "closed to off",
			input:    "closed",
			expected: "off",
		},
		{
			name:     "OFF to off (lowercase)",
			input:    "OFF",
			expected: "off",
		},
		{
			name:     "already off",
			input:    "off",
			expected: "off",
		},
		{
			name:     "Off to off (mixed case)",
			input:    "Off",
			expected: "off",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oh, err := New(tt.input)
			if err != nil {
				t.Fatalf("failed to parse input %q: %v", tt.input, err)
			}
			result := oh.PrettifyValue()
			if result != tt.expected {
				t.Errorf("PrettifyValue() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestPrettify_Complex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "complex with multiple time ranges and rules",
			input:    "Mo-Fr 9:00-12:00,14:00-18:00;Sa 10:00-14:00",
			expected: "Mo-Fr 09:00-12:00,14:00-18:00; Sa 10:00-14:00",
		},
		{
			name:     "complex with full weekday names",
			input:    "Monday-Friday 9:00-12:00,14:00-18:00;Saturday 10:00-14:00",
			expected: "Mo-Fr 09:00-12:00,14:00-18:00; Sa 10:00-14:00",
		},
		{
			name:     "complex with spaces in time ranges",
			input:    "Mo-Fr 9:00 - 12:00 , 14:00 - 18:00 ; Sa 10:00 - 14:00",
			expected: "Mo-Fr 09:00-12:00,14:00-18:00; Sa 10:00-14:00",
		},
		{
			name:     "complex with multiple weekday groups",
			input:    "Mo,We,Fr 9:00-17:00;Tu,Th 10:00-18:00",
			expected: "Mo,We,Fr 09:00-17:00; Tu,Th 10:00-18:00",
		},
		{
			name:     "complex with off days",
			input:    "Mo-Fr 09:00-17:00;Sa-Su off",
			expected: "Mo-Fr 09:00-17:00; Sa-Su off",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oh, err := New(tt.input)
			if err != nil {
				t.Fatalf("failed to parse input %q: %v", tt.input, err)
			}
			result := oh.PrettifyValue()
			if result != tt.expected {
				t.Errorf("PrettifyValue() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestPrettify_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty weekdays with time range",
			input:    "09:00-17:00",
			expected: "09:00-17:00",
		},
		{
			name:     "midnight crossing",
			input:    "Mo 22:00-02:00",
			expected: "Mo 22:00-02:00",
		},
		{
			name:     "open end time",
			input:    "Mo 09:00+",
			expected: "Mo 09:00+",
		},
		{
			name:     "single digit hour with open end",
			input:    "Mo 9:00+",
			expected: "Mo 09:00+",
		},
		{
			name:     "normalize spaces with comma-separated times",
			input:    "Mo 09:00-12:00 , 14:00-18:00",
			expected: "Mo 09:00-12:00,14:00-18:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oh, err := New(tt.input)
			if err != nil {
				t.Fatalf("failed to parse input %q: %v", tt.input, err)
			}
			result := oh.PrettifyValue()
			if result != tt.expected {
				t.Errorf("PrettifyValue() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestPrettify_PreserveComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "preserve comment with quoted string",
			input:    "Mo-Fr 09:00-17:00 \"lunch break\"",
			expected: "Mo-Fr 09:00-17:00 \"lunch break\"",
		},
		{
			name:     "normalize time with comment",
			input:    "Mo-Fr 9:00-17:00 \"office hours\"",
			expected: "Mo-Fr 09:00-17:00 \"office hours\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oh, err := New(tt.input)
			if err != nil {
				t.Fatalf("failed to parse input %q: %v", tt.input, err)
			}
			result := oh.PrettifyValue()
			if result != tt.expected {
				t.Errorf("PrettifyValue() = %q, want %q", result, tt.expected)
			}
		})
	}
}
