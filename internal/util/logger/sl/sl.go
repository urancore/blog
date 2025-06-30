package sl

import "log/slog"

func Error(value error) slog.Attr {
	return slog.Attr{Key: "err",
		Value: slog.AnyValue(value),
	}
}
