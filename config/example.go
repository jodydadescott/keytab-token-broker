package config

import (
	"encoding/json"
	"time"

	"gopkg.in/yaml.v2"
)

var examplePolicy = `
package main

default auth_get_nonce = false
default auth_get_keytab = false
default auth_get_secret = false

auth_base {
   # Match Issuer
   input.claims.iss == "abc123"
}

auth_get_nonce {
   auth_base
}

auth_nonce {
   # Verify that the request nonce matches the expected nonce. Our token provider
   # has the nonce in the audience field under claims
   input.claims.aud == input.nonce
}

auth_get_keytab {
   # The nonce must be validated and then the principal. This is done by splitting the
   # principals in the claim service.keytab by the comma into a set and checking for
   # match with requested principal
   auth_base
   auth_nonce
   split(input.claims.service.keytab,",")[_] == input.principal
}

auth_get_secret {
   # Verify that the request nonce matches the expected nonce. Our token provider
   # has the nonce in the audience field under claims
   auth_base
   auth_nonce
}
`

var exampleTLSCert = `-----BEGIN CERTIFICATE-----
................................................................
................................................................
................................................................
................................................................
................................................................
................................................................
................................................................
................................................................
................................................................
................................................................
................................................................
....................................
-----END CERTIFICATE-----`

var exampleTLSKey = `-----BEGIN EC PRIVATE KEY-----
................................................................
................................................................
................................................................
................................
-----END EC PRIVATE KEY-----`

// NewV1ExampleConfig New example config
func NewV1ExampleConfig() *Config {
	return &Config{
		APIVersion: "V1",
		Network: &Network{
			Listen:    "any",
			HTTPPort:  8080,
			HTTPSPort: 8443,
			TLSCert:   exampleTLSCert,
			TLSKey:    exampleTLSKey,
		},
		Policy: &Policy{
			Policy:         examplePolicy,
			NonceLifetime:  time.Duration(60) * time.Second,
			KeytabLifetime: time.Duration(60) * time.Second,
			Seed:           "this is not a good seed",
		},
		Logging: &Logging{
			LogLevel:         "info",
			LogFormat:        "json",
			OutputPaths:      []string{"stderr"},
			ErrorOutputPaths: []string{"stderr"},
		},
		Data: &Data{
			KeytabPrincipals: []string{"superman@EXAMPLE.COM", "birdman@EXAMPLE.COM"},
			Secrets: []*Secret{
				&Secret{
					Name:     "secret1",
					Seed:     "this is not a good seed for secret1",
					Lifetime: time.Duration(10) * time.Minute,
				},
				&Secret{
					Name:     "secret2",
					Seed:     "this is not a good seed for secret2",
					Lifetime: time.Duration(10) * time.Minute,
				},
				&Secret{
					Name:     "secret3",
					Seed:     "this is not a good seed for secret3",
					Lifetime: time.Duration(10) * time.Minute,
				},
			},
		},
	}
}

// ExampleConfigJSON Return example config as YAML
func ExampleConfigJSON() string {
	j, _ := json.Marshal(NewV1ExampleConfig())
	return string(j)
}

// ExampleConfigYAML Return example config as YAML
func ExampleConfigYAML() string {
	j, _ := yaml.Marshal(NewV1ExampleConfig())
	return string(j)
}
