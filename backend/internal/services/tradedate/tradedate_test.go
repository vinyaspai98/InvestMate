package tradedate

import (
	"testing"
	"time"
)

func TestIsTradingDay(t *testing.T) {
	tests := []struct {
		name     string
		date     time.Time
		expected bool
	}{
		{
			name:     "Normal Wednesday trading day",
			date:     time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "Saturday is not a trading day",
			date:     time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "Sunday is not a trading day",
			date:     time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "Mahatma Gandhi Jayanti (2026-10-02) is a trade holiday",
			date:     time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "Christmas (2026-12-25) is a trade holiday",
			date:     time.Date(2026, 12, 25, 0, 0, 0, 0, time.UTC),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsTradingDay(tt.date)
			if got != tt.expected {
				t.Errorf("IsTradingDay(%s) = %v; want %v", tt.date.Format("2006-01-02"), got, tt.expected)
			}
		})
	}
}

func TestPreviousWorkingDay(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string // YYYY-MM-DD
	}{
		{
			name:     "Tuesday mail notification -> Monday trade date",
			input:    time.Date(2026, 10, 6, 10, 0, 0, 0, time.UTC), // Tuesday
			expected: "2026-10-05",                                   // Monday
		},
		{
			name:     "Monday mail notification -> Previous working day skipping weekend and holiday",
			input:    time.Date(2026, 10, 5, 9, 30, 0, 0, time.UTC), // Monday Oct 5
			// Oct 4 is Sun, Oct 3 is Sat, Oct 2 is Gandhi Jayanti (holiday), Oct 1 is Thursday (trading day)
			expected: "2026-10-01",
		},
		{
			name:     "Sunday mail notification -> Previous Friday",
			input:    time.Date(2026, 9, 20, 15, 0, 0, 0, time.UTC), // Sunday Sept 20
			expected: "2026-09-18",                                   // Friday Sept 18
		},
		{
			name:     "Saturday mail notification -> Previous Friday",
			input:    time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC), // Saturday Sept 19
			expected: "2026-09-18",                                   // Friday Sept 18
		},
		{
			name:     "Mail on Jan 27, 2026 (day after Republic Day) -> Jan 23 (Friday)",
			// Jan 26 is Republic Day (Mon), Jan 25 (Sun), Jan 24 (Sat), Jan 23 (Fri)
			input:    time.Date(2026, 1, 27, 10, 0, 0, 0, time.UTC),
			expected: "2026-01-23",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PreviousWorkingDay(tt.input)
			got := result.Format("2006-01-02")
			if got != tt.expected {
				t.Errorf("PreviousWorkingDay(%s) = %s; want %s", tt.input.Format("2006-01-02"), got, tt.expected)
			}
		})
	}
}
