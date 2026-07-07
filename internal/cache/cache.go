package cache

import (
	"context"
	"sync"
	"time"

	gocache "github.com/patrickmn/go-cache"
)

type Cache struct {
	c       *gocache.Cache
	flights sync.Map // key string -> *flight
}

type flight struct {
	wg  sync.WaitGroup
	val any
	err error
}

type entry struct {
	val      any
	freshTTL time.Duration
	freshAt  time.Time // when the value was placed into cache
}

func New(defaultTTL time.Duration) *Cache {
	// We use go-cache only as a TTL store (entry-level TTL passed on Set).
	return &Cache{c: gocache.New(defaultTTL, 2*defaultTTL)}
}

func (c *Cache) Get(key string) (any, bool) {
	v, ok := c.c.Get(key)
	if !ok {
		return nil, false
	}
	if e, ok := v.(*entry); ok {
		return e.val, true
	}
	return v, true
}

func (c *Cache) Set(key string, val any, fresh time.Duration) {
	if fresh <= 0 {
		c.c.SetDefault(key, &entry{val: val, freshTTL: 0, freshAt: time.Now()})
		return
	}
	c.c.Set(key, &entry{val: val, freshTTL: fresh, freshAt: time.Now()}, fresh*3)
}

func (c *Cache) GetOrFetch(ctx context.Context, key string, ttl time.Duration, fn func(ctx context.Context) (any, error)) (any, error) {
	if v, ok := c.Get(key); ok {
		// Check freshness; if stale, kick a background refresh.
		if raw, found := c.c.Get(key); found {
			if e, ok := raw.(*entry); ok && e.freshTTL > 0 && time.Since(e.freshAt) > e.freshTTL {
				go c.refresh(key, ttl, fn)
			}
		}
		return v, nil
	}
	actual, loaded := c.flights.LoadOrStore(key, &flight{})
	f := actual.(*flight)
	if !loaded {
		f.wg.Add(1)
		defer func() { f.wg.Done(); c.flights.Delete(key) }()
		v, err := fn(ctx)
		f.val, f.err = v, err
		if err == nil {
			c.Set(key, v, ttl)
		}
		return v, err
	}
	f.wg.Wait()
	return f.val, f.err
}

func (c *Cache) refresh(key string, ttl time.Duration, fn func(ctx context.Context) (any, error)) {
	if _, loaded := c.flights.LoadOrStore(key+":refresh", &flight{}); loaded {
		return // a refresh is already running
	}
	defer c.flights.Delete(key + ":refresh")
	if v, err := fn(context.Background()); err == nil {
		c.Set(key, v, ttl)
	}
}

func (c *Cache) Delete(key string) { c.c.Delete(key) }
func (c *Cache) Flush()             { c.c.Flush() }
