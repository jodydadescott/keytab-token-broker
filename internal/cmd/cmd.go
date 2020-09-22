package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/jodydadescott/kerberos-bridge/internal/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

var createConfigCmd = &cobra.Command{
	Use:   "create",
	Short: "create example configuration",

	RunE: func(cmd *cobra.Command, args []string) error {

		config := server.NewExampleConfig()

		out := ""
		switch strings.ToLower(viper.GetString("format")) {

		case "yaml":
			out = config.YAML()
			break

		case "json":
			out = config.JSON()
			break

		case "":
			out = config.YAML()
			break

		default:
			return fmt.Errorf(fmt.Sprintf("Output format %s is unknown. Must be yaml or json", viper.GetString("format")))
		}

		switch strings.ToLower(viper.GetString("out")) {

		case "", "stdout":
			fmt.Print(out)
			return nil

		case "stderr":
			fmt.Fprint(os.Stderr, out)
			return nil

		}

		return ioutil.WriteFile(viper.GetString("out"), []byte(out), 0644)
	},
}

var exportConfigCmd = &cobra.Command{
	Use:   "export",
	Short: "export config from Windows registry to file",

	RunE: func(cmd *cobra.Command, args []string) error {

		config, err := getConfigFromRegistry()
		if err != nil {
			return err
		}

		out := ""
		switch strings.ToLower(viper.GetString("format")) {

		case "yaml":
			out = config.YAML()
			break

		case "json":
			out = config.JSON()
			break

		case "":
			out = config.YAML()
			break

		default:
			return fmt.Errorf(fmt.Sprintf("Output format %s is unknown. Must be yaml or json", viper.GetString("format")))
		}

		switch strings.ToLower(viper.GetString("out")) {

		case "", "stdout":
			fmt.Print(out)
			return nil

		case "stderr":
			fmt.Fprint(os.Stderr, out)
			return nil

		}

		return ioutil.WriteFile(viper.GetString("out"), []byte(out), 0644)
	},
}

var importConfigCmd = &cobra.Command{
	Use:   "import",
	Short: "import config from file or url into Windows registry",

	RunE: func(cmd *cobra.Command, args []string) error {

		in := viper.GetString("in")
		if in == "" {
			return fmt.Errorf("Missing --in (filename or url)")
		}

		if strings.HasPrefix(in, "https://") || strings.HasPrefix(in, "http://") {
			config, err := getConfigFromURI(in)
			if err != nil {
				return err
			}
			err = setRegistryConfig(config)
			if err != nil {
				return err
			}
			return nil
		}

		config, err := getConfigFromFile(in)
		if err != nil {
			return err
		}

		err = setRegistryConfig(config)
		if err != nil {
			return err
		}
		return nil
	},
}

func getRuntimeConfig() (*server.Config, error) {

	var err error
	var config *server.Config

	in := viper.GetString("run-config")

	if in != "" {
		fmt.Fprint(os.Stderr, fmt.Sprintf("Runtime config is %s from arg", in))

		if strings.HasPrefix(in, "https://") || strings.HasPrefix(in, "http://") {
			config, err := getConfigFromURI(in)
			if err != nil {
				return nil, err
			}
			return config, nil
		}

		config, err = getConfigFromFile(in)
		if err != nil {
			return nil, err
		}
		return config, err
	}

	if runtime.GOOS == "windows" {

		in, err = getRuntimeConfigString()
		if err != nil {
			return nil, err
		}

		if in == "registry" || in == "" {
			fmt.Fprint(os.Stderr, fmt.Sprintf("Runtime config is %s from Windows registry", in))
			config, err = getConfigFromRegistry()
			if err != nil {
				return nil, err
			}
			return config, nil
		}

		if strings.HasPrefix(in, "https://") || strings.HasPrefix(in, "http://") {
			fmt.Fprint(os.Stderr, fmt.Sprintf("Runtime config is remote %s", in))
			config, err := getConfigFromURI(in)
			if err != nil {
				return nil, err
			}
			return config, nil
		}

		config, err := getConfigFromFile(in)
		if err != nil {
			return nil, err
		}
		return config, err
	}

	return nil, fmt.Errorf("")

	return nil, fmt.Errorf("")
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "run server",

	RunE: func(cmd *cobra.Command, args []string) error {

		// getRuntimeConfig

		// config := server.NewConfig()

		// if runtime.GOOS == "windows" {
		// 	if viper.GetBool("registry") {
		// 		fmt.Fprintln(os.Stderr, "Getting and merging config from Windows registry")
		// 		newConfig, err := registryGetConfig()
		// 		if err != nil {
		// 			return err
		// 		}
		// 		config.MergeConfig(newConfig)
		// 	}
		// }

		// if x := viper.GetString("config"); x != "" {
		// 	fmt.Fprintln(os.Stderr, fmt.Sprintf("Getting and merging config from file %s", x))
		// 	newConfig, err := getConfigFromFile(x)
		// 	if err != nil {
		// 		return err
		// 	}
		// 	config.MergeConfig(newConfig)
		// }

		// fmt.Fprintln(os.Stderr, fmt.Sprintf("Getting and merging config from args"))
		// config.MergeConfig(getConfigFromArgs())

		// fmt.Println(fmt.Sprintf("Config:%s", config.JSON()))

		// sig := make(chan os.Signal, 2)
		// signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

		// server, err := server.NewServer(config)
		// if err != nil {
		// 	fmt.Println(fmt.Sprintf("Unable to start server; err=%s", err))
		// 	os.Exit(3)
		// }

		// <-sig

		// server.Shutdown()

		return nil

	},
}

// Execute ...
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {

	configCmd.AddCommand(createConfigCmd)
	// import/export are only for windows registry
	if runtime.GOOS == "windows" {
		configCmd.AddCommand(importConfigCmd, exportConfigCmd)
	}

	// Config Command

	configCmd.PersistentFlags().StringP("format", "", "", "output format in yaml or json; default is yaml")
	viper.BindPFlag("format", configCmd.PersistentFlags().Lookup("format"))

	configCmd.PersistentFlags().StringP("in", "", "", "input (file or url)")
	viper.BindPFlag("in", configCmd.PersistentFlags().Lookup("in"))

	configCmd.PersistentFlags().StringP("out", "", "", "output file")
	viper.BindPFlag("out", configCmd.PersistentFlags().Lookup("out"))

	// ServerCmd

	serverCmd.PersistentFlags().StringP("run-config", "", "", "file or url for runtime config. overrides default")
	viper.BindPFlag("run-config", serverCmd.PersistentFlags().Lookup("run-config"))

	rootCmd.AddCommand(configCmd, serverCmd)
}

func getConfigFromFile(filename string) (*server.Config, error) {

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

func getConfigFromString(input string) (*server.Config, error) {
	return getConfigFromBytes([]byte(input))
}

func getConfigFromBytes(input []byte) (*server.Config, error) {

	var yamlErr error
	var err error
	var config *server.Config

	// File could be YAML or JSON. Try both. Return yaml error if neither work
	config, yamlErr = server.ConfigFromYAML(input)
	if yamlErr == nil {
		return config, nil
	}

	config, err = server.ConfigFromJSON(input)
	if err == nil {
		return config, nil
	}

	return nil, yamlErr
}

func getConfigFromURI(uri string) (*server.Config, error) {

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
