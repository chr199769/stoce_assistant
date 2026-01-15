package llm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetMarketPhase(t *testing.T) {
	tests := []struct {
		name     string
		timeStr  string // Format: "15:04"
		expected MarketPhase
		weekend  bool
	}{
		{
			name:     "Pre-market (08:30)",
			timeStr:  "08:30",
			expected: PhasePreMarket,
		},
		{
			name:     "Pre-market boundary (09:24)",
			timeStr:  "09:24",
			expected: PhasePreMarket,
		},
		{
			name:     "Intraday start (09:25)",
			timeStr:  "09:25",
			expected: PhaseIntraday,
		},
		{
			name:     "Intraday (10:00)",
			timeStr:  "10:00",
			expected: PhaseIntraday,
		},
		{
			name:     "Intraday end (15:00)",
			timeStr:  "15:00",
			expected: PhaseIntraday,
		},
		{
			name:     "Post-market (15:01)",
			timeStr:  "15:01",
			expected: PhasePostMarket,
		},
		{
			name:     "Post-market night (20:00)",
			timeStr:  "20:00",
			expected: PhasePostMarket,
		},
		{
			name:     "Early morning (06:00)",
			timeStr:  "06:00",
			expected: PhasePostMarket,
		},
		{
			name:     "Weekend (Saturday 10:00)",
			timeStr:  "10:00",
			weekend:  true,
			expected: PhasePostMarket,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Construct a base time (e.g., a Monday)
			baseDate := time.Date(2023, 5, 22, 0, 0, 0, 0, time.Local) // Monday
			if tt.weekend {
				baseDate = time.Date(2023, 5, 20, 0, 0, 0, 0, time.Local) // Saturday
			}

			parsedTime, _ := time.Parse("15:04", tt.timeStr)
			testTime := time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(),
				parsedTime.Hour(), parsedTime.Minute(), 0, 0, time.Local)

			phase := GetMarketPhase(testTime)
			assert.Equal(t, tt.expected, phase)
		})
	}
}
