package controller

import (
	"kbridge/internal/api"
	"kbridge/internal/httpserver"
	"kbridge/internal/keytabstore"
	"kbridge/internal/noncestore"
	"kbridge/internal/tokenstore"

	"go.uber.org/zap"
)

// Controller ...
type Controller struct {
	tokenStore  *tokenstore.TokenStore
	keytabStore *keytabstore.KeytabStore
	nonceStore  *noncestore.NonceStore
	httpServer  *httpserver.HTTPServer
	api         *api.API
}

// Config ...
type Config struct {
	NonceLife  int64
	HTTPListen string
}

// NewController ...
func NewController(c *Config) *Controller {

	zap.L().Debug("Starting")

	tokenStoreConfig := tokenstore.NewConfig()
	nonceStoreConfig := noncestore.NewConfig()

	keytabstoreConfig := &keytabstore.Config{}

	if c.NonceLife > 0 {
		nonceStoreConfig.NonceLife = c.NonceLife
	}

	controller := &Controller{
		tokenStore:  tokenStoreConfig.Build(),
		keytabStore: keytabstore.NewKeytabStore(keytabstoreConfig),
		nonceStore:  nonceStoreConfig.Build(),
	}

	controller.api = api.NewAPI(controller.tokenStore, controller.keytabStore, controller.nonceStore)

	if c.HTTPListen != "" {
		controller.httpServer = httpserver.NewHTTPServer(c.HTTPListen, controller.api)
	}

	return controller
}

// NewControllerDefault ...
func NewControllerDefault() *Controller {

	zap.L().Info("Starting")

	c := &Config{
		HTTPListen: ":8080",
	}
	return NewController(c)
}

// GetAPI ...
func (t *Controller) GetAPI() *api.API {
	return t.api
}

// GetTokenStore ...
func (t *Controller) GetTokenStore() *tokenstore.TokenStore {
	return t.tokenStore
}

// GetKeytabStore ...
func (t *Controller) GetKeytabStore() *keytabstore.KeytabStore {
	return t.keytabStore
}

// Shutdown ...
func (t *Controller) Shutdown() {

	zap.L().Info("Shutting down")

	if t.httpServer != nil {
		t.httpServer.Shutdown()
	}
	t.keytabStore.Shutdown()
	t.tokenStore.Shutdown()
	t.nonceStore.Shutdown()
}
