package slogcontext

import (
	"context"
	"runtime"
)

type pcKeyType struct{}

var pcKey = pcKeyType{}

// PCをcontextに追加する
// slog.xxxContextのラッパー関数内で呼び出して、ログが呼び出されたPCを取得する
func WithPC(ctx context.Context) context.Context {
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	pc := pcs[0]
	return context.WithValue(ctx, pcKey, pc)
}
