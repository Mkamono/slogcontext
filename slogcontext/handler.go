package slogcontext

import (
	"context"
	"log/slog"
	"sync"
)

var _ slog.Handler = Handler{}

type Handler struct {
	handler slog.Handler
}

func NewHandler(handler slog.Handler) slog.Handler {
	return Handler{
		handler: handler,
	}
}

func (h Handler) Handle(ctx context.Context, record slog.Record) error {
	if v, ok := ctx.Value(logContextKey{}).(*sync.Map); ok {
		v.Range(func(key, val any) bool {
			if keyString, ok := key.(string); ok {
				record.AddAttrs(slog.Any(keyString, val))
			}
			return true
		})
	}

	// ctxからpcを取得し、Recordを更新する
	pc, ok := ctx.Value(pcKey).(uintptr)
	if ok {
		record.PC = pc
	}

	return h.handler.Handle(ctx, record)
}

// implement Handler interface

func (h Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return Handler{h.handler.WithAttrs(attrs)}
}

func (h Handler) WithGroup(name string) slog.Handler {
	return h.handler.WithGroup(name)
}
