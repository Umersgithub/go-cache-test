package main

import (
	"container/list"
	"fmt"
	"sync"
)

// Cache holds our LRU cache data
type Cache struct {
	max_size  int
	curr_size int
	cache     map[string]*list.Element
	lru       *list.List
	lock      sync.Mutex
}

type cacheItem struct {
	key string
	val interface{}
}

// NewCache creates a new cache with given size
func NewCache(size int) *Cache {
	return &Cache{
		max_size:  size,
		curr_size: 0,
		cache:     make(map[string]*list.Element),
		lru:       list.New(),
	}
}

// Set adds or updates an item in the cache
func (c *Cache) Set(key string, value interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Check if key already exists
	if ele, ok := c.cache[key]; ok {
		c.lru.MoveToFront(ele)
		ele.Value.(*cacheItem).val = value
		return
	}

	// If we're at capacity, remove the LRU item
	if c.curr_size >= c.max_size {
		c.removeLRU()
	}

	// Add new item
	ele := c.lru.PushFront(&cacheItem{key, value})
	c.cache[key] = ele
	c.curr_size++
}

// Get retrieves an item from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if ele, ok := c.cache[key]; ok {
		c.lru.MoveToFront(ele)
		return ele.Value.(*cacheItem).val, true
	}
	return nil, false
}

// removeLRU removes the least recently used item
func (c *Cache) removeLRU() {
	if c.curr_size == 0 {
		return // shouldn't happen, but just in case
	}

	ele := c.lru.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

// removeElement is a helper to remove an element from the cache
func (c *Cache) removeElement(e *list.Element) {
	c.lru.Remove(e)
	kv := e.Value.(*cacheItem)
	delete(c.cache, kv.key)
	c.curr_size--
}

// Len returns the number of items in the cache
func (c *Cache) Len() int {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.curr_size
}

// Clear empties the cache
func (c *Cache) Clear() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.curr_size = 0
	c.cache = make(map[string]*list.Element)
	c.lru.Init()
}

func main() {
	// Create a small cache
	c := NewCache(3)

	// Add some items
	c.Set("name", "John")
	c.Set("age", 30)
	c.Set("city", "New York")

	// Try to get an item
	if val, ok := c.Get("age"); ok {
		fmt.Printf("Age: %v\n", val)
	} else {
		fmt.Println("Age not found")
	}

	// Add one more item, which should evict the LRU item
	c.Set("country", "USA")

	// Try to get the evicted item
	if _, ok := c.Get("name"); !ok {
		fmt.Println("Name was evicted")
	}

	fmt.Printf("Cache size: %d\n", c.Len())

	// Clear the cache
	c.Clear()
	fmt.Printf("Cache size after clear: %d\n", c.Len())
}
