package adapter

import "log/slog"

// implement Handler interface
var _ = slog.HandlerOptions{
	ReplaceAttr: replacer(nil),
}

type replacer func(groups []string, a slog.Attr) slog.Attr

type ReplaceRule func(a slog.Attr) slog.Attr

func NewReplacer(rules ...ReplaceRule) replacer {
	return func(groups []string, a slog.Attr) slog.Attr {
		for _, rule := range rules {
			a = rule(a)
		}
		return a
	}
}
