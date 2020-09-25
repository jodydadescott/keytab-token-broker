package config

import (
	"encoding/json"

	"gopkg.in/yaml.v2"
)

var exampleQuery string = "grant_new_nonce = data.kbridge.grant_new_nonce; data.kbridge.get_principals[get_principals]"

var examplePolicy string = `
package kbridge

default grant_new_nonce = false
grant_new_nonce {
	input.iss == "https://api.console.aporeto.com/v/1/namespaces/5ddc396b9facec0001d3c886/oauthinfo"
}
get_principals[grant] {
	grant := split(input.claims.service.keytab,",")
}
`

func v1ExampleConfig() *Config {
	return &Config{
		APIVersion: "V1",
		Network: &Network{
			Listen:    "any",
			HTTPPort:  8080,
			HTTPSPort: 8443,
		},
		Policy: &Policy{
			Query:          exampleQuery,
			Policy:         examplePolicy,
			NonceLifetime:  60,
			KeytabLifetime: 120,
		},
		Logging: &Logging{
			LogLevel:         "info",
			LogFormat:        "json",
			OutputPaths:      []string{"stderr"},
			ErrorOutputPaths: []string{"stderr"},
		},
		Data: &Data{
			KeytabPrincipals: []string{"superman@EXAMPLE.COM", "birdman@EXAMPLE.COM"},
		},
	}
}

// ExampleConfigJSON Return example config as YAML
func ExampleConfigJSON() string {
	j, _ := json.Marshal(v1ExampleConfig())
	return string(j)
}

// ExampleConfigYAML Return example config as YAML
func ExampleConfigYAML() string {
	j, _ := yaml.Marshal(v1ExampleConfig())
	return string(j)
}
