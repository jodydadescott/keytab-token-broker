package cachemap

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// CacheMap ...
type CacheMap struct {
	mutex    sync.RWMutex
	internal map[string]Entity
	closed   chan struct{}
	wg       sync.WaitGroup
	ticker   *time.Ticker
}

// NewCacheMap returns new cachemap. CacheMap holds structs that implement the Entity
// interface. Perodically (cleanupIntervalSeconds) all entities will be polled by calling
// Valid(). If the entity returns true it will be left. If it returns false it will be
// removed.
func NewCacheMap(cleanupIntervalSeconds int) *CacheMap {

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
				zap.L().Debug("Shutting down")
				return
			case <-cacheMap.ticker.C:
				zap.L().Debug("Cleanup-> running")
				cacheMap.cleanup()
				zap.L().Debug("Cleanup-> completed")
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
func (t *CacheMap) Put(key string, e Entity) {

	if e == nil {
		panic("Entity must not be nil")
	}

	if key == "" {
		panic("Key must not be empty")
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.internal[key] = e

	zap.L().Info(fmt.Sprintf("Adding entity %s", e.JSON()))
}

// Get Returns the entity if found and true or nil and false if entity is not found
func (t *CacheMap) Get(key string) (Entity, bool) {

	if key == "" {
		panic("key must not be empty")
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
			zap.L().Info(fmt.Sprintf("Ejecting entity %s", e.JSON()))
		}
	}

	if len(removes) > 0 {
		for _, key := range removes {
			delete(t.internal, key)
		}
	}

}
