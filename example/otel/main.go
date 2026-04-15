package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/Mkamono/slogcontext/slogcontext"
	slogcontextotel "github.com/Mkamono/slogcontext/slogcontext/otel"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	tp := sdktrace.NewTracerProvider()
	defer tp.Shutdown(context.Background())
	otel.SetTracerProvider(tp)

	setupLogger()

	http.HandleFunc("/", rootHandler)
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tracer := otel.Tracer("example")
	ctx, span := tracer.Start(ctx, "rootHandler")
	defer span.End()

	ctx = slogcontext.WithValue(ctx, slogcontext.Attrs{"user_id": "u-123"})

	slog.InfoContext(ctx, "handling request", "key2", "value2")
	w.Write([]byte("ok"))
}

func setupLogger() {
	base := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true})
	handler := slogcontext.NewHandler(slogcontextotel.NewHandler(base))
	slog.SetDefault(slog.New(handler))
}
