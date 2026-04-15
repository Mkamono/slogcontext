# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

Each sub-module has its own `go.mod` and must be built/tested independently.

```sh
# Core library (stdlib only)
cd slogcontext && go test ./...
cd slogcontext && go test -run TestName ./...

# Sub-modules
cd slogcontext/otel && go build ./...
cd slogcontext/wrapper/echo && go build ./...
cd slogcontext/wrapper/http && go build ./...
cd slogcontext/wrapper/strict && go build ./...

# Examples
cd example/standard && go run main.go
cd example/echo && go run main.go
cd example/googleCloud && go run main.go
cd example/strict && go run main.go
cd example/otel && go run main.go
cd example/googlecloud-otel-echo && go run main.go
```

After adding dependencies to any sub-module, run `go mod tidy` inside that module's directory.

## Module Structure

This repo uses multiple Go modules. Each has its own `go.mod` and local `replace` directives for development:

| Module path | Directory |
|---|---|
| `github.com/Mkamono/slogcontext/slogcontext` | `slogcontext/` — core, stdlib only |
| `github.com/Mkamono/slogcontext/slogcontext/otel` | `slogcontext/otel/` |
| `github.com/Mkamono/slogcontext/slogcontext/wrapper/echo` | `slogcontext/wrapper/echo/` |
| `github.com/Mkamono/slogcontext/slogcontext/wrapper/http` | `slogcontext/wrapper/http/` |
| `github.com/Mkamono/slogcontext/slogcontext/wrapper/strict` | `slogcontext/wrapper/strict/` |

**The core module must remain stdlib-only.** External dependencies belong in sub-modules.

`replace` directives in sub-module `go.mod` files are ignored by downstream consumers — they only affect local development builds.

## Architecture

A Go library that extends `log/slog` to propagate log attributes through `context.Context`.

### Core flow

1. **`slogcontext.WithValue(ctx, Attrs)`** — stores key-value pairs in a `sync.Map` held in context. Each call copies the map to preserve immutability across goroutines.
2. **`slogcontext.NewHandler(base)`** — wraps any `slog.Handler`. On `Handle()`, reads the `sync.Map` from context and appends attrs to the `slog.Record`, then delegates to the inner handler.
3. **`slogcontext.WithPC(ctx)`** — captures `runtime.Callers(3)` and stores the PC in context. `Handler.Handle()` overrides `record.PC` so source location points to the real call site, not the wrapper. Wrapper packages call this on every log function.

### Handler composition

Handlers are stacked innermost-first. The recommended full setup:

```go
// outer → inner → base
slogcontext.NewHandler(otel.NewHandler(base))
```

`slogcontext.Handler` injects context attrs, then delegates to `otel.Handler` which injects trace/span, then delegates to `base`.

### Sub-packages

- **`adapter/`** — `NewReplacer(rules ...ReplaceRule)` composes `slog.HandlerOptions.ReplaceAttr` functions.
- **`adapter/googleCloud/`** — `KeyRule()` remaps slog keys to GCP Structured Logging field names. `OtelRule(projectID)` remaps `trace_id`/`span_id` to `logging.googleapis.com/trace` (with project prefix) and `logging.googleapis.com/spanId`. `OtelRuleWithKeys(projectID, traceKey, spanKey)` for custom key names.
- **`otel/`** — `NewHandler(base, ...Option)` extracts trace/span from an active OTel span via `trace.SpanFromContext`. Uses `IsRecording()` (not `IsValid()`) to exclude ended spans. Options: `WithTraceKey`, `WithSpanKey`, `WithTraceFlagsKey`, `WithBaggage(prefix)`.
- **`wrapper/echo/`** — `InfoContext/ErrorContext/WarnContext(eCtx echo.Context, msg string, args ...any)` mirror the standard `slog` API. `WithValue(eCtx, attrs)` attaches attrs to the request context. All logging calls invoke `WithPC` internally.
- **`wrapper/http/`** — `WithValue(req, attrs)` attaches attrs to an `*http.Request`'s context.
- **`wrapper/strict/`** — pre-defined typed keys (`UserIDKey`, `ReqIDKey`, `TraceIDKey`, `SessionIDKey`) to enforce a fixed logging vocabulary.

### GCP + OTel setup pattern

```go
base := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    AddSource: true,
    ReplaceAttr: adapter.NewReplacer(
        googlecloud.KeyRule(),
        googlecloud.OtelRule("my-gcp-project"),
    ),
})
slog.SetDefault(slog.New(slogcontext.NewHandler(otel.NewHandler(base))))
```
