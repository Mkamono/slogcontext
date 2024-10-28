package slogcontext

import (
	"context"
	"sync"
)

type logContextKey struct{}

type Attrs map[string]any

func WithValue(parent context.Context, attrs Attrs) context.Context {
	if parent == nil {
		parent = context.Background()
	}
	if v, ok := parent.Value(logContextKey{}).(*sync.Map); ok {
		mapCopy := copySyncMap(v)
		for key, val := range attrs {
			mapCopy.Store(key, val)
		}
		return context.WithValue(parent, logContextKey{}, mapCopy)
	}
	v := &sync.Map{}
	for key, val := range attrs {
		v.Store(key, val)
	}
	return context.WithValue(parent, logContextKey{}, v)
}

func copySyncMap(m *sync.Map) *sync.Map {
	var cp sync.Map
	m.Range(func(k, v interface{}) bool {
		cp.Store(k, v)
		return true
	})
	return &cp
}
