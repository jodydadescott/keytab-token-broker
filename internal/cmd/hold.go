package cmd

// import (
// 	"bufio"
// 	"fmt"
// 	"io/ioutil"
// 	"os"
// 	"runtime"
// 	"strings"

// 	"github.com/jodydadescott/kerberos-bridge/internal/server"
// 	"github.com/spf13/cobra"
// 	"github.com/spf13/viper"
// )

// var rootCmd = &cobra.Command{
// 	Use:   "kerberos-bridge",
// 	Short: "get kerberos keytabs with oauth tokens",
// 	Long: `Provides expiring kerberos keytabs to holders of bearer tokens
// 	by validating token is permitted keytab by policy. Policy is
// 	in the form of Open Policy Agent (OPA). Keytabs may be used
// 	to generate kerberos tickets and then discarded.`,
// }

// func getInputConfig() (*server.Config, error) {

// 	// TODO: Need to handle url

// 	config := server.NewConfig()
// 	inputs := viper.GetString("input")

// 	var err error
// 	var newConfig *server.Config
// 	if inputs != "" {
// 		for _, input := range strings.Split(inputs, ",") {

// 			if runtime.GOOS == "windows" && input == ":registry" {
// 				newConfig, err = getConfigFromRegistry()
// 			} else {
// 				newConfig, err = getConfigFromFile(input)
// 			}
// 			if err != nil {
// 				return nil, err
// 			}
// 			config.MergeConfig(newConfig)
// 		}
// 	}

// 	return config, nil
// }

// var configCmdWindows = `Create configuration from args and zero or more configuration
// files. Configuration files can be from filesystem, URL or Windows
// Registry (Windows only). Input and output format can be JSON or YAML.
// `

// func getconfigCmdLong() string {

// 	if runtime.GOOS == "windows" {
// 		return `Create configuration from args and zero or more configuration
// 	files. Configuration files can be from filesystem, URL or Windows
// 	Registry. Input and output format can be JSON or YAML. Use this to create,
// 	edit and store the configuration in the Windows registry or a local file.
// 	`
// 	}

// 	return `Create configuration from args and zero or more configuration
// 	files. Configuration files can be from filesystem or URL. Input and
// 	output format can be JSON or YAML. Use this to create, edit and store
// 	the configuration in a local file
// 	`
// }

// var createExampleConfig = &cobra.Command{
// 	Use:   "create-example",
// 	Short: "create example configuration",

// 	RunE: func(cmd *cobra.Command, args []string) error {

// 		if viper.GetString("output") == "" {
// 			return fmt.Errorf("Missing output")
// 		}

// 		config := server.NewExampleConfig()

// 		output := ""
// 		switch strings.ToLower(viper.GetString("output-format")) {

// 		case "yaml":
// 			output = config.YAML()
// 			break

// 		case "json":
// 			output = config.JSON()
// 			break

// 		case "":
// 			output = config.YAML()
// 			break

// 		default:
// 			return fmt.Errorf(fmt.Sprintf("Output format %s is unknown. Must be yaml or json", viper.GetString("output-format")))
// 		}

// 		switch strings.ToLower(viper.GetString("output")) {

// 		case "stdout":
// 			fmt.Println(output)
// 			return nil

// 		case "stderr":
// 			fmt.Fprintln(os.Stderr, output)
// 			return nil

// 		}

// 		return ioutil.WriteFile(viper.GetString("output"), []byte(output), 0644)
// 	},
// }

// var configCmd = &cobra.Command{
// 	Use:   "config",
// 	Short: "get and create configuration",
// 	Long:  getconfigCmdLong(),

// 	RunE: func(cmd *cobra.Command, args []string) error {

// 		if viper.GetString("output") == "" {
// 			return fmt.Errorf("Missing output")
// 		}

// 		config, err := getInputConfig()
// 		if err != nil {
// 			return err
// 		}

// 		config.MergeConfig(getConfigFromArgs())

// 		if runtime.GOOS == "windows" {
// 			if viper.GetString("output") == ":registry" {
// 				err = setRegistryConfig(config)
// 				if err != nil {
// 					return err
// 				}
// 				return nil
// 			}
// 		}

// 		output := ""
// 		switch strings.ToLower(viper.GetString("output-format")) {

// 		case "yaml":
// 			output = config.YAML()
// 			break

// 		case "json":
// 			output = config.JSON()
// 			break

// 		case "":
// 			output = config.YAML()
// 			break

// 		default:
// 			return fmt.Errorf(fmt.Sprintf("Output format %s is unknown. Must be yaml or json", viper.GetString("output-format")))
// 		}

// 		switch strings.ToLower(viper.GetString("output")) {

// 		case "", "stdout":
// 			fmt.Println(output)
// 			return nil

// 		case "stderr":
// 			fmt.Fprintln(os.Stderr, output)
// 			return nil

// 		}

// 		return ioutil.WriteFile(viper.GetString("output"), []byte(output), 0644)
// 	},
// }

// var serverCmd = &cobra.Command{
// 	Use:   "server",
// 	Short: "run server",

// 	RunE: func(cmd *cobra.Command, args []string) error {

// 		// config := server.NewConfig()

// 		// if runtime.GOOS == "windows" {
// 		// 	if viper.GetBool("registry") {
// 		// 		fmt.Fprintln(os.Stderr, "Getting and merging config from Windows registry")
// 		// 		newConfig, err := registryGetConfig()
// 		// 		if err != nil {
// 		// 			return err
// 		// 		}
// 		// 		config.MergeConfig(newConfig)
// 		// 	}
// 		// }

// 		// if x := viper.GetString("config"); x != "" {
// 		// 	fmt.Fprintln(os.Stderr, fmt.Sprintf("Getting and merging config from file %s", x))
// 		// 	newConfig, err := getConfigFromFile(x)
// 		// 	if err != nil {
// 		// 		return err
// 		// 	}
// 		// 	config.MergeConfig(newConfig)
// 		// }

// 		// fmt.Fprintln(os.Stderr, fmt.Sprintf("Getting and merging config from args"))
// 		// config.MergeConfig(getConfigFromArgs())

// 		// fmt.Println(fmt.Sprintf("Config:%s", config.JSON()))

// 		// sig := make(chan os.Signal, 2)
// 		// signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

// 		// server, err := server.NewServer(config)
// 		// if err != nil {
// 		// 	fmt.Println(fmt.Sprintf("Unable to start server; err=%s", err))
// 		// 	os.Exit(3)
// 		// }

// 		// <-sig

// 		// server.Shutdown()

// 		return nil

// 	},
// }

// // Execute ...
// func Execute() {
// 	if err := rootCmd.Execute(); err != nil {
// 		fmt.Println(err)
// 		os.Exit(1)
// 	}
// }

// func init() {

// 	rootCmd.AddCommand(configCmd)
// 	// rootCmd.AddCommand(serverCmd, configCmd)

// 	// Server Command

// 	// if runtime.GOOS == "windows" {
// 	// 	serverCmd.PersistentFlags().StringP("config", "", "", "Config: may be 'registry', file, or url. Default=registry")
// 	// } else {
// 	// 	serverCmd.PersistentFlags().StringP("config", "", "", "Config: May be file or url. Default=/etc/kerberos-bridge.conf")
// 	// }
// 	// viper.BindPFlag("config", serverCmd.PersistentFlags().Lookup("config"))

// 	// Config Command

// 	if runtime.GOOS == "windows" {
// 		configCmd.PersistentFlags().StringP("input", "", "", "input config (optional); if set must be ':registry', file, or url")
// 		configCmd.PersistentFlags().StringP("output", "", "", "output config (required); must be ':registry', file, stderr, or stdout")
// 	} else {
// 		configCmd.PersistentFlags().StringP("input", "", "", "input config (optional); if set must be file or url")
// 		configCmd.PersistentFlags().StringP("output", "", "", "output config (required); must be file, stderr, or stdout")
// 	}
// 	viper.BindPFlag("input", configCmd.PersistentFlags().Lookup("input"))
// 	viper.BindPFlag("output", configCmd.PersistentFlags().Lookup("output"))
// 	// configCmd.MarkFlagRequired("output_config")

// 	configCmd.PersistentFlags().StringP("output-format", "", "", "output format in yaml or json; default is yaml")
// 	viper.BindPFlag("output-format", configCmd.PersistentFlags().Lookup("output-format"))

// 	configCmd.PersistentFlags().StringP("log-level", "", "", "log level (debug, info, warn, error); default is info")
// 	viper.BindPFlag("log-level", configCmd.PersistentFlags().Lookup("log-level"))

// 	configCmd.PersistentFlags().StringP("log-format", "", "", "log format (json, text) default os json")
// 	viper.BindPFlag("log-format", configCmd.PersistentFlags().Lookup("log-format"))

// 	configCmd.PersistentFlags().StringP("listen", "", "", "network listen interface(s); default empty (all)")
// 	viper.BindPFlag("listen", configCmd.PersistentFlags().Lookup("listen"))

// 	configCmd.PersistentFlags().StringP("log-to", "", "", "file(s) and/or device(s) to log to (comma delimited); default is stderr")
// 	viper.BindPFlag("log-to", configCmd.PersistentFlags().Lookup("log-to"))

// 	configCmd.PersistentFlags().IntP("httpport", "", 0, "http port")
// 	viper.BindPFlag("httpport", configCmd.PersistentFlags().Lookup("httpport"))

// 	configCmd.PersistentFlags().IntP("httpsport", "", 0, "https port")
// 	viper.BindPFlag("httpsport", configCmd.PersistentFlags().Lookup("httpsport"))

// 	configCmd.PersistentFlags().IntP("nonce-lifetime", "", 0, "nonce lifetime; default is 60s")
// 	viper.BindPFlag("nonce-lifetime", configCmd.PersistentFlags().Lookup("nonce-lifetime"))

// 	configCmd.PersistentFlags().StringP("policy-query", "", "", "Policy query statement")
// 	viper.BindPFlag("policy-query", configCmd.PersistentFlags().Lookup("policy-query"))

// 	configCmd.PersistentFlags().StringP("policy-regoscript", "", "", "Policy rego script")
// 	viper.BindPFlag("policy-regoscript", configCmd.PersistentFlags().Lookup("policy-regoscript"))

// 	configCmd.PersistentFlags().IntP("keytab-lifetime", "", 0, "keytab lifetime; default is 120s")
// 	viper.BindPFlag("keytab-lifetime", configCmd.PersistentFlags().Lookup("keytab-lifetime"))

// 	configCmd.PersistentFlags().StringP("keytab-principals", "", "", "keytab principals (comma delimited)")
// 	viper.BindPFlag("keytab-principals", configCmd.PersistentFlags().Lookup("keytab-principals"))

// 	// viper.AutomaticEnv()

// }

// func getConfigFromArgs() *server.Config {

// 	config := server.NewConfig()

// 	if x := viper.GetString("log_level"); x != "" {
// 		config.LogLevel = x
// 	}

// 	if x := viper.GetString("log_format"); x != "" {
// 		config.LogFormat = x
// 	}

// 	if x := viper.GetString("log_to"); x != "" {
// 		for _, s := range strings.Split(x, ",") {
// 			config.LogTo = append(config.LogTo, s)
// 		}
// 	}

// 	if x := viper.GetString("listen"); x != "" {
// 		config.Listen = x
// 	}

// 	if x := viper.GetInt("httpport"); x > 0 {
// 		config.HTTPPort = x
// 	}

// 	if x := viper.GetInt("httpsport"); x > 0 {
// 		config.HTTPSPort = x
// 	}

// 	if x := viper.GetInt("nonce-lifetime"); x > 0 {
// 		config.Nonce.Lifetime = x
// 	}

// 	if x := viper.GetString("policy-query"); x != "" {
// 		config.Policy.Query = x
// 	}

// 	if x := viper.GetString("policy-regoscript"); x != "" {
// 		config.Policy.RegoScript = x
// 	}

// 	if x := viper.GetInt("keytab-lifetime"); x > 0 {
// 		config.Keytab.Lifetime = x
// 	}

// 	if x := viper.GetString("keytab-principals"); x != "" {
// 		for _, s := range strings.Split(x, ",") {
// 			config.Keytab.Principals = append(config.Keytab.Principals, strings.TrimSpace(s))
// 		}
// 	}

// 	return config
// }

// func getConfigFromFile(filename string) (*server.Config, error) {

// 	f, err := os.Open(filename)
// 	if err != nil {
// 		return nil, err
// 	}

// 	defer f.Close()

// 	reader := bufio.NewReader(f)
// 	content, err := ioutil.ReadAll(reader)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return getConfigFromBytes(content)
// }

// func getConfigFromString(input string) (*server.Config, error) {
// 	return getConfigFromBytes([]byte(input))
// }

// func getConfigFromBytes(input []byte) (*server.Config, error) {

// 	var yamlErr error
// 	var err error
// 	var config *server.Config

// 	// File could be YAML or JSON. Try both. Return yaml error if neither work
// 	config, yamlErr = server.ConfigFromYAML(input)
// 	if yamlErr == nil {
// 		return config, nil
// 	}

// 	config, err = server.ConfigFromJSON(input)
// 	if err == nil {
// 		return config, nil
// 	}

// 	return nil, yamlErr
// }
