package cmd

// import (
// 	"errors"
// 	"strings"

// 	"github.com/spf13/cobra"
// 	"github.com/spf13/viper"
// 	"go.uber.org/zap"
// )

// func init() {

// 	EnforcerdCmd.PersistentFlags().StringVar(&cmdConfig.LogLevel, "log-level", "info", "Set the log-level between info, debug, trace.")

// 	RunCmd.Flags().StringP("cgroup", "", "", "cgroup")
// 	RunCmd.Flags().StringSliceP("label", "l", []string{}, "Label used when the command is running.")
// 	RunCmd.Flags().StringSliceP("tag", "T", []string{}, "Tag used when the command is running.")
// 	RunCmd.Flags().StringP("service-name", "", "", "Service name for the command.")
// 	RunCmd.Flags().StringSliceP("param", "p", []string{}, "Additional parameter to the command, in the form of key=value.")
// 	RunCmd.Flags().StringSliceP("port", "", []string{}, "Ports for the command.")
// 	RunCmd.Flags().StringP("ports", "", "", "Ports for the command.")
// 	RunCmd.Flags().Bool("host-policy", false, "Enforce policy for the host.")
// 	RunCmd.Flags().Bool("network-only", false, "Enforce policy for a receiving service.")
// 	RunCmd.Flags().Bool("autoport", false, "Auto port discovery for the PU.")

// 	// Deprecated
// 	RunCmd.Flags().MarkDeprecated("label", "please use --tag instead")  // nolint: errcheck
// 	RunCmd.Flags().MarkDeprecated("ports", "please use --port instead") // nolint: errcheck
// }

// var runCmdExample = []string{
// 	`  enforcerd run /bin/bash`,
// 	`  enforcerd run --tag role=server -- python -m SimpleHTTPServer`,
// }

// // RunCmd represents the `enforcerd run` command
// var RunCmd = &cobra.Command{
// 	Use:     "run <command>",
// 	Short:   "Run a command in a secure context.",
// 	Long:    `Run a command in a secure context.`,
// 	Example: strings.Join(runCmdExample, "\n"),
// 	PreRunE: func(cmd *cobra.Command, args []string) error {
// 		viper.BindPFlags(cmd.Flags()) // nolint: errcheck

// 		if len(args) == 0 {
// 			return errors.New("no command specified")
// 		}

// 		if buildflags.IsLegacyKernel() {
// 			return errors.New("enforcerd run not supported on older kernels")
// 		}

// 		return nil
// 	},
// 	RunE: func(cmd *cobra.Command, args []string) error { // nolint

// 		command := args[0]
// 		parameters := args[1:]
// 		cgroup := viper.GetString("cgroup")
// 		serviceName := viper.GetString("service-name")
// 		ports := viper.GetStringSlice("port")
// 		hostPolicy := viper.GetBool("host-policy")
// 		networkOnly := viper.GetBool("network-only")
// 		autoPort := viper.GetBool("autoport")

// 		if viper.GetString("ports") != "" { // deprecated: remove this block if flag --ports is removed
// 			ports = append(ports, strings.Split(viper.GetString("ports"), ",")...)
// 		}

// 		tags := viper.GetStringSlice("tag")
// 		tags = append(tags, viper.GetStringSlice("label")...) // deprecated: remove this line if flag --label is removed

// 		zap.L().Debug("Execute command",
// 			zap.String("command", command),
// 			zap.Strings("parameters", args[1:]),
// 			zap.String("cgroup", cgroup),
// 			zap.String("serviceName", serviceName),
// 			zap.Strings("ports", ports),
// 			zap.Strings("tags", tags),
// 		)

// 		services, err := systemdutil.ParseServices(ports)
// 		if err != nil {
// 			return err
// 		}

// 		c := &systemdutil.CLIRequest{
// 			Request:     systemdutil.CreateRequest,
// 			Cgroup:      cgroup,
// 			Labels:      tags,
// 			ServiceName: serviceName,
// 			Services:    services,
// 			HostPolicy:  hostPolicy,
// 			NetworkOnly: networkOnly,
// 			Executable:  command,
// 			Parameters:  parameters,
// 			AutoPort:    autoPort,
// 		}

// 		p := systemdutil.NewRequestProcessor()
// 		return p.CreateAndRun(c)
// 	},
// }
