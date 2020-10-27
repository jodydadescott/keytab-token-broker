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
	"fmt"
	"time"

	"github.com/jodydadescott/tokens2secrets/internal/keytab"
	"github.com/jodydadescott/tokens2secrets/internal/nonce"
	"github.com/jodydadescott/tokens2secrets/internal/policy"
	"github.com/jodydadescott/tokens2secrets/internal/publickey"
	"github.com/jodydadescott/tokens2secrets/internal/secret"
	"github.com/jodydadescott/tokens2secrets/internal/token"
	"go.uber.org/zap"
)

// Config ...
type Config struct {
	Policy         string
	NonceLifetime  time.Duration
	SecretSecrets  []*secret.Secret
	KeytabKeytabs  []*keytab.Keytab
	KeytabLifetime time.Duration
}

// Cache ...
type Cache struct {
	token     *token.Cache
	keytab    *keytab.Cache
	nonce     *nonce.Cache
	secret    *secret.Cache
	publickey publickey.Cache
	policy    *policy.Policy
}

// Build Returns a new Server
func (config *Config) Build() (*Cache, error) {

	zap.L().Info(fmt.Sprintf("Starting"))

	policyConfig := &policy.Config{}
	publickeyConfig := &publickey.Config{}
	tokenConfig := &token.Config{}
	keytabConfig := &keytab.Config{}
	nonceConfig := &nonce.Config{}
	secretConfig := &secret.Config{}

	if config.Policy != "" {
		policyConfig.Policy = config.Policy
	}

	if config.NonceLifetime > 0 {
		nonceConfig.Lifetime = config.NonceLifetime
	}

	if config.SecretSecrets != nil {
		secretConfig.Secrets = config.SecretSecrets
	}

	if config.KeytabKeytabs != nil {
		keytabConfig.Keytabs = config.KeytabKeytabs
	}

	policy, err := policyConfig.Build()
	if err != nil {
		return nil, err
	}

	publickey, err := publickeyConfig.Build()
	if err != nil {
		return nil, err
	}

	token, err := tokenConfig.Build(publickey)
	if err != nil {
		return nil, err
	}

	keytab, err := keytabConfig.Build()
	if err != nil {
		return nil, err
	}

	nonce, err := nonceConfig.Build()
	if err != nil {
		return nil, err
	}

	secret, err := secretConfig.Build()
	if err != nil {
		return nil, err
	}

	return &Cache{
		token:     token,
		keytab:    keytab,
		nonce:     nonce,
		secret:    secret,
		publickey: publickey,
		policy:    policy,
	}, nil

}

// Shutdown shutdown
func (t *Cache) Shutdown() {

	if t.secret != nil {
		t.secret.Shutdown()
	}

	if t.keytab != nil {
		t.keytab.Shutdown()
	}

	if t.nonce != nil {
		t.nonce.Shutdown()
	}

	if t.token != nil {
		t.token.Shutdown()
	}

	if t.publickey != nil {
		t.publickey.Shutdown()
	}

}

// GetNonce returns Nonce if provided token is authorized
func (t *Cache) GetNonce(ctx context.Context, tokenString string) (*nonce.Nonce, error) {

	token, err := t.token.ParseToken(tokenString)
	if err != nil {
		zap.L().Debug(fmt.Sprintf("GetNonce(tokenString=%s)->%s", tokenString, "Error:"+err.Error()))
		return nil, err
	}

	// Validate that token is allowed to pull nonce
	err = t.policy.AuthGetNonce(ctx, token.Claims)
	if err != nil {
		zap.L().Debug(fmt.Sprintf("GetNonce(tokenString=%s)->%s", tokenString, "Error:"+err.Error()))
		return nil, err
	}

	nonce, err := t.nonce.NewNonce()
	if err != nil {
		zap.L().Debug(fmt.Sprintf("GetNonce(tokenString=%s)->%s", tokenString, "Error:"+err.Error()))
		return nil, err
	}

	zap.L().Debug(fmt.Sprintf("GetNonce(tokenString=%s)->%s", tokenString, "Granted"))
	return nonce, nil
}

// GetKeytab returns Keytab if provided token is authorized
func (t *Cache) GetKeytab(ctx context.Context, tokenString, principal string) (*keytab.Keytab, error) {

	token, err := t.token.ParseToken(tokenString)
	if err != nil {
		zap.L().Debug(fmt.Sprintf("GetKeytab(tokenString=%s,principal=%s)->%s", tokenString, principal, "Error:"+err.Error()))
		return nil, err
	}

	nonce, err := t.nonce.GetNonce(token.Aud)
	if err != nil {
		zap.L().Debug(fmt.Sprintf("GetKeytab(tokenString=%s,principal=%s)->%s", tokenString, principal, "Error:"+err.Error()))
		return nil, err
	}

	err = t.policy.AuthGetKeytab(ctx, token.Claims, nonce.Value, principal)
	if err != nil {
		zap.L().Debug(fmt.Sprintf("GetKeytab(tokenString=%s,principal=%s)->%s", tokenString, principal, "Error:"+err.Error()))
		return nil, err
	}

	keytab, err := t.keytab.GetKeytab(principal)

	if err != nil {
		zap.L().Debug(fmt.Sprintf("GetKeytab(tokenString=%s,principal=%s)->%s", tokenString, principal, "Error:"+err.Error()))
		return nil, err
	}

	zap.L().Debug(fmt.Sprintf("GetKeytab(tokenString=%s,principal=%s)->%s", tokenString, principal, "Granted"))
	return keytab, nil
}

// GetSecret returns Secret if provided token is authorized
func (t *Cache) GetSecret(ctx context.Context, tokenString, name string) (*secret.Secret, error) {

	token, err := t.token.ParseToken(tokenString)
	if err != nil {
		zap.L().Debug(fmt.Sprintf("GetSecret(tokenString=%s,name=%s)->%s", tokenString, name, "Error:"+err.Error()))
		return nil, err
	}

	nonce, err := t.nonce.GetNonce(token.Aud)
	if err != nil {
		zap.L().Debug(fmt.Sprintf("GetSecret(tokenString=%s,principal=%s)->%s", tokenString, name, "Error:"+err.Error()))
		return nil, err
	}

	err = t.policy.AuthGetSecret(ctx, token.Claims, nonce.Value, name)
	if err != nil {
		zap.L().Debug(fmt.Sprintf("GetSecret(tokenString=%s,name=%s)->%s", tokenString, name, "Error:"+err.Error()))
		return nil, err
	}

	secret, err := t.secret.GetSecret(name)
	if err != nil {
		zap.L().Debug(fmt.Sprintf("GetSecret(tokenString=%s,name=%s)->%s", tokenString, name, "Error:"+err.Error()))
		return nil, err
	}

	zap.L().Debug(fmt.Sprintf("GetSecret(tokenString=%s,name=%s)->%s", tokenString, name, "Granted"))
	return secret, nil
}
