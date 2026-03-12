package service

import (
	"strings"
	"time"
)

// Tree status constants.
const (
	statusGrowing = "growing"
	statusFading  = "fading"
	statusDying   = "dying"
	statusDied    = "died"
)

// computeTreeStatus returns the live status of a tree based on how long ago
// the last reflection was submitted, the tree's difficulty level, and whether
// the tree is currently paused.
//
// Decay schedule:
//
//	easy   => fading after 30 d, dying after 60 d, died after 90 d
//	medium => fading after  7 d, dying after 14 d, died after 21 d
//	hard   => fading after  1 d, dying after  2 d, died after  3 d
//
// When is_pause = true, time is frozen at the moment pause started (pausedAt),
// so the tree does not decay further until it is unpaused.
// On unpause the repository shifts last_reflect_at forward by the pause
// duration, so afterwards this function can always use time.Now() safely.
func computeTreeStatus(difficulties string, lastReflectAt *time.Time, isPause bool, pausedAt *time.Time) string {
	if lastReflectAt == nil {
		return statusGrowing
	}

	// While paused: freeze elapsed time at the moment pause started.
	reference := time.Now()
	if isPause && pausedAt != nil {
		reference = *pausedAt
	}

	elapsed := reference.Sub(*lastReflectAt)
	if elapsed < 0 {
		elapsed = 0
	}

	var fading, dying, died time.Duration
	switch strings.ToLower(difficulties) {
	case "easy":
		fading = 30 * 24 * time.Hour
		dying = 60 * 24 * time.Hour
		died = 90 * 24 * time.Hour
	case "medium":
		fading = 7 * 24 * time.Hour
		dying = 14 * 24 * time.Hour
		died = 21 * 24 * time.Hour
	case "hard":
		fading = 24 * time.Hour
		dying = 48 * time.Hour
		died = 72 * time.Hour
	default:
		return statusGrowing
	}

	switch {
	case elapsed >= died:
		return statusDied
	case elapsed >= dying:
		return statusDying
	case elapsed >= fading:
		return statusFading
	default:
		return statusGrowing
	}
}
