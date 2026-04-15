package otel_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"go.opentelemetry.io/otel/baggage"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/Mkamono/slogcontext/slogcontext/otel"
)

// captureHandler stores attrs from each handled record.
type captureHandler struct {
	attrs []slog.Attr
}

func (h *captureHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }
func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	r.Attrs(func(a slog.Attr) bool {
		h.attrs = append(h.attrs, a)
		return true
	})
	return nil
}
func (h *captureHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(name string) slog.Handler       { return h }

func attrMap(attrs []slog.Attr) map[string]string {
	m := make(map[string]string, len(attrs))
	for _, a := range attrs {
		m[a.Key] = a.Value.String()
	}
	return m
}

// newRecordingSpanCtx starts a real SDK span (IsRecording() == true).
func newRecordingSpanCtx(t *testing.T) (context.Context, trace.Span) {
	t.Helper()
	tp := sdktrace.NewTracerProvider()
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
	return tp.Tracer("test").Start(context.Background(), "op")
}

func TestHandler_Handle_TraceAndSpanInjected(t *testing.T) {
	ctx, span := newRecordingSpanCtx(t)
	defer span.End()

	inner := &captureHandler{}
	h := otel.NewHandler(inner)
	r := slog.NewRecord(time.Time{}, slog.LevelInfo, "msg", 0)
	if err := h.Handle(ctx, r); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	got := attrMap(inner.attrs)
	if got[otel.DefaultTraceKey] == "" {
		t.Errorf("expected trace_id to be set, got attrs: %v", inner.attrs)
	}
	if got[otel.DefaultSpanKey] == "" {
		t.Errorf("expected span_id to be set, got attrs: %v", inner.attrs)
	}
}

func TestHandler_Handle_NoAttrsWhenNoSpan(t *testing.T) {
	// Background context has no active span → IsRecording() == false.
	inner := &captureHandler{}
	h := otel.NewHandler(inner)
	r := slog.NewRecord(time.Time{}, slog.LevelInfo, "msg", 0)
	if err := h.Handle(context.Background(), r); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(inner.attrs) != 0 {
		t.Errorf("expected no attrs with no span, got: %v", inner.attrs)
	}
}

func TestHandler_Handle_EndedSpanNotInjected(t *testing.T) {
	ctx, span := newRecordingSpanCtx(t)
	span.End() // ended → IsRecording() == false

	inner := &captureHandler{}
	h := otel.NewHandler(inner)
	r := slog.NewRecord(time.Time{}, slog.LevelInfo, "msg", 0)
	if err := h.Handle(ctx, r); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	got := attrMap(inner.attrs)
	if got[otel.DefaultTraceKey] != "" || got[otel.DefaultSpanKey] != "" {
		t.Errorf("ended span must not inject trace attrs, got: %v", inner.attrs)
	}
}

func TestHandler_Handle_CustomKeys(t *testing.T) {
	ctx, span := newRecordingSpanCtx(t)
	defer span.End()

	inner := &captureHandler{}
	h := otel.NewHandler(inner, otel.WithTraceKey("tid"), otel.WithSpanKey("sid"))
	r := slog.NewRecord(time.Time{}, slog.LevelInfo, "msg", 0)
	if err := h.Handle(ctx, r); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	got := attrMap(inner.attrs)
	if got["tid"] == "" {
		t.Errorf("expected custom trace key 'tid', got attrs: %v", inner.attrs)
	}
	if got["sid"] == "" {
		t.Errorf("expected custom span key 'sid', got attrs: %v", inner.attrs)
	}
	if got[otel.DefaultTraceKey] != "" || got[otel.DefaultSpanKey] != "" {
		t.Errorf("default keys must not appear when custom keys set, got: %v", inner.attrs)
	}
}

func TestHandler_Handle_TraceFlagsKey(t *testing.T) {
	ctx, span := newRecordingSpanCtx(t)
	defer span.End()

	inner := &captureHandler{}
	h := otel.NewHandler(inner, otel.WithTraceFlagsKey("flags"))
	r := slog.NewRecord(time.Time{}, slog.LevelInfo, "msg", 0)
	if err := h.Handle(ctx, r); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	got := attrMap(inner.attrs)
	if _, ok := got["flags"]; !ok {
		t.Errorf("expected flags key in attrs, got: %v", inner.attrs)
	}
}

func TestHandler_WithAttrs_OptsPreserved(t *testing.T) {
	// WithAttrs must return a Handler that still injects trace/span with the
	// same options (custom keys) as the original.
	ctx, span := newRecordingSpanCtx(t)
	defer span.End()

	inner := &captureHandler{}
	h := otel.NewHandler(inner, otel.WithTraceKey("tid"), otel.WithSpanKey("sid"))
	h = h.WithAttrs([]slog.Attr{slog.String("static", "s")})

	r := slog.NewRecord(time.Time{}, slog.LevelInfo, "msg", 0)
	if err := h.Handle(ctx, r); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	got := attrMap(inner.attrs)
	if got["tid"] == "" {
		t.Errorf("WithAttrs lost custom traceKey option, got attrs: %v", inner.attrs)
	}
	if got["sid"] == "" {
		t.Errorf("WithAttrs lost custom spanKey option, got attrs: %v", inner.attrs)
	}
}

func TestHandler_WithGroup_OptsPreserved(t *testing.T) {
	// WithGroup must return a Handler that still injects trace/span.
	ctx, span := newRecordingSpanCtx(t)
	defer span.End()

	inner := &captureHandler{}
	h := otel.NewHandler(inner, otel.WithTraceKey("tid"), otel.WithSpanKey("sid"))
	h = h.WithGroup("grp")

	r := slog.NewRecord(time.Time{}, slog.LevelInfo, "msg", 0)
	if err := h.Handle(ctx, r); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	got := attrMap(inner.attrs)
	if got["tid"] == "" {
		t.Errorf("WithGroup lost custom traceKey option, got attrs: %v", inner.attrs)
	}
	if got["sid"] == "" {
		t.Errorf("WithGroup lost custom spanKey option, got attrs: %v", inner.attrs)
	}
}

func TestHandler_Handle_Baggage(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		members  map[string]string
		wantKeys map[string]string
	}{
		{
			name:     "baggage member without prefix",
			prefix:   "",
			members:  map[string]string{"tenant": "acme"},
			wantKeys: map[string]string{"tenant": "acme"},
		},
		{
			name:     "baggage member with prefix",
			prefix:   "baggage_",
			members:  map[string]string{"env": "prod"},
			wantKeys: map[string]string{"baggage_env": "prod"},
		},
		{
			name:     "multiple baggage members",
			prefix:   "b_",
			members:  map[string]string{"a": "1", "c": "3"},
			wantKeys: map[string]string{"b_a": "1", "b_c": "3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bag, _ := baggage.New()
			for k, v := range tt.members {
				m, _ := baggage.NewMember(k, v)
				bag, _ = bag.SetMember(m)
			}
			ctx := baggage.ContextWithBaggage(context.Background(), bag)

			inner := &captureHandler{}
			h := otel.NewHandler(inner, otel.WithBaggage(tt.prefix))
			r := slog.NewRecord(time.Time{}, slog.LevelInfo, "msg", 0)
			if err := h.Handle(ctx, r); err != nil {
				t.Fatalf("Handle: %v", err)
			}

			got := attrMap(inner.attrs)
			for k, v := range tt.wantKeys {
				if got[k] != v {
					t.Errorf("key %q: want %q, got %q (all attrs: %v)", k, v, got[k], inner.attrs)
				}
			}
		})
	}
}
