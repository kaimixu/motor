package conf

import (
	"strings"
	"sync/atomic"
)

type Storage struct {
	values atomic.Value
}

func KeyName(key string) string {
	return strings.ToLower(key)
}

func (p *Storage) Store(s map[string]*Value) {
	m := make(map[string]*Value, len(s))

	for k, v := range s {
		m[KeyName(k)] = v
	}

	p.values.Store(m)
}

func (p *Storage) Load() map[string]*Value {
	src := p.values.Load().(map[string]*Value)
	dst := make(map[string]*Value)
	for k, v := range src {
		dst[k] = v
	}

	return dst
}

func (p *Storage) Get(key string) *Value {
	return p.Load()[KeyName(key)]
}
