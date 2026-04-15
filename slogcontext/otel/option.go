package otel

const (
	defaultTraceKey = "trace_id"
	defaultSpanKey  = "span_id"
)

// Option is a functional option for configuring Handler.
type Option func(*options)

type options struct {
	traceKey       string
	spanKey        string
	traceFlagsKey  string
	includeBaggage bool
	baggagePrefix  string
}

func defaultOptions() options {
	return options{
		traceKey: defaultTraceKey,
		spanKey:  defaultSpanKey,
	}
}

// WithTraceKey sets the slog attribute key used for the trace ID.
func WithTraceKey(key string) Option {
	return func(o *options) { o.traceKey = key }
}

// WithSpanKey sets the slog attribute key used for the span ID.
func WithSpanKey(key string) Option {
	return func(o *options) { o.spanKey = key }
}

// WithTraceFlagsKey enables injection of W3C trace flags under the given key.
// Pass an empty string to disable (the default).
func WithTraceFlagsKey(key string) Option {
	return func(o *options) { o.traceFlagsKey = key }
}

// WithBaggage enables automatic injection of OTel baggage members as log attributes.
// Each member key is optionally prefixed with prefix.
func WithBaggage(prefix string) Option {
	return func(o *options) {
		o.includeBaggage = true
		o.baggagePrefix = prefix
	}
}
