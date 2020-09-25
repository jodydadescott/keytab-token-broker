package config

import (
	"fmt"

	"github.com/jodydadescott/kerberos-bridge/internal/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	Listen    string `json:"Listen,omitempty" yaml:"Listen,omitempty"`
	HTTPPort  int    `json:"httpPort,omitempty" yaml:"httpPort,omitempty"`
	HTTPSPort int    `json:"httpsPort,omitempty" yaml:"httpsPort,omitempty"`
}

// Policy Config
type Policy struct {
	Query          string `json:"query,omitempty" yaml:"query,omitempty"`
	Policy         string `json:"policy,omitempty" yaml:"policy,omitempty"`
	NonceLifetime  int    `json:"nonceLifetime,omitempty" yaml:"nonceLifetime,omitempty"`
	KeytabLifetime int    `json:"keytabLifetime,omitempty" yaml:"keytabLifetime,omitempty"`
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

// ServerConfig Return Server Config
func (t *Config) ServerConfig() (*server.Config, error) {

	if t.APIVersion == "" {
		return nil, fmt.Errorf("Missing APIVersion")
	}

	if t.APIVersion != "V1" {
		return nil, fmt.Errorf(fmt.Sprintf("APIVersion %s not supported", t.APIVersion))
	}

	serverConfig := server.NewConfig()

	if t.Network != nil {
		serverConfig.Listen = t.Network.Listen
		serverConfig.HTTPPort = t.Network.HTTPPort
		serverConfig.HTTPSPort = t.Network.HTTPSPort
	}

	if t.Policy != nil {
		serverConfig.Query = t.Policy.Query
		serverConfig.Policy = t.Policy.Policy
		serverConfig.Nonce.Lifetime = t.Policy.NonceLifetime
		serverConfig.Keytab.Lifetime = t.Policy.KeytabLifetime
	}

	if t.Data != nil {
		if t.Data.KeytabPrincipals != nil {
			for _, s := range t.Data.KeytabPrincipals {
				serverConfig.Keytab.Principals = append(serverConfig.Keytab.Principals, s)
			}
		}
	}

	return serverConfig, nil
}

// ZapConfig Return Zap Config
func (t *Config) ZapConfig() (*zap.Config, error) {

	zapConfig := &zap.Config{
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		EncoderConfig: zap.NewProductionEncoderConfig(),
	}

	if t.Logging != nil {

		switch t.Logging.LogLevel {

		case "debug":
			zapConfig.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
			break

		case "info":
			zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
			break

		case "warn":
			zapConfig.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
			break

		case "error":
			zapConfig.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
			break

		case "":
			zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)

		default:
			return nil, fmt.Errorf("logging level must be debug, info (default), warn or error")
		}

		switch t.Logging.LogFormat {

		case "json":
			zapConfig.Encoding = "json"
			break

		case "console":
			zapConfig.Encoding = "console"
			break

		case "":
			zapConfig.Encoding = "json"
			break

		default:
			return nil, fmt.Errorf("logging format must be json (default) or console")

		}

		if t.Logging.OutputPaths == nil || len(t.Logging.OutputPaths) <= 0 {
			zapConfig.OutputPaths = append(zapConfig.OutputPaths, "stderr")
		} else {
			for _, s := range t.Logging.OutputPaths {
				zapConfig.OutputPaths = append(zapConfig.OutputPaths, s)
			}
		}

		if t.Logging.ErrorOutputPaths == nil || len(t.Logging.ErrorOutputPaths) <= 0 {
			zapConfig.ErrorOutputPaths = append(zapConfig.ErrorOutputPaths, "stderr")
		} else {
			for _, s := range t.Logging.ErrorOutputPaths {
				zapConfig.ErrorOutputPaths = append(zapConfig.ErrorOutputPaths, s)
			}
		}

	} else {

		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
		zapConfig.Encoding = "json"
		zapConfig.OutputPaths = append(zapConfig.OutputPaths, "stderr")
		zapConfig.ErrorOutputPaths = append(zapConfig.ErrorOutputPaths, "stderr")

	}

	return zapConfig, nil
}
