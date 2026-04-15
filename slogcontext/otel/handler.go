package otel

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
)

var _ slog.Handler = Handler{}

// Handler is a slog.Handler that automatically injects OpenTelemetry trace
// context into every log record. It reads span information from the context
// on each Handle() call — no manual WithValue() is needed.
//
// Compose with slogcontext.NewHandler as the outer layer:
//
//	slogcontext.NewHandler(otel.NewHandler(base))
type Handler struct {
	handler slog.Handler
	opts    options
}

// NewHandler wraps base with OTel trace and optional baggage injection.
func NewHandler(base slog.Handler, opt ...Option) slog.Handler {
	o := defaultOptions()
	for _, fn := range opt {
		fn(&o)
	}
	return Handler{handler: base, opts: o}
}

func (h Handler) Handle(ctx context.Context, record slog.Record) error {
	span := trace.SpanFromContext(ctx)
	// IsRecording() returns true only while the span is active (not yet ended).
	// This intentionally excludes ended spans whose context may still linger,
	// preventing stale trace/span IDs from appearing in logs.
	if span.IsRecording() {
		sc := span.SpanContext()
		if sc.HasTraceID() {
			record.AddAttrs(slog.String(h.opts.traceKey, sc.TraceID().String()))
		}
		if sc.HasSpanID() {
			record.AddAttrs(slog.String(h.opts.spanKey, sc.SpanID().String()))
		}
		if h.opts.traceFlagsKey != "" {
			record.AddAttrs(slog.String(h.opts.traceFlagsKey, sc.TraceFlags().String()))
		}
	}

	if h.opts.includeBaggage {
		bag := baggage.FromContext(ctx)
		for _, member := range bag.Members() {
			record.AddAttrs(slog.String(h.opts.baggagePrefix+member.Key(), member.Value()))
		}
	}

	return h.handler.Handle(ctx, record)
}

func (h Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return Handler{handler: h.handler.WithAttrs(attrs), opts: h.opts}
}

func (h Handler) WithGroup(name string) slog.Handler {
	return Handler{handler: h.handler.WithGroup(name), opts: h.opts}
}
