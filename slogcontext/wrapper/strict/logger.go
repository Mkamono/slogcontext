package strict

import (
	"context"

	"github.com/Mkamono/slogcontext/slogcontext"
)

type strictKey struct{ keyStr string }

var (
	UserIDKey    = strictKey{keyStr: "user_id"}
	ReqIDKey     = strictKey{keyStr: "req_id"}
	TraceIDKey   = strictKey{keyStr: "trace_id"}
	SessionIDKey = strictKey{keyStr: "session_id"}
)

func WithValue(ctx context.Context, key strictKey, val string) context.Context {
	return slogcontext.WithValue(ctx, slogcontext.Attrs{key.keyStr: val})
}
