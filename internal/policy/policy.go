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
	"reflect"

	"github.com/open-policy-agent/opa/rego"
	"go.uber.org/zap"
)

// Config ...
type Config struct {
	Query  string
	Policy string
}

// Policy ...
type Policy struct {
	query rego.PreparedEvalQuery
}

// Build ...
func (config *Config) Build() (*Policy, error) {

	if config.Query == "" {
		return nil, fmt.Errorf("Query is empty")
	}

	if config.Policy == "" {
		return nil, fmt.Errorf("Policy is empty")
	}

	ctx := context.Background()

	query, err := rego.New(
		rego.Query(config.Query),
		rego.Module("kerberos.rego", config.Policy),
	).PrepareForEval(ctx)

	if err != nil {
		return nil, err
	}

	return &Policy{
		query: query,
	}, nil
}

// RenderDecision Return rendered decision
func (t *Policy) RenderDecision(ctx context.Context, input interface{}) (*Decision, error) {

	results, err := t.query.Eval(ctx, rego.EvalInput(input))

	if err != nil {
		zap.L().Error(fmt.Sprintf("error->%s", err))
		return nil, err
	}

	if len(results) == 0 {
		fmt.Println("No results")
		return nil, fmt.Errorf("No results")
	}

	ok := false
	auth := false

	auth, ok = results[0].Bindings["auth"].(bool)

	if !ok {
		return nil, fmt.Errorf("Received unexpected type %s; expected type bool", reflect.TypeOf(results[0].Bindings["auth"]))
	}

	var tmpprincipals []interface{}

	tmpprincipals, ok = results[0].Bindings["principals"].([]interface{})

	if !ok {
		return nil, fmt.Errorf("Received unexpected type %s; expected type []interface{}", reflect.TypeOf(results[0].Bindings["principals"]))
	}

	principals := []string{}
	for _, principal := range tmpprincipals {
		principals = append(principals, fmt.Sprintf("%s", principal))
	}

	//xtype := reflect.TypeOf(results[0].Bindings["principals"])
	// fmt.Println(xtype)

	return &Decision{
		Auth:       auth,
		Principals: principals,
	}, nil
}
