package policyengine

import (
	"context"
	"fmt"
	"reflect"

	"github.com/open-policy-agent/opa/rego"
	"go.uber.org/zap"
)

var exampleQuery = "grant_new_nonce = data.kbridge.grant_new_nonce; data.kbridge.get_principals[get_principals]"

var exampleRegoScript string = `
package kbridge
	
default grant_new_nonce = false

grant_new_nonce {
	input.iss == "https://api.console.aporeto.com/v/1/namespaces/5ddc396b9facec0001d3c886/oauthinfo"
}

get_principals[grant] {
	grant := split(input.claims.service.keytab,",")
}
`

// Config The Config
type Config struct {
	Query      string `json:"query,omitempty" yaml:"query,omitempty"`
	RegoScript string `json:"policy,omitempty" yaml:"policy,omitempty"`
}

// NewConfig ...
func NewConfig() *Config {
	return &Config{}
}

// NewExampleConfig ...
func NewExampleConfig() *Config {
	return &Config{
		Query:      exampleQuery,
		RegoScript: exampleRegoScript,
	}
}

// PolicyEngine ...
type PolicyEngine struct {
	query                   rego.PreparedEvalQuery
	disableNonceRequirement bool
}

// NewPolicyEngine Returns New Rego
func NewPolicyEngine(config *Config) (*PolicyEngine, error) {

	if config.Query == "" {
		return nil, fmt.Errorf("Query is empty")
	}

	if config.RegoScript == "" {
		return nil, fmt.Errorf("RegoScript is empty")
	}

	ctx := context.Background()

	query, err := rego.New(
		rego.Query(config.Query),
		rego.Module("kerberos.rego", config.RegoScript),
	).PrepareForEval(ctx)

	if err != nil {
		return nil, err
	}

	return &PolicyEngine{
		query: query,
	}, nil

}

// RenderDecision ...
func (t *PolicyEngine) RenderDecision(ctx context.Context, input interface{}) (*PolicyDecision, error) {

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
	getNonce := false

	getNonce, ok = results[0].Bindings["grant_new_nonce"].(bool)

	if !ok {
		return nil, fmt.Errorf("Received unexpected type %s; expected type bool", reflect.TypeOf(results[0].Bindings["grant_new_nonce"]))
	}

	var tmpprincipals []interface{}

	tmpprincipals, ok = results[0].Bindings["get_principals"].([]interface{})

	if !ok {
		return nil, fmt.Errorf("Received unexpected type %s; expected type []interface{}", reflect.TypeOf(results[0].Bindings["get_principals"]))
	}

	principals := []string{}
	for _, principal := range tmpprincipals {
		principals = append(principals, fmt.Sprintf("%s", principal))
	}

	//xtype := reflect.TypeOf(results[0].Bindings["principals"])
	// fmt.Println(xtype)

	return &PolicyDecision{
		GetNonce:   getNonce,
		Principals: principals,
	}, nil
}
