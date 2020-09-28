package app

import (
	"context"
	"fmt"
	"reflect"

	"github.com/open-policy-agent/opa/rego"
	"go.uber.org/zap"
)

type policyConfig struct {
	Query  string `json:"query,omitempty" yaml:"query,omitempty"`
	Policy string `json:"policy,omitempty" yaml:"policy,omitempty"`
}

type policyDecision struct {
	GetNonce   bool     `json:"getNonce,omitempty" yaml:"getNonce,omitempty"`
	Principals []string `json:"principals,omitempty" yaml:"principals,omitempty"`
}

type policy struct {
	query rego.PreparedEvalQuery
}

func (config *policyConfig) build() (*policy, error) {

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

	return &policy{
		query: query,
	}, nil
}

func (t *policy) renderDecision(ctx context.Context, input interface{}) (*policyDecision, error) {

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

	return &policyDecision{
		GetNonce:   getNonce,
		Principals: principals,
	}, nil
}

// HasPrincipal Returns true if principal is present in entity
func (t *policyDecision) hasPrincipal(principal string) bool {
	if principal == "" {
		return false
	}
	for _, s := range t.Principals {
		if s == principal {
			return true
		}
	}
	return false
}
