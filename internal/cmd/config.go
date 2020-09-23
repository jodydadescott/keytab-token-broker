package cmd

// import (
// 	"encoding/json"
// 	"fmt"

// 	"github.com/jodydadescott/kerberos-bridge/internal/server"
// 	"go.uber.org/zap"
// 	"go.uber.org/zap/zapcore"
// 	"gopkg.in/yaml.v2"
// )

// // Config Application Configuration
// type Config struct {
// 	LogLevel         string         `json:"logLevel,omitempty" yaml:"logLevel,omitempty"`
// 	LogFormat        string         `json:"logFormat,omitempty" yaml:"logFormat,omitempty"`
// 	OutputPaths      []string       `json:"outputPaths,omitempty" yaml:"outputPaths,omitempty"`
// 	ErrorOutputPaths []string       `json:"errorOutputPaths,omitempty" yaml:"errorOutputPaths,omitempty"`
// 	ServerConfig     *server.Config `json:"serverConfig,omitempty" yaml:"serverConfig,omitempty"`
// }

// // ZapConfig Return new instance of ZapConfig
// func (t *Config) ZapConfig() (*zap.Config, error) {

// 	zapConfig := &zap.Config{
// 		Development: false,
// 		Sampling: &zap.SamplingConfig{
// 			Initial:    100,
// 			Thereafter: 100,
// 		},
// 		EncoderConfig: zap.NewProductionEncoderConfig(),
// 	}

// 	switch t.LogLevel {

// 	case "debug":
// 		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
// 		break

// 	case "info":
// 		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
// 		break

// 	case "warn":
// 		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
// 		break

// 	case "error":
// 		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
// 		break

// 	case "":
// 		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)

// 	default:
// 		return nil, fmt.Errorf("logging level must be debug, info (default), warn or error")
// 	}

// 	switch t.LogFormat {

// 	case "json":
// 		zapConfig.Encoding = "json"
// 		break

// 	case "console":
// 		zapConfig.Encoding = "console"
// 		break

// 	case "":
// 		zapConfig.Encoding = "json"
// 		break

// 	default:
// 		return nil, fmt.Errorf("logging format must be json (default) or console")

// 	}

// 	if t.OutputPaths == nil || len(t.OutputPaths) <= 0 {
// 		zapConfig.OutputPaths = append(zapConfig.OutputPaths, "stderr")
// 	} else {
// 		zapConfig.OutputPaths = t.OutputPaths
// 	}

// 	if t.ErrorOutputPaths == nil || len(t.ErrorOutputPaths) <= 0 {
// 		zapConfig.ErrorOutputPaths = append(zapConfig.ErrorOutputPaths, "stderr")
// 	} else {
// 		zapConfig.ErrorOutputPaths = t.ErrorOutputPaths
// 	}

// 	return zapConfig, nil
// }

// // JSON return JSON string representation of entity
// func (t *Config) JSON() string {
// 	e, err := json.Marshal(t)
// 	if err != nil {
// 		panic(err.Error())
// 	}
// 	return string(e)
// }

// // YAML return JSON string representation of entity
// func (t *Config) YAML() string {
// 	e, err := yaml.Marshal(t)
// 	if err != nil {
// 		panic(err.Error())
// 	}
// 	return string(e)
// }
