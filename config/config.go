package config

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jodydadescott/kerberos-bridge/internal/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
)

// V1 Version 1 Config
type V1 struct {
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

// ConfigsFromBytes ...
func ConfigsFromBytes(input []byte) (*server.Config, *zap.Config, error) {

	// Must check version
	var config *V1

	err := yaml.Unmarshal(input, &config)
	if err != nil {
		err = json.Unmarshal(input, &config)
		if err != nil {
			return nil, nil, fmt.Errorf("Config is not valid json or yaml")
		}
	}

	return getConfigsFromV1(config)
}

func getConfigsFromV1(u *V1) (*server.Config, *zap.Config, error) {

	serverConfig := server.NewConfig()

	zapConfig := &zap.Config{
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		EncoderConfig: zap.NewProductionEncoderConfig(),
	}

	if u.APIVersion == "" {
		return nil, nil, fmt.Errorf("Missing APIVersion")
	}

	if strings.ToUpper(u.APIVersion) != "V1" {
		return nil, nil, fmt.Errorf(fmt.Sprintf("APIVersion %s not supported", u.APIVersion))
	}

	if u.Network != nil {
		serverConfig.Listen = u.Network.Listen
		serverConfig.HTTPPort = u.Network.HTTPPort
		serverConfig.HTTPSPort = u.Network.HTTPSPort
	}

	if u.Policy != nil {
		serverConfig.Query = u.Policy.Query
		serverConfig.Policy = u.Policy.Policy
		serverConfig.Nonce.Lifetime = u.Policy.NonceLifetime
		serverConfig.Keytab.Lifetime = u.Policy.KeytabLifetime
	}

	if u.Data != nil {
		if u.Data.KeytabPrincipals != nil {
			for _, s := range u.Data.KeytabPrincipals {
				serverConfig.Keytab.Principals = append(serverConfig.Keytab.Principals, s)
			}
		}
	}

	if u.Logging != nil {

		switch u.Logging.LogLevel {

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
			return nil, nil, fmt.Errorf("logging level must be debug, info (default), warn or error")
		}

		switch u.Logging.LogFormat {

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
			return nil, nil, fmt.Errorf("logging format must be json (default) or console")

		}

		if u.Logging.OutputPaths == nil || len(u.Logging.OutputPaths) <= 0 {
			zapConfig.OutputPaths = append(zapConfig.OutputPaths, "stderr")
		} else {
			for _, s := range u.Logging.OutputPaths {
				zapConfig.OutputPaths = append(zapConfig.OutputPaths, s)
			}
		}

		if u.Logging.ErrorOutputPaths == nil || len(u.Logging.ErrorOutputPaths) <= 0 {
			zapConfig.ErrorOutputPaths = append(zapConfig.ErrorOutputPaths, "stderr")
		} else {
			for _, s := range u.Logging.ErrorOutputPaths {
				zapConfig.ErrorOutputPaths = append(zapConfig.ErrorOutputPaths, s)
			}
		}

	} else {

		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
		zapConfig.Encoding = "json"
		zapConfig.OutputPaths = append(zapConfig.OutputPaths, "stderr")
		zapConfig.ErrorOutputPaths = append(zapConfig.ErrorOutputPaths, "stderr")

	}

	return serverConfig, zapConfig, nil
}
