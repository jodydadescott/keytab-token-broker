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

package policy

import (
	"context"
	"fmt"

	"github.com/open-policy-agent/opa/rego"
	"go.uber.org/zap"
)

// Config ...
type Config struct {
	Policy string
}

// Policy ...
type Policy struct {
	query rego.PreparedEvalQuery
}

// Build ...
func (config *Config) Build() (*Policy, error) {

	if config.Policy == "" {
		return nil, fmt.Errorf("Policy is empty")
	}

	ctx := context.Background()

	query, err := rego.New(
		rego.Query("auth_get_nonce = data.main.auth_get_nonce; auth_get_keytab = data.main.auth_get_keytab"),
		rego.Module("kerberos.rego", config.Policy),
	).PrepareForEval(ctx)

	if err != nil {
		return nil, err
	}

	return &Policy{
		query: query,
	}, nil
}

// AuthGetNonce Auth that claims are allowed to get nonce
func (t *Policy) AuthGetNonce(ctx context.Context, claims map[string]interface{}) bool {

	input := &Input{
		Claims: claims,
	}

	results, err := t.query.Eval(ctx, rego.EvalInput(input))

	if err != nil {
		zap.L().Error(fmt.Sprintf("Unexpected error on Rego policy execution; err->%s", err))
		return false
	}

	if len(results) == 0 {
		zap.L().Error(fmt.Sprintf("Unexpected error on Rego policy execution; results are empty"))
		return false
	}

	// data.kbridge.auth_get_keytab

	if auth, ok := results[0].Bindings["auth_get_nonce"].(bool); ok {
		zap.L().Error(fmt.Sprintf("Got result %t", auth))
		return auth
	}

	zap.L().Error(fmt.Sprintf("Unexpected error on Rego policy execution; unexpected result type"))
	return false
}

// AuthGetKeytab Auth that claims, nonce and principals are allowed to get requested keytab
func (t *Policy) AuthGetKeytab(ctx context.Context, claims map[string]interface{}, nonce, principal string) bool {

	input := &Input{
		Claims:    claims,
		Nonce:     nonce,
		Principal: principal,
	}

	results, err := t.query.Eval(ctx, rego.EvalInput(input))

	if err != nil {
		zap.L().Error(fmt.Sprintf("Unexpected error on Rego policy execution; err->%s", err))
		return false
	}

	if len(results) == 0 {
		zap.L().Error(fmt.Sprintf("Unexpected error on Rego policy execution; results are empty"))
		return false
	}

	if auth, ok := results[0].Bindings["auth_get_keytab"].(bool); ok {
		zap.L().Error(fmt.Sprintf("Got result %t", auth))
		return auth
	}

	zap.L().Error(fmt.Sprintf("Unexpected error on Rego policy execution; unexpected result type"))
	return false
}
