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

package keytabs

import (
	"fmt"
	"regexp"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	principalRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

const (
	defaultCacheRefreshInterval int = 30

	defaultSoftLifetime = 120
	minSoftLifetime     = 30
	maxSoftLifetime     = 86400

	defaultHardLifetime = 600
	minHardLifetime     = 30
	maxHardLifetime     = 86400
)

// Config Configuration
type Config struct {
	CacheRefreshInterval, SoftLifetime, HardLifetime int
	Principals                                       []string
}

// KeytabCache holds and manages Kerberos Keytabs. Upon start all Keytabs will be
// created new and stored in the cache. When a request is made for a keytab its
// soft expiration and hard expiration field will be set. Once the soft
// expiration time is exceeded a new Keytab will be created hence invalidating
// the previous Keytab. In the situation where a Keytab is requested when the
// soft expiration time has already been set the expiration timer will be
// increased. This has the side effect of extending the lifetime of the already
// granted Keytab. To prevent the situation where continous retireval of the
// keytab could result in the Keytab never expiring once the hard expiration
// is exceeded the Keytab will be regenerated. It is possible that a client
// having obtained a keytab towards the end of the hard expiration time may not
// have enough time to make use of the Keytab before its expiration. It is left
// to the client to handle this rare situation by requesting a new Keytab.
type KeytabCache struct {
	internal                   map[string]*wrapper
	softLifetime, hardLifetime int64
	closed                     chan struct{}
	wg                         sync.WaitGroup
	ticker                     *time.Ticker
	mutex                      sync.RWMutex
}

// Internal struct to hold data in map
type wrapper struct {
	principal string
	keytab    *Keytab
	mutex     sync.Mutex
}

// Build Returns a new Keytab Cache
func (config *Config) Build() (*KeytabCache, error) {

	zap.L().Debug("Starting Keytab Cache")

	if config.Principals == nil {
		return nil, fmt.Errorf("principals is nil")
	}

	cacheRefreshInterval := defaultCacheRefreshInterval
	softLifetime := defaultSoftLifetime
	hardLifetime := defaultHardLifetime

	if config.CacheRefreshInterval > 0 {
		cacheRefreshInterval = config.CacheRefreshInterval
	}

	if config.SoftLifetime > 0 {
		softLifetime = config.SoftLifetime
	}

	if config.HardLifetime > 0 {
		hardLifetime = config.HardLifetime
	}

	if softLifetime > maxSoftLifetime || softLifetime < minSoftLifetime {
		return nil, fmt.Errorf(fmt.Sprintf("%s must be greater then %d and less then %d", "Lifetime", minSoftLifetime, maxSoftLifetime))
	}

	if hardLifetime > maxHardLifetime || hardLifetime < minHardLifetime {
		return nil, fmt.Errorf(fmt.Sprintf("%s must be greater then %d and less then %d", "HardLifetime", minHardLifetime, maxHardLifetime))
	}

	if softLifetime > hardLifetime {
		return nil, fmt.Errorf("HardLifetime must be greater then (soft) Lifetime")
	}

	t := &KeytabCache{
		internal:     make(map[string]*wrapper),
		softLifetime: int64(softLifetime),
		hardLifetime: int64(hardLifetime),
		closed:       make(chan struct{}),
		ticker:       time.NewTicker(time.Duration(cacheRefreshInterval) * time.Second),
	}

	if len(config.Principals) <= 0 {
		zap.L().Warn("principals is empty. This is probably a mistake")
	} else {

		zap.L().Info("Loading principals")
		t.mutex.Lock()
		defer t.mutex.Unlock()

		for _, principal := range config.Principals {

			if len(principal) < 3 && len(principal) > 254 {
				if len(principal) < 3 {
					return nil, fmt.Errorf("Principal %s is to short", principal)
				}
				return nil, fmt.Errorf("Principal %s is to long", principal)
			}

			if !principalRegex.MatchString(principal) {
				return nil, fmt.Errorf("Principal %s is invalid", principal)
			}

			t.internal[principal] = &wrapper{
				principal: principal,
			}

			zap.L().Debug(fmt.Sprintf("Loaded principal %s", principal))

		}
	}

	go func() {

		// Initial
		t.cacheRefresh()

		for {
			select {
			case <-t.closed:
				zap.L().Debug("Shutting down Keytab Cache")
				return
			case <-t.ticker.C:
				t.cacheRefresh()
			}
		}
	}()

	return t, nil
}

// GetKeytab returns keytab If wrapper does not exist then principal does not exist
// If the wrapper does exist then we check if it has a valid
// keytab and if it does we return it. If it does not then we
// generate a new keytab and return it. We set the flag dirty
// to true so that we know someone has the keytab
func (t *KeytabCache) GetKeytab(principal string) *Keytab {

	if principal == "" {
		zap.L().Debug(fmt.Sprintf("Principal is empty"))
		return nil
	}

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if wrapper, ok := t.internal[principal]; ok {

		wrapper.mutex.Lock()
		defer wrapper.mutex.Unlock()

		if wrapper.keytab == nil {
			// It must be a local system problem. Probably authorization.
			return nil
		}

		if wrapper.keytab.SoftExp == 0 {
			// In this situation the Keytab is clean
			// Set the expiration time and the hard time
			now := time.Now().Unix()
			wrapper.keytab.SoftExp = now + t.softLifetime
			wrapper.keytab.HardExp = now + t.hardLifetime
			zap.L().Debug(fmt.Sprintf("Keytab with principal %s changed from clean to dirty; hard and soft expiration set", wrapper.principal))
			// Func is exported. Return clone to untrusted outsiders
			return wrapper.keytab.Clone()
		}

		// In this situation the keytab is dirty. The (soft) expiration is increased
		// but the hard expiration is not modified

		wrapper.keytab.SoftExp = time.Now().Unix() + t.softLifetime
		zap.L().Debug(fmt.Sprintf("Keytab with principal %s is already dirty; soft expiration increased", wrapper.principal))

		// Func is exported. Return clone to untrusted outsiders
		return wrapper.keytab.Clone()
	}

	return nil
}

func (t *KeytabCache) cacheRefresh() {

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	// Process all keytabs with time now
	now := time.Now().Unix()

	for _, wrapper := range t.internal {
		wrapper.mutex.Lock()
		defer wrapper.mutex.Unlock()

		if wrapper.keytab == nil {
			zap.L().Debug(fmt.Sprintf("Keytab with principal %s does not have a keytab; creating new", wrapper.principal))
			t.wrapperRefresh(wrapper)
			continue
		}

		if wrapper.keytab.SoftExp > 0 {

			if wrapper.keytab.HardExp <= 0 {
				zap.L().Error(fmt.Sprintf("Keytab with principal %s has softExp of %d and hardExp of %d. This is not expected", wrapper.principal, wrapper.keytab.SoftExp, wrapper.keytab.HardExp))
				panic("This should not happen")
			}

			if now > wrapper.keytab.SoftExp {
				zap.L().Info(fmt.Sprintf("Keytab with principal %s is dirty and soft expiration exceeded; renewing", wrapper.principal))
				t.wrapperRefresh(wrapper)
				continue
			}

			if now > wrapper.keytab.HardExp {
				zap.L().Info(fmt.Sprintf("Keytab with principal %s is dirty and hard expiration exceeded; renewing", wrapper.principal))
				t.wrapperRefresh(wrapper)
				continue
			}

			zap.L().Info(fmt.Sprintf("Keytab with principal %s is dirty and expiration (hard & soft) not exceeded; nothing to do", wrapper.principal))
			continue

		}

		zap.L().Debug(fmt.Sprintf("Keytab with principal %s is clean; nothing to do", wrapper.principal))

	}

}

func (t *KeytabCache) wrapperRefresh(wrapper *wrapper) {

	var err error
	base64File := ""
	// Principal:  "HTTP/" + principal,

	if runtime.GOOS == "windows" {

		base64File, err = windowsNewKeytab(wrapper.principal)
		if err != nil {
			zap.L().Error(fmt.Sprintf("Unable to create Keytab for principal %s, err:%s ", wrapper.principal, err))
			return
		}

	} else {

		base64File, err = unixNewKeytab(wrapper.principal)
		if err != nil {
			zap.L().Error(fmt.Sprintf("Unable to create Keytab for principal %s, err:%s ", wrapper.principal, err))
			return
		}

	}

	wrapper.keytab = &Keytab{
		Principal:  "HTTP/" + wrapper.principal,
		Base64File: base64File,
	}

	zap.L().Debug(fmt.Sprintf("New Keytab created for principal %s", wrapper.principal))

}

// Shutdown Cache
func (t *KeytabCache) Shutdown() {
	close(t.closed)
	t.wg.Wait()
}
