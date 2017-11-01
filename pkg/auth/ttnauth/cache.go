// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package ttnauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"sync"
	"time"
)

type cache struct {
	expire time.Duration
	mu     sync.RWMutex
	cache  map[string]cachedResult
	salt   []byte
}

// newCache returns a new cache and starts a cleanup goroutine.
func newCache(expire time.Duration) *cache {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		panic(err)
	}
	c := &cache{
		expire: expire,
		cache:  make(map[string]cachedResult),
		salt:   salt,
	}
	go func() {
		for {
			time.Sleep(c.expire)
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
	Access
	expires time.Time
}

func (c *cache) key(username string, password []byte) string {
	hash := sha256.Sum256(append(c.salt, password...))
	return username + "." + base64.RawStdEncoding.EncodeToString(hash[:])
}

func (c *cache) Set(username string, password []byte, access Access) {
	c.mu.Lock()
	c.cache[c.key(username, password)] = cachedResult{
		Access:  access,
		expires: time.Now().Add(c.expire),
	}
	c.mu.Unlock()
}

func (c *cache) Get(username string, password []byte) (access *Access) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := c.key(username, password)
	if cached, ok := c.cache[key]; ok {
		if cached.expires.After(time.Now()) {
			return &cached.Access
		}
		delete(c.cache, key)
	}
	return nil
}
