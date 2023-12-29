package syncmap

import "sync"

type M[K comparable, V any] struct {
	m sync.Map
}

func (m *M[K, V]) Load(key K) (value V, ok bool) {
	v, ok := m.m.Load(key)
	if !ok {
		return value, false
	}
	return v.(V), true
}

func (m *M[K, V]) Store(key K, value V) {
	m.m.Store(key, value)
}

func (m *M[K, V]) Delete(key K) {
	m.m.Delete(key)
}

func (m *M[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	ac, ok := m.m.LoadOrStore(key, value)
	if !ok {
		return actual, false
	}
	return ac.(V), true
}

func (m *M[K, V]) LoadAndDelete(key K) (value V, ok bool) {
	v, ok := m.m.LoadAndDelete(key)
	if !ok {
		return value, false
	}
	return v.(V), true
}

func (m *M[K,V]) Swap(key K, value V) (previous V, loaded bool) {
	p, ok := m.m.Swap(key, value)
	if !ok {
		return previous, false
	}
	return p.(V), true
}

func (m *M[K, V]) CompareAndDelete(key K, value V) (deleted bool) {
	return m.m.CompareAndDelete(key, value)
}

func (m *M[K, V]) CompareAndSwap(key K, old, new V) bool {
	return m.m.CompareAndSwap(key, old, new)
}

func (m *M[K, V]) Range(f func(key K, value V) bool) {
	m.m.Range(func (k any, v any) bool {
		return f(k.(K), v.(V))
	})
}

