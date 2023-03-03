package kit

import (
	"sync"
)


type MapSync[K comparable, V any] struct {
	Data sync.Map
}

func NewMapSync[K comparable, V any]() *MapSync[K, V] {
	return &MapSync[K, V]{}
}

func (m *MapSync[K, V]) Set(k K, v V) {
	m.Data.Store(k, v)
}

func (m *MapSync[K, V]) Load(k K) (V, bool) {
	value, ok := m.Data.Load(k)
	return value.(V), ok
}

func (m *MapSync[K, V]) LoadOrStore(k K, v V) (actual V, loaded bool) {
	actualAny, loaded1 := m.Data.LoadOrStore(k, v)
	return actualAny.(V), loaded1
}

func (m *MapSync[K, V]) Exist(k K) bool {
	_, ok := m.Data.Load(k)
	return ok
}

func (m *MapSync[K, V]) Delete(k K) {
	m.Data.Delete(k)
}

func (m *MapSync[K, V]) Range(f func(key K, value V) bool) {
	m.Data.Range(func(key, value any) bool {
		return f(key.(K), value.(V))
	})
}
