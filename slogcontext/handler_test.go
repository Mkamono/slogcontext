package slogcontext_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/Mkamono/slogcontext/slogcontext"
)

// collectingHandler is a minimal slog.Handler that stores attrs added to the record.
type collectingHandler struct {
	collect *[]slog.Attr
}

func (h *collectingHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }
func (h *collectingHandler) Handle(_ context.Context, r slog.Record) error {
	r.Attrs(func(a slog.Attr) bool {
		*h.collect = append(*h.collect, a)
		return true
	})
	return nil
}
func (h *collectingHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return h }
func (h *collectingHandler) WithGroup(name string) slog.Handler       { return h }

func attrsToMap(attrs []slog.Attr) map[string]string {
	m := make(map[string]string, len(attrs))
	for _, a := range attrs {
		m[a.Key] = a.Value.String()
	}
	return m
}

func TestHandler_Handle_ContextAttrs(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		wantKeys map[string]string
	}{
		{
			name:     "empty context adds no attrs",
			ctx:      context.Background(),
			wantKeys: map[string]string{},
		},
		{
			name: "single attr propagated",
			ctx:  slogcontext.WithValue(context.Background(), slogcontext.Attrs{"request_id": "abc"}),
			wantKeys: map[string]string{
				"request_id": "abc",
			},
		},
		{
			name: "multiple attrs propagated",
			ctx: slogcontext.WithValue(context.Background(), slogcontext.Attrs{
				"user_id": "99",
				"env":     "prod",
			}),
			wantKeys: map[string]string{
				"user_id": "99",
				"env":     "prod",
			},
		},
		{
			name: "chained WithValue both visible",
			ctx: func() context.Context {
				ctx := slogcontext.WithValue(context.Background(), slogcontext.Attrs{"a": "1"})
				return slogcontext.WithValue(ctx, slogcontext.Attrs{"b": "2"})
			}(),
			wantKeys: map[string]string{"a": "1", "b": "2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []slog.Attr
			h := slogcontext.NewHandler(&collectingHandler{collect: &got})
			r := slog.NewRecord(time.Time{}, slog.LevelInfo, "msg", 0)
			if err := h.Handle(tt.ctx, r); err != nil {
				t.Fatalf("Handle: %v", err)
			}
			gotMap := attrsToMap(got)
			for k, v := range tt.wantKeys {
				if gotMap[k] != v {
					t.Errorf("key %q: want %q, got %q", k, v, gotMap[k])
				}
			}
			if len(got) != len(tt.wantKeys) {
				t.Errorf("attr count: want %d, got %d (%v)", len(tt.wantKeys), len(got), got)
			}
		})
	}
}

func TestHandler_Handle_PCOverride(t *testing.T) {
	var capturedPC uintptr
	inner := &pcCapturingHandler{&capturedPC}
	h := slogcontext.NewHandler(inner)

	ctx := slogcontext.WithPC(context.Background())
	r := slog.NewRecord(time.Time{}, slog.LevelInfo, "msg", 0)
	r.PC = 0 // explicitly zero

	if err := h.Handle(ctx, r); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if capturedPC == 0 {
		t.Error("expected PC to be overridden by WithPC, got 0")
	}
}

func TestHandler_Handle_NoPCWithoutWithPC(t *testing.T) {
	var capturedPC uintptr
	inner := &pcCapturingHandler{&capturedPC}
	h := slogcontext.NewHandler(inner)

	// No WithPC — record.PC must remain at whatever was set
	const originalPC uintptr = 12345
	r := slog.NewRecord(time.Time{}, slog.LevelInfo, "msg", 0)
	r.PC = originalPC

	if err := h.Handle(context.Background(), r); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if capturedPC != originalPC {
		t.Errorf("expected PC=%d unchanged, got %d", originalPC, capturedPC)
	}
}

// pcCapturingHandler captures the PC from the slog.Record passed to Handle.
type pcCapturingHandler struct {
	pc *uintptr
}

func (h *pcCapturingHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }
func (h *pcCapturingHandler) Handle(_ context.Context, r slog.Record) error {
	*h.pc = r.PC
	return nil
}
func (h *pcCapturingHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return h }
func (h *pcCapturingHandler) WithGroup(name string) slog.Handler       { return h }

func TestHandler_WithAttrs_ContextAttrsStillInjected(t *testing.T) {
	// WithAttrs must return a slogcontext.Handler (not the bare inner handler),
	// so context attrs are still injected on every Handle call.
	tests := []struct {
		name     string
		ctxAttrs slogcontext.Attrs
		wantKey  string
	}{
		{
			name:     "context attr visible after WithAttrs",
			ctxAttrs: slogcontext.Attrs{"ctx_key": "c"},
			wantKey:  "ctx_key",
		},
		{
			name:     "multiple context attrs visible after WithAttrs",
			ctxAttrs: slogcontext.Attrs{"a": "1", "b": "2"},
			wantKey:  "a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []slog.Attr
			base := slogcontext.NewHandler(&collectingHandler{collect: &got})
			h := base.WithAttrs([]slog.Attr{slog.String("static", "s")})

			ctx := slogcontext.WithValue(context.Background(), tt.ctxAttrs)
			r := slog.NewRecord(time.Time{}, slog.LevelInfo, "msg", 0)
			if err := h.Handle(ctx, r); err != nil {
				t.Fatalf("Handle: %v", err)
			}

			found := false
			for _, a := range got {
				if a.Key == tt.wantKey {
					found = true
				}
			}
			if !found {
				t.Errorf("context attr %q missing after WithAttrs, got: %v", tt.wantKey, got)
			}
		})
	}
}

func TestHandler_WithGroup_ContextAttrsStillInjected(t *testing.T) {
	// WithGroup must return a Handler that still injects context attrs
	// at the top level (context attrs are not grouped).
	var got []slog.Attr
	base := slogcontext.NewHandler(&collectingHandler{collect: &got})
	h := base.WithGroup("grp")

	ctx := slogcontext.WithValue(context.Background(), slogcontext.Attrs{"req_id": "xyz"})
	r := slog.NewRecord(time.Time{}, slog.LevelInfo, "msg", 0)
	if err := h.Handle(ctx, r); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	found := false
	for _, a := range got {
		if a.Key == "req_id" {
			found = true
		}
	}
	if !found {
		t.Errorf("context attr missing after WithGroup, got: %v", got)
	}
}
