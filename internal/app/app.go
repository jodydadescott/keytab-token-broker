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
	"github.com/jodydadescott/keytab-token-broker/internal/policy"
	"github.com/jodydadescott/keytab-token-broker/internal/tokens"
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
		Nonce:  &nonces.Config{},
		Token:  &tokens.Config{},
		Keytab: &keytabs.Config{},
	}
}

// Config ...
type Config struct {
	Listen    string
	HTTPPort  int
	HTTPSPort int
	Query     string
	Policy    string
	Nonce     *nonces.Config
	Token     *tokens.Config
	Keytab    *keytabs.Config
}

// Server ...
type Server struct {
	closed      chan struct{}
	wg          sync.WaitGroup
	tokenCache  *tokens.TokenCache
	keytabCache *keytabs.KeytabCache
	nonceCache  *nonces.NonceCache
	httpServer  *http.Server
	policy      *policy.Policy
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
		Query:  config.Query,
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

func (t *Server) newNonce(ctx context.Context, token string) (*nonces.Nonce, error) {

	if token == "" {
		zap.L().Debug(fmt.Sprintf("newNonce(token=)->Denied:Fail : err=%s", "Token is empty"))
		return nil, ErrDataValidationFail
	}

	shortToken := token[1:8] + "..."

	xtoken, err := t.tokenCache.GetToken(token)
	if err != nil {
		zap.L().Debug(fmt.Sprintf("newNonce(%s)->[err=%s]", shortToken, err))
		return nil, ErrAuthFail
	}

	// Validate that token is allowed to pull nonce
	decision, err := t.policy.RenderDecision(ctx, xtoken.Claims)
	if err != nil {
		zap.L().Debug(fmt.Sprintf("newNonce(%s)->[err=%s]", shortToken, err))
		return nil, ErrAuthFail
	}

	if !decision.Auth {
		err = fmt.Errorf("Authorization denied")
		zap.L().Debug(fmt.Sprintf("newNonce(%s)->[err=%s]", shortToken, err))
		return nil, ErrAuthFail
	}

	nonce := t.nonceCache.NewNonce()
	shortNonce := nonce.Value[1:8] + "..."

	zap.L().Debug(fmt.Sprintf("newNonce(%s)->[%s]", shortToken, shortNonce))
	return nonce, nil
}

func (t *Server) getKeytab(ctx context.Context, token, principal string) (*keytabs.Keytab, error) {

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

	decision, err := t.policy.RenderDecision(ctx, xtoken.Claims)
	if err != nil {
		zap.L().Debug(fmt.Sprintf("getKeytab(token=%s,principal=%s)->Denied:Fail : err=%s", shortToken, principal, err))
		return nil, ErrAuthFail
	}

	if !decision.Auth {
		err = fmt.Errorf("Authorization denied")
		zap.L().Debug(fmt.Sprintf("getKeytab(token=%s,principal=%s)->Denied:Decision [Auth] : err=%s", shortToken, principal, err))
		return nil, ErrAuthFail
	}

	if !decision.HasPrincipal(principal) {
		zap.L().Debug(fmt.Sprintf("getKeytab(token=%s,principal=%s)->Denied:Decision [No matching principal] : err=%s", shortToken, principal, err))
		return nil, ErrAuthFail
	}

	keytab := t.keytabCache.GetKeytab(principal)

	if keytab == nil {
		zap.L().Debug(fmt.Sprintf("getKeytab(token=%s,principal=%s)->Denied:Fail : err=%s", shortToken, principal, err))
		return nil, ErrNotFound
	}

	zap.L().Debug(fmt.Sprintf("getKeytab(token=%s,principal=%s)->Granted", shortToken, principal))
	return keytab, nil

}

// Shutdown Server
func (t *Server) Shutdown() {
	close(t.closed)
	t.wg.Wait()
}
