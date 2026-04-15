package googlecloud_test

import (
	"log/slog"
	"testing"

	googlecloud "github.com/Mkamono/slogcontext/slogcontext/adapter/googleCloud"
)

func TestKeyRule(t *testing.T) {
	tests := []struct {
		name    string
		in      slog.Attr
		wantKey string
		wantVal string
	}{
		{
			name:    "level INFO stays as severity",
			in:      slog.String(slog.LevelKey, slog.LevelInfo.String()),
			wantKey: googlecloud.LevelKey,
			wantVal: "INFO",
		},
		{
			name:    "level WARN becomes WARNING",
			in:      slog.String(slog.LevelKey, slog.LevelWarn.String()),
			wantKey: googlecloud.LevelKey,
			wantVal: "WARNING",
		},
		{
			name:    "level ERROR stays as severity",
			in:      slog.String(slog.LevelKey, slog.LevelError.String()),
			wantKey: googlecloud.LevelKey,
			wantVal: "ERROR",
		},
		{
			name:    "level DEBUG stays as severity",
			in:      slog.String(slog.LevelKey, slog.LevelDebug.String()),
			wantKey: googlecloud.LevelKey,
			wantVal: "DEBUG",
		},
		{
			name:    "msg key remapped to message",
			in:      slog.String(slog.MessageKey, "hello"),
			wantKey: googlecloud.MessageKey,
			wantVal: "hello",
		},
		{
			name:    "source key remapped",
			in:      slog.String(slog.SourceKey, "file.go:10"),
			wantKey: googlecloud.SourceKey,
			wantVal: "file.go:10",
		},
		{
			name:    "unknown key passed through unchanged",
			in:      slog.String("custom", "value"),
			wantKey: "custom",
			wantVal: "value",
		},
	}

	rule := googlecloud.KeyRule()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rule(tt.in)
			if got.Key != tt.wantKey {
				t.Errorf("key: want %q, got %q", tt.wantKey, got.Key)
			}
			if got.Value.String() != tt.wantVal {
				t.Errorf("value: want %q, got %q", tt.wantVal, got.Value.String())
			}
		})
	}
}

func TestOtelRule(t *testing.T) {
	const project = "my-project"
	const traceID = "4bf92f3577b34da6a3ce929d0e0e4736"
	const spanID = "00f067aa0ba902b7"

	tests := []struct {
		name    string
		in      slog.Attr
		wantKey string
		wantVal string
	}{
		{
			name:    "trace_id remapped with project prefix",
			in:      slog.String("trace_id", traceID),
			wantKey: googlecloud.TraceKey,
			wantVal: "projects/" + project + "/traces/" + traceID,
		},
		{
			name:    "span_id key remapped",
			in:      slog.String("span_id", spanID),
			wantKey: googlecloud.SpanKey,
			wantVal: spanID,
		},
		{
			name:    "unrelated key unchanged",
			in:      slog.String("user_id", "42"),
			wantKey: "user_id",
			wantVal: "42",
		},
	}

	rule := googlecloud.OtelRule(project)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rule(tt.in)
			if got.Key != tt.wantKey {
				t.Errorf("key: want %q, got %q", tt.wantKey, got.Key)
			}
			if got.Value.String() != tt.wantVal {
				t.Errorf("value: want %q, got %q", tt.wantVal, got.Value.String())
			}
		})
	}
}

func TestOtelRuleWithKeys(t *testing.T) {
	const project = "proj"

	tests := []struct {
		name     string
		traceKey string
		spanKey  string
		in       slog.Attr
		wantKey  string
		wantVal  string
	}{
		{
			name:     "custom trace key matched",
			traceKey: "my_trace",
			spanKey:  "my_span",
			in:       slog.String("my_trace", "traceval"),
			wantKey:  googlecloud.TraceKey,
			wantVal:  "projects/" + project + "/traces/traceval",
		},
		{
			name:     "custom span key matched",
			traceKey: "my_trace",
			spanKey:  "my_span",
			in:       slog.String("my_span", "spanval"),
			wantKey:  googlecloud.SpanKey,
			wantVal:  "spanval",
		},
		{
			name:     "default keys not matched when custom keys set",
			traceKey: "my_trace",
			spanKey:  "my_span",
			in:       slog.String("trace_id", "should_not_match"),
			wantKey:  "trace_id",
			wantVal:  "should_not_match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := googlecloud.OtelRuleWithKeys(project, tt.traceKey, tt.spanKey)
			got := rule(tt.in)
			if got.Key != tt.wantKey {
				t.Errorf("key: want %q, got %q", tt.wantKey, got.Key)
			}
			if got.Value.String() != tt.wantVal {
				t.Errorf("value: want %q, got %q", tt.wantVal, got.Value.String())
			}
		})
	}
}
