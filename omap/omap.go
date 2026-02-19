package omap

import (
	"cmp"
	"iter"
	"maps"
	"slices"
)

// Map is an ordered map.
type Map[K cmp.Ordered, V any] struct {
	entries map[K]V
	order   []K
}

// Entry represents an entry in a map.
type Entry[K cmp.Ordered, V any] struct {
	Key   K
	Value V
}

// New creates an ordered map from a list of entries.
func New[K cmp.Ordered, V any](entries ...Entry[K, V]) *Map[K, V] {
	m := NewWithCapacity[K, V](len(entries))
	for _, el := range entries {
		m.Set(el.Key, el.Value)
	}
	return m
}

// NewWithCapacity creates an empty ordered map with the given capacity.
func NewWithCapacity[K cmp.Ordered, V any](capacity int) *Map[K, V] {
	return &Map[K, V]{
		entries: make(map[K]V, capacity),
		order:   make([]K, 0, capacity),
	}
}

// Collect collects key-value pairs from seq into a new ordered map and returns it.
func Collect[K cmp.Ordered, V any](seq iter.Seq2[K, V]) *Map[K, V] {
	m := New[K, V]()
	m.Insert(seq)
	return m
}

// FromMap creates an ordered map from an existing map.
// The existing entries are ordered by sorting the keys.
func FromMap[K cmp.Ordered, V any](values map[K]V) *Map[K, V] {
	return &Map[K, V]{
		entries: values,
		order:   slices.Sorted(maps.Keys(values)),
	}
}

// Insert adds the key-value pairs from seq to m. If a key in seq already exists in m,
// its value will be overwritten and its insertion order will be preserved.
func (m *Map[K, V]) Insert(seq iter.Seq2[K, V]) {
	for key, value := range seq {
		m.Set(key, value)
	}
}

// IsZero reports if map is empty.
func (m *Map[K, V]) IsZero() bool {
	return m == nil || len(m.entries) == 0
}

// Get returns the value for a key. If the key does not exist,
// ok will be false and value with be the zero value of its type.
func (m *Map[K, V]) Get(key K) (value V, ok bool) {
	if m == nil || m.entries == nil {
		var zero V
		return zero, false
	}
	v, ok := m.entries[key]
	return v, ok
}

// Value returns the value for a key. If the key does not exist,
// value with be the zero value of its type.
func (m *Map[K, V]) Value(key K) (value V) {
	if m == nil || m.entries == nil {
		var zero V
		return zero
	}
	return m.entries[key]
}

// Has reports if the key is in the map.
func (m *Map[K, V]) Has(key K) (ok bool) {
	if m == nil || m.entries == nil {
		return false
	}
	_, ok = m.entries[key]
	return ok
}

// Set sets the value for a key. If the key already exists in the map,
// its value will be overwritten and its insertion order will be preserved.
func (m *Map[K, V]) Set(key K, value V) {
	if m == nil {
		m = NewWithCapacity[K, V](1)
	}

	if m.entries == nil {
		m.entries = map[K]V{}
	}

	// Check if key exists
	_, ok := m.entries[key]
	if ok {
		// Update existing value and return
		m.entries[key] = value
		return
	}

	// Set value and add key to order
	m.entries[key] = value
	m.order = append(m.order, key)
}

// Len returns the number of elements in the map.
func (m *Map[K, V]) Len() int {
	if m == nil {
		return 0
	}
	return len(m.entries)
}

// Delete removes a key from the map.
func (m *Map[K, V]) Delete(key K) {
	if m == nil {
		return
	}
	// Delete from the map
	delete(m.entries, key)
	// Delete the key from the order
	m.order = slices.DeleteFunc(m.order, func(v K) bool {
		return v == key
	})
}

// SetOrder overwrites the order of the ordered map. The provided order
// is sanitized by removing any keys not present in the map and
// adding any additional keys in the map to the end of the order.
func (m *Map[K, V]) SetOrder(order []K) {
	// Sanitize input by removing any keys that
	// do not match a value in the map
	order = slices.DeleteFunc(order, func(key K) bool {
		return m.Has(key)
	})

	// Mark each key in order as visited
	visited := map[K]struct{}{}
	for _, key := range order {
		visited[key] = struct{}{}
	}

	// Add keys not in order to the end
	for _, key := range slices.Sorted(maps.Keys(m.entries)) {
		if _, ok := visited[key]; !ok {
			// Add the key to the slice
			order = append(order, key)
			// Mark as visited
			visited[key] = struct{}{}
		}
	}

	// Store the slice
	m.order = order
}

// Clone returns a copy of the ordered map. This is a shallow clone:
// the new keys and values are set using ordinary assignment.
func (m *Map[K, V]) Clone() *Map[K, V] {
	if m == nil {
		return nil
	}
	return &Map[K, V]{
		entries: maps.Clone(m.entries),
		order:   slices.Clone(m.order),
	}
}

// Backward returns a copy of the ordered map with the order reversed.
func (m *Map[K, V]) Backward() *Map[K, V] {
	bm := m.Clone()
	slices.Reverse(bm.order)
	return bm
}

// All returns an iterator over key-value pairs from m in insertion order.
func (m *Map[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(key K, value V) bool) {
		for key := range m.Keys() {
			if !yield(key, m.Value(key)) {
				return
			}
		}
	}
}

// AllBackward returns an iterator over key-value pairs from m in reverse insertion order.
func (m *Map[K, V]) AllBackward() iter.Seq2[K, V] {
	return func(yield func(key K, value V) bool) {
		for key := range m.KeysBackward() {
			if !yield(key, m.Value(key)) {
				return
			}
		}
	}
}

// Keys returns an iterator over keys in m in insertion order.
func (m *Map[K, V]) Keys() iter.Seq[K] {
	return func(yield func(key K) bool) {
		if m == nil {
			return
		}
		for _, key := range m.order {
			if !yield(key) {
				return
			}
		}
	}
}

// KeysBackward returns an iterator over keys in m in reverse insertion order.
func (m *Map[K, V]) KeysBackward() iter.Seq[K] {
	return func(yield func(key K) bool) {
		if m == nil {
			return
		}
		for _, key := range slices.Backward(m.order) {
			if !yield(key) {
				return
			}
		}
	}
}

// Values returns an iterator over values in m in insertion order.
func (m *Map[K, V]) Values() iter.Seq[V] {
	return func(yield func(value V) bool) {
		for _, value := range m.All() {
			if !yield(value) {
				return
			}
		}
	}
}

// ValuesBackward returns an iterator over values in m in reverse insertion order.
func (m *Map[K, V]) ValuesBackward() iter.Seq[V] {
	return func(yield func(value V) bool) {
		for _, value := range m.AllBackward() {
			if !yield(value) {
				return
			}
		}
	}
}
