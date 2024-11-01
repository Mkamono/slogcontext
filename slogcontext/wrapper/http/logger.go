package http

import (
	"net/http"

	"github.com/Mkamono/slogcontext/slogcontext"
)

func WithValue(req *http.Request, attrs slogcontext.Attrs) *http.Request {
	ctx := slogcontext.WithValue(req.Context(), attrs)
	return req.WithContext(ctx)
}
