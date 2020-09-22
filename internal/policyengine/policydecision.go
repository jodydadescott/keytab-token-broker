package policyengine

import (
	"encoding/json"

	"github.com/jodydadescott/kerberos-bridge/internal/tokenstore"
)

// PolicyDecision Holds decision to permit or deny request
type PolicyDecision struct {
	GetNonce   bool     `json:"getNonce,omitempty" yaml:"getNonce,omitempty"`
	Principals []string `json:"principals,omitempty" yaml:"principals,omitempty"`
}

// JSON return JSON string representation of entity
func (t *PolicyDecision) JSON() string {
	e, err := json.Marshal(t)
	if err != nil {
		panic(err.Error())
	}
	return string(e)
}

// PolicyDecisionFromJSON Returns entity from JSON string
func PolicyDecisionFromJSON(b []byte) (*tokenstore.Token, error) {
	var t tokenstore.Token
	err := json.Unmarshal(b, &t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// HasPrincipal Returns true if principal is present in entity
func (t *PolicyDecision) HasPrincipal(principal string) bool {
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
