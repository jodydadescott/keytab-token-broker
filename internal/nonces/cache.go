package nonces

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/jinzhu/copier"
	"github.com/jodydadescott/kerberos-bridge/internal/cachemap"
)

const (
	charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	defaultCacheRefreshInterval int = 30
	minCacheRefreshInterval         = 15
	maxCacheRefreshInterval         = 3600

	defaultLifetime int = 60
	minLifetime         = 30
	maxLifetime         = 86400 // 1 Day
)

// ErrNotFound Not Found
var ErrNotFound error = errors.New("Not Found")

// Config Config
type Config struct {
	CacheRefreshInterval int `json:"cacheRefreshInterval,omitempty" yaml:"cacheRefreshInterval,omitempty"`
	Lifetime             int `json:"lifetime,omitempty" yaml:"lifetime,omitempty"`
}

// Cache Holds expiring nonces
type Cache struct {
	cacheMap   *cachemap.CacheMap
	seededRand *rand.Rand
	lifetime   int64
}

// Build Returns a new Cache
func (config *Config) Build() (*Cache, error) {

	cacheRefreshInterval := defaultCacheRefreshInterval
	lifetime := defaultLifetime

	if config.CacheRefreshInterval > 0 {
		cacheRefreshInterval = config.CacheRefreshInterval
	}

	if config.Lifetime > 0 {
		lifetime = config.Lifetime
	}

	if cacheRefreshInterval < minCacheRefreshInterval || cacheRefreshInterval > maxCacheRefreshInterval {
		return nil, fmt.Errorf(fmt.Sprintf("%s must be greater then %d and less then %d", "CacheRefreshInterval", minCacheRefreshInterval, maxCacheRefreshInterval))
	}

	if lifetime > maxLifetime || lifetime < minLifetime {
		return nil, fmt.Errorf(fmt.Sprintf("%s must be greater then %d and less then %d", "Lifetime", minLifetime, maxLifetime))
	}

	return &Cache{
		cacheMap: cachemap.NewCacheMap("nonce", cacheRefreshInterval),
		seededRand: rand.New(
			rand.NewSource(time.Now().Unix())),
		lifetime: int64(lifetime),
	}, nil

}

// Shutdown shutdowns the cache map
func (t *Cache) Shutdown() {
	t.cacheMap.Shutdown()
}

// NewNonce Returns a new nonce
func (t *Cache) NewNonce() *Nonce {

	b := make([]byte, 64)
	for i := range b {
		b[i] = charset[t.seededRand.Intn(len(charset))]
	}

	nonce := &Nonce{
		Exp:   time.Now().Unix() + t.lifetime,
		Value: string(b),
	}

	t.cacheMap.Put(nonce.Value, nonce)

	// Func is exported. Return clone to untrusted outsiders
	clone := &Nonce{}
	err := copier.Copy(&clone, &nonce)
	if err != nil {
		panic(err)
	}

	return clone
}

// GetNonce Gets nonce by value and returns it. If the nonce is not found then
// nil and an error are returned
func (t *Cache) GetNonce(value string) (*Nonce, error) {

	if value == "" {
		panic("String 'value' is required")
	}

	if e, exist := t.cacheMap.Get(value); exist {
		nonce := e.(*Nonce)

		// Func is exported. Return clone to untrusted outsiders
		clone := &Nonce{}
		err := copier.Copy(&clone, &nonce)
		if err != nil {
			panic(err)
		}

		return clone, nil
	}

	return nil, ErrNotFound
}
