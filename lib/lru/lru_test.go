package lru

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	cache, err := New[string, int](10)
	require.NoError(t, err)
	assert.NotNil(t, cache)
}

func TestNewZeroSize(t *testing.T) {
	_, err := New[string, int](0)
	assert.Error(t, err)
}

func TestSetAndGet(t *testing.T) {
	cache, _ := New[string, int](10)

	cache.Set("a", 1)
	cache.Set("b", 2)

	v, ok := cache.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, v)

	v, ok = cache.Get("b")
	assert.True(t, ok)
	assert.Equal(t, 2, v)
}

func TestGetMissing(t *testing.T) {
	cache, _ := New[string, int](10)

	_, ok := cache.Get("missing")
	assert.False(t, ok)
}

func TestSetOverwrite(t *testing.T) {
	cache, _ := New[string, int](10)

	cache.Set("a", 1)
	cache.Set("a", 2)

	v, ok := cache.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 2, v)
}

func TestEviction(t *testing.T) {
	cache, _ := New[string, int](3)

	cache.Set("a", 1)
	cache.Set("b", 2)
	cache.Set("c", 3)
	cache.Set("d", 4) // should evict "a"

	_, ok := cache.Get("a")
	assert.False(t, ok, "expected 'a' to be evicted")

	_, ok = cache.Get("b")
	assert.True(t, ok, "expected 'b' to still exist")

	_, ok = cache.Get("c")
	assert.True(t, ok, "expected 'c' to still exist")

	_, ok = cache.Get("d")
	assert.True(t, ok, "expected 'd' to still exist")
}

func TestEvictionCallback(t *testing.T) {
	var evictedKey string
	var evictedValue int

	cache, _ := NewWithEvict[string, int](2, func(k string, v int) {
		evictedKey = k
		evictedValue = v
	})

	cache.Set("a", 1)
	cache.Set("b", 2)
	cache.Set("c", 3) // should evict "a"

	assert.Equal(t, "a", evictedKey)
	assert.Equal(t, 1, evictedValue)
}

func TestGetPromotesToFront(t *testing.T) {
	cache, _ := New[string, int](3)

	cache.Set("a", 1)
	cache.Set("b", 2)
	cache.Set("c", 3)

	cache.Get("a") // promote "a" to front

	cache.Set("d", 4) // should evict "b" (oldest after "a" was promoted)

	_, ok := cache.Get("a")
	assert.True(t, ok, "expected 'a' to still exist after promotion")

	_, ok = cache.Get("b")
	assert.False(t, ok, "expected 'b' to be evicted")
}

func TestPeekDoesNotPromote(t *testing.T) {
	cache, _ := New[string, int](3)

	cache.Set("a", 1)
	cache.Set("b", 2)
	cache.Set("c", 3)

	cache.Peek("a") // peek doesn't promote

	cache.Set("d", 4) // should evict "a" (still oldest)

	_, ok := cache.Get("a")
	assert.False(t, ok, "expected 'a' to be evicted (Peek should not promote)")
}

func TestContains(t *testing.T) {
	cache, _ := New[string, int](10)

	cache.Set("a", 1)

	assert.True(t, cache.Contains("a"))
	assert.False(t, cache.Contains("b"))
}

func TestContainsEvicted(t *testing.T) {
	cache, _ := New[string, int](2)

	cache.Set("a", 1)
	cache.Set("b", 2)

	assert.True(t, cache.Contains("a"))

	cache.Set("c", 3) // evicts "a"

	assert.False(t, cache.Contains("a"), "expected evicted key to return false from Contains")
}

func TestDelete(t *testing.T) {
	cache, _ := New[string, int](10)

	cache.Set("a", 1)

	assert.True(t, cache.Delete("a"))
	assert.False(t, cache.Delete("a"))

	_, ok := cache.Get("a")
	assert.False(t, ok)
}

func TestDeleteCallsEvictionCallback(t *testing.T) {
	var evictedKey string

	cache, _ := NewWithEvict[string, int](10, func(k string, v int) {
		evictedKey = k
	})

	cache.Set("a", 1)
	cache.Delete("a")

	assert.Equal(t, "a", evictedKey)
}
