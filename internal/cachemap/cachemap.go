/*
Copyright © 2020 Jody Scott <jody@thescottsweb.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cachemap

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Entity ...
type Entity interface {
	Expiration() int64
	Key() string
	JSON() string
}

// Process ...
// type Process func(Entity)

// Config ...
type Config struct {
	CacheRefreshInterval int
	Name                 string
	//Process              Process
}

// CacheMap ...
type CacheMap struct {
	name     string
	mutex    sync.RWMutex
	internal map[string]Entity
	closed   chan struct{}
	ticker   *time.Ticker
	wg       sync.WaitGroup
}

// Build ...
func (c *Config) Build() (*CacheMap, error) {

	if c.Name == "" {
		return nil, fmt.Errorf("name is empty")
	}

	if c.CacheRefreshInterval <= 0 {
		return nil, fmt.Errorf("cacheRefreshInterval must be greater then zero")
	}

	t := &CacheMap{
		name:     c.Name,
		internal: make(map[string]Entity),
		closed:   make(chan struct{}),
		ticker:   time.NewTicker(time.Duration(c.CacheRefreshInterval) * time.Second),
		wg:       sync.WaitGroup{},
	}

	go func() {
		t.wg.Add(1)
		for {
			select {
			case <-t.closed:
				zap.L().Debug(fmt.Sprintf("Shutting down token cache %s", t.name))
				t.wg.Done()
				return
			case <-t.ticker.C:
				t.processCache()

			}
		}
	}()

	return t, nil
}

// Shutdown ...
func (t *CacheMap) Shutdown() {
	close(t.closed)
	t.wg.Wait()
}

// Put ...
func (t *CacheMap) Put(e Entity) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.internal[e.Key()] = e
	zap.L().Info(fmt.Sprintf("Cache %s: added to cache->%s", t.name, e.JSON()))
}

// Get ...
func (t *CacheMap) Get(key string) Entity {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	w, _ := t.internal[key]
	return w
}

func (t *CacheMap) processCache() {

	zap.L().Debug(fmt.Sprintf("Processing %s cache", t.name))

	var removes []string
	t.mutex.Lock()
	defer t.mutex.Unlock()

	for key, e := range t.internal {

		if time.Now().Unix() > e.Expiration() {
			removes = append(removes, key)
			zap.L().Info(fmt.Sprintf("Cache %s: ejecting->%s", t.name, e.JSON()))
		} else {
			zap.L().Debug(fmt.Sprintf("Cache %s: preserving->%s", t.name, e.JSON()))
		}
	}

	if len(removes) > 0 {
		for _, key := range removes {
			delete(t.internal, key)
		}
	}

	zap.L().Debug(fmt.Sprintf("Completed cleanup on token cache %s", t.name))

}
