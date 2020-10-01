package policy

import (
	"context"
	"encoding/json"
	"testing"
)

func Test1(t *testing.T) {

	exampleToken := `
{
	"service": {
	  "@cloud:aws:ami-id": "ami-098f16afa9edf40be",
	  "ilove": "the80s",
	  "keytab": "principal1@EXAMPLE.COM, principal2@EXAMPLE.COM"
	},
	"aud": "initial",
	"exp": 1601482326,
	"iat": 1601478726,
	"iss": "https://api.console.aporeto.com/v/1/namespaces/5ddc396b9facec0001d3c886/oauthinfo",
	"sub": "5f7495d9a2057f00012669a0"
  }
`

	var input interface{}

	// Unmarshal or Decode the JSON to the interface.
	json.Unmarshal([]byte(exampleToken), &input)

	exampleQuery := "auth = data.kbridge.auth; data.kbridge.principals[principals]"

	examplePolicy := `
	package kbridge
		
	default auth = false
	
	auth {
		input.iss == "https://api.console.aporeto.com/v/1/namespaces/5ddc396b9facec0001d3c886/oauthinfo"
	}
	
	principals[grant] {
		grant := split(input.service.keytab,",")
	}
	
	`

	config := &Config{
		Query:  exampleQuery,
		Policy: examplePolicy,
	}

	ctx := context.Background()

	policyEngine, err := config.Build()

	if err != nil {
		t.Errorf("Unexpected error:%s", err)
	}

	decision, err := policyEngine.RenderDecision(ctx, input)
	if err != nil {
		t.Errorf("Unexpected error:%s", err)
	}

	if decision.Auth != true {
		t.Errorf("GetNonce should be true")
	}

	if !decision.HasPrincipal("principal1@EXAMPLE.COM") {
		t.Errorf("Principal should exist")
	}

}
