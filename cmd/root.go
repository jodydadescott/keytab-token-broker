package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/jodydadescott/kerberos-bridge/internal/model"
	"github.com/jodydadescott/kerberos-bridge/internal/server"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "kerberos-bridge",
	Short: "get kerberos keytabs with oauth tokens",
	Long: `Provides expiring kerberos keytabs to holders of bearer tokens
	by validating token is permitted keytab by policy. Policy is
	in the form of Open Policy Agent (OPA). Keytabs may be used
	to generate kerberos tickets and then discarded.`,
}

var cmdServer = &cobra.Command{
	Use:   "server",
	Short: "start server instance",
	//Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		cobra.OnInitialize(initConfig)

		config, err := getServerConfigFromFile(cfgFile)
		if err != nil {
			panic(err)
		}

		logger, _ := zap.NewDevelopment()
		zap.ReplaceGlobals(logger)

		sig := make(chan os.Signal, 2)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

		server, err := server.NewServer(config)
		if err != nil {
			panic(err)
		}

		<-sig

		zap.L().Debug("Shutting Down")

		server.Shutdown()

	},
}

var cmdClient = &cobra.Command{
	Use:   "client",
	Short: "start client daemon",
	//Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Client")

		tokenString := "eyJhbGciOiJFUzI1NiIsImtpZCI6IjVlM2RjNWNhYjE1NjhjMDAwMWU3YzJjMSIsInR5cCI6IkpXVCJ9.eyJzZXJ2aWNlIjp7IkBjbG91ZDphd3M6YW1pLWlkIjoiYW1pLTA5OGYxNmFmYTllZGY0MGJlIiwiaWxvdmUiOiJ0aGU4MHMiLCJrZXl0YWIiOiJzdXBlcm1hbkBFWEFNUExFLkNPTSJ9LCJhdWQiOiJiIiwiZXhwIjoxNjAwMzgwMDA1LCJpYXQiOjE2MDAzNzY0MDUsImlzcyI6Imh0dHBzOi8vYXBpLmNvbnNvbGUuYXBvcmV0by5jb20vdi8xL25hbWVzcGFjZXMvNWRkYzM5NmI5ZmFjZWMwMDAxZDNjODg2L29hdXRoaW5mbyIsInN1YiI6IjVmNjNjZTIxYTIwNTdmMDAwMTI1ZmI3MSJ9.HJvVSTjH_YQb7_dS78HpfYHgmTYsnwuFzCb2j71oWtBg9eIefL3OBPMVhJAtCcTesPc9Rymlz91Gj0qxWA0x7A"

		token, err := model.TokenFromBase64(tokenString)
		if err != nil {
			panic(err)
		}

		fmt.Println(token.Valid())

	},
}

var cmdConfig = &cobra.Command{
	Use:   "config",
	Short: "generate example config",
	//Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		config := server.NewExampleConfig()
		fmt.Print(config.YAML())
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {

	cmdServer.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kerberos-bridge.yaml)")
	//cmdClient.PersistentFlags().Bool("viper", true, "use Viper for configuration")

	rootCmd.AddCommand(cmdServer, cmdClient, cmdConfig)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".kerberos-bridge")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func getServerConfigFromFile(filename string) (*server.Config, error) {

	fileBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config, err := server.ConfigFromYAML(fileBytes)
	if err != nil {
		config, err = server.ConfigFromJSON(fileBytes)
		if err != nil {
			return nil, err
		}
	}
	return config, nil
}
