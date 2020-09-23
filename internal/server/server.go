package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jodydadescott/kerberos-bridge/internal/keytabs"
	"github.com/jodydadescott/kerberos-bridge/internal/nonces"
	"github.com/jodydadescott/kerberos-bridge/internal/tokens"
	"go.uber.org/zap"
)

// ErrExpired ...
var ErrExpired error = errors.New("Expired")

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
	tokenStore  *tokens.Cache
	keytabStore *keytabs.Cache
	nonceStore  *nonces.Cache
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

	tokenstore, err := config.Token.Build()
	if err != nil {
		return nil, err
	}

	keytabStore, err := config.Keytab.Build()
	if err != nil {
		return nil, err
	}

	nonceStore, err := config.Nonce.Build()
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
		tokenStore:  tokenstore,
		keytabStore: keytabStore,
		nonceStore:  nonceStore,
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
				zap.L().Info("Shutting down")

				if server.httpServer != nil {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					server.httpServer.Shutdown(ctx)
				}

				server.keytabStore.Shutdown()
				server.tokenStore.Shutdown()
				server.nonceStore.Shutdown()

			}
		}
	}()

	return server, nil
}

// ServeHTTP HTTP/HTTPS Handler
func (t *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	zap.L().Debug(fmt.Sprintf("Entering ServeHTTP path=%s method=%s", r.URL.Path, r.Method))

	defer zap.L().Debug(fmt.Sprintf("Exiting ServeHTTP path=%s method=%s", r.URL.Path, r.Method))

	w.Header().Set("Content-Type", "application/json")

	token := getBearerToken(r)

	if token == "" {
		sendERR(w, "Token required")
		return
	}

	switch r.URL.Path {
	case "/getnonce":
		nonce, err := t.newNonce(r.Context(), token)
		if handleERR(w, err) {
			return
		}
		fmt.Fprintf(w, toJSON(nonce)+"\n")
		w.WriteHeader(http.StatusOK)
		return

	case "/getkeytab":

		principal := getKey(r, "principal")
		if principal == "" {
			sendERR(w, "Principal required")
			return
		}

		keytab, err := t.getKeytab(r.Context(), token, principal)
		if handleERR(w, err) {
			return
		}

		fmt.Fprintf(w, toJSON(keytab)+"\n")
		w.WriteHeader(http.StatusOK)
		return
	}

	fmt.Fprintf(w, newErrorResponse("Path "+r.URL.Path+" not mapped")+"\n")
	w.WriteHeader(http.StatusConflict)

	zap.L().Debug(fmt.Sprintf("Exiting ServeHTTP"))

}

func (t *Server) newNonce(ctx context.Context, token string) (*nonces.Nonce, error) {

	if token == "" {
		err := fmt.Errorf("Authorization denied")
		zap.L().Debug("Token is empty")
		return nil, err
	}

	shortToken := token[1:8] + "..."

	xtoken, err := t.tokenStore.GetToken(token)

	if err != nil {
		zap.L().Debug(fmt.Sprintf("NewNonce(%s)->[err=%s]", shortToken, err))
		return nil, err
	}

	// Token may exist in cache but be expired
	if !xtoken.Valid() {
		zap.L().Debug(fmt.Sprintf("NewNonce(%s)->[err=%s]", shortToken, ErrExpired))
		return nil, ErrExpired
	}

	// Validate that token is allowed to pull nonce
	decision, err := t.policy.renderDecision(ctx, xtoken)
	if err != nil {
		zap.L().Debug(fmt.Sprintf("NewNonce(%s)->[err=%s]", shortToken, err))
		return nil, err
	}

	if !decision.GetNonce {
		err = fmt.Errorf("Authorization denied")
		zap.L().Debug(fmt.Sprintf("NewNonce(%s)->[err=%s]", shortToken, err))
		return nil, err
	}

	nonce := t.nonceStore.NewNonce()
	shortNonce := nonce.Value[1:8] + "..."

	zap.L().Debug(fmt.Sprintf("NewNonce(%s)->[%s]", shortToken, shortNonce))
	return nonce, nil
}

func newErrorResponse(message string) string {
	return "{\"error\":\"" + message + "\"}"
}

func sendERR(w http.ResponseWriter, message string) {
	fmt.Fprintf(w, newErrorResponse(message)+"\n")
	w.WriteHeader(http.StatusConflict)
}

func handleERR(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	fmt.Fprintf(w, newErrorResponse(err.Error())+"\n")
	w.WriteHeader(http.StatusConflict)
	return true
}

func getBearerToken(r *http.Request) string {
	token := r.Header.Get("Authorization")
	if token != "" {
		return token
	}
	return getKey(r, "bearertoken")
}

func getKey(r *http.Request, name string) string {
	keys, ok := r.URL.Query()[name]
	if !ok || len(keys[0]) < 1 {
		return ""
	}
	return string(keys[0])
}

func (t *Server) getKeytab(ctx context.Context, token, principal string) (*keytabs.Keytab, error) {

	shortToken := ""

	if token == "" || principal == "" {
		var err error

		if token == "" && principal == "" {
			err = fmt.Errorf("Token and Principal are empty")
		} else if token == "" {
			err = fmt.Errorf("Token is empty")
		} else {
			err = fmt.Errorf("Principal is empty")
		}

		shortToken = token[1:8] + ".."
		zap.L().Debug(fmt.Sprintf("GetKeytab(token=%s,principal=%s)->[err=%s]", shortToken, principal, err))
		return nil, err
	}

	shortToken = token[1:8] + ".."

	xtoken, err := t.tokenStore.GetToken(token)
	if err != nil {
		zap.L().Debug(fmt.Sprintf("GetKeytab(token=%s,principal=%s)->[err=%s]", shortToken, principal, err))
		return nil, err
	}

	// Token may exist in cahce but be expired
	if !xtoken.Valid() {
		zap.L().Debug(fmt.Sprintf("GetKeytab(token=%s,principal=%s)->[err=%s]", shortToken, principal, ErrExpired))
		return nil, ErrExpired
	}

	if xtoken.Aud == "" {
		err = fmt.Errorf("Audience is empty")
		zap.L().Debug(fmt.Sprintf("GetKeytab(token=%s,principal=%s)->[err=%s]", shortToken, principal, err))
		return nil, err
	}

	_, err = t.nonceStore.GetNonce(xtoken.Aud)
	if err != nil {
		zap.L().Debug(fmt.Sprintf("GetKeytab(token=%s,principal=%s)->[err=%s]", shortToken, principal, err))
		return nil, err
	}

	decision, err := t.policy.renderDecision(ctx, xtoken)
	if err != nil {
		zap.L().Debug(fmt.Sprintf("GetKeytab(token=%s,principal=%s)->[err=%s]", shortToken, principal, err))
		return nil, err
	}

	if !decision.hasPrincipal(principal) {
		err = fmt.Errorf("Authorization denied")
		zap.L().Debug(fmt.Sprintf("GetKeytab(token=%s,principal=%s)->[err=%s]", shortToken, principal, err))
		return nil, err
	}

	keytab, err := t.keytabStore.GetKeytab(principal)

	if err != nil {
		zap.L().Debug(fmt.Sprintf("GetKeytab(token=%s,principal=%s)->[err=%s]", shortToken, principal, err))
		return nil, err
	}

	zap.L().Debug(fmt.Sprintf("GetKeytab(token=%s,principal=%s)->[valid keytab]", shortToken, principal))
	return keytab, nil

}

func toJSON(e interface{}) string {
	// Log error and return valid JSON with err in it
	j, _ := json.Marshal(e)
	return string(j)
}

// Shutdown Server
func (t *Server) Shutdown() {
	close(t.closed)
	t.wg.Wait()
}
