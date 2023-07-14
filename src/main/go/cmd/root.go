package cmd

import (
	"fmt"
	"github.com/engswee/flashpipe/config"
	"github.com/spf13/pflag"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewCmdRoot() *cobra.Command {
	var version = "3.0.0" // FLASHPIPE_VERSION

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
	rootCmd.PersistentFlags().String("tmn-host", "", "Host for tenant management node of Cloud Integration excluding https:// [or set environment HOST_TMN]")
	rootCmd.PersistentFlags().String("tmn-userid", "", "User ID for Basic Auth [or set environment BASIC_USERID]")
	rootCmd.PersistentFlags().String("tmn-password", "", "Password for Basic Auth [or set environment BASIC_PASSWORD]")
	rootCmd.PersistentFlags().String("oauth-host", "", "Host for OAuth token server excluding https:// [or set environment HOST_OAUTH]")
	rootCmd.PersistentFlags().String("oauth-clientid", "", "Client ID for using OAuth [or set environment OAUTH_CLIENTID]")
	rootCmd.PersistentFlags().String("oauth-clientsecret", "", "Client Secret for using OAuth [or set environment OAUTH_CLIENTSECRET]")
	rootCmd.PersistentFlags().String("oauth-path", "/oauth/token", "Path for OAuth token server, e.g /oauth2/api/v1/token for Neo [or set environment HOST_OAUTH_PATH]")

	rootCmd.PersistentFlags().Bool("debug", false, "Show debug logs")

	// TODO - to be removed once fully ported from Java to Go
	rootCmd.PersistentFlags().String("location.mavenrepo", "", "Maven Repository Location [or set environment MVN_REPO_LOCATION]")
	rootCmd.PersistentFlags().String("location.flashpipe", "", "FlashPipe Location [or set environment FLASHPIPE_LOCATION]")

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
	if err != nil {
		os.Exit(1)
	}
}

//func init() {

// Execution sequence
//deploy init() -> execute init() in alphabetically order of file - init bind flags to viper config
//root init()
//root Execute()
//root initconfig()
//rootCmd PersistentPreRun - runs with the called command, i.e. sync, deploy
//deployCmd Run

//https://github.com/carolynvs/stingoftheviper/blob/main/main.go
//https://carolynvanslyck.com/blog/2020/08/sting-of-the-viper/

//cobra.OnInitialize(initConfig)
// Here you will define your flags and configuration settings.
// Cobra supports persistent flags, which, if defined here,
// will be global for your application.

//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/flashpipe.yaml)")
//
//// Define cobra flags, the default value has the lowest (least significant) precedence
//rootCmd.PersistentFlags().String("tmn-host", "", "Host for tenant management node of Cloud Integration excluding https:// [or set environment HOST_TMN]")
//rootCmd.PersistentFlags().String("tmn-userid", "", "User ID for Basic Auth [or set environment BASIC_USERID]")
//rootCmd.PersistentFlags().String("tmn-password", "", "Password for Basic Auth [or set environment BASIC_PASSWORD]")
//rootCmd.PersistentFlags().String("oauth-host", "", "Host for OAuth token server excluding https:// [or set environment HOST_OAUTH]")
//rootCmd.PersistentFlags().String("oauth-clientid", "", "Client ID for using OAuth [or set environment OAUTH_CLIENTID]")
//rootCmd.PersistentFlags().String("oauth-clientsecret", "", "Client Secret for using OAuth [or set environment OAUTH_CLIENTSECRET]")
//rootCmd.PersistentFlags().String("oauth-path", "/oauth/token", "Path for OAuth token server, e.g /oauth2/api/v1/token for Neo [or set environment HOST_OAUTH_PATH]")
//

//flagName := strings.ReplaceAll("location.mavenrepo", ".", "-")
//rootCmd.PersistentFlags().String(flagName, "", "Maven Repository Location [or set environment MVN_REPO_LOCATION]")
//rootViper.BindPFlag("location.mavenrepo", rootCmd.PersistentFlags().Lookup(flagName))

//setPersistentStringFlagAndBind(rootViper, rootCmd, "location.mavenrepo", "", "Maven Repository Location [or set environment MVN_REPO_LOCATION]")
//setPersistentStringFlagAndBind(rootViper, rootCmd, "location.flashpipe", "", "FlashPipe Location [or set environment FLASHPIPE_LOCATION]")
//setPersistentStringFlagAndBind(rootViper, rootCmd, "debug.level", "", "Debug level - FLASHPIPE, APACHE, ALL")

//rootCmd.PersistentFlags().MarkHidden("config")
//rootCmd.PersistentFlags().MarkHidden("location-mavenrepo")
//rootCmd.PersistentFlags().MarkHidden("location-flashpipe")
//rootCmd.PersistentFlags().MarkHidden("debug-level")

//rootCmd.MarkFlagsRequiredTogether("tmn.userid", "tmn.password")
//rootCmd.MarkFlagsRequiredTogether("oauth.host", "oauth.clientid", "oauth.clientsecret")

// Cobra also supports local flags, which will only run
// when this action is called directly.
//}

func initializeConfig(cmd *cobra.Command) error {
	//v := viper.New()
	cfgFile := config.GetString(cmd, "config")
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name "flashpipe" (without extension).
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

	// TODO - handle environment variable prefix to FLASHPIPE?
	viper.SetEnvPrefix("FLASHPIPE")

	// Environment variables can't have dashes in them, so bind them to their equivalent
	// keys with underscores, e.g. --favorite-color to STING_FAVORITE_COLOR
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Bind to environment variables
	// Works great for simple config names, but needs help for names
	// like --favorite-color which we fix in the bindFlags function
	viper.AutomaticEnv()

	// Bind the current command's flags to viper
	bindFlags(cmd)
	return nil
}

// Bind each cobra flag to its associated viper configuration (config file and environment variable)
func bindFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// TODO - handle environment variable name??
		// Determine the naming convention of the flags when represented in the config file
		configName := f.Name

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && viper.IsSet(configName) {
			val := viper.Get(configName)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
	// Set debug flag from command line to viper
	viper.Set("debug", config.GetBool(cmd, "debug"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// This is only executed after root.Execute() is called
	//if cfgFile != "" {
	//	// Use config file from the flag.
	//	rootViper.SetConfigFile(cfgFile)
	//} else {
	//	// Find home directory.
	//	home, err := os.UserHomeDir()
	//	cobra.CheckErr(err)
	//
	//	// Search config in home directory with name "flashpipe" (without extension).
	//	rootViper.AddConfigPath(home)
	//	rootViper.SetConfigType("yaml")
	//	rootViper.SetConfigName("flashpipe")
	//}
	//
	//// If a config file is found, read it in.
	//if err := rootViper.ReadInConfig(); err == nil {
	//	fmt.Fprintln(os.Stderr, "Using config file:", rootViper.ConfigFileUsed())
	//}
	//
	//// Order config is read - CLI flag, env, config file, default
	//
	//rootViper.SetDefault("location.mavenrepo", "/usr/share/maven/ref/repository")
	//rootViper.BindEnv("location.mavenrepo", "MVN_REPO_LOCATION")
	//mavenRepoLocation = rootViper.GetString("location.mavenrepo")
	//
	//rootViper.SetDefault("location.flashpipe", fmt.Sprintf("%v/io/github/engswee/flashpipe/%v/flashpipe-%viper.jar", mavenRepoLocation, version, version))
	//rootViper.BindEnv("location.flashpipe", "FLASHPIPE_LOCATION")
	//flashpipeLocation = rootViper.GetString("location.flashpipe")
	//
	////tmnHost = setMandatoryVariable(rootViper, "tmn.host", "HOST_TMN")
	//oauthHost = setOptionalVariable(rootViper, "oauth.host", "HOST_OAUTH")
	//if oauthHost == "" {
	//	// Basic Authentication
	//	basicUserId = setMandatoryVariable(rootViper, "tmn.userid", "BASIC_USERID")
	//	basicPassword = setMandatoryVariable(rootViper, "tmn.password", "BASIC_PASSWORD")
	//} else {
	//	// OAuth
	//	oauthClientId = setMandatoryVariable(rootViper, "oauth.clientid", "OAUTH_CLIENTID")
	//	oauthClientSecret = setMandatoryVariable(rootViper, "oauth.clientsecret", "OAUTH_CLIENTSECRET")
	//	oauthTokenPath = setOptionalVariable(rootViper, "oauth.path", "HOST_OAUTH_PATH")
	//}
	//
	//rootViper.SetDefault("debug.flashpipe", "/tmp/log4j2-config/log4j2-debug-flashpipe.xml")
	//rootViper.SetDefault("debug.apache", "/tmp/log4j2-config/log4j2-debug-apache.xml")
	//rootViper.SetDefault("debug.all", "/tmp/log4j2-config/log4j2-debug-all.xml")
	//
	//debugLevel := rootViper.GetString("debug.level")
	//if debugLevel != "" {
	//	log4jFile = rootViper.GetString("debug." + strings.ToLower(debugLevel))
	//}

	//for _, key := range rootViper.AllKeys() {
	//	fmt.Println(key, "=", rootViper.GetString(key))
	//}
}
