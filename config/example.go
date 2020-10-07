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

var exampleTLSCert = `-----BEGIN CERTIFICATE-----
MIICMDCCAbUCCQDfhpjqkd8ewjAKBggqhkjOPQQDAjCBgDELMAkGA1UEBhMCVVMx
CzAJBgNVBAgMAlRYMRIwEAYDVQQHDAlTb3V0aGxha2UxDTALBgNVBAoMBEpvZHkx
DTALBgNVBAsMBEpvZHkxEDAOBgNVBAMMB2V4YW1wbGUxIDAeBgkqhkiG9w0BCQEW
EWFkbWluQGV4YW1wbGUuY29tMB4XDTIwMTAwNjIxMjUwOVoXDTMwMTAwNDIxMjUw
OVowgYAxCzAJBgNVBAYTAlVTMQswCQYDVQQIDAJUWDESMBAGA1UEBwwJU291dGhs
YWtlMQ0wCwYDVQQKDARKb2R5MQ0wCwYDVQQLDARKb2R5MRAwDgYDVQQDDAdleGFt
cGxlMSAwHgYJKoZIhvcNAQkBFhFhZG1pbkBleGFtcGxlLmNvbTB2MBAGByqGSM49
AgEGBSuBBAAiA2IABH5IWn5ZRoCBxiaoWcHxsR8ozPN5zsnoBRT/aI+b5kQYxnOr
/Qd3tFnuq955BpfAbuGMfieqrrMop1wkcysz3KcglqlzUTy/Kk+FmmNWWUA/KE1W
Z70r0u9nG1FSQgHLtTAKBggqhkjOPQQDAgNpADBmAjEAueGyOviFCHJDqJxv+sZA
oSXPQqUnGzTfiOBT+e/iSXPuedrq2aTl9iDxjX3I/kLQAjEAnzZM29NgpI1D/28G
HyMaJavNYa31o1fSMKGW67zNV0LNy3X3FSyPLbVX4KT3Jy13
-----END CERTIFICATE-----`

var exampleTLSKey = `-----BEGIN EC PRIVATE KEY-----
MIGkAgEBBDBWh9l1d6MvJeEU9wfo4GcLULM8NU01h5fej7h8NMBjLdjWEwt4Ted8
AxAnwk018C6gBwYFK4EEACKhZANiAAR+SFp+WUaAgcYmqFnB8bEfKMzzec7J6AUU
/2iPm+ZEGMZzq/0Hd7RZ7qveeQaXwG7hjH4nqq6zKKdcJHMrM9ynIJapc1E8vypP
hZpjVllAPyhNVme9K9LvZxtRUkIBy7U=
-----END EC PRIVATE KEY-----`

// TimePeriod
// OneMinute
// FiveMinute
// QuarterHour
// HalfHour
// Hour
// QuarterDay
// HalfDay
// Day

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
			Query:          exampleQuery,
			Policy:         examplePolicy,
			NonceLifetime:  60,
			KeytabLifetime: 60,
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
