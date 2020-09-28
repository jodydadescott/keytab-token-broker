package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jodydadescott/keytab-token-broker/internal/keytabs"
	"github.com/jodydadescott/keytab-token-broker/internal/nonces"
	"github.com/jodydadescott/keytab-token-broker/internal/tokens"
	"go.uber.org/zap"
)

var (
	// ErrAuthFail ...
	ErrAuthFail error = errors.New("Authorization Failure")
	// ErrNotFound ...
	ErrNotFound error = errors.New("Matching attribute not found")
)

// NewConfig Returns new config
func NewConfig() *Config {
	return &Config{
		Nonce:  &nonces.Config{},
		Token:  &tokens.Config{},
		Keytab: &keytabs.Config{},
	}
}

// Config ...
type Config struct {
	Listen    string          `json:"Listen,omitempty" yaml:"Listen,omitempty"`
	HTTPPort  int             `json:"httpPort,omitempty" yaml:"httpPort,omitempty"`
	HTTPSPort int             `json:"httpsPort,omitempty" yaml:"httpsPort,omitempty"`
	Query     string          `json:"query,omitempty" yaml:"query,omitempty"`
	Policy    string          `json:"policy,omitempty" yaml:"policy,omitempty"`
	Nonce     *nonces.Config  `json:"nonce,omitempty" yaml:"nonce,omitempty"`
	Token     *tokens.Config  `json:"token,omitempty" yaml:"token,omitempty"`
	Keytab    *keytabs.Config `json:"keytab,omitempty" yaml:"keytab,omitempty"`
}

// Server ...
type Server struct {
	closed      chan struct{}
	wg          sync.WaitGroup
	tokenCache  *tokens.TokenCache
	keytabCache *keytabs.KeytabCache
	nonceCache  *nonces.NonceCache
	httpServer  *http.Server
	policy      *policy
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

	policyConfig := &policyConfig{
		Query:  config.Query,
		Policy: config.Policy,
	}
	policy, err := policyConfig.build()
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

	go func() {

		if config.HTTPPort > 0 {
			go func() {
				listen := config.Listen
				if strings.ToLower(listen) == "any" {
					listen = ""
				}
				listener := listen + ":" + strconv.Itoa(config.HTTPPort)
				zap.L().Debug("Starting HTTP")
				server.httpServer = &http.Server{Addr: listener, Handler: server}
				server.httpServer.ListenAndServe()
			}()
		}

		if config.HTTPSPort > 0 {
			zap.L().Debug("Starting HTTPS - Not implemented")
		}

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

// Shutdown Server
func (t *Server) Shutdown() {
	close(t.closed)
	t.wg.Wait()
}
