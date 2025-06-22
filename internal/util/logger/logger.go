package logger

import (
	"io"
	"log/slog"
	"os"

	"blog/internal/config"
	"blog/internal/util/logger/handler/prettylog"
)

type slogWrapper struct {
	*slog.Logger
}

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	With(args ...any) slogWrapper
	// Log(ctx context.Context, level slog.Level, msg string, args ...any)
}

func NewLogger(cfg *config.Config) Logger {
	var handler slog.Handler
	var level slog.Leveler = slog.LevelDebug
	var addSource bool = !cfg.Logger.DisableSrc
	var output io.Writer = os.Stdout

	// local | dev | prod
	switch cfg.Logger.Mode {
	case "local":
		level = slog.LevelDebug
		handler = prettylog.NewPrettyHandler(output, &slog.HandlerOptions{
			Level:     level,
			AddSource: addSource,
		})

	case "dev":
		level = slog.LevelInfo
		handler = prettylog.NewPrettyHandler(output, &slog.HandlerOptions{
			Level:     level,
			AddSource: addSource,
		})

	case "prod":
		level = slog.LevelError
		handler = slog.NewJSONHandler(output, &slog.HandlerOptions{
			Level:     level,
			AddSource: addSource,
		})

	default:
		level = slog.LevelInfo
		handler = prettylog.NewPrettyHandler(output, &slog.HandlerOptions{Level: level})
		slog.New(handler).Warn("unknown logger mode specified", "mode", cfg.Logger.Mode)
	}

	/*
	* slogLogger := slog.New(handler).With(
	*	slog.String("service", cfg.App.Name),
	*	slog.String("app_version", cfg.App.Version),
	* )
	 */

	slogLogger := slog.New(handler)
	slogLogger.Info("logger init",
		slog.String("logger-mode", cfg.Logger.Mode),
	)

	return &slogWrapper{slogLogger}
}

func (s *slogWrapper) With(args ...any) slogWrapper {
	newLogger := s.Logger.With(args...)
	return slogWrapper{newLogger}
}
