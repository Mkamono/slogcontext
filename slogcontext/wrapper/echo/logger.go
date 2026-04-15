package echo

import (
	"log/slog"

	"github.com/Mkamono/slogcontext/slogcontext"
	"github.com/labstack/echo/v4"
)

func InfoContext(eCtx echo.Context, msg string, args ...any) {
	ctx := slogcontext.WithPC(eCtx.Request().Context())
	slog.InfoContext(ctx, msg, args...)
}

func ErrorContext(eCtx echo.Context, msg string, args ...any) {
	ctx := slogcontext.WithPC(eCtx.Request().Context())
	slog.ErrorContext(ctx, msg, args...)
}

func WarnContext(eCtx echo.Context, msg string, args ...any) {
	ctx := slogcontext.WithPC(eCtx.Request().Context())
	slog.WarnContext(ctx, msg, args...)
}

func WithValue(eCtx echo.Context, attrs slogcontext.Attrs) {
	ctx := slogcontext.WithValue(eCtx.Request().Context(), attrs)
	eCtx.SetRequest(eCtx.Request().WithContext(ctx))
}
