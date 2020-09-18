package cmd

// import (
// 	"fmt"
// 	"io/ioutil"
// 	"os"
// 	"os/signal"
// 	"runtime"
// 	"syscall"

// 	"github.com/jodydadescott/kerberos-bridge/internal/server"
// 	"github.com/spf13/cobra"
// 	"github.com/spf13/viper"
// 	"go.uber.org/zap"
// )

// var rootCmd = &cobra.Command{
// 	Use:   "kerberos-bridge",
// 	Short: "get kerberos keytabs with oauth tokens",
// 	Long: `Provides expiring kerberos keytabs to holders of bearer tokens
// 	by validating token is permitted keytab by policy. Policy is
// 	in the form of Open Policy Agent (OPA). Keytabs may be used
// 	to generate kerberos tickets and then discarded.`,

// 	Run: func(cmd *cobra.Command, args []string) {

// 		config, err := getConfig()
// 		if err != nil {
// 			panic(err)
// 		}

// 		logger, _ := zap.NewDevelopment()
// 		zap.ReplaceGlobals(logger)

// 		sig := make(chan os.Signal, 2)
// 		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

// 		server, err := server.NewServer(config)
// 		if err != nil {
// 			panic(err)
// 		}

// 		<-sig

// 		zap.L().Debug("Shutting Down")

// 		server.Shutdown()

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
// 	rootCmd.Flags().StringP("config", "", "", "config file")
// 	rootCmd.Flags().StringP("nonce-lifetime", "", "", "nonce lifetime")
// 	rootCmd.Flags().StringP("policy-query", "", "", "Policy query")
// 	rootCmd.Flags().StringP("policy-rego", "", "", "Policy rego (raw script or filename to script")
// 	rootCmd.Flags().StringP("keytab-lifetime", "", "", "keytab lifetime")
// 	rootCmd.Flags().StringP("keytab-principals", "", "", "keytab principals (comma delimited)")
// 	rootCmd.Flags().StringP("http-port", "", "", "http port (if enabled)")
// 	rootCmd.Flags().StringP("https-port", "", "", "https port (if enabled)")
// 	rootCmd.Flags().StringP("listen-interface", "", "", "interface to listen on (http/https)")
// 	rootCmd.Flags().StringP("log-level", "", "info", "log level (info, debug, trace)")
// 	rootCmd.Flags().StringP("log-format", "", "json", "log format (json, text)")
// 	rootCmd.Flags().BoolP("log-to-console", "", true, "log to console")
// }

// func initConfig() {
// 	viper.AutomaticEnv()
// }

// func getConfig() (*server.Config, error) {

// 	configFile := viper.GetString("config")

// 	if configFile != "" {
// 		if fileExist(configFile) {
// 			return getConfigFromFile(configFile)
// 		}
// 		return nil, fmt.Errorf("File does not exist")
// 	}

// 	if runtime.GOOS != "windows" {
// 		configFile = "C:\\Program Files\\kerberos-bridge\\config.yaml"
// 		if fileExist(configFile) {
// 			return getConfigFromFile(configFile)
// 		}
// 	} else {
// 		configFile = "/etc/kerberos-bridge/config.yaml"
// 		if fileExist(configFile) {
// 			return getConfigFromFile(configFile)
// 		}

// 		configFile = "/etc/kerberos-bridge.yaml"
// 		if fileExist(configFile) {
// 			return getConfigFromFile(configFile)
// 		}
// 	}

// 	return nil, fmt.Errorf("Config not found")
// }

// func fileExist(filename string) bool {
// 	if _, err := os.Stat(filename); err != nil {
// 		return false
// 	}
// 	return true
// }

// func getConfigFromFile(filename string) (*server.Config, error) {

// 	fileBytes, err := ioutil.ReadFile(filename)
// 	if err != nil {
// 		return nil, err
// 	}

// 	config, err := server.ConfigFromYAML(fileBytes)
// 	if err != nil {
// 		config, err = server.ConfigFromJSON(fileBytes)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	return config, nil
// }
