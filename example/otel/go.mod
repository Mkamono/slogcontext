module otel-example

go 1.23.2

require (
	github.com/Mkamono/slogcontext/slogcontext v0.0.0
	github.com/Mkamono/slogcontext/slogcontext/otel v0.0.0
	go.opentelemetry.io/otel v1.33.0
	go.opentelemetry.io/otel/sdk v1.33.0
)

require (
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel/metric v1.33.0 // indirect
	go.opentelemetry.io/otel/trace v1.33.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
)

replace (
	github.com/Mkamono/slogcontext/slogcontext => ../../slogcontext
	github.com/Mkamono/slogcontext/slogcontext/otel => ../../slogcontext/otel
)
