package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/Mkamono/slogcontext/slogcontext"
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
		echoContextLogger.InfoContext(c, "Hello, World!", slogcontext.Attrs{"key2": "value2"})
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.Logger.Fatal(e.Start(":1323"))
}

func setupLogger() {
	baseLogHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	})
	logHandler := slogcontext.NewHandler(baseLogHandler)
	slog.SetDefault(slog.New(logHandler))
}
