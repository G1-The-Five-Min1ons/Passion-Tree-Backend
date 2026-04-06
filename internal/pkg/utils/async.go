package utils

import (
	"context"
	"log/slog"
	"time"
)

// SafeGo รับหน้าที่รัน Background Task อย่างปลอดภัย พร้อมระบบ Timeout ตัดจบถ้างานค้าง
func SafeGo(ctx context.Context, logger *slog.Logger, taskName string, timeout time.Duration, task func(ctx context.Context) error) {
	detachedCtx := context.WithoutCancel(ctx)

	go func() {
		timeoutCtx, cancel := context.WithTimeout(detachedCtx, timeout)
		defer cancel()

		defer func() {
			if r := recover(); r != nil {
				logger.ErrorContext(timeoutCtx, "panic recovered in background task", "task", taskName, "panic_info", r)
			}
		}()

		if err := task(timeoutCtx); err != nil {
			logger.ErrorContext(timeoutCtx, "background task failed", "task", taskName, "error", err)
		} else {
			logger.InfoContext(timeoutCtx, "background task completed successfully", "task", taskName)
		}
	}()
}