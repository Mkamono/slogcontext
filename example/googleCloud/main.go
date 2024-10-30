package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/Mkamono/slogcontext/slogcontext"
	"github.com/Mkamono/slogcontext/slogcontext/adapter"
	googlecloud "github.com/Mkamono/slogcontext/slogcontext/adapter/googleCloud"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = slogcontext.WithValue(ctx, slogcontext.Attrs{"key": "value"})
	r = r.WithContext(ctx)
	slog.InfoContext(r.Context(), "Hello, World!", "key2", "value2")

	hello := []byte("Hello World!!!")
	_, err := w.Write(hello)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	setupLogger()
	http.HandleFunc("/", rootHandler)
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

func setupLogger() {
	baseLogHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		ReplaceAttr: adapter.NewReplacer(
			googlecloud.KeyRule(),
		),
	})
	logHandler := slogcontext.NewHandler(baseLogHandler)
	slog.SetDefault(slog.New(logHandler))
}
