package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/Mkamono/slogcontext/slogcontext"
	"github.com/Mkamono/slogcontext/slogcontext/adapter"
	googlecloud "github.com/Mkamono/slogcontext/slogcontext/adapter/googleCloud"
	slogcontextotel "github.com/Mkamono/slogcontext/slogcontext/otel"
	echoLogger "github.com/Mkamono/slogcontext/slogcontext/wrapper/echo"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const projectID = "my-project"

func main() {
	setupOtel()
	setupLogger()

	e := echo.New()
	e.Use(tracingMiddleware())

	// /hello: ルートspanのみ。trace_id・span_idが出力されることを確認
	e.GET("/hello", helloHandler)

	// /nested: 子spanを切る。同じtrace_idで異なるspan_idになることを確認
	e.GET("/nested", nestedHandler)

	e.Logger.Fatal(e.Start(":1323"))
}

func setupOtel() {
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)
	// W3C traceparentヘッダーからトレースコンテキストを伝播
	otel.SetTextMapPropagator(propagation.TraceContext{})
}

func setupLogger() {
	base := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		ReplaceAttr: adapter.NewReplacer(
			googlecloud.KeyRule(),
			googlecloud.OtelRule(projectID),
		),
	})
	handler := slogcontext.NewHandler(slogcontextotel.NewHandler(base))
	slog.SetDefault(slog.New(handler))
}

// tracingMiddleware はW3C traceparentヘッダーからコンテキストを抽出し、
// ルートspanを開始する。リクエスト全体のtrace_id・span_idがここで決まる。
func tracingMiddleware() echo.MiddlewareFunc {
	tracer := otel.Tracer("server")
	propagator := otel.GetTextMapPropagator()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			// 上流サービスからtraceparentヘッダーで伝播されていれば続きのspanになる
			ctx = propagator.Extract(ctx, propagation.HeaderCarrier(c.Request().Header))
			ctx, span := tracer.Start(ctx, c.Request().URL.Path)
			defer span.End()
			c.SetRequest(c.Request().WithContext(ctx))

			echoLogger.WithValue(c, slogcontext.Attrs{
				"method": c.Request().Method,
				"path":   c.Request().URL.Path,
			})
			echoLogger.InfoContext(c, "request received")

			err := next(c)

			echoLogger.InfoContext(c, "request completed", "status", http.StatusOK)
			return err
		}
	}
}

// helloHandler: ルートspanのまま処理。ミドルウェアと同じtrace_id・span_id
func helloHandler(c echo.Context) error {
	echoLogger.InfoContext(c, "hello handler", "greeting", "hello")
	return c.String(http.StatusOK, "Hello!")
}

// nestedHandler: 子spanを切ってspan_idが変わることを示す
func nestedHandler(c echo.Context) error {
	tracer := otel.Tracer("handler")

	// 子span開始前: ミドルウェアのルートspanのspan_id
	echoLogger.InfoContext(c, "before child span")

	parentCtx := c.Request().Context()
	childCtx, childSpan := tracer.Start(parentCtx, "child-operation")

	// 子spanのコンテキストをリクエストにセット
	c.SetRequest(c.Request().WithContext(childCtx))
	// trace_id は同じ、span_id だけ変わる
	echoLogger.InfoContext(c, "inside child span", "step", "child")

	childSpan.End()

	// 親コンテキストに戻す
	c.SetRequest(c.Request().WithContext(parentCtx))
	// span_id がルートspanに戻る
	echoLogger.InfoContext(c, "after child span")

	return c.String(http.StatusOK, "Nested done!")
}
