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

package server

import (
	"fmt"
	"time"

	"github.com/jodydadescott/tokens2secrets/internal/app"
	"github.com/jodydadescott/tokens2secrets/internal/http"
	"github.com/jodydadescott/tokens2secrets/internal/keytab"
	"github.com/jodydadescott/tokens2secrets/internal/secret"
	"go.uber.org/zap"
)

// Config ...
type Config struct {
	Policy                                              string
	NonceLifetime, SecretMaxLifetime, SecretMinLifetime time.Duration
	SecretSecrets                                       []*secret.Secret
	KeytabKeytabs                                       []*keytab.Keytab
	KeytabLifetime                                      time.Duration

	Listen, TLSCert, TLSKey string
	HTTPPort, HTTPSPort     int
}

// Server ...
type Server struct {
	app  *app.Cache
	http *http.Server
}

// Build Returns a new Server
func (config *Config) Build() (*Server, error) {

	zap.L().Debug("Starting")

	if config.HTTPPort <= 0 && config.HTTPSPort <= 0 {
		return nil, fmt.Errorf("Either HTTPPort or HTTPSPort must be set")
	}

	appConfig := &app.Config{
		Policy:         config.Policy,
		NonceLifetime:  config.NonceLifetime,
		SecretSecrets:  config.SecretSecrets,
		KeytabKeytabs:  config.KeytabKeytabs,
		KeytabLifetime: config.KeytabLifetime,
	}

	app, err := appConfig.Build()
	if err != nil {
		return nil, err
	}

	httpConfig := &http.Config{
		Listen:    config.Listen,
		TLSCert:   config.TLSCert,
		TLSKey:    config.TLSKey,
		HTTPPort:  config.HTTPPort,
		HTTPSPort: config.HTTPSPort,
	}

	http, err := httpConfig.Build(app)
	if err != nil {
		return nil, err
	}

	return &Server{
		app:  app,
		http: http,
	}, nil
}

// Shutdown shutdown
func (t *Server) Shutdown() {
	zap.L().Debug("Stopping")

	if t.http != nil {
		t.http.Shutdown()
	}
	t.app.Shutdown()
}
