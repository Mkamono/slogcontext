package adapter

import "log/slog"

// implement Handler interface
type replacer func(groups []string, a slog.Attr) slog.Attr

var _ = slog.HandlerOptions{
	ReplaceAttr: replacer(nil),
}

type ReplaceRule func(a slog.Attr) slog.Attr

func NewReplacer(rules ...ReplaceRule) replacer {
	return func(groups []string, a slog.Attr) slog.Attr {
		for _, rule := range rules {
			a = rule(a)
		}
		return a
	}
}
