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
) string {
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
