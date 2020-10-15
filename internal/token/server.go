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
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jodydadescott/tokens2keytabs/internal/publickey"
	"go.uber.org/zap"
)

const (
	maxCacheRefreshInterval     = 3600
	defaultCacheRefreshInterval = time.Duration(30) * time.Second
)

var (
	// ErrTokenExpired Token not found
	ErrTokenExpired error = errors.New("Token expired")
	// ErrTokenInvalid Token not valid
	ErrTokenInvalid error = errors.New("Token invalid")
)

// Config The config
type Config struct {
	CacheRefreshInterval time.Duration
}

// Server ...
type Server struct {
	mutex      sync.RWMutex
	internal   map[string]*Token
	closed     chan struct{}
	ticker     *time.Ticker
	wg         sync.WaitGroup
	publicKeys *publickey.Server
}

// Build Returns a new Token Cache
func (config *Config) Build() (*Server, error) {

	zap.L().Debug("Starting Token Cache")

	cacheRefreshInterval := defaultCacheRefreshInterval

	if config.CacheRefreshInterval > 0 {
		cacheRefreshInterval = config.CacheRefreshInterval
	}

	publicKeysConfig := &publickey.Config{}

	publicKeys, err := publicKeysConfig.Build()
	if err != nil {
		return nil, err
	}

	t := &Server{
		internal:   make(map[string]*Token),
		closed:     make(chan struct{}),
		ticker:     time.NewTicker(cacheRefreshInterval),
		wg:         sync.WaitGroup{},
		publicKeys: publicKeys,
	}

	go func() {
		t.wg.Add(1)
		for {
			select {
			case <-t.closed:
				zap.L().Debug("PublicKeys Stopping")
				t.wg.Done()
				return
			case <-t.ticker.C:
				t.processCache()

			}
		}
	}()

	return t, nil

}

func (t *Server) processCache() {

	zap.L().Debug("Processing cache start")

	var removes []string
	t.mutex.Lock()
	defer t.mutex.Unlock()

	for key, e := range t.internal {

		if time.Now().Unix() > e.Exp {
			removes = append(removes, key)
			zap.L().Info(fmt.Sprintf("Ejecting->%s", e.JSON()))
		} else {
			zap.L().Debug(fmt.Sprintf("Preserving->%s", e.JSON()))
		}
	}

	if len(removes) > 0 {
		for _, key := range removes {
			delete(t.internal, key)
		}
	}

	zap.L().Debug("Processing cache completed")

}

// GetToken ...
func (t *Server) GetToken(token string) (*Token, error) {

	if token == "" {
		zap.L().Warn("request for empty token")
		return nil, ErrTokenInvalid
	}

	t.mutex.RLock()

	xtoken, exist := t.internal[token]
	if exist {
		t.mutex.RUnlock()
		if xtoken.Exp < time.Now().Unix() {
			zap.L().Debug(fmt.Sprintf("Token exist in cache and is invalid %s", token))
			return nil, ErrTokenExpired
		}
		// Func is exported. Return clone to untrusted outsiders
		return xtoken.Clone(), nil
	}

	t.mutex.RUnlock()
	t.mutex.Lock()
	defer t.mutex.Unlock()

	zap.L().Debug(fmt.Sprintf("Token=%s not found in cache", token))

	xtoken = &Token{}

	_, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {

		// SigningMethodRSA

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, fmt.Errorf("Token claims have unexpected format")
		}

		for k, v := range claims {

			if k == "iss" {
				xtoken.Iss, _ = v.(string)
			}

			if k == "exp" {
				floatValue := v.(float64)
				xtoken.Exp = int64(floatValue)
			}

			if k == "aud" {
				xtoken.Aud, _ = v.(string)
			}

		}

		xtoken.Claims = claims
		xtoken.Kid = token.Header["kid"].(string)
		xtoken.Alg = token.Header["alg"].(string)

		if xtoken.Iss == "" {
			return nil, fmt.Errorf("Issuer(iss) not found in claims")
		}

		if !strings.HasPrefix(xtoken.Iss, "https://") {
			return nil, fmt.Errorf("Issuer is not valid")
		}

		if xtoken.Exp == 0 {
			return nil, fmt.Errorf("Expiration(exp) not found")
		}

		publicKey, err := t.publicKeys.GetKey(xtoken.Iss, xtoken.Kid)
		if err != nil {
			return nil, err
		}

		if publicKey.Kty == "" {
			return nil, fmt.Errorf("kty is empty. should be EC or RSA")
		}

		switch token.Method.(type) {
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
		if err == publickey.ErrPublicKeyInvalid {
			return nil, publickey.ErrPublicKeyInvalid
		} else if err.Error() == "Token is expired" {
			// We will be the judge of that
		} else if err.Error() == "Token used before issued" {
			// Slight drift in clock. We will be the judge of that
		} else {
			zap.L().Debug(fmt.Sprintf("%s", err))
			return nil, ErrTokenInvalid
		}
	}

	// Token may be expired

	if xtoken.Exp < time.Now().Unix() {
		zap.L().Debug(fmt.Sprintf("Token not in cache and is invalid %s", token))
		return nil, ErrTokenExpired
	}

	t.internal[token] = xtoken

	return xtoken.Clone(), nil
}

// Shutdown Cache
func (t *Server) Shutdown() {
	close(t.closed)
	t.wg.Wait()
}
