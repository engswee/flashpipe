package cmd

import (
	"fmt"
	"log"
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

	setPersistentStringFlagAndBind(rootCmd, "location.mavenrepo", "", "Maven Repository Location [or set environment MVN_REPO_LOCATION]")
	setPersistentStringFlagAndBind(rootCmd, "location.flashpipe", "", "FlashPipe Location [or set environment FLASHPIPE_LOCATION]")
	setPersistentStringFlagAndBind(rootCmd, "tmn.host", "", "Host for tenant management node of Cloud Integration excluding https:// [or set environment HOST_TMN]")
	setPersistentStringFlagAndBind(rootCmd, "tmn.userid", "", "User ID for Basic Auth [or set environment BASIC_USERID]")
	setPersistentStringFlagAndBind(rootCmd, "tmn.password", "", "Password for Basic Auth [or set environment BASIC_PASSWORD]")
	setPersistentStringFlagAndBind(rootCmd, "oauth.host", "", "Host for OAuth token server excluding https:// [or set environment HOST_OAUTH]")
	setPersistentStringFlagAndBind(rootCmd, "oauth.clientid", "", "Client ID for using OAuth [or set environment OAUTH_CLIENTID]")
	setPersistentStringFlagAndBind(rootCmd, "oauth.clientsecret", "", "Client Secret for using OAuth [or set environment OAUTH_CLIENTSECRET]")
	setPersistentStringFlagAndBind(rootCmd, "oauth.path", "/oauth/token", "Path for OAuth token server, e.g /oauth2/api/v1/token for Neo [or set environment HOST_OAUTH_PATH]")
	setPersistentStringFlagAndBind(rootCmd, "debug.level", "", "Debug level - FLASHPIPE, APACHE, ALL")

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

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	// Order config is read - CLI flag, env, config file, default

	viper.SetDefault("location.mavenrepo", "/usr/share/maven/ref/repository")
	viper.BindEnv("location.mavenrepo", "MVN_REPO_LOCATION")
	mavenRepoLocation = viper.GetString("location.mavenrepo")

	viper.SetDefault("location.flashpipe", fmt.Sprintf("%v/io/github/engswee/flashpipe/%v/flashpipe-%v.jar", mavenRepoLocation, version, version))
	viper.BindEnv("location.flashpipe", "FLASHPIPE_LOCATION")
	flashpipeLocation = viper.GetString("location.flashpipe")

	setMandatoryVariable("tmn.host", "HOST_TMN")
	if setOptionalVariable("oauth.host", "HOST_OAUTH") == "" {
		// Basic Authentication
		setMandatoryVariable("tmn.userid", "BASIC_USERID")
		setMandatoryVariable("tmn.password", "BASIC_PASSWORD")
	} else {
		// OAuth
		setMandatoryVariable("oauth.clientid", "OAUTH_CLIENTID")
		setMandatoryVariable("oauth.clientsecret", "OAUTH_CLIENTSECRET")
		//viper.SetDefault("oauth.path", "/oauth/token")
		setOptionalVariable("oauth.path", "HOST_OAUTH_PATH")
	}

	viper.SetDefault("debug.flashpipe", "/tmp/log4j2-config/log4j2-debug-flashpipe.xml")
	viper.SetDefault("debug.apache", "/tmp/log4j2-config/log4j2-debug-apache.xml")
	viper.SetDefault("debug.all", "/tmp/log4j2-config/log4j2-debug-all.xml")

	debugLevel := viper.GetString("debug.level")
	if debugLevel != "" {
		log4jFile = viper.GetString("debug." + strings.ToLower(debugLevel))
	}

	//for _, key := range viper.AllKeys() {
	//	fmt.Println(key, "=", viper.GetString(key))
	//}
}

func setMandatoryVariable(viperKey string, envVarName string) string {
	viper.BindEnv(viperKey, envVarName)
	val := viper.GetString(viperKey)
	if val == "" {
		log.Fatalf("[ERROR] 🛑 Mandatory environment variable %v is not populated", envVarName)
	} else {
		os.Setenv(envVarName, val) // TODO - remove when Java switch to CLI arguments
	}
	return val
}

func setOptionalVariable(viperKey string, envVarName string) string {
	viper.BindEnv(viperKey, envVarName)
	val := viper.GetString(viperKey)
	if val != "" {
		os.Setenv(envVarName, val) // TODO - remove when Java switch to CLI arguments
	}
	return val
}

func setPersistentStringFlagAndBind(cmd *cobra.Command, viperKey string, defaultValue string, usage string) {
	flagName := strings.ReplaceAll(viperKey, ".", "-")
	cmd.PersistentFlags().String(flagName, defaultValue, usage)
	viper.BindPFlag(viperKey, cmd.PersistentFlags().Lookup(flagName))
}

func setStringFlagAndBind(cmd *cobra.Command, viperKey string, defaultValue string, usage string) {
	flagName := strings.ReplaceAll(viperKey, ".", "-")
	cmd.Flags().String(flagName, defaultValue, usage)
	viper.BindPFlag(viperKey, cmd.Flags().Lookup(flagName))
}

func setIntFlagAndBind(cmd *cobra.Command, viperKey string, defaultValue int, usage string) {
	flagName := strings.ReplaceAll(viperKey, ".", "-")
	cmd.Flags().Int(flagName, defaultValue, usage)
	viper.BindPFlag(viperKey, cmd.Flags().Lookup(flagName))
}

func setBoolFlagAndBind(cmd *cobra.Command, viperKey string, defaultValue bool, usage string) {
	flagName := strings.ReplaceAll(viperKey, ".", "-")
	cmd.Flags().Bool(flagName, defaultValue, usage)
	viper.BindPFlag(viperKey, cmd.Flags().Lookup(flagName))
}
