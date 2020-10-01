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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jodydadescott/keytab-token-broker/config"
	"github.com/jodydadescott/keytab-token-broker/internal/app"
	"github.com/open-policy-agent/opa/rego"
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
	Config *config.Config
}

// NewConfigLoader Return new ConfigLoader instance
func NewConfigLoader() *ConfigLoader {
	return &ConfigLoader{
		Config: config.NewConfig(),
	}
}

// ServerConfig Returns Server Config
func (t *ConfigLoader) ServerConfig() (*app.Config, error) {

	serverConfig := app.NewConfig()

	if t.Config.Network != nil {
		serverConfig.Listen = t.Config.Network.Listen
		serverConfig.HTTPPort = t.Config.Network.HTTPPort
		serverConfig.HTTPSPort = t.Config.Network.HTTPSPort
	}

	if t.Config.Policy != nil {
		serverConfig.Query = t.Config.Policy.Query
		serverConfig.Policy = t.Config.Policy.Policy
		serverConfig.Nonce.Lifetime = t.Config.Policy.NonceLifetime
		serverConfig.Keytab.SoftLifetime = t.Config.Policy.KeytabSoftLifetime
		serverConfig.Keytab.HardLifetime = t.Config.Policy.KeytabHardLifetime
	}

	if t.Config.Data != nil {
		if t.Config.Data.KeytabPrincipals != nil {
			for _, s := range t.Config.Data.KeytabPrincipals {
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

	if t.Config.Logging != nil {

		switch t.Config.Logging.LogLevel {

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

		switch t.Config.Logging.LogFormat {

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

		if t.Config.Logging.OutputPaths == nil || len(t.Config.Logging.OutputPaths) <= 0 {
			zapConfig.OutputPaths = append(zapConfig.OutputPaths, "stderr")
		} else {
			for _, s := range t.Config.Logging.OutputPaths {
				zapConfig.OutputPaths = append(zapConfig.OutputPaths, s)
			}
		}

		if t.Config.Logging.ErrorOutputPaths == nil || len(t.Config.Logging.ErrorOutputPaths) <= 0 {
			zapConfig.ErrorOutputPaths = append(zapConfig.ErrorOutputPaths, "stderr")
		} else {
			for _, s := range t.Config.Logging.ErrorOutputPaths {
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

// LoadeFromBytes Load data from bytes
func (t *ConfigLoader) LoadeFromBytes(input []byte) error {

	// Input could be JSON, YAML or REGO Policy

	var config *config.Config

	err := yaml.Unmarshal(input, &config)
	if err != nil {
		err = json.Unmarshal(input, &config)
		if err != nil {

			ctx := context.Background()

			policyString := string(input)

			_, err := rego.New(
				rego.Query("grant_new_nonce = data.kbridge.grant_new_nonce; data.kbridge.get_principals[get_principals]"),
				rego.Module("kerberos.rego", policyString),
			).PrepareForEval(ctx)

			if err == nil {
				t.Config.Policy.Policy = policyString
				return nil
			}

			return fmt.Errorf("Input is not valid YAML, JSON or Rego config")

		}
	}

	// This should be done before the unmarshalling by reading the first
	if config.APIVersion == "" {
		return fmt.Errorf("Missing APIVersion")
	}

	if config.APIVersion != "V1" {
		return fmt.Errorf(fmt.Sprintf("APIVersion %s not supported", config.APIVersion))
	}

	t.Config.Merge(config)

	return nil
}

// LoadFrom Load config(s) from one or more files or URLs (comma delimited)
func (t *ConfigLoader) LoadFrom(input string) error {

	var err error

	for _, s := range strings.Split(input, ",") {
		if strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "http://") {
			err = t.loadFromURL(s)
		} else {
			err = t.loadFromFile(s)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func (t *ConfigLoader) loadFromURL(input string) error {

	req, err := http.NewRequest("GET", input, nil)
	if err != nil {
		return err
	}

	resp, err := getHTTPClient().Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(fmt.Sprintf("%s returned status code %d", input, resp.StatusCode))
	}

	return t.LoadeFromBytes(b)

}

func (t *ConfigLoader) loadFromFile(input string) error {

	f, err := os.Open(input)
	if err != nil {
		return err
	}

	defer f.Close()

	reader := bufio.NewReader(f)
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	return t.LoadeFromBytes(b)

}

// LoadFromLocal Load data from registry or static file
func (t *ConfigLoader) LoadFromLocal() error {

	configString, err := GetRuntimeConfigString()
	if err != nil {
		return err
	}

	if configString == "" {
		return errors.New("config location not found")
	}

	err = t.LoadFrom(configString)
	if err != nil {
		return err
	}

	return nil
}

func getHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: maxIdleConnections,
		},
		Timeout: time.Duration(requestTimeout) * time.Second,
	}
}
