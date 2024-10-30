package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/Mkamono/slogcontext/slogcontext"
	"github.com/Mkamono/slogcontext/slogcontext/adapter"
	echoContextLogger "github.com/Mkamono/slogcontext/slogcontext/wrapper/echo"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	setupLogger()

	// Set up server
	e.GET("/", func(c echo.Context) error {
		// Log with context
		echoContextLogger.WithValue(c, slogcontext.Attrs{"key": "value"})
		echoContextLogger.InfoContext(c, "Hello, World!", slogcontext.Attrs{"key": "value"})
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.Logger.Fatal(e.Start(":1323"))
}

func setupLogger() {
	baseHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		ReplaceAttr: adapter.NewReplacer(
			adapter.SourceHierarchyRule(echoContextLogger.WrapHierarchy),
		),
	})
	slogcontext.NewHandler(baseHandler)
	slog.SetDefault(slog.New(baseHandler))
}
