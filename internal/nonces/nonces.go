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

package nonces

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/jodydadescott/keytab-token-broker/internal/cachemap"
	"go.uber.org/zap"
)

const (
	charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	defaultCacheRefreshInterval int = 30
	minCacheRefreshInterval         = 15
	maxCacheRefreshInterval         = 3600

	defaultLifetime int = 60
)

// Config Config
type Config struct {
	CacheRefreshInterval, Lifetime int
}

// NonceCache Manages nonces. For our purposes a nonce is defined as a random
// string with an expiration time. Upon request a new nonce is generated
// and returned along with the expiration time to the caller. This allows
// the caller to hand the nonce to a remote party. The remote party can then
// present the nonce back in the future (before the expiration time is reached)
// and the nonce can be validated that it originated with us.
type NonceCache struct {
	cacheMap   *cachemap.CacheMap
	seededRand *rand.Rand
	lifetime   int64
}

// Build Returns a new Cache
func (config *Config) Build() (*NonceCache, error) {

	zap.L().Debug("Starting Nonce Cache")

	cacheRefreshInterval := defaultCacheRefreshInterval
	lifetime := defaultLifetime

	if config.CacheRefreshInterval > 0 {
		cacheRefreshInterval = config.CacheRefreshInterval
	}

	if config.Lifetime > 0 {
		lifetime = config.Lifetime
	}

	cacheMapConfig := &cachemap.Config{
		CacheRefreshInterval: cacheRefreshInterval,
		Name:                 "nonce",
	}

	cacheMap, err := cacheMapConfig.Build()
	if err != nil {
		return nil, err
	}

	return &NonceCache{
		cacheMap: cacheMap,
		lifetime: int64(lifetime),
		seededRand: rand.New(
			rand.NewSource(time.Now().Unix())),
	}, nil

}

// Shutdown shutdowns the cache map
func (t *NonceCache) Shutdown() {
	t.cacheMap.Shutdown()
}

// NewNonce Returns a new nonce
func (t *NonceCache) NewNonce() *Nonce {

	b := make([]byte, 64)
	for i := range b {
		b[i] = charset[t.seededRand.Intn(len(charset))]
	}

	nonce := &Nonce{
		Exp:   time.Now().Unix() + t.lifetime,
		Value: string(b),
	}

	t.cacheMap.Put(nonce)

	// Func is exported. Return clone to untrusted outsiders
	return nonce.Clone()
}

// GetNonce returns nonce if found and not expired
func (t *NonceCache) GetNonce(value string) *Nonce {

	if value == "" {
		zap.L().Warn("request for empty nonce")
		return nil
	}

	e := t.cacheMap.Get(value)
	if e == nil {
		zap.L().Debug(fmt.Sprintf("Nonce not found; nonce value:%s", value))
		return nil
	}

	nonce := e.(*Nonce)

	if time.Now().Unix() > nonce.Expiration() {
		zap.L().Debug(fmt.Sprintf("Nonce expired; nonce value:%s", value))
		return nil
	}
	// Func is exported. Return clone to untrusted outsiders

	zap.L().Debug(fmt.Sprintf("Nonce found and not expired; nonce value:%s", value))
	return nonce.Clone()
}
