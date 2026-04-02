package utils

import (
	"context"
	"log/slog"
)

// SafeGo รับหน้าที่รัน Background Task อย่างปลอดภัย
func SafeGo(ctx context.Context, logger *slog.Logger, taskName string, task func(ctx context.Context) error) {
	detachedCtx := context.WithoutCancel(ctx)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.ErrorContext(detachedCtx, "panic recovered in background task", "task", taskName, "panic_info", r)
			}
		}()

		if err := task(detachedCtx); err != nil {
			logger.ErrorContext(detachedCtx, "background task failed", "task", taskName, "error", err)
		}
	}()
}
