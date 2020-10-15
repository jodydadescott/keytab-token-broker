package config

import (
	"encoding/json"
	"time"

	"github.com/jinzhu/copier"
	"gopkg.in/yaml.v2"
)

// Config Config
type Config struct {
	APIVersion string   `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	Network    *Network `json:"network,omitempty" yaml:"network,omitempty"`
	Policy     *Policy  `json:"policy,omitempty" yaml:"policy,omitempty"`
	Logging    *Logging `json:"logging,omitempty" yaml:"logging,omitempty"`
	Data       *Data    `json:"data,omitempty" yaml:"data,omitempty"`
}

// Network Config
type Network struct {
	Listen    string `json:"listen,omitempty" yaml:"listen,omitempty"`
	HTTPPort  int    `json:"httpPort,omitempty" yaml:"httpPort,omitempty"`
	HTTPSPort int    `json:"httpsPort,omitempty" yaml:"httpsPort,omitempty"`
	TLSCert   string `json:"tlscert,omitempty" yaml:"tlscert,omitempty"`
	TLSKey    string `json:"tlsKey,omitempty" yaml:"tlsKey,omitempty"`
}

// Policy Config
type Policy struct {
	Policy         string        `json:"policy,omitempty" yaml:"policy,omitempty"`
	NonceLifetime  time.Duration `json:"nonceLifetime,omitempty" yaml:"nonceLifetime,omitempty"`
	KeytabLifetime time.Duration `json:"keytabLifetime,omitempty" yaml:"keytabLifetime,omitempty"`
	Seed           string        `json:"seed,omitempty" yaml:"seed,omitempty"`
}

// Logging Config
type Logging struct {
	LogLevel         string   `json:"logLevel,omitempty" yaml:"logLevel,omitempty"`
	LogFormat        string   `json:"logFormat,omitempty" yaml:"logFormat,omitempty"`
	OutputPaths      []string `json:"outputPaths,omitempty" yaml:"outputPaths,omitempty"`
	ErrorOutputPaths []string `json:"errorOutputPaths,omitempty" yaml:"errorOutputPaths,omitempty"`
}

// Data Config
type Data struct {
	KeytabPrincipals []string  `json:"principals,omitempty" yaml:"principals,omitempty"`
	Secrets          []*Secret `json:"secrets,omitempty" yaml:"secrets,omitempty"`
}

// Secret Config
type Secret struct {
	Name     string        `json:"name,omitempty" yaml:"name,omitempty"`
	Seed     string        `json:"seed,omitempty" yaml:"seed,omitempty"`
	Lifetime time.Duration `json:"lifetime,omitempty" yaml:"lifetime,omitempty"`
}

// NewConfig Returns new V1 Config
func NewConfig() *Config {
	return &Config{
		APIVersion: "V1",
		Network: &Network{
			Listen: "any",
		},
		Policy:  &Policy{},
		Logging: &Logging{},
		Data:    &Data{},
	}
}

// JSON Return JSON String representation
func (t *Config) JSON() string {
	j, _ := json.Marshal(t)
	return string(j)
}

// YAML Return YAML String representation
func (t *Config) YAML() string {
	j, _ := yaml.Marshal(t)
	return string(j)
}

// Clone return copy of entity
func (t *Config) Clone() *Config {
	clone := &Config{}
	copier.Copy(&clone, &t)
	return clone
}

// Merge Config into existing config
func (t *Config) Merge(config *Config) {

	config = config.Clone()

	if config.Network != nil {

		if t.Network == nil {
			t.Network = &Network{}
		}

		if config.Network.Listen != "" {
			t.Network.Listen = config.Network.Listen
		}

		if config.Network.HTTPPort > 0 {
			t.Network.HTTPPort = config.Network.HTTPPort
		}

		if config.Network.HTTPSPort > 0 {
			t.Network.HTTPSPort = config.Network.HTTPSPort
		}

		if config.Network.TLSKey != "" {
			t.Network.TLSKey = config.Network.TLSKey
		}

		if config.Network.TLSCert != "" {
			t.Network.TLSCert = config.Network.TLSCert
		}

	}

	if config.Policy != nil {

		if t.Policy == nil {
			t.Policy = &Policy{}
		}

		if config.Policy.Policy != "" {
			t.Policy.Policy = config.Policy.Policy
		}

		if config.Policy.NonceLifetime > 0 {
			t.Policy.NonceLifetime = config.Policy.NonceLifetime
		}

		if config.Policy.KeytabLifetime > 0 {
			t.Policy.KeytabLifetime = config.Policy.KeytabLifetime
		}

		if config.Policy.Seed != "" {
			t.Policy.Seed = config.Policy.Seed
		}

	}

	if config.Logging != nil {

		if t.Logging == nil {
			t.Logging = &Logging{}
		}

		if config.Logging.LogLevel != "" {
			t.Logging.LogLevel = config.Logging.LogLevel
		}

		if config.Logging.LogFormat != "" {
			t.Logging.LogFormat = config.Logging.LogFormat
		}

		if config.Logging.OutputPaths != nil {
			for _, s := range config.Logging.OutputPaths {
				if s != "" {
					t.Logging.OutputPaths = append(t.Logging.OutputPaths, s)
				}
			}
		}

		if config.Logging.ErrorOutputPaths != nil {
			for _, s := range config.Logging.ErrorOutputPaths {
				if s != "" {
					t.Logging.ErrorOutputPaths = append(t.Logging.ErrorOutputPaths, s)
				}
			}
		}

	}

	if config.Data != nil {

		if t.Data == nil {
			t.Data = &Data{}
		}

		if config.Data.KeytabPrincipals != nil {
			for _, s := range config.Data.KeytabPrincipals {
				t.Data.KeytabPrincipals = append(t.Data.KeytabPrincipals, s)
			}
		}

		if config.Data.Secrets != nil {
			for _, s := range config.Data.Secrets {
				t.Data.Secrets = append(t.Data.Secrets, s)
			}
		}

	}

}
