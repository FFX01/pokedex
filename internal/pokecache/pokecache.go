package pokecache

import (
	"sync"
	"time"
)

type Cache struct {
	Data     map[string]cacheEntry
	mu       sync.Mutex
	interval time.Duration
}

type cacheEntry struct {
	createdAt time.Time
	value     []byte
}

func NewCache(ttl time.Duration) *Cache {
    cache := Cache{
        interval: ttl,
        Data: map[string]cacheEntry{},
    }
    go cache.reapLoop()
    return &cache
}

func (c *Cache) Add(key string, value []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := cacheEntry{
		createdAt: time.Now(),
		value:     value,
	}
	c.Data[key] = entry

	return nil
}

func (c *Cache) Get(key string) (value []byte, found bool) {
    c.mu.Lock()
    defer c.mu.Unlock()
	entry, ok := c.Data[key]
	if !ok {
		return []byte{}, false
	}
	return entry.value, true
}

func (c *Cache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.Data, key)
}

func (c *Cache) reapLoop() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		currentTime := <-ticker.C
		clearOldEntries(c, currentTime)
	}
}

func clearOldEntries(c *Cache, currentTime time.Time) {
	for key, entry := range c.Data {
		diff := currentTime.Sub(entry.createdAt)
		if diff > c.interval {
			c.Remove(key)
		}
	}
}
