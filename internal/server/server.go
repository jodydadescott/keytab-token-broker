package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/jinzhu/copier"
	"github.com/jodydadescott/kerberos-bridge/internal/keytabstore"
	"github.com/jodydadescott/kerberos-bridge/internal/model"
	"github.com/jodydadescott/kerberos-bridge/internal/noncestore"
	"github.com/jodydadescott/kerberos-bridge/internal/policyengine"
	"github.com/jodydadescott/kerberos-bridge/internal/tokenstore"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

// ErrExpired ...
var ErrExpired error = errors.New("Expired")

// Config ...
type Config struct {
	Nonce           *noncestore.Config   `json:"nonce,omitempty" yaml:"nonce,omitempty"`
	Policy          *policyengine.Config `json:"policy,omitempty" yaml:"policy,omitempty"`
	Token           *tokenstore.Config   `json:"token,omitempty" yaml:"token,omitempty"`
	Keytab          *keytabstore.Config  `json:"keytab,omitempty" yaml:"keytab,omitempty"`
	HTTPPort        int                  `json:"httpPort,omitempty" yaml:"httpPort,omitempty"`
	HTTPSPort       int                  `json:"httpsPort,omitempty" yaml:"httpsPort,omitempty"`
	ListenInterface string               `json:"listenInterface,omitempty" yaml:"listenInterface,omitempty"`
}

// NewConfig ...
func NewConfig() *Config {
	return &Config{
		Nonce:  noncestore.NewConfig(),
		Policy: policyengine.NewConfig(),
		Token:  tokenstore.NewConfig(),
		Keytab: keytabstore.NewConfig(),
	}
}

// NewExampleConfig ...
func NewExampleConfig() *Config {
	return &Config{
		HTTPPort:  8080,
		HTTPSPort: 8443,
		Nonce:     noncestore.NewExampleConfig(),
		Policy:    policyengine.NewExampleConfig(),
		Token:     tokenstore.NewExampleConfig(),
		Keytab:    keytabstore.NewExampleConfig(),
	}
}

// JSON Returns JSON string representation of entity
func (t *Config) JSON() string {
	e, err := json.Marshal(t)
	if err != nil {
		panic(err.Error())
	}
	return string(e)
}

// YAML Returns YAML string representation of entity
func (t *Config) YAML() string {
	e, err := yaml.Marshal(t)
	if err != nil {
		panic(err.Error())
	}
	return string(e)
}

// ConfigFromJSON Returns entity from JSON string
func ConfigFromJSON(b []byte) (*Config, error) {
	var t Config
	err := json.Unmarshal(b, &t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// ConfigFromYAML Returns entity from YAML string
func ConfigFromYAML(b []byte) (*Config, error) {
	var t Config
	err := yaml.Unmarshal(b, &t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Server ...
type Server struct {
	closed       chan struct{}
	wg           sync.WaitGroup
	tokenStore   *tokenstore.TokenStore
	keytabStore  *keytabstore.KeytabStore
	nonceStore   *noncestore.NonceStore
	policyEngine *policyengine.PolicyEngine
	httpServer   *http.Server
}

// NewServer ...
func NewServer(config *Config) (*Server, error) {

	if config.HTTPPort < 0 {
		return nil, fmt.Errorf("HTTPPort must be 0 or greater")
	}

	if config.HTTPSPort < 0 {
		return nil, fmt.Errorf("HTTPSPort must be 0 or greater")
	}

	if config.HTTPPort == 0 && config.HTTPSPort == 0 {
		return nil, fmt.Errorf("Must enable http or https")
	}

	tokenstore, err := tokenstore.NewTokenStore(config.Token)
	if err != nil {
		return nil, err
	}

	keytabStore, err := keytabstore.NewKeytabStore(config.Keytab)
	if err != nil {
		return nil, err
	}

	nonceStore, err := noncestore.NewNonceStore(config.Nonce)
	if err != nil {
		return nil, err
	}

	policyEngine, err := policyengine.NewPolicyEngine(config.Policy)
	if err != nil {
		return nil, err
	}

	server := &Server{
		closed:       make(chan struct{}),
		tokenStore:   tokenstore,
		keytabStore:  keytabStore,
		nonceStore:   nonceStore,
		policyEngine: policyEngine,
	}

	go func() {

		zap.L().Debug("Starting")

		if config.HTTPPort > 0 {
			go func() {
				listener := config.ListenInterface + ":" + strconv.Itoa(config.HTTPPort)
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
		fmt.Fprintf(w, nonce.JSON()+"\n")
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

		fmt.Fprintf(w, keytab.JSON()+"\n")
		w.WriteHeader(http.StatusOK)
		return
	}

	fmt.Fprintf(w, newErrorResponse("Path "+r.URL.Path+" not mapped")+"\n")
	w.WriteHeader(http.StatusConflict)

	zap.L().Debug(fmt.Sprintf("Exiting ServeHTTP"))

}

func (t *Server) newNonce(ctx context.Context, token string) (*model.Nonce, error) {

	if token == "" {
		err := fmt.Errorf("Authorization denied")
		zap.L().Debug("Token is empty")
		return nil, err
	}

	shortToken := token[1:8] + "..."

	xtoken, err := t.tokenStore.GetToken(token)

	fmt.Println(xtoken.JSON())

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
	decision, err := t.policyEngine.RenderDecision(ctx, xtoken)
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

	// Make a copy of the entity to hand out so that encapsulation is preserved
	clone := &model.Nonce{}
	err = copier.Copy(&clone, &nonce)
	if err != nil {
		panic(err)
	}
	zap.L().Debug(fmt.Sprintf("NewNonce(%s)->[%s]", shortToken, shortNonce))
	return clone, nil
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

func (t *Server) getKeytab(ctx context.Context, token, principal string) (*model.Keytab, error) {

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

	decision, err := t.policyEngine.RenderDecision(ctx, xtoken)
	if err != nil {
		zap.L().Debug(fmt.Sprintf("GetKeytab(token=%s,principal=%s)->[err=%s]", shortToken, principal, err))
		return nil, err
	}

	if !decision.HasPrincipal(principal) {
		err = fmt.Errorf("Authorization denied")
		zap.L().Debug(fmt.Sprintf("GetKeytab(token=%s,principal=%s)->[err=%s]", shortToken, principal, err))
		return nil, err
	}

	keytab, err := t.keytabStore.GetKeytab(principal)

	if err != nil {
		zap.L().Debug(fmt.Sprintf("GetKeytab(token=%s,principal=%s)->[err=%s]", shortToken, principal, err))
		return nil, err
	}

	clone := &model.Keytab{}
	err = copier.Copy(&clone, &keytab)
	if err != nil {
		panic(err)
	}

	zap.L().Debug(fmt.Sprintf("GetKeytab(token=%s,principal=%s)->[valid keytab]", shortToken, principal))
	return clone, nil

}

// Shutdown ...
func (t *Server) Shutdown() {
	close(t.closed)
	t.wg.Wait()
}
