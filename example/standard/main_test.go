package main

import (
	"context"
	"net/http"
	"testing"
)

func Test_rootHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		args args
	}{
		{"", args{
			nil,
			func() *http.Request {
				r, _ := http.NewRequest("GET", "/", nil)
				return r.WithContext(context.Background())
			}(),
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootHandler(tt.args.w, tt.args.r)
		})
	}
}
