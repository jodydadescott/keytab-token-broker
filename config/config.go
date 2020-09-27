package config

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
	Listen    string `json:"Listen,omitempty" yaml:"Listen,omitempty"`
	HTTPPort  int    `json:"httpPort,omitempty" yaml:"httpPort,omitempty"`
	HTTPSPort int    `json:"httpsPort,omitempty" yaml:"httpsPort,omitempty"`
}

// Policy Config
type Policy struct {
	Query              string `json:"query,omitempty" yaml:"query,omitempty"`
	Policy             string `json:"policy,omitempty" yaml:"policy,omitempty"`
	NonceLifetime      int    `json:"nonceLifetime,omitempty" yaml:"nonceLifetime,omitempty"`
	KeytabSoftLifetime int    `json:"keytabSoftLifetime,omitempty" yaml:"keytabSoftLifetime,omitempty"`
	KeytabHardLifetime int    `json:"keytabHardLifetime,omitempty" yaml:"keytabHardLifetime,omitempty"`
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
	KeytabPrincipals []string `json:"principals,omitempty" yaml:"principals,omitempty"`
}

// NewV1Config Returns new V1 Config
func NewV1Config() *Config {
	return &Config{
		APIVersion: "V1",
		Network: &Network{
			Listen: "any",
		},
		Policy: &Policy{},
		Logging: &Logging{
			OutputPaths:      []string{""},
			ErrorOutputPaths: []string{""},
		},
		Data: &Data{
			KeytabPrincipals: []string{""},
		},
	}
}
