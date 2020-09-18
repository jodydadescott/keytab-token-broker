package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jodydadescott/kerberos-bridge/internal/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var rootCmd = &cobra.Command{
	Use:   "kerberos-bridge",
	Short: "get kerberos keytabs with oauth tokens",
	Long: `Provides expiring kerberos keytabs to holders of bearer tokens
	by validating token is permitted keytab by policy. Policy is
	in the form of Open Policy Agent (OPA). Keytabs may be used
	to generate kerberos tickets and then discarded.`,

	Run: func(cmd *cobra.Command, args []string) {

		// fmt.Println("log level:", viper.GetString("log_level"))
		// fmt.Println("log format:", viper.GetString("log_format"))
		// fmt.Println("log to:", viper.GetString("log_to"))

		zapConfig := &zap.Config{
			Development: false,
			Sampling: &zap.SamplingConfig{
				Initial:    100,
				Thereafter: 100,
			},
			EncoderConfig: zap.NewProductionEncoderConfig(),
		}

		switch viper.GetString("log_level") {

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

		default:
			fmt.Println("logging level must be debug, info, warn or error")
			os.Exit(2)
		}

		switch viper.GetString("log_format") {

		case "json":
			zapConfig.Encoding = "json"
			break

		case "console":
			zapConfig.Encoding = "console"
			break

		default:
			fmt.Println("logging format must be json or console")
			os.Exit(2)

		}

		for _, s := range strings.Split(viper.GetString("log_to"), ",") {
			zapConfig.OutputPaths = append(zapConfig.OutputPaths, s)
			zapConfig.ErrorOutputPaths = append(zapConfig.ErrorOutputPaths, s)
		}

		config := server.NewConfig()
		config.HTTPPort = viper.GetInt("httpport")
		config.HTTPSPort = viper.GetInt("httpsport")
		config.ListenInterface = viper.GetString("listen")
		config.Nonce.Lifetime = viper.GetInt("nonce_lifetime")
		config.Policy.Query = viper.GetString("policy_query")

		// Policy may be the policy or a file
		// policy := viper.GetString("policy_rego")

		// if policy == "" {
		// 	fmt.Println("Policy must specify a rego policy directly or reference a file that holds policy")
		// 	os.Exit(2)
		// }

		// if isFile(policy) {
		// 	zap.L().Debug(fmt.Sprintf("Policy is a file"))
		// 	f, err := os.Open(policy)
		// 	if err != nil {
		// 		panic(err)
		// 	}
		// 	defer f.Close()

		// 	reader := bufio.NewReader(f)
		// 	content, err := ioutil.ReadAll(reader)
		// 	if err != nil {
		// 		panic(err)
		// 	}
		// 	config.Policy.Policy = string(content)

		// } else {
		// 	config.Policy.Policy = policy
		// }

		config.Keytab.Lifetime = viper.GetInt("keytab_lifetime")

		for _, s := range strings.Split(viper.GetString("keytab_principals"), ",") {
			config.Keytab.Principals = append(config.Keytab.Principals, s)
		}

		logger, err := zapConfig.Build()
		if err != nil {
			panic(err)
		}
		zap.ReplaceGlobals(logger)

		zap.L().Debug(fmt.Sprintf("Running with config %s", config.JSON()))

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

func isFile(s string) bool {
	if _, err := os.Stat(s); err != nil {
		return false
	}
	return true
}

// Execute ...
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().IntP("nonce-lifetime", "", 0, "nonce lifetime")
	rootCmd.PersistentFlags().StringP("policy-query", "", "", "Policy query")
	rootCmd.PersistentFlags().StringP("policy-rego", "", "", "Policy rego (raw script or filename to script")
	rootCmd.PersistentFlags().IntP("keytab-lifetime", "", 0, "keytab lifetime")
	rootCmd.PersistentFlags().StringP("keytab-principals", "", "", "keytab principals (comma delimited)")
	rootCmd.PersistentFlags().IntP("httpport", "", 0, "http port (if enabled)")
	rootCmd.PersistentFlags().IntP("httpsport", "", 0, "https port (if enabled)")
	rootCmd.PersistentFlags().StringP("listen", "", "", "interface to listen on (http/https)")
	rootCmd.PersistentFlags().StringP("log-level", "", "info", "log level (debug, info=default, warn, error)")
	rootCmd.PersistentFlags().StringP("log-format", "", "json", "log format (json, text)")
	rootCmd.PersistentFlags().StringP("log-to", "", "stderr", "log to stderr or file")

	viper.BindPFlag("nonce_lifetime", rootCmd.PersistentFlags().Lookup("nonce-lifetime"))
	viper.BindPFlag("policy_query", rootCmd.PersistentFlags().Lookup("policy-query"))
	viper.BindPFlag("policy_rego", rootCmd.PersistentFlags().Lookup("policy-rego"))
	viper.BindPFlag("keytab_lifetime", rootCmd.PersistentFlags().Lookup("keytab-lifetime"))
	viper.BindPFlag("keytab_principals", rootCmd.PersistentFlags().Lookup("keytab-principals"))
	viper.BindPFlag("httpport", rootCmd.PersistentFlags().Lookup("httpport"))
	viper.BindPFlag("httpsport", rootCmd.PersistentFlags().Lookup("httpsport"))
	viper.BindPFlag("listen", rootCmd.PersistentFlags().Lookup("listen"))
	viper.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("log_format", rootCmd.PersistentFlags().Lookup("log-format"))
	viper.BindPFlag("log_to", rootCmd.PersistentFlags().Lookup("log-to"))

	viper.SetEnvPrefix("kbridge")
	viper.AutomaticEnv()

}
