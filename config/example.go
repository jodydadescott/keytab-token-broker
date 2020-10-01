package config

import (
	"encoding/json"

	"gopkg.in/yaml.v2"
)

var exampleQuery = "auth = data.kbridge.auth; data.kbridge.principals[principals]"

var examplePolicy = `
package kbridge
	
default auth = false

auth {
	input.iss == "https://api.console.aporeto.com/v/1/namespaces/5ddc396b9facec0001d3c886/oauthinfo"
}

principals[grant] {
	grant := split(input.service.keytab,",")
}

`

// NewV1ExampleConfig New example config
func NewV1ExampleConfig() *Config {
	return &Config{
		APIVersion: "V1",
		Network: &Network{
			Listen:    "any",
			HTTPPort:  8080,
			HTTPSPort: 8443,
		},
		Policy: &Policy{
			Query:              exampleQuery,
			Policy:             examplePolicy,
			NonceLifetime:      60,
			KeytabSoftLifetime: 120,
			KeytabHardLifetime: 600,
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
	j, _ := json.Marshal(NewV1ExampleConfig())
	return string(j)
}

// ExampleConfigYAML Return example config as YAML
func ExampleConfigYAML() string {
	j, _ := yaml.Marshal(NewV1ExampleConfig())
	return string(j)
}
