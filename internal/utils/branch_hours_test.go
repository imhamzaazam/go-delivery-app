package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBranchClock(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int16
		wantErr bool
	}{
		{"09:00", "09:00", 9*60 + 0, false},
		{"22:30", "22:30", 22*60 + 30, false},
		{"00:00", "00:00", 0, false},
		{"23:59", "23:59", 23*60 + 59, false},
		{"invalid hour", "25:00", 0, true},
		{"bad format", "not-a-time", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBranchClock(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatBranchClock(t *testing.T) {
	tests := []struct {
		name   string
		minutes int16
		want   string
	}{
		{"09:00", 9*60 + 0, "09:00"},
		{"22:30", 22*60 + 30, "22:30"},
		{"00:00", 0, "00:00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatBranchClock(tt.minutes)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsBranchOpenAt(t *testing.T) {
	// Branch hours: 09:00 - 22:30 (Karachi timezone)
	openingMinutes := int16(9*60 + 0)   // 09:00
	closingMinutes := int16(22*60 + 30) // 22:30

	karachi := BranchLocation()

	tests := []struct {
		name     string
		now      time.Time // in Karachi
		expected bool
	}{
		{
			name:     "during business hours - midday",
			now:      time.Date(2026, 3, 14, 14, 0, 0, 0, karachi),
			expected: true,
		},
		{
			name:     "at opening time",
			now:      time.Date(2026, 3, 14, 9, 0, 0, 0, karachi),
			expected: true,
		},
		{
			name:     "just before closing",
			now:      time.Date(2026, 3, 14, 22, 29, 0, 0, karachi),
			expected: true,
		},
		{
			name:     "at closing time - closed",
			now:      time.Date(2026, 3, 14, 22, 30, 0, 0, karachi),
			expected: false,
		},
		{
			name:     "after closing - 22:59 Karachi (matches 20:59 UTC+3 scenario)",
			now:      time.Date(2026, 3, 14, 22, 59, 0, 0, karachi),
			expected: false,
		},
		{
			name:     "before opening",
			now:      time.Date(2026, 3, 14, 8, 0, 0, 0, karachi),
			expected: false,
		},
		{
			name:     "late night",
			now:      time.Date(2026, 3, 14, 23, 30, 0, 0, karachi),
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsBranchOpenAt(openingMinutes, closingMinutes, tt.now)
			assert.Equal(t, tt.expected, got, "IsBranchOpenAt(09:00, 22:30, %s)", tt.now.Format("15:04 MST"))
		})
	}
}

func TestIsBranchOpenAt_CrossesMidnight(t *testing.T) {
	// Branch open 22:00 - 02:00 (overnight)
	openingMinutes := int16(22 * 60)
	closingMinutes := int16(2 * 60)
	karachi := BranchLocation()

	tests := []struct {
		name     string
		now      time.Time
		expected bool
	}{
		{"23:00 - open", time.Date(2026, 3, 14, 23, 0, 0, 0, karachi), true},
		{"01:00 - open", time.Date(2026, 3, 15, 1, 0, 0, 0, karachi), true},
		{"03:00 - closed", time.Date(2026, 3, 15, 3, 0, 0, 0, karachi), false},
		{"21:00 - closed", time.Date(2026, 3, 14, 21, 0, 0, 0, karachi), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsBranchOpenAt(openingMinutes, closingMinutes, tt.now)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestIsBranchOpenAt_24Hours(t *testing.T) {
	// Same open/close = always open
	minutes := int16(9 * 60)
	karachi := BranchLocation()
	now := time.Date(2026, 3, 14, 15, 0, 0, 0, karachi)
	assert.True(t, IsBranchOpenAt(minutes, minutes, now))
}
