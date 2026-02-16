package lru

import (
	"errors"
	"sync"
)

type LRUValue[K comparable, V any] struct {
	value      V
	key        K
	next, prev *LRUValue[K, V]
}

type LRU[K comparable, V any] struct {
	items    map[K]*LRUValue[K, V]
	root     *LRUValue[K, V]
	onEvict  func(k K, v V)
	cap, len int
	mu       sync.RWMutex
}

func New[K comparable, V any](size int) (*LRU[K, V], error) {
	return NewWithEvict[K, V](size, nil)
}

func NewWithEvict[K comparable, V any](size int, onEvict func(key K, value V)) (*LRU[K, V], error) {
	if size == 0 {
		return nil, errors.New("Size needs to be a positive non-zero integer")
	}

	v := make(map[K]*LRUValue[K, V], size)
	root := &LRUValue[K, V]{}
	root.next = root
	root.prev = root
	return &LRU[K, V]{
		items:   v,
		root:    root,
		cap:     size,
		onEvict: onEvict,
		mu:      sync.RWMutex{},
	}, nil
}

func (c *LRU[K, V]) moveItemToFront(item *LRUValue[K, V]) {
	item.prev.next = item.next
	item.next.prev = item.prev

	item.prev = c.root
	item.next = c.root.next

	c.root.next.prev = item
	c.root.next = item
}

func (c *LRU[K, V]) insertFront(item *LRUValue[K, V]) {
	item.next = c.root.next
	item.prev = c.root
	c.root.next.prev = item
	c.root.next = item
	c.len++
}

func (c *LRU[K, V]) removeItem(item *LRUValue[K, V]) {
	item.prev.next = item.next
	item.next.prev = item.prev
	item.prev = nil
	item.next = nil
	delete(c.items, item.key)
	c.len--
	if c.onEvict != nil {
		c.onEvict(item.key, item.value)
	}
}

func (c *LRU[K, V]) removeOldest() {
	c.removeItem(c.root.prev)
}

func (c *LRU[K, V]) Set(k K, v V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, ok := c.items[k]; ok {
		c.moveItemToFront(item)
		item.value = v
		return
	}

	value := &LRUValue[K, V]{value: v, key: k}
	c.items[k] = value
	c.insertFront(value)

	if c.len > c.cap {
		c.removeOldest()
	}
}

func (c *LRU[K, V]) Get(k K) (value V, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, ok := c.items[k]; ok {
		c.moveItemToFront(item)
		return item.value, true
	}
	return
}

func (c *LRU[K, V]) Peek(k K) (value V, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if item, ok := c.items[k]; ok {
		return item.value, true
	}
	return
}

func (c *LRU[K, V]) Contains(k K) (present bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, ok := c.items[k]
	return ok
}

func (c *LRU[K, V]) Delete(k K) (present bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, ok := c.items[k]; ok {
		c.removeItem(item)
		return true
	}
	return false
}
