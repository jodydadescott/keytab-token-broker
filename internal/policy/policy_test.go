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
	"encoding/json"
	"testing"
)

var (
	examplePolicy = `
package main

default auth_get_nonce = false
default auth_get_keytab = false
default auth_get_secret = false

auth_base {
   # Here we match that the token issuer is an authorized issuer
   input.claims.iss == "abc123"
}

auth_get_nonce {
  # For now all we are doing is calling auth_base. This could be expanded as needed.
   auth_base
}

auth_nonce {
   # To prevent replay attack we compare the nonce from the user with the nonce in
   # the token claims. Here we expect the nonce from the user to match the audience
   # (aud) field. If our token issuer uses a different claim we will need to adjust
   # as necessary.
   input.claims.aud == input.nonce
}

auth_get_keytab {
   # Here we auth the principals requested by the user. We use claims from the token
   # provider to determine is the bearer should be authorized. Our token provider has
   # the authorized principals in a comma delineated string under the string array
   # service which is under the claims. We split the comma string into a set and
   # check for a match
   auth_base
   split(input.claims.service.keytab,",")[_] == input.principal
}

auth_get_secret {
   # Verify that the request nonce matches the expected nonce. Our token provider
   # has the nonce in the audience field under claims
   auth_base
   auth_nonce
}
`

	exampleInput = `
{
	"alg": "EC",
	"kid": "donut",
	"iss": "abc123",
	"exp": 1599844897,
	"aud": "drpepper",
	"service": {
	  "keytab": "user1@example.com,user2@example.com"
	}
  }
`
)

func Test1(t *testing.T) {

	var claims map[string]interface{}

	// Unmarshal or Decode the JSON to the interface.
	json.Unmarshal([]byte(exampleInput), &claims)

	config := &Config{
		Policy: examplePolicy,
	}

	ctx := context.Background()

	policyEngine, err := config.Build()

	if err != nil {
		t.Errorf("Unexpected error:%s", err)
	}

	if !policyEngine.AuthGetNonce(ctx, claims) {
		t.Errorf("AuthGetNonce should be true")
	}

	// AuthGetKeytab(ctx context.Context, claims map[string]interface{}, nonce string, principals []string)
	if !policyEngine.AuthGetKeytab(ctx, claims, "drpepper", "user1@example.com") {
		t.Errorf("AuthGetNonce should be true")
	}

	if !policyEngine.AuthGetSecret(ctx, claims, "drpepper", "secret1") {
		t.Errorf("AuthGetSecret should be true")
	}

}
