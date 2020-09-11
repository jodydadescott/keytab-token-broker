package noncestore

import (
	"fmt"
	"kbridge/internal/cachemap"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

const (
	charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	defaultNonceCleanup int   = 30
	defaultNonceLife    int64 = 10
)

// Config ...
type Config struct {
	NonceLife    int64
	NonceCleanup int
}

// NonceStore ...
type NonceStore struct {
	cacheMap   *cachemap.CacheMap
	seededRand *rand.Rand
	nonceLife  int64
}

// NewConfig ...
func NewConfig() *Config {
	return &Config{}
}

// Build ...
func (c *Config) Build() *NonceStore {

	nonceCleanup := defaultNonceCleanup
	nonceLife := defaultNonceLife

	if c.NonceCleanup > 0 {
		nonceCleanup = c.NonceCleanup
	}

	if c.NonceLife > 0 {
		nonceLife = c.NonceLife
	}

	nonceStore := &NonceStore{
		cacheMap: cachemap.NewCacheMap(nonceCleanup),
		seededRand: rand.New(
			rand.NewSource(time.Now().Unix())),
		nonceLife: nonceLife,
	}

	return nonceStore
}

// Shutdown ...
func (t *NonceStore) Shutdown() {
	t.cacheMap.Shutdown()
}

// NewNonce ...
func (t *NonceStore) NewNonce() *Nonce {

	b := make([]byte, 64)
	for i := range b {
		b[i] = charset[t.seededRand.Intn(len(charset))]
	}

	nonce := &Nonce{
		Exp:   time.Now().Unix() + t.nonceLife,
		Value: string(b),
	}

	zap.L().Debug(fmt.Sprintf("NewNonce()->%s", nonce.Value))
	t.cacheMap.Put(nonce.Value, nonce)
	return nonce
}

// GetNonce ...
func (t *NonceStore) GetNonce(value string) (*Nonce, error) {

	if value == "" {
		panic("String 'value' is required")
	}

	if e, exist := t.cacheMap.Get(value); exist {
		nonce := e.(*Nonce)
		zap.L().Debug(fmt.Sprintf("GetNonce(%s)->[Found]", value))
		return nonce, nil
	}

	zap.L().Debug(fmt.Sprintf("GetNonce(%s)->[Not Found]", value))
	return nil, ErrNotFound
}
