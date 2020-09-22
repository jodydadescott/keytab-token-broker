package noncestore

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

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

// NewConfig ...
func NewConfig() *Config {
	return &Config{}
}

// NewExampleConfig ...
func NewExampleConfig() *Config {
	return &Config{
		Lifetime: defaultLifetime,
	}
}

// NonceStore Holds expiring nonces
type NonceStore struct {
	cacheMap   *cachemap.CacheMap
	seededRand *rand.Rand
	lifetime   int64
}

// NewNonceStore Returns a new NonceStore
func NewNonceStore(config *Config) (*NonceStore, error) {

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

	nonceStore := &NonceStore{
		cacheMap: cachemap.NewCacheMap("nonce", cacheRefreshInterval),
		seededRand: rand.New(
			rand.NewSource(time.Now().Unix())),
		lifetime: int64(lifetime),
	}

	return nonceStore, nil
}

// Shutdown shutdowns the cache map
func (t *NonceStore) Shutdown() {
	t.cacheMap.Shutdown()
}

// NewNonce Returns a new nonce
func (t *NonceStore) NewNonce() *Nonce {

	b := make([]byte, 64)
	for i := range b {
		b[i] = charset[t.seededRand.Intn(len(charset))]
	}

	nonce := &Nonce{
		Exp:   time.Now().Unix() + t.lifetime,
		Value: string(b),
	}

	t.cacheMap.Put(nonce.Value, nonce)
	return nonce
}

// GetNonce Gets nonce by value and returns it. If the nonce is not found then
// nil and an error are returned
func (t *NonceStore) GetNonce(value string) (*Nonce, error) {

	if value == "" {
		panic("String 'value' is required")
	}

	if e, exist := t.cacheMap.Get(value); exist {
		nonce := e.(*Nonce)
		return nonce, nil
	}

	return nil, ErrNotFound
}
