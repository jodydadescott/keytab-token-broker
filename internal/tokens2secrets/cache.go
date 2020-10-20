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

package tokens2secrets

import (
	"context"
	"fmt"
	"sync"

	"github.com/jodydadescott/tokens2keytabs/internal/keytab"
	"github.com/jodydadescott/tokens2keytabs/internal/nonce"
	"github.com/jodydadescott/tokens2keytabs/internal/policy"
	"github.com/jodydadescott/tokens2keytabs/internal/secret"
	"github.com/jodydadescott/tokens2keytabs/internal/token"
	"go.uber.org/zap"
)

// NewConfig Returns new config
func NewConfig() *Config {
	return &Config{
		Nonce:  &nonce.Config{},
		Token:  &token.Config{},
		Keytab: &keytab.Config{},
		Secret: &secret.Config{},
	}
}

// Config ...
type Config struct {
	Query, Policy, Seed string
	Nonce               *nonce.Config
	Token               *token.Config
	Keytab              *keytab.Config
	Secret              *secret.Config
}

// Cache ...
type Cache struct {
	closed chan struct{}
	wg     sync.WaitGroup
	token  *token.Cache
	keytab *keytab.Cache
	nonce  *nonce.Cache
	secret *secret.Cache
	policy *policy.Policy
}

// Build Returns a new Cache
func (config *Config) Build() (*Cache, error) {

	zap.L().Info(fmt.Sprintf("Starting"))

	token, err := config.Token.Build()
	if err != nil {
		return nil, err
	}

	keytab, err := config.Keytab.Build()
	if err != nil {
		return nil, err
	}

	nonce, err := config.Nonce.Build()
	if err != nil {
		return nil, err
	}

	secret, err := config.Secret.Build()
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

	server := &Cache{
		closed: make(chan struct{}),
		token:  token,
		keytab: keytab,
		nonce:  nonce,
		secret: secret,
		policy: policy,
	}

	go func() {
		for {
			select {
			case <-server.closed:
				zap.L().Debug("Shutting down")
				server.keytab.Shutdown()
				server.secret.Shutdown()
				server.token.Shutdown()
				server.nonce.Shutdown()
			}
		}
	}()

	return server, nil
}

// GetNonce returns Nonce if provided token is authorized
func (t *Cache) GetNonce(ctx context.Context, tokenString string) (*nonce.Nonce, error) {

	if tokenString == "" {
		zap.L().Debug(fmt.Sprintf("newNonce(tokenString=)->Denied:Fail : err=%s", "tokenString is empty"))
		return nil, ErrDataValidationFail
	}

	token, err := t.token.ParseToken(tokenString)
	if err != nil {
		zap.L().Debug(fmt.Sprintf("newNonce(%s)->[err=%s]", tokenString, err))
		return nil, ErrAuthFail
	}

	// Validate that token is allowed to pull nonce
	if t.policy.AuthGetNonce(ctx, token.Claims) {
		nonce := t.nonce.NewNonce()
		zap.L().Debug(fmt.Sprintf("newNonce(%s)->[%s]", tokenString, nonce.Value))
		return nonce, nil
	}

	zap.L().Debug(fmt.Sprintf("newNonce(%s)->[auth denied by policy]", tokenString))
	return nil, ErrAuthFail
}

// GetKeytab returns Keytab if provided token is authorized
func (t *Cache) GetKeytab(ctx context.Context, token, principal string) (*keytab.Keytab, error) {

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

	xtoken, err := t.token.ParseToken(token)
	if err != nil {
		zap.L().Debug(fmt.Sprintf("getKeytab(token=%s,principal=%s)->Denied:Fail : err=%s", shortToken, principal, err))
		return nil, ErrAuthFail
	}

	if xtoken.Aud == "" {
		err = fmt.Errorf("Audience is empty")
		zap.L().Debug(fmt.Sprintf("getKeytab(token=%s,principal=%s)->Denied:Fail : err=%s", shortToken, principal, err))
		return nil, ErrAuthFail
	}

	nonce := t.nonce.GetNonce(xtoken.Aud)
	if nonce == nil {
		zap.L().Debug(fmt.Sprintf("getKeytab(token=%s,principal=%s)->Denied:Fail : err=%s", shortToken, principal, err))
		return nil, ErrAuthFail
	}

	if t.policy.AuthGetKeytab(ctx, xtoken.Claims, nonce.Value, principal) {

		keytab := t.keytab.GetKeytab(principal)

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

// GetSecret returns Secret if provided token is authorized
func (t *Cache) GetSecret(ctx context.Context, token, name string) (*secret.Secret, error) {

	shortToken := ""

	if token == "" {
		zap.L().Debug(fmt.Sprintf("getSecret(token=,name=%s)->Denied:Fail : err=%s", name, "Token is empty"))
		return nil, ErrDataValidationFail
	}

	shortToken = token[1:8] + ".."

	if name == "" {
		zap.L().Debug(fmt.Sprintf("getSecret(token=%s,name=%s)->Denied:Fail : err=%s", shortToken, name, "Principal is empty"))
		return nil, ErrDataValidationFail
	}

	xtoken, err := t.token.ParseToken(token)
	if err != nil {
		zap.L().Debug(fmt.Sprintf("getSecret(token=%s,name=%s)->Denied:Fail : err=%s", shortToken, name, err))
		return nil, ErrAuthFail
	}

	if xtoken.Aud == "" {
		err = fmt.Errorf("Audience is empty")
		zap.L().Debug(fmt.Sprintf("getSecret(token=%s,name=%s)->Denied:Fail : err=%s", shortToken, name, err))
		return nil, ErrAuthFail
	}

	nonce := t.nonce.GetNonce(xtoken.Aud)
	if nonce == nil {
		zap.L().Debug(fmt.Sprintf("getSecret(token=%s,name=%s)->Denied:Fail : err=%s", shortToken, name, err))
		return nil, ErrAuthFail
	}

	if t.policy.AuthGetSecret(ctx, xtoken.Claims, nonce.Value, name) {

		secret := t.secret.GetSecret(name)

		if secret == nil {
			zap.L().Debug(fmt.Sprintf("getSecret(token=%s,name=%s)->Denied:Fail : err=%s", shortToken, name, err))
			return nil, ErrNotFound
		}

		zap.L().Debug(fmt.Sprintf("getSecret(token=%s,name=%s)->Granted", shortToken, name))
		return secret, nil

	}

	zap.L().Debug(fmt.Sprintf("getSecret(token=%s,name=%s)->Denied:Policy : err=%s", shortToken, name, err))
	return nil, ErrAuthFail

}

// Shutdown Server
func (t *Cache) Shutdown() {
	close(t.closed)
	t.wg.Wait()
}
