package cachemap

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// CacheMap ...
type CacheMap struct {
	name     string
	mutex    sync.RWMutex
	internal map[string]Entity
	closed   chan struct{}
	wg       sync.WaitGroup
	ticker   *time.Ticker
}

// Entity Interface type held in cache. Entity must implement Valid() so that cache
// map knows when to eject entity. JSON() is used for logging transactions.
type Entity interface {
	Valid() bool
	JSON() string
}

// NewCacheMap returns new cachemap. CacheMap holds structs that implement the Entity
// interface. Perodically (cleanupIntervalSeconds) all entities will be polled by calling
// Valid(). If the entity returns true it will be left. If it returns false it will be
// removed.
func NewCacheMap(name string, cleanupIntervalSeconds int) *CacheMap {

	zap.L().Debug("Starting")

	if cleanupIntervalSeconds < 10 {
		panic("cleanupIntervalSeconds must be 10 or greater")
	}

	cacheMap := &CacheMap{
		internal: make(map[string]Entity),
		closed:   make(chan struct{}),
		ticker:   time.NewTicker(time.Duration(cleanupIntervalSeconds) * time.Second),
	}

	go func() {
		for {
			select {
			case <-cacheMap.closed:
				zap.L().Debug(fmt.Sprintf("Shutting down cache %s", cacheMap.name))
				return
			case <-cacheMap.ticker.C:
				zap.L().Debug(fmt.Sprintf("Running cleanup on cache %s", cacheMap.name))
				cacheMap.cleanup()
				zap.L().Debug(fmt.Sprintf("Completed cleanup on cache %s", cacheMap.name))
			}
		}
	}()

	return cacheMap
}

// Shutdown ...
func (t *CacheMap) Shutdown() {
	close(t.closed)
	t.wg.Wait()
}

// Put Puts entity into map. Entity must have a non-empty key.
func (t *CacheMap) Put(key string, e Entity) error {

	if e == nil {
		return fmt.Errorf("entity is nil")
	}

	if key == "" {
		return fmt.Errorf("key is nil")
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.internal[key] = e

	zap.L().Info(fmt.Sprintf("Adding entity %s to cache %s", e.JSON(), t.name))
	return nil
}

// Get Returns the entity if found and true or nil and false if entity is not found
func (t *CacheMap) Get(key string) (Entity, bool) {

	if key == "" {
		zap.L().Debug(fmt.Sprintf("Key is empty"))
		return nil, false
	}

	t.mutex.RLock()
	defer t.mutex.RUnlock()
	if e, exist := t.internal[key]; exist {
		return e, true
	}
	return nil, false
}

func (t *CacheMap) cleanup() {

	var removes []string
	t.mutex.Lock()
	defer t.mutex.Unlock()

	for key, e := range t.internal {
		if e.Valid() {
			// zap.L().Debug(fmt.Sprintf("Preserving entity %s", e.JSON()))
		} else {
			removes = append(removes, key)
			zap.L().Info(fmt.Sprintf("Ejecting entity %s from cache %s", e.JSON(), t.name))
		}
	}

	if len(removes) > 0 {
		for _, key := range removes {
			delete(t.internal, key)
		}
	}

}
