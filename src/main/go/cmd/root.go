package cmd

import (
	"fmt"
	"github.com/engswee/flashpipe/logger"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var version = "2.7.2-SNAPSHOT" // FLASHPIPE_VERSION
var mavenRepoLocation string
var flashpipeLocation string
var log4jFile string
var rootViper = viper.New()

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "flashpipe",
	Version: version,
	Short:   "FlashPipe - The CI/CD Companion for SAP Integration Suite",
	Long: `FlashPipe - The CI/CD Companion for SAP Integration Suite

FlashPipe is a CLI that is used to simplify the Build-To-Deploy cycle
for SAP Integration Suite by providing CI/CD capabilities for 
automating time-consuming manual tasks like:
- synchronising integration artifacts to Git
- uploading/updating integration artifacts to SAP Integration Suite
- deploy integration artifacts on SAP Integration Suite`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/flashpipe.yaml)")

	setPersistentStringFlagAndBind(rootViper, rootCmd, "location.mavenrepo", "", "Maven Repository Location [or set environment MVN_REPO_LOCATION]")
	setPersistentStringFlagAndBind(rootViper, rootCmd, "location.flashpipe", "", "FlashPipe Location [or set environment FLASHPIPE_LOCATION]")
	setPersistentStringFlagAndBind(rootViper, rootCmd, "tmn.host", "", "Host for tenant management node of Cloud Integration excluding https:// [or set environment HOST_TMN]")
	setPersistentStringFlagAndBind(rootViper, rootCmd, "tmn.userid", "", "User ID for Basic Auth [or set environment BASIC_USERID]")
	setPersistentStringFlagAndBind(rootViper, rootCmd, "tmn.password", "", "Password for Basic Auth [or set environment BASIC_PASSWORD]")
	setPersistentStringFlagAndBind(rootViper, rootCmd, "oauth.host", "", "Host for OAuth token server excluding https:// [or set environment HOST_OAUTH]")
	setPersistentStringFlagAndBind(rootViper, rootCmd, "oauth.clientid", "", "Client ID for using OAuth [or set environment OAUTH_CLIENTID]")
	setPersistentStringFlagAndBind(rootViper, rootCmd, "oauth.clientsecret", "", "Client Secret for using OAuth [or set environment OAUTH_CLIENTSECRET]")
	setPersistentStringFlagAndBind(rootViper, rootCmd, "oauth.path", "/oauth/token", "Path for OAuth token server, e.g /oauth2/api/v1/token for Neo [or set environment HOST_OAUTH_PATH]")
	setPersistentStringFlagAndBind(rootViper, rootCmd, "debug.level", "", "Debug level - FLASHPIPE, APACHE, ALL")

	rootCmd.PersistentFlags().MarkHidden("config")
	rootCmd.PersistentFlags().MarkHidden("location-mavenrepo")
	rootCmd.PersistentFlags().MarkHidden("location-flashpipe")
	rootCmd.PersistentFlags().MarkHidden("debug-level")

	//rootCmd.MarkFlagsRequiredTogether("tmn.userid", "tmn.password")
	//rootCmd.MarkFlagsRequiredTogether("oauth.host", "oauth.clientid", "oauth.clientsecret")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		rootViper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name "flashpipe" (without extension).
		rootViper.AddConfigPath(home)
		rootViper.SetConfigType("yaml")
		rootViper.SetConfigName("flashpipe")
	}

	// If a config file is found, read it in.
	if err := rootViper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", rootViper.ConfigFileUsed())
	}

	// Order config is read - CLI flag, env, config file, default

	rootViper.SetDefault("location.mavenrepo", "/usr/share/maven/ref/repository")
	rootViper.BindEnv("location.mavenrepo", "MVN_REPO_LOCATION")
	mavenRepoLocation = rootViper.GetString("location.mavenrepo")

	rootViper.SetDefault("location.flashpipe", fmt.Sprintf("%v/io/github/engswee/flashpipe/%v/flashpipe-%v.jar", mavenRepoLocation, version, version))
	rootViper.BindEnv("location.flashpipe", "FLASHPIPE_LOCATION")
	flashpipeLocation = rootViper.GetString("location.flashpipe")

	setMandatoryVariable(rootViper, "tmn.host", "HOST_TMN")
	if setOptionalVariable(rootViper, "oauth.host", "HOST_OAUTH") == "" {
		// Basic Authentication
		setMandatoryVariable(rootViper, "tmn.userid", "BASIC_USERID")
		setMandatoryVariable(rootViper, "tmn.password", "BASIC_PASSWORD")
	} else {
		// OAuth
		setMandatoryVariable(rootViper, "oauth.clientid", "OAUTH_CLIENTID")
		setMandatoryVariable(rootViper, "oauth.clientsecret", "OAUTH_CLIENTSECRET")
		setOptionalVariable(rootViper, "oauth.path", "HOST_OAUTH_PATH")
	}

	rootViper.SetDefault("debug.flashpipe", "/tmp/log4j2-config/log4j2-debug-flashpipe.xml")
	rootViper.SetDefault("debug.apache", "/tmp/log4j2-config/log4j2-debug-apache.xml")
	rootViper.SetDefault("debug.all", "/tmp/log4j2-config/log4j2-debug-all.xml")

	debugLevel := rootViper.GetString("debug.level")
	if debugLevel != "" {
		log4jFile = rootViper.GetString("debug." + strings.ToLower(debugLevel))
	}

	//for _, key := range rootViper.AllKeys() {
	//	fmt.Println(key, "=", rootViper.GetString(key))
	//}
}

func setMandatoryVariable(viperInstance *viper.Viper, viperKey string, envVarName string) string {
	viperInstance.BindEnv(viperKey, envVarName)
	val := viperInstance.GetString(viperKey)
	if val == "" {
		logger.Error("Mandatory environment variable", envVarName, "is not populated")
		os.Exit(1)
	} else {
		os.Setenv(envVarName, val) // TODO - remove when Java switch to CLI arguments
	}
	return val
}

func setOptionalVariable(viperInstance *viper.Viper, viperKey string, envVarName string) string {
	viperInstance.BindEnv(viperKey, envVarName)
	val := viperInstance.GetString(viperKey)
	if val != "" {
		os.Setenv(envVarName, val) // TODO - remove when Java switch to CLI arguments
	}
	return val
}

func setPersistentStringFlagAndBind(viperInstance *viper.Viper, cmd *cobra.Command, viperKey string, defaultValue string, usage string) {
	flagName := strings.ReplaceAll(viperKey, ".", "-")
	cmd.PersistentFlags().String(flagName, defaultValue, usage)
	viperInstance.BindPFlag(viperKey, cmd.PersistentFlags().Lookup(flagName))
}

func setStringFlagAndBind(viperInstance *viper.Viper, cmd *cobra.Command, viperKey string, defaultValue string, usage string) {
	flagName := strings.ReplaceAll(viperKey, ".", "-")
	cmd.Flags().String(flagName, defaultValue, usage)
	viperInstance.BindPFlag(viperKey, cmd.Flags().Lookup(flagName))
}

func setIntFlagAndBind(viperInstance *viper.Viper, cmd *cobra.Command, viperKey string, defaultValue int, usage string) {
	flagName := strings.ReplaceAll(viperKey, ".", "-")
	cmd.Flags().Int(flagName, defaultValue, usage)
	viperInstance.BindPFlag(viperKey, cmd.Flags().Lookup(flagName))
}

func setBoolFlagAndBind(viperInstance *viper.Viper, cmd *cobra.Command, viperKey string, defaultValue bool, usage string) {
	flagName := strings.ReplaceAll(viperKey, ".", "-")
	cmd.Flags().Bool(flagName, defaultValue, usage)
	viperInstance.BindPFlag(viperKey, cmd.Flags().Lookup(flagName))
}
