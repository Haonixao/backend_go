package simple_cacher

import (
	"sync"
	"time"
)

type CacheEl[E any] struct {
	el        *E
	expiresAt time.Time
}

type Cacher[K comparable, E any] struct {
	mu    sync.RWMutex
	cache map[K]*CacheEl[E]
	// 0 means no limit
	maxSize int
}

func NewCacher[K comparable, E any](maxSize int) *Cacher[K, E] {
	return &Cacher[K, E]{
		cache:   make(map[K]*CacheEl[E]),
		maxSize: maxSize,
	}
}

func (c *Cacher[K, E]) SetEl(key K, el *E, expiresAt time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.maxSize > 0 && len(c.cache) > c.maxSize {
		c.cache = make(map[K]*CacheEl[E])
	}
	cacheEl := &CacheEl[E]{
		el:        el,
		expiresAt: expiresAt,
	}
	c.cache[key] = cacheEl
}

func (c *Cacher[K, E]) GetEl(key K) *E {
	c.mu.RLock()
	cacheEl, exists := c.cache[key]
	if !exists {
		c.mu.RUnlock()
		return nil
	}

	if cacheEl.expiresAt.After(time.Now()) {
		c.mu.RUnlock()
		return cacheEl.el
	}

	c.mu.RUnlock()
	c.mu.Lock()
	defer c.mu.Unlock()

	if cacheEl, exists = c.cache[key]; exists {
		if cacheEl.expiresAt.After(time.Now()) {
			return cacheEl.el
		}
		delete(c.cache, key)
	}

	return nil
}

func (c *Cacher[K, E]) DelEl(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, key)
}

