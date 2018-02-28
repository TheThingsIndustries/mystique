// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package ttnauth

import (
	"encoding/base64"
	"sync"
	"time"
)

type Cache interface {
	GetOrFetch(username string, password []byte, fetch func(username string, password []byte) (*Access, error)) (*Access, error)
}

type cache struct {
	expires time.Duration
	mu      sync.Mutex
	cache   map[string]*cachedResult
}

// newCache returns a new cache and starts a cleanup goroutine.
func newCache(expires time.Duration) *cache {
	c := &cache{
		expires: expires,
		cache:   make(map[string]*cachedResult),
	}
	go func() {
		for {
			time.Sleep(c.expires)
			now := time.Now()
			c.mu.Lock()
			for key, cached := range c.cache {
				if cached.expires.Before(now) {
					delete(c.cache, key) // yes, you can delete from within a range
				}
			}
			c.mu.Unlock()
		}
	}()
	return c
}

type cachedResult struct {
	*Access
	err     error
	expires time.Time
	wg      sync.WaitGroup
}

func (c *cache) key(username string, password []byte) string {
	return username + "." + base64.RawStdEncoding.EncodeToString(password)
}

func (c *cache) GetOrFetch(username string, password []byte, fetch func(username string, password []byte) (*Access, error)) (*Access, error) {
	key := c.key(username, password)
	c.mu.Lock()
	cached, ok := c.cache[key]
	if !ok {
		cached = &cachedResult{}
		c.cache[key] = cached
	}
	if cached.expires.Before(time.Now()) {
		cached.expires = time.Now().Add(c.expires)
		cached.wg.Add(1)
		go func() {
			cached.Access, cached.err = fetch(username, password)
			cached.wg.Done()
		}()
	}
	c.mu.Unlock()
	cached.wg.Wait()
	return cached.Access, cached.err
}
