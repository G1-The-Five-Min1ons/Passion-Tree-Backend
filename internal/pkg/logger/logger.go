package logger

import (
	"log/slog"
	"os"
)

func SetupLogger(isDev bool) *slog.Logger {
	// detail Log (Level)
	logLevel := slog.LevelInfo
	if isDev {
		logLevel = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		AddSource: true, 
		Level:     logLevel,
	}

	var handler slog.Handler
	if isDev {
		// on Development: use TextHandler
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		// on Production (Azure): use JSONHandler
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	logger := slog.New(handler).With(
		slog.String("service", "passion-tree-backend"),
		slog.String("env", func() string {
			if isDev {
				return "development"
			}
			return "production"
		}()),
	)

	slog.SetDefault(logger)

	return logger
}