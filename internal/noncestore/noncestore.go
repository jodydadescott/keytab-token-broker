package noncestore

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/jodydadescott/kerberos-bridge/internal/cachemap"
	"github.com/jodydadescott/kerberos-bridge/internal/model"
)

const (
	charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	defaultLifetime int = 60
	maxLifetime         = 86400 // 1 Day
	minLifetime         = 30

	// Interval between cache cleanup
	cacheCleanup = 60
)

// ErrNotFound Not Found
var ErrNotFound error = errors.New("Not Found")

// Config Config
type Config struct {
	Lifetime int `json:"lifetime,omitempty" yaml:"lifetime,omitempty"`
}

// MergeConfig ...
func (t *Config) MergeConfig(newConfig *Config) {
	if newConfig.Lifetime > 0 {
		t.Lifetime = newConfig.Lifetime
	}
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

	lifetime := defaultLifetime

	if config.Lifetime > 0 {
		lifetime = config.Lifetime
	}

	if lifetime > maxLifetime || lifetime < minLifetime {
		return nil, fmt.Errorf("Lifetime %d is invalid. Must be greater then %d and less then %d", lifetime, maxLifetime, minLifetime)
	}

	nonceStore := &NonceStore{
		cacheMap: cachemap.NewCacheMap("nonce", cacheCleanup),
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
func (t *NonceStore) NewNonce() *model.Nonce {

	b := make([]byte, 64)
	for i := range b {
		b[i] = charset[t.seededRand.Intn(len(charset))]
	}

	nonce := &model.Nonce{
		Exp:   time.Now().Unix() + t.lifetime,
		Value: string(b),
	}

	t.cacheMap.Put(nonce.Value, nonce)
	return nonce
}

// GetNonce Gets nonce by value and returns it. If the nonce is not found then
// nil and an error are returned
func (t *NonceStore) GetNonce(value string) (*model.Nonce, error) {

	if value == "" {
		panic("String 'value' is required")
	}

	if e, exist := t.cacheMap.Get(value); exist {
		nonce := e.(*model.Nonce)
		return nonce, nil
	}

	return nil, ErrNotFound
}
