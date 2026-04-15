package googlecloud

import (
	"log/slog"

	"github.com/Mkamono/slogcontext/slogcontext/adapter"
)

// OtelRule returns a ReplaceRule that remaps the default otel.Handler key names
// ("trace_id", "span_id") to Google Cloud Logging structured log fields.
// projectID is the GCP project ID used to form the full trace resource name.
//
// The hardcoded key names match otel.DefaultTraceKey and otel.DefaultSpanKey.
// If you configure otel.Handler with otel.WithTraceKey or otel.WithSpanKey,
// use OtelRuleWithKeys instead to keep the two in sync.
//
// Compose with KeyRule via adapter.NewReplacer:
//
//	adapter.NewReplacer(googlecloud.KeyRule(), googlecloud.OtelRule("my-project"))
func OtelRule(projectID string) adapter.ReplaceRule {
	return OtelRuleWithKeys(projectID, "trace_id", "span_id")
}

// OtelRuleWithKeys is like OtelRule but allows specifying custom attribute key names
// to match the otel.Handler configuration (via otel.WithTraceKey / otel.WithSpanKey).
func OtelRuleWithKeys(projectID, traceKey, spanKey string) adapter.ReplaceRule {
	return func(a slog.Attr) slog.Attr {
		switch a.Key {
		case traceKey:
			return slog.String(TraceKey, "projects/"+projectID+"/traces/"+a.Value.String())
		case spanKey:
			a.Key = SpanKey
		}
		return a
	}
}

func KeyRule() adapter.ReplaceRule {
	return func(a slog.Attr) slog.Attr {
		switch a.Key {
		case slog.LevelKey:
			if a.Value.String() == slog.LevelWarn.String() {
				return slog.String(LevelKey, "WARNING")
			}
			a.Key = LevelKey
		case slog.MessageKey:
			a.Key = MessageKey
		case slog.SourceKey:
			a.Key = SourceKey
		}
		return a
	}
}
