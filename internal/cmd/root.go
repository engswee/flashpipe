package cmd

import (
	"context"
	"fmt"
	"github.com/engswee/flashpipe/internal/analytics"
	"github.com/engswee/flashpipe/internal/config"
	"github.com/engswee/flashpipe/internal/logger"
	"github.com/go-errors/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"strings"
)

func NewCmdRoot() *cobra.Command {
	var version = "3.2.0" // FLASHPIPE_VERSION

	// rootCmd represents the base command when called without any subcommands
	rootCmd := &cobra.Command{
		Use:     "flashpipe",
		Version: version,
		Short:   "FlashPipe - The CI/CD Companion for SAP Integration Suite",
		Long: `FlashPipe - The CI/CD Companion for SAP Integration Suite

FlashPipe is a CLI that is used to simplify the Build-To-Deploy cycle
for SAP Integration Suite by providing CI/CD capabilities for 
automating time-consuming manual tasks like:
- synchronising integration artifacts to Git
- creating/updating integration artifacts to SAP Integration Suite
- deploying integration artifacts on SAP Integration Suite`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// You can bind cobra and viper in a few locations, but PersistencePreRunE on the root command works well
			return initializeConfig(cmd)
		},
	}

	rootCmd.PersistentFlags().String("config", "", "config file (default is $HOME/flashpipe.yaml)")

	// Define cobra flags, the default value has the lowest (least significant) precedence
	rootCmd.PersistentFlags().String("tmn-host", "", "Host for tenant management node of Cloud Integration excluding https://")
	rootCmd.PersistentFlags().String("tmn-userid", "", "User ID for Basic Auth")
	rootCmd.PersistentFlags().String("tmn-password", "", "Password for Basic Auth")
	rootCmd.PersistentFlags().String("oauth-host", "", "Host for OAuth token server excluding https:// ")
	rootCmd.PersistentFlags().String("oauth-clientid", "", "Client ID for using OAuth")
	rootCmd.PersistentFlags().String("oauth-clientsecret", "", "Client Secret for using OAuth")
	rootCmd.PersistentFlags().String("oauth-path", "/oauth/token", "Path for OAuth token server")

	rootCmd.PersistentFlags().Bool("debug", false, "Show debug logs")

	_ = rootCmd.MarkPersistentFlagRequired("tmn-host")
	rootCmd.MarkFlagsRequiredTogether("tmn-userid", "tmn-password")
	rootCmd.MarkFlagsRequiredTogether("oauth-host", "oauth-clientid", "oauth-clientsecret")

	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	rootCmd := NewCmdRoot()
	rootCmd.AddCommand(NewDeployCommand())
	rootCmd.AddCommand(NewSyncCommand())
	updateCmd := NewUpdateCommand()
	updateCmd.AddCommand(NewArtifactCommand())
	updateCmd.AddCommand(NewPackageCommand())
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(NewSnapshotCommand())

	err := rootCmd.Execute()
	// TODO - Log to analytics - don't log if it's usage is printed out.... need to probably segregate based on error
	analytics.Log(rootCmd)

	if err != nil {
		// Display stack trace based on type of error
		var msg string
		switch err.(type) {
		case *errors.Error:
			msg = err.(*errors.Error).ErrorStack()
		default:
			msg = err.Error()
		}
		log.Fatal().Msg(msg)
	}
}

func getRootCommand(cmd *cobra.Command) *cobra.Command {
	return getParentCommand(cmd)
}

func getParentCommand(cmd *cobra.Command) *cobra.Command {
	if cmd.Parent() == nil {
		return cmd
	}
	return getParentCommand(cmd.Parent())
}

func initializeConfig(cmd *cobra.Command) error {
	root := getRootCommand(cmd)
	root.SetContext(context.WithValue(cmd.Context(), "command", cmd.Name()))
	cfgFile := config.GetString(cmd, "config")
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name "flashpipe.yaml".
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName("flashpipe")
	}

	if err := viper.ReadInConfig(); err != nil {
		// It's okay if there isn't a config file
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	viper.SetEnvPrefix("FLASHPIPE")

	// Environment variables can't have dashes in them, so bind them to their equivalent
	// keys with underscores, e.g. --artifact-id to FLASHPIPE_ARTIFACT_ID
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Bind to environment variables
	viper.AutomaticEnv()

	// Bind the current command's flags to viper
	bindFlags(cmd)

	// Set debug flag from command line to viper
	if !viper.IsSet("debug") {
		viper.Set("debug", config.GetBool(cmd, "debug"))
	}

	if config.GetString(cmd, "oauth-host") == "" && config.GetString(cmd, "tmn-userid") == "" {
		return fmt.Errorf("required flag \"tmn-userid\" (Basic Auth) or \"oauth-host\" (OAuth) not set")
	}

	logger.InitConsoleLogger(viper.GetBool("debug"))

	return nil
}

// Bind each cobra flag to its associated viper configuration (config file and environment variable)
func bindFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		configName := f.Name
		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && viper.IsSet(configName) {
			val := viper.Get(configName)
			cmd.Flags().Set(configName, fmt.Sprintf("%v", val))
		}
	})
}
