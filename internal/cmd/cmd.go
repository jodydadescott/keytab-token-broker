package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/jodydadescott/kerberos-bridge/internal/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

const (
	maxIdleConnections int = 2
	requestTimeout     int = 60
)

var rootCmd = &cobra.Command{
	Use:   "kerberos-bridge",
	Short: "get kerberos keytabs with oauth tokens",
	Long: `Provides expiring kerberos keytabs to holders of bearer tokens
	by validating token is permitted keytab by policy. Policy is
	in the form of Open Policy Agent (OPA). Keytabs may be used
	to generate kerberos tickets and then discarded.`,
}

var _httpClient *http.Client

func getHTTPClient() *http.Client {
	if _httpClient == nil {
		_httpClient = &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: maxIdleConnections,
			},
			Timeout: time.Duration(requestTimeout) * time.Second,
		}
	}
	return _httpClient
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "manage configuration",
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "set runtime configuration",

	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) < 1 {
			return errors.New("config string required")
		}

		runtimeConfigString := args[0]
		// Need to verify string

		err := setRuntimeConfigString(runtimeConfigString)
		if err != nil {
			return err
		}
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get",
	Short: "get runtime configuration",

	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := getRuntimeConfigString()
		if err != nil {
			return err
		}
		fmt.Println(config)
		return nil
	},
}

var configExampleCmd = &cobra.Command{
	Use:   "example",
	Short: "create example configuration",

	RunE: func(cmd *cobra.Command, args []string) error {

		outputFileOrDev := "stdout"

		if len(args) > 0 {
			outputFileOrDev = args[0]
		}

		config := NewExampleConfig()

		configString := ""
		switch strings.ToLower(viper.GetString("format")) {

		case "yaml":
			configString = config.YAML()
			break

		case "json":
			configString = config.JSON()
			break

		case "":
			configString = config.YAML()
			break

		default:
			return fmt.Errorf(fmt.Sprintf("Output format %s is unknown. Must be yaml or json", viper.GetString("format")))
		}

		if outputFileOrDev == "stdout" {
			fmt.Print(configString)
			return nil
		}

		return ioutil.WriteFile(outputFileOrDev, []byte(configString), 0644)
	},
}

func getRuntimeConfig() (*Config, error) {

	runtimeConfigString := viper.GetString("config")
	runtimeConfigStringSource := "arg"

	if runtimeConfigString == "" {
		tmp, err := getRuntimeConfigString()
		if err != nil {
			return nil, err
		}
		runtimeConfigString = tmp
		runtimeConfigStringSource = "system"
	}

	if runtimeConfigString == "" {
		return nil, errors.New("runtime config string not found")
	}

	fmt.Fprintln(os.Stderr, fmt.Sprintf("Using runtime config string %s from %s", runtimeConfigString, runtimeConfigStringSource))

	if strings.HasPrefix(runtimeConfigString, "https://") || strings.HasPrefix(runtimeConfigString, "http://") {
		config, err := getConfigFromURI(runtimeConfigString)
		if err != nil {
			return nil, err
		}
		return config, nil
	}

	config, err := getConfigFromFile(runtimeConfigString)
	if err != nil {
		return nil, err
	}
	return config, err
}

var serverCmd = &cobra.Command{
	Use:   "start",
	Short: "start server",

	RunE: func(cmd *cobra.Command, args []string) error {

		config, err := getRuntimeConfig()
		if err != nil {
			return err
		}

		if config.ServerConfig == nil {
			return errors.New("ServerConfig is missing from config")
		}

		zapConfig, err := config.ZapConfig()
		if err != nil {
			return err
		}

		logger, err := zapConfig.Build()
		if err != nil {
			return err
		}

		if runtime.GOOS == "windows" {
			hook, err := getZapHook()
			if err != nil {
				return err
			}
			logger = logger.WithOptions(zap.Hooks(hook))
		}

		zap.ReplaceGlobals(logger)
		//defer logger.Sync()

		//fmt.Println(fmt.Sprintf("Config:%s", config.JSON()))

		sig := make(chan os.Signal, 2)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

		server, err := server.NewServer(config.ServerConfig)
		if err != nil {
			return err
		}

		<-sig

		server.Shutdown()

		return nil
	},
}

// Execute ...
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {

	configCmd.AddCommand(configSetCmd, configGetCmd, configExampleCmd)
	rootCmd.AddCommand(configCmd, serverCmd)

	configExampleCmd.PersistentFlags().StringP("format", "", "", "output format in yaml or json; default is yaml")
	viper.BindPFlag("format", configExampleCmd.PersistentFlags().Lookup("format"))

	serverCmd.PersistentFlags().StringP("config", "", "", "configuration file")
	viper.BindPFlag("config", serverCmd.PersistentFlags().Lookup("config"))
}

func getConfigFromFile(filename string) (*Config, error) {

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	reader := bufio.NewReader(f)
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return getConfigFromBytes(content)
}

func getConfigFromBytes(input []byte) (*Config, error) {

	var config *Config
	err0 := yaml.Unmarshal(input, &config)
	if err0 == nil {
		return config, nil
	}

	err := json.Unmarshal(input, &config)
	if err == nil {
		return config, nil
	}

	return nil, err0
}

func getConfigFromURI(uri string) (*Config, error) {

	fmt.Fprintln(os.Stderr, fmt.Sprintf("Getting config from %s", uri))

	req, err := http.NewRequest("GET", uri, nil)
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

	config, err := getConfigFromBytes(b)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(fmt.Sprintf("%s returned status code %d", uri, resp.StatusCode))
	}

	return config, nil
}
