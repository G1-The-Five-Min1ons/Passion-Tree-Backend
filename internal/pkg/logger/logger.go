package logger

import (
	"log/slog"
	"os"
)

func SetupLogger(isDev bool) *slog.Logger {
	opts := &slog.HandlerOptions{
		AddSource: true, 
		
		Level: func() slog.Level {
			if isDev {
				return slog.LevelDebug
			}
			return slog.LevelInfo
		}(),
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)

	return slog.New(handler).With(
		slog.String("service", "passion-tree-backend"),
		slog.String("env", func() string {
			if isDev {
				return "development"
			}
			return "production"
		}()),
	)
}