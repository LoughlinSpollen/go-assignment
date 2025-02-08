package cache

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

type OrderedMap interface {
	Set(key, value string) bool
	Get(key string) (string, bool)
	Items() []struct {
		Key   string
		Value string
	}
	Delete(key string) bool
}

type node[K comparable, V string] struct {
	key   K
	value V
	// to allow scannning all items based on the order they were inserted
	prev *node[K, V]
	next *node[K, V]
}

func newNode[K comparable, V string](key K, value V) *node[K, V] {
	log.Trace("map newNode")

	return &node[K, V]{
		key:   key,
		value: value,
	}
}

type orderedMap[K comparable, V string] struct {
	lock sync.RWMutex
	dict map[K]*node[K, V]
	head *node[K, V]
	tail *node[K, V]
}

func NewOrderedMap[K comparable, V string]() *orderedMap[K, V] {
	log.Trace("map NewMap")

	return &orderedMap[K, V]{
		dict: make(map[K]*node[K, V]),
		head: nil,
		tail: nil,
	}
}

func (m *orderedMap[K, V]) Set(key K, value V) bool {
	log.Trace("map Set")

	m.lock.Lock()
	defer m.lock.Unlock()

	if n, ok := m.dict[key]; ok {
		n.value = value
		return true
	}

	n := newNode(key, value)
	m.dict[key] = n

	if m.head == nil {
		m.head = n
		m.tail = n
	} else {
		m.tail.next = n
		n.prev = m.tail
		m.tail = n
	}
	return true
}

func (m *orderedMap[K, V]) Get(key K) (V, bool) {
	log.Trace("map Get")

	m.lock.RLock()
	defer m.lock.RUnlock()

	var n *node[K, V]
	var ok bool
	if n, ok = m.dict[key]; !ok {
		var zeroValue V
		return zeroValue, false
	}
	return n.value, true
}

func (m *orderedMap[K, V]) Items() []struct {
	Key   K
	Value V
} {
	log.Trace("map Items")

	m.lock.RLock()
	defer m.lock.RUnlock()

	result := make([]struct {
		Key   K
		Value V
	}, 0, len(m.dict))

	for curr := m.head; curr != nil; curr = curr.next {
		result = append(result, struct {
			Key   K
			Value V
		}{
			Key:   curr.key,
			Value: curr.value,
		})
	}
	return result
}

func (m *orderedMap[K, V]) Delete(key K) bool {
	log.Trace("map Delete")

	m.lock.Lock()
	defer m.lock.Unlock()

	var n *node[K, V]
	var ok bool
	if n, ok = m.dict[key]; !ok {
		return false
	}

	delete(m.dict, key)
	if n.prev != nil {
		n.prev.next = n.next
	} else {
		m.head = n.next
	}

	if n.next != nil {
		n.next.prev = n.prev
	} else {
		m.tail = n.prev
	}

	return true
}
