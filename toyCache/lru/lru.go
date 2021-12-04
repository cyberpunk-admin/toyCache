package lru

import (
	"container/list"
)

// Cache is an LRU cache, it is not safe for concurrent access
type Cache struct {
	maxBytes int64
	nBytes   int64
	ll       *list.List
	cache    map[string]*list.Element
	// optional and executed when an entry is purged
	OnEvicted func(key string, value Value)
}

type entry struct {
	key   string
	value Value
}

// Value used Len() to get how many bytes is takes
type Value interface {
	Len() int
}

// New is the constructor of Cache
func New(maxBytes int64, onEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Add adds a value to the cache
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		kv := ele.Value.(*entry)
		c.nBytes += int64(value.Len()) - int64(kv.value.Len())
		c.ll.MoveToFront(ele)
		kv.value = value
	} else {
		c.nBytes += int64(value.Len()) + int64(len(key))
		ele := c.ll.PushFront(&entry{key: key, value: value})
		c.cache[key] = ele
	}
	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}

// Get look ups a key's value
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// RemoveOldest remove oldest item
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Len return the number of cache entries
func (c *Cache) Len() int {
	return c.ll.Len()
}
