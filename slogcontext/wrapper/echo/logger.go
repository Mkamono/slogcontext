package echo

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/Mkamono/slogcontext/slogcontext"
)

// implement echo.Context interface
type customContext interface {
	Request() *http.Request
	SetRequest(r *http.Request)
}

func getContext(eCtx customContext) context.Context {
	ctx := eCtx.Request().Context()
	return ctx
}

func setContent(eCtx customContext, ctx context.Context) {
	eCtx.SetRequest(eCtx.Request().WithContext(ctx))
}

func split(attrs slogcontext.Attrs) []any {
	args := make([]any, 0, len(attrs)*2)
	for k, v := range attrs {
		args = append(args, k, v)
	}
	return args
}

func InfoContext(eCtx customContext, msg string, attrs slogcontext.Attrs) {
	ctx := getContext(eCtx)
	ctx = slogcontext.WithPC(ctx)
	slog.InfoContext(ctx, msg, split(attrs)...)
}

func ErrorContext(eCtx customContext, msg string, attrs slogcontext.Attrs) {
	ctx := getContext(eCtx)
	ctx = slogcontext.WithPC(ctx)
	slog.ErrorContext(ctx, msg, split(attrs)...)
}

func WarnContext(eCtx customContext, msg string, attrs slogcontext.Attrs) {
	ctx := getContext(eCtx)
	ctx = slogcontext.WithPC(ctx)
	slog.WarnContext(ctx, msg, split(attrs)...)
}

func WithValue(eCtx customContext, attrs slogcontext.Attrs) {
	ctx := slogcontext.WithValue(getContext(eCtx), attrs)
	setContent(eCtx, ctx)
}
