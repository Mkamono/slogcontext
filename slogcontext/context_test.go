package slogcontext_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/Mkamono/slogcontext/slogcontext"
)

// recordAttrs runs Handle and returns the attrs that were added to the record
// by the slogcontext.Handler.
func collectAttrs(ctx context.Context) []slog.Attr {
	var got []slog.Attr
	h := slogcontext.NewHandler(&collectingHandler{collect: &got})
	r := slog.NewRecord(time.Time{}, slog.LevelInfo, "msg", 0)
	_ = h.Handle(ctx, r)
	return got
}

func TestWithValue(t *testing.T) {
	tests := []struct {
		name     string
		parent   context.Context
		attrs    slogcontext.Attrs
		wantKeys []string
	}{
		{
			name:     "nil parent does not panic",
			parent:   nil,
			attrs:    slogcontext.Attrs{"k": "v"},
			wantKeys: []string{"k"},
		},
		{
			name:     "fresh context stores single attr",
			parent:   context.Background(),
			attrs:    slogcontext.Attrs{"user_id": "42"},
			wantKeys: []string{"user_id"},
		},
		{
			name:     "fresh context stores multiple attrs",
			parent:   context.Background(),
			attrs:    slogcontext.Attrs{"a": 1, "b": "two"},
			wantKeys: []string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := slogcontext.WithValue(tt.parent, tt.attrs)
			got := collectAttrs(ctx)
			keySet := make(map[string]bool, len(got))
			for _, a := range got {
				keySet[a.Key] = true
			}
			for _, k := range tt.wantKeys {
				if !keySet[k] {
					t.Errorf("missing key %q in attrs %v", k, got)
				}
			}
		})
	}
}

func TestWithValue_CopyOnWrite(t *testing.T) {
	parent := context.Background()
	ctx1 := slogcontext.WithValue(parent, slogcontext.Attrs{"x": "1"})
	ctx2 := slogcontext.WithValue(ctx1, slogcontext.Attrs{"y": "2"})

	// ctx1 must NOT see key "y"
	for _, a := range collectAttrs(ctx1) {
		if a.Key == "y" {
			t.Errorf("parent context was mutated: found key %q", a.Key)
		}
	}

	// ctx2 must see both keys
	keySet := make(map[string]bool)
	for _, a := range collectAttrs(ctx2) {
		keySet[a.Key] = true
	}
	for _, k := range []string{"x", "y"} {
		if !keySet[k] {
			t.Errorf("derived context missing key %q", k)
		}
	}
}

func TestWithValue_Overwrite(t *testing.T) {
	ctx := context.Background()
	ctx = slogcontext.WithValue(ctx, slogcontext.Attrs{"key": "first"})
	ctx = slogcontext.WithValue(ctx, slogcontext.Attrs{"key": "second"})

	for _, a := range collectAttrs(ctx) {
		if a.Key == "key" && a.Value.String() != "second" {
			t.Errorf("expected overwritten value %q, got %q", "second", a.Value.String())
		}
	}
}
