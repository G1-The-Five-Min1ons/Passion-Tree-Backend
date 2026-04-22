package service

import (
	"context"
	"strings"
	"time"
)

func (s *serviceImpl) syncTreeStatus(
	ctx context.Context,
	treeID string,
	currentStatus string,
	difficulties string,
	lastReflectAt *time.Time,
	isPause bool,
	pausedAt *time.Time,
	isReflectionClosed bool,
) string {
	if isReflectionClosed {
		return normalizeTreeStatus(currentStatus)
	}

	if !isPause {
		activated, activatedPausedAt, err := s.refRepo.TryActivateScheduledPause(ctx, treeID)
		if err != nil {
			s.logger.WarnContext(ctx, "failed to activate scheduled pause",
				"tree_id", treeID,
				"error", err,
			)
		} else if activated {
			isPause = true
			pausedAt = activatedPausedAt
		}
	}

	if isPause && pausedAt != nil && !time.Now().Before(*pausedAt) {
		if err := s.refRepo.UnpauseTree(ctx, treeID, ""); err != nil {
			s.logger.WarnContext(ctx, "failed to auto-unpause tree",
				"tree_id", treeID,
				"paused_at", pausedAt,
				"error", err,
			)
		} else {
			isPause = false
			pausedAt = nil
		}
	}

	computedStatus := normalizeTreeStatus(computeTreeStatus(difficulties, lastReflectAt, isPause, pausedAt))
	storedStatus := strings.TrimSpace(strings.ToLower(currentStatus))

	if treeID != "" && storedStatus != computedStatus {
		if err := s.refRepo.UpdateTreeStatus(ctx, treeID, computedStatus); err != nil {
			s.logger.WarnContext(ctx, "failed to persist computed tree status",
				"tree_id", treeID,
				"stored_status", currentStatus,
				"computed_status", computedStatus,
				"error", err,
			)
		}
	}

	return computedStatus
}
