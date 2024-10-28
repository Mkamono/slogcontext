package googlecloud

import (
	"log/slog"

	"github.com/Mkamono/slogcontext/slogcontext/adapter"
)

func KeyRule() adapter.ReplaceRule {
	return func(a slog.Attr) slog.Attr {
		switch a.Key {
		case slog.LevelKey:
			if a.Value.String() == slog.LevelWarn.String() {
				return slog.String(LevelKey, "WARNING")
			}
			a.Key = LabelKey
		case slog.MessageKey:
			a.Key = MessageKey
		case slog.SourceKey:
			a.Key = SourceKey
		}
		return a
	}
}
