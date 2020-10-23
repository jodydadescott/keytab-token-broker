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

package token

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jodydadescott/tokens2secrets/internal/publickey"
	"go.uber.org/zap"
)

const (
	defaultCacheRefresh = time.Duration(30) * time.Second
)

// PublicKeyCache Interface
type PublicKeyCache interface {
	GetKey(iss, kid string) (*publickey.PublicKey, error)
}

// Config config
// IdleConnections is the idle connections for the HTTP client
// CacheRefresh is the time interval between cache refresh
// PublicKeyLifetime is the lifetime of Public Keys as they do not have a defined life
// NonceLifetime is the lifetime of a Nonce
// RequestTimeout is the request timeout for the HTTP client
type Config struct {
	CacheRefresh time.Duration
}

// Cache Parses and verifies tokens by fetching public keys from the token issuer and caching
// public keys for future use. Tokens that are verified are also stored in the cache for
// quicker validation in the future. Replay attack is also provided by a Nonce implementation.
// The nonce implementation works by generating random strings that may be fetched by a
// token bearer. The bearer should use the nonce to get a new token from their token provider
// with the audience (aud) field set to the nonce value. When then token is parsed
type Cache struct {
	tokenMapMutex       sync.RWMutex
	tokenMap            map[string]*Token
	closed              chan struct{}
	ticker              *time.Ticker
	wg                  sync.WaitGroup
	seededRand          *rand.Rand
	permitPublicKeyHTTP bool
	publicKeyCache      PublicKeyCache
}

func (t *Cache) mapGetToken(key string) *Token {
	t.tokenMapMutex.RLock()
	defer t.tokenMapMutex.RUnlock()
	return t.tokenMap[key]
}

func (t *Cache) mapPutToken(entity *Token) {
	t.tokenMapMutex.Lock()
	defer t.tokenMapMutex.Unlock()
	t.tokenMap[entity.TokenString] = entity
}

// Default returns default instance with default config
func Default(publicKeyCache PublicKeyCache) (*Cache, error) {
	c := &Config{}
	return c.Build(publicKeyCache)
}

// Build Returns a new Token Cache
func (config *Config) Build(publicKeyCache PublicKeyCache) (*Cache, error) {

	zap.L().Debug("Starting")

	cacheRefresh := defaultCacheRefresh

	if config.CacheRefresh > 0 {
		cacheRefresh = config.CacheRefresh
	}

	if publicKeyCache == nil {
		return nil, fmt.Errorf("publicKeyCache is nil")
	}

	t := &Cache{
		tokenMap:       make(map[string]*Token),
		closed:         make(chan struct{}),
		ticker:         time.NewTicker(cacheRefresh),
		wg:             sync.WaitGroup{},
		publicKeyCache: publicKeyCache,
	}

	go func() {
		t.wg.Add(1)
		for {
			select {
			case <-t.closed:
				t.wg.Done()
				return
			case <-t.ticker.C:
				zap.L().Debug("Processing cache start")
				t.processTokenCache()
				zap.L().Debug("Processing cache completed")
			}
		}
	}()

	return t, nil

}

// ParseToken ...
func (t *Cache) ParseToken(tokenString string) (*Token, error) {

	if tokenString == "" {
		zap.L().Debug("tokenString is empty")
		return nil, ErrInvalid
	}

	token := t.mapGetToken(tokenString)

	if token != nil {
		zap.L().Debug(fmt.Sprintf("Token %s found in cache", token.ShortName))

		if token.Exp > time.Now().Unix() {
			zap.L().Debug(fmt.Sprintf("Token %s is expired", token.ShortName))
			return nil, ErrExpired
		}

		return token.Copy(), nil
	}

	var err error
	token, err = ParseToken(tokenString)
	if err != nil {
		zap.L().Debug(fmt.Sprintf("Unable to parse token %s", token.ShortName))
		return nil, err
	}

	zap.L().Debug(fmt.Sprintf("Token not found in cache; token=%s", token.ShortName))

	if token.Alg == "" {
		zap.L().Debug(fmt.Sprintf("Token %s is missing required field alg", token.ShortName))
		return nil, ErrMissingField
	}

	if token.Kid == "" {
		zap.L().Debug(fmt.Sprintf("Token %s is missing required field kid", token.ShortName))
		return nil, ErrMissingField
	}

	if token.Typ == "" {
		zap.L().Debug(fmt.Sprintf("Token %s is missing required field typ", token.ShortName))
		return nil, ErrMissingField
	}

	if token.Iss == "" {
		zap.L().Debug(fmt.Sprintf("Token %s is missing required field iss", token.ShortName))
		return nil, ErrMissingField
	}

	if !strings.HasPrefix(token.Iss, "http") {
		zap.L().Debug(fmt.Sprintf("Token %s has field iss but value %s is not expected", token.ShortName, token.Iss))
		return nil, ErrMissingField
	}

	if !t.permitPublicKeyHTTP {
		if !strings.HasPrefix(token.Iss, "https") {
			zap.L().Debug(fmt.Sprintf("Token %s has field iss but value %s is not permitted as https is required", token.ShortName, token.Iss))
			return nil, ErrMissingField
		}
	}

	_, err = jwt.Parse(token.TokenString, func(jwtToken *jwt.Token) (interface{}, error) {

		// SigningMethodRSA

		publicKey, err := t.publicKeyCache.GetKey(token.Iss, token.Kid)
		if err != nil {
			return nil, err
		}

		if publicKey.Kty == "" {
			return nil, fmt.Errorf("kty is empty. should be EC or RSA")
		}

		switch jwtToken.Method.(type) {
		case *jwt.SigningMethodECDSA:
			if publicKey.Kty != "EC" {
				return nil, fmt.Errorf("Expected value for kty is EC not %s", publicKey.Kty)
			}
			break
		case *jwt.SigningMethodRSA:
			if publicKey.Kty != "RSA" {
				return nil, fmt.Errorf("Expected value for kty is RSA not %s", publicKey.Kty)
			}
			break
		default:
			return nil, fmt.Errorf("Signing type %s unsupported", publicKey.Kty)
		}

		return publicKey.EcdsaPublicKey, nil

	})

	if err != nil {
		if err.Error() == "Token is expired" {
			// We will be the judge of that
		} else if err.Error() == "Token used before issued" {
			// Slight drift in clock. We will be the judge of that
		} else {
			zap.L().Debug(fmt.Sprintf("Unable to verify signature for token %s; error=%s", token.ShortName, err.Error()))
			return nil, ErrSignatureInvalid
		}
	}

	if token.Exp > time.Now().Unix() {
		zap.L().Debug(fmt.Sprintf("Token %s is expired", token.ShortName))
		return nil, ErrExpired
	}

	t.mapPutToken(token)
	zap.L().Debug(fmt.Sprintf("Token %s added to cache", token.ShortName))
	return token.Copy(), nil
}

func (t *Cache) processTokenCache() {

	zap.L().Debug("Processing Token cache")

	var removes []string
	t.tokenMapMutex.Lock()
	defer t.tokenMapMutex.Unlock()

	for key, e := range t.tokenMap {

		if time.Now().Unix() > e.Exp {
			removes = append(removes, key)
			zap.L().Info(fmt.Sprintf("Ejecting->%s", e.JSON()))
		} else {
			zap.L().Debug(fmt.Sprintf("Preserving->%s", e.JSON()))
		}
	}

	if len(removes) > 0 {
		for _, key := range removes {
			delete(t.tokenMap, key)
		}
	}

	zap.L().Debug("Processing Token cache completed")

}

// Shutdown Cache
func (t *Cache) Shutdown() {
	zap.L().Debug("Stopping")
	close(t.closed)
	t.wg.Wait()
}
