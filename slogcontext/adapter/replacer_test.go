package adapter_test

import (
	"log/slog"
	"testing"

	"github.com/Mkamono/slogcontext/slogcontext/adapter"
)

func TestNewReplacer(t *testing.T) {
	upper := func(a slog.Attr) slog.Attr {
		if a.Key == "level" {
			return slog.String(a.Key, "UPPER")
		}
		return a
	}
	prefix := func(a slog.Attr) slog.Attr {
		return slog.Attr{Key: "pfx_" + a.Key, Value: a.Value}
	}

	tests := []struct {
		name    string
		groups  []string
		attr    slog.Attr
		rules   []adapter.ReplaceRule
		wantKey string
		wantVal string
	}{
		{
			name:    "no rules returns attr unchanged",
			groups:  nil,
			attr:    slog.String("foo", "bar"),
			rules:   nil,
			wantKey: "foo",
			wantVal: "bar",
		},
		{
			name:    "single rule applied",
			groups:  nil,
			attr:    slog.String("level", "info"),
			rules:   []adapter.ReplaceRule{upper},
			wantKey: "level",
			wantVal: "UPPER",
		},
		{
			name:    "rules applied in order",
			groups:  nil,
			attr:    slog.String("level", "info"),
			rules:   []adapter.ReplaceRule{upper, prefix},
			wantKey: "pfx_level",
			wantVal: "UPPER",
		},
		{
			name:    "non-empty groups skips all rules",
			groups:  []string{"grp"},
			attr:    slog.String("level", "info"),
			rules:   []adapter.ReplaceRule{upper},
			wantKey: "level",
			wantVal: "info",
		},
		{
			name:    "nested groups skips all rules",
			groups:  []string{"grp", "sub"},
			attr:    slog.String("k", "v"),
			rules:   []adapter.ReplaceRule{prefix},
			wantKey: "k",
			wantVal: "v",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := adapter.NewReplacer(tt.rules...)
			got := r(tt.groups, tt.attr)
			if got.Key != tt.wantKey {
				t.Errorf("key: want %q, got %q", tt.wantKey, got.Key)
			}
			if got.Value.String() != tt.wantVal {
				t.Errorf("value: want %q, got %q", tt.wantVal, got.Value.String())
			}
		})
	}
}
