package service

import (
	"strconv"
	"time"
)

func parseDurationOrHours(raw string, fallback time.Duration) time.Duration {
	if raw == "" {
		return fallback
	}

	if durationValue, err := time.ParseDuration(raw); err == nil {
		return durationValue
	}

	// Backward compatibility: legacy configuration stores integer hours.
	if hours, err := strconv.Atoi(raw); err == nil {
		return time.Duration(hours) * time.Hour
	}

	return fallback
}

func clampExpiry(expireAt time.Time, maxExpiresAt *time.Time) time.Time {
	if maxExpiresAt != nil && expireAt.After(*maxExpiresAt) {
		return *maxExpiresAt
	}
	return expireAt
}
