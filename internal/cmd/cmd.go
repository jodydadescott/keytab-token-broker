package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/jodydadescott/kerberos-bridge/config"
	"github.com/jodydadescott/kerberos-bridge/internal/configloader"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var rootCmd = &cobra.Command{
	Use:   "kerberos-bridge",
	Short: "get kerberos keytabs with oauth tokens",
	Long: `Provides expiring kerberos keytabs to holders of bearer tokens
	by validating token is permitted keytab by policy. Policy is
	in the form of Open Policy Agent (OPA). Keytabs may be used
	to generate kerberos tickets and then discarded.`,
}

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "manage service",
}

var serviceInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "install service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return installService()
	},
}

var serviceRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return removeService()
	},
}

var serviceStartCmd = &cobra.Command{
	Use:   "start",
	Short: "start service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return startService()
	},
}

var serviceStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return stopService()
	},
}

var servicePauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "pause service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return pauseService()
	},
}

var serviceContinueCmd = &cobra.Command{
	Use:   "continue",
	Short: "continue service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return continueService()
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "manage configuration",
}

var configImportCmd = &cobra.Command{
	Use:   "import",
	Short: "import config from file or url",

	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) < 1 {
			return errors.New("config string required")
		}

		runtimeConfigString := args[0]
		// Need to verify string

		err := configloader.SetRuntimeConfigString(runtimeConfigString)
		if err != nil {
			return err
		}
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "set configuration",

	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) < 1 {
			return errors.New("config string required")
		}

		runtimeConfigString := args[0]
		// Need to verify string

		err := configloader.SetRuntimeConfigString(runtimeConfigString)
		if err != nil {
			return err
		}
		return nil
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "show config",

	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := configloader.GetRuntimeConfigString()
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

		configString := ""
		switch strings.ToLower(viper.GetString("format")) {

		case "yaml":
			configString = config.ExampleConfigYAML()
			break

		case "json":
			configString = config.ExampleConfigJSON()
			break

		case "":
			configString = config.ExampleConfigYAML()
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

var serverCmd = &cobra.Command{
	Use:   "start",
	Short: "start server",

	RunE: func(cmd *cobra.Command, args []string) error {

		config, err := configloader.GetConfigs()
		if err != nil {
			return err
		}

		serverConfig, err := config.ServerConfig()
		if err != nil {
			return err
		}

		zapConfig, err := config.ZapConfig()
		if err != nil {
			return err
		}

		logger, err := zapConfig.Build()
		if err != nil {
			return err
		}

		zap.ReplaceGlobals(logger)
		//defer logger.Sync()

		// 	//fmt.Println(fmt.Sprintf("Config:%s", config.JSON()))

		server, err := serverConfig.Build()
		if err != nil {
			return err
		}

		sig := make(chan os.Signal, 2)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig

		server.Shutdown()

		return nil
	},
}

// Execute ...
func Execute() {

	if runtime.GOOS == "windows" {
		isIntSess, err := isAnInteractiveSession()
		if err != nil {
			fmt.Fprintln(os.Stderr, fmt.Sprintf("failed to determine if we are running in an interactive session: %v", err))
			os.Exit(1)
		}
		if !isIntSess {
			runService()
			return
		}
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// serviceCmd

func init() {

	if runtime.GOOS == "windows" {
		serviceCmd.AddCommand(serviceInstallCmd, serviceRemoveCmd, serviceStartCmd, serviceStopCmd, servicePauseCmd, serviceContinueCmd)
		rootCmd.AddCommand(serviceCmd)
	}

	configCmd.AddCommand(configSetCmd, configShowCmd, configExampleCmd)
	rootCmd.AddCommand(configCmd, serverCmd)

	configExampleCmd.PersistentFlags().StringP("format", "", "", "output format in yaml or json; default is yaml")
	viper.BindPFlag("format", configExampleCmd.PersistentFlags().Lookup("format"))

	serverCmd.PersistentFlags().StringP("config", "", "", "configuration file")
	viper.BindPFlag("config", serverCmd.PersistentFlags().Lookup("config"))
}
