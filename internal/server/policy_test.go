package server

import (
	"context"
	"testing"
)

type token struct {
	Claims *claims `json:"claims"`
	Aud    string  `json:"aud"`
	Exp    int     `json:"exp"`
	Iat    int     `json:"iat"`
	Iss    string  `json:"iss"`
	Sub    string  `json:"sub"`
}

type claims struct {
	Service *service `json:"service"`
}

// TestService ...
type service struct {
	Keytab string `json:"keytab"`
}

func Test1(t *testing.T) {

	exampleQuery := "grant_new_nonce = data.kbridge.grant_new_nonce; data.kbridge.get_principals[get_principals]"

	examplePolicy := `
	package kbridge
		
	default grant_new_nonce = false
	grant_new_nonce {
		input.iss == "https://api.console.aporeto.com/v/1/namespaces/5ddc396b9facec0001d3c886/oauthinfo"
	}
	get_principals[grant] {
		grant := split(input.claims.service.keytab,",")
	}
	`

	input := &token{
		Iss: "https://api.console.aporeto.com/v/1/namespaces/5ddc396b9facec0001d3c886/oauthinfo",
		Exp: 1599844897,
		Aud: "daisy",
		Claims: &claims{
			Service: &service{
				Keytab: "principal1@EXAMPLE.COM, principal2@EXAMPLE.COM",
			},
		},
	}

	config := &policyConfig{
		Query:  exampleQuery,
		Policy: examplePolicy,
	}

	ctx := context.Background()

	policyEngine, err := config.build()

	if err != nil {
		t.Errorf("Unexpected error:%s", err)
	}

	decision, err := policyEngine.renderDecision(ctx, input)
	if err != nil {
		t.Errorf("Unexpected error:%s", err)
	}

	if decision.GetNonce != true {
		t.Errorf("GetNonce should be true")
	}

	if !decision.hasPrincipal("principal1@EXAMPLE.COM") {
		t.Errorf("Principal should exist")
	}

}
