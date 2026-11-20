package ranking

import (
	"container/list"
	"time"
)

type cacheEntry struct {
	key       string
	value     float64
	expiresAt time.Time
}

type ShortTermCache struct {
	capacity int
	ttl      time.Duration
	items    map[string]*list.Element
	order    *list.List
}

func NewShortTermCache(capacity int, ttl time.Duration) *ShortTermCache {
	if capacity <= 0 {
		capacity = 256
	}
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &ShortTermCache{
		capacity: capacity,
		ttl:      ttl,
		items:    make(map[string]*list.Element, capacity),
		order:    list.New(),
	}
}

func preferenceKey(context, code, candidate string) string {
	tail := unicodeTailString(context, 4)
	return code + "\x00" + tail + "\x00" + candidate
}

func (c *ShortTermCache) RememberLocked(context, code, candidate string, value float64, now time.Time) {
	if c == nil || candidate == "" {
		return
	}
	key := preferenceKey(context, code, candidate)
	if el, ok := c.items[key]; ok {
		entry := el.Value.(*cacheEntry)
		entry.value = value
		entry.expiresAt = now.Add(c.ttl)
		c.order.MoveToFront(el)
		return
	}
	for c.order.Len() >= c.capacity {
		back := c.order.Back()
		if back == nil {
			break
		}
		old := back.Value.(*cacheEntry)
		delete(c.items, old.key)
		c.order.Remove(back)
	}
	entry := &cacheEntry{key: key, value: value, expiresAt: now.Add(c.ttl)}
	c.items[key] = c.order.PushFront(entry)
}

func (c *ShortTermCache) LookupLocked(context, code, candidate string, now time.Time) float64 {
	if c == nil {
		return 0
	}
	key := preferenceKey(context, code, candidate)
	el, ok := c.items[key]
	if !ok {
		return 0
	}
	entry := el.Value.(*cacheEntry)
	if now.After(entry.expiresAt) {
		delete(c.items, key)
		c.order.Remove(el)
		return 0
	}
	remaining := entry.expiresAt.Sub(now).Seconds() / c.ttl.Seconds()
	if remaining < 0 {
		remaining = 0
	}
	if remaining > 1 {
		remaining = 1
	}
	return entry.value * remaining
}

func (c *ShortTermCache) ClearLocked() {
	if c == nil {
		return
	}
	c.items = make(map[string]*list.Element, c.capacity)
	c.order = list.New()
}

func (c *ShortTermCache) Len() int {
	if c == nil {
		return 0
	}
	return c.order.Len()
}
