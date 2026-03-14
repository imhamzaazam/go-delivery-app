package utils

import (
	"fmt"
	"time"
)

const RestaurantTimezone = "Asia/Karachi"

func ParseBranchClock(value string) (int16, error) {
	parsed, err := time.Parse("15:04", value)
	if err != nil {
		return 0, fmt.Errorf("invalid branch time %q: %w", value, err)
	}

	minutes := (parsed.Hour() * 60) + parsed.Minute()
	return int16(minutes), nil
}

func FormatBranchClock(minutes int16) string {
	totalMinutes := int(minutes)
	if totalMinutes < 0 {
		totalMinutes = 0
	}
	totalMinutes = totalMinutes % (24 * 60)

	hours := totalMinutes / 60
	mins := totalMinutes % 60
	return fmt.Sprintf("%02d:%02d", hours, mins)
}

func BranchLocation() *time.Location {
	location, err := time.LoadLocation(RestaurantTimezone)
	if err != nil {
		return time.FixedZone(RestaurantTimezone, 5*60*60)
	}

	return location
}

func IsBranchOpenAt(openingMinutes int16, closingMinutes int16, now time.Time) bool {
	if openingMinutes == closingMinutes {
		return true
	}

	localizedNow := now.In(BranchLocation())
	currentMinutes := int16((localizedNow.Hour() * 60) + localizedNow.Minute())

	if openingMinutes < closingMinutes {
		return currentMinutes >= openingMinutes && currentMinutes < closingMinutes
	}

	return currentMinutes >= openingMinutes || currentMinutes < closingMinutes
}
