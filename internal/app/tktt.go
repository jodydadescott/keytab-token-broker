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

package app

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jodydadescott/tokens2keytabs/internal/keytab"
	"github.com/jodydadescott/tokens2keytabs/internal/nonce"
	"github.com/jodydadescott/tokens2keytabs/internal/policy"
	"github.com/jodydadescott/tokens2keytabs/internal/token"
	"go.uber.org/zap"
)

var (
	// ErrDataValidationFail ...
	ErrDataValidationFail error = errors.New("Data validation failure")
	// ErrAuthFail ...
	ErrAuthFail error = errors.New("Authorization failure")
	// ErrNotFound ...
	ErrNotFound error = errors.New("Entity not found")
)

// NewConfig Returns new config
func NewConfig() *Config {
	return &Config{
		Nonce:  &nonce.Config{},
		Token:  &token.Config{},
		Keytab: &keytab.Config{},
	}
}

// Config ...
type Config struct {
	Listen, TLSCert, TLSKey, Query, Policy, Seed string
	HTTPPort, HTTPSPort                          int
	Nonce                                        *nonce.Config
	Token                                        *token.Config
	Keytab                                       *keytab.Config
}

// Server ...
type Server struct {
	closed                  chan struct{}
	wg                      sync.WaitGroup
	tokenCache              *token.Tokens
	keytabCache             *keytab.Keytabs
	nonceCache              *nonce.Nonces
	httpServer, httpsServer *http.Server
	policy                  *policy.Policy
}

// Build Returns a new Server
func (config *Config) Build() (*Server, error) {

	zap.L().Info(fmt.Sprintf("Starting"))

	if config.HTTPPort < 0 {
		return nil, fmt.Errorf("HTTPPort must be 0 or greater")
	}

	if config.HTTPSPort < 0 {
		return nil, fmt.Errorf("HTTPSPort must be 0 or greater")
	}

	if config.HTTPPort == 0 && config.HTTPSPort == 0 {
		return nil, fmt.Errorf("Must enable http or https")
	}

	tokenCache, err := config.Token.Build()
	if err != nil {
		return nil, err
	}

	keytabCache, err := config.Keytab.Build()
	if err != nil {
		return nil, err
	}

	nonceCache, err := config.Nonce.Build()
	if err != nil {
		return nil, err
	}

	policyConfig := &policy.Config{
		Policy: config.Policy,
	}
	policy, err := policyConfig.Build()
	if err != nil {
		return nil, err
	}

	server := &Server{
		closed:      make(chan struct{}),
		tokenCache:  tokenCache,
		keytabCache: keytabCache,
		nonceCache:  nonceCache,
		policy:      policy,
	}

	if config.HTTPPort > 0 {
		listen := config.Listen
		if strings.ToLower(listen) == "any" {
			listen = ""
		}
		listener := listen + ":" + strconv.Itoa(config.HTTPPort)
		zap.L().Debug("Starting HTTP")
		server.httpServer = &http.Server{Addr: listener, Handler: server}
		go func() {
			server.httpServer.ListenAndServe()
		}()
	}

	if config.HTTPSPort > 0 {
		listen := config.Listen
		if strings.ToLower(listen) == "any" {
			listen = ""
		}
		listener := listen + ":" + strconv.Itoa(config.HTTPSPort)

		zap.L().Debug("Starting HTTPS")

		if config.TLSCert == "" {
			return nil, fmt.Errorf("TLSCert is required when HTTPS port is set")
		}

		if config.TLSKey == "" {
			return nil, fmt.Errorf("TLSKey is required when HTTPS port is set")
		}

		cert, err := tls.X509KeyPair([]byte(config.TLSCert), []byte(config.TLSKey))
		if err != nil {
			return nil, err
		}

		server.httpsServer = &http.Server{Addr: listener, Handler: server, TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}}}

		go func() {
			server.httpsServer.ListenAndServeTLS("", "")
		}()

	}

	go func() {

		for {
			select {
			case <-server.closed:
				zap.L().Debug("Shutting down")

				if server.httpServer != nil {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					server.httpServer.Shutdown(ctx)
				}

				server.keytabCache.Shutdown()
				server.tokenCache.Shutdown()
				server.nonceCache.Shutdown()

			}
		}
	}()

	return server, nil
}

func (t *Server) newNonce(ctx context.Context, tokenString string) (*nonce.Nonce, error) {

	if tokenString == "" {
		zap.L().Debug(fmt.Sprintf("newNonce(tokenString=)->Denied:Fail : err=%s", "tokenString is empty"))
		return nil, ErrDataValidationFail
	}

	xtoken, err := t.tokenCache.GetToken(tokenString)
	if err != nil {
		zap.L().Debug(fmt.Sprintf("newNonce(%s)->[err=%s]", tokenString, err))
		return nil, ErrAuthFail
	}

	// Validate that token is allowed to pull nonce
	if t.policy.AuthGetNonce(ctx, xtoken.Claims) {
		nonce := t.nonceCache.NewNonce()
		zap.L().Debug(fmt.Sprintf("newNonce(%s)->[%s]", tokenString, nonce.Value))
		return nonce, nil
	}

	zap.L().Debug(fmt.Sprintf("newNonce(%s)->[auth denied by policy]", tokenString))
	return nil, ErrAuthFail
}

func (t *Server) getKeytab(ctx context.Context, token, principal string) (*keytab.Keytab, error) {

	shortToken := ""

	if token == "" {
		zap.L().Debug(fmt.Sprintf("getKeytab(token=,principal=%s)->Denied:Fail : err=%s", principal, "Token is empty"))
		return nil, ErrDataValidationFail
	}

	shortToken = token[1:8] + ".."

	if principal == "" {
		zap.L().Debug(fmt.Sprintf("getKeytab(token=%s,principal=%s)->Denied:Fail : err=%s", shortToken, principal, "Principal is empty"))
		return nil, ErrDataValidationFail
	}

	xtoken, err := t.tokenCache.GetToken(token)
	if err != nil {
		zap.L().Debug(fmt.Sprintf("getKeytab(token=%s,principal=%s)->Denied:Fail : err=%s", shortToken, principal, err))
		return nil, ErrAuthFail
	}

	if xtoken.Aud == "" {
		err = fmt.Errorf("Audience is empty")
		zap.L().Debug(fmt.Sprintf("getKeytab(token=%s,principal=%s)->Denied:Fail : err=%s", shortToken, principal, err))
		return nil, ErrAuthFail
	}

	nonce := t.nonceCache.GetNonce(xtoken.Aud)
	if nonce == nil {
		zap.L().Debug(fmt.Sprintf("getKeytab(token=%s,principal=%s)->Denied:Fail : err=%s", shortToken, principal, err))
		return nil, ErrAuthFail
	}

	if t.policy.AuthGetKeytab(ctx, xtoken.Claims, nonce.Value, principal) {

		keytab := t.keytabCache.GetKeytab(principal)

		if keytab == nil {
			zap.L().Debug(fmt.Sprintf("getKeytab(token=%s,principal=%s)->Denied:Fail : err=%s", shortToken, principal, err))
			return nil, ErrNotFound
		}

		zap.L().Debug(fmt.Sprintf("getKeytab(token=%s,principal=%s)->Granted", shortToken, principal))
		return keytab, nil

	}

	zap.L().Debug(fmt.Sprintf("getKeytab(token=%s,principal=%s)->Denied:Policy : err=%s", shortToken, principal, err))
	return nil, ErrAuthFail

}

// Shutdown Server
func (t *Server) Shutdown() {
	close(t.closed)
	t.wg.Wait()
}
