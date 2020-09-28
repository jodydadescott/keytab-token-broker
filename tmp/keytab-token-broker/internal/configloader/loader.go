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

package configloader

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jodydadescott/kerberos-bridge/config"
	"github.com/jodydadescott/kerberos-bridge/internal/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
)

const (
	maxIdleConnections int = 2
	requestTimeout     int = 60
)

// ConfigLoader Config
type ConfigLoader struct {
	config *config.Config
}

// ServerConfig Returns Server Config
func (t *ConfigLoader) ServerConfig() (*server.Config, error) {

	serverConfig := server.NewConfig()

	if t.config.Network != nil {
		serverConfig.Listen = t.config.Network.Listen
		serverConfig.HTTPPort = t.config.Network.HTTPPort
		serverConfig.HTTPSPort = t.config.Network.HTTPSPort
	}

	if t.config.Policy != nil {
		serverConfig.Query = t.config.Policy.Query
		serverConfig.Policy = t.config.Policy.Policy
		serverConfig.Nonce.Lifetime = t.config.Policy.NonceLifetime
		serverConfig.Keytab.SoftLifetime = t.config.Policy.KeytabSoftLifetime
		serverConfig.Keytab.HardLifetime = t.config.Policy.KeytabHardLifetime
	}

	if t.config.Data != nil {
		if t.config.Data.KeytabPrincipals != nil {
			for _, s := range t.config.Data.KeytabPrincipals {
				serverConfig.Keytab.Principals = append(serverConfig.Keytab.Principals, s)
			}
		}
	}

	return serverConfig, nil
}

// ZapConfig Returns Zap Config
func (t *ConfigLoader) ZapConfig() (*zap.Config, error) {

	zapConfig := &zap.Config{
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		EncoderConfig: zap.NewProductionEncoderConfig(),
	}

	if t.config.Logging != nil {

		switch t.config.Logging.LogLevel {

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

		switch t.config.Logging.LogFormat {

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

		if t.config.Logging.OutputPaths == nil || len(t.config.Logging.OutputPaths) <= 0 {
			zapConfig.OutputPaths = append(zapConfig.OutputPaths, "stderr")
		} else {
			for _, s := range t.config.Logging.OutputPaths {
				zapConfig.OutputPaths = append(zapConfig.OutputPaths, s)
			}
		}

		if t.config.Logging.ErrorOutputPaths == nil || len(t.config.Logging.ErrorOutputPaths) <= 0 {
			zapConfig.ErrorOutputPaths = append(zapConfig.ErrorOutputPaths, "stderr")
		} else {
			for _, s := range t.config.Logging.ErrorOutputPaths {
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

// NewConfigLoaderFromBytes Return new ConfigLoader from bytes
func NewConfigLoaderFromBytes(input []byte) (*ConfigLoader, error) {

	var config *config.Config

	err := yaml.Unmarshal(input, &config)
	if err != nil {
		err = json.Unmarshal(input, &config)
		if err != nil {
			return nil, fmt.Errorf("Input is not valid YAML or JSON config")
		}
	}

	// This should be done before the unmarshalling by reading the first
	if config.APIVersion == "" {
		return nil, fmt.Errorf("Missing APIVersion")
	}

	if config.APIVersion != "V1" {
		return nil, fmt.Errorf(fmt.Sprintf("APIVersion %s not supported", config.APIVersion))
	}

	return &ConfigLoader{
		config: config,
	}, nil

}

// NewConfigLoaderFromFileOrURL Return new ConfigLoader from file
func NewConfigLoaderFromFileOrURL(input string) (*ConfigLoader, error) {

	if strings.HasPrefix(input, "https://") || strings.HasPrefix(input, "http://") {

		req, err := http.NewRequest("GET", input, nil)
		if err != nil {
			return nil, err
		}

		resp, err := getHTTPClient().Do(req)
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf(fmt.Sprintf("%s returned status code %d", input, resp.StatusCode))
		}

		return NewConfigLoaderFromBytes(b)
	}

	f, err := os.Open(input)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	reader := bufio.NewReader(f)
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return NewConfigLoaderFromBytes(content)

}

// NewConfigLoader Return new ConfigLoader from local settings
func NewConfigLoader() (*ConfigLoader, error) {

	fileOrURL, err := GetRuntimeConfigString()
	if err != nil {
		return nil, err
	}

	if fileOrURL == "" {
		return nil, errors.New("config location not found")
	}

	return NewConfigLoaderFromFileOrURL(fileOrURL)

}

func getHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: maxIdleConnections,
		},
		Timeout: time.Duration(requestTimeout) * time.Second,
	}
}
