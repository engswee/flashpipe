/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
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
	Short:   "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
	rootCmd.PersistentFlags().MarkHidden("config")

	rootCmd.PersistentFlags().String("location-mavenrepo", "", "Maven Repository Location (or set environment MVN_REPO_LOCATION)")
	rootCmd.PersistentFlags().MarkHidden("location-mavenrepo")
	viper.BindPFlag("location.mavenrepo", rootCmd.PersistentFlags().Lookup("location-mavenrepo"))

	rootCmd.PersistentFlags().String("location-flashpipe", "", "FlashPipe Location (or set environment FLASHPIPE_LOCATION)")
	rootCmd.PersistentFlags().MarkHidden("location-flashpipe")
	viper.BindPFlag("location.flashpipe", rootCmd.PersistentFlags().Lookup("location-flashpipe"))

	rootCmd.PersistentFlags().String("tmn-host", "", "Host for tenant management node of Cloud Integration excluding https:// (or set environment HOST_TMN)")
	viper.BindPFlag("tmn.host", rootCmd.PersistentFlags().Lookup("tmn-host"))
	rootCmd.PersistentFlags().String("tmn-userid", "", "User ID for Basic Auth (or set environment BASIC_USERID)")
	viper.BindPFlag("tmn.userid", rootCmd.PersistentFlags().Lookup("tmn-userid"))
	rootCmd.PersistentFlags().String("tmn-password", "", "Password for Basic Auth (or set environment BASIC_PASSWORD)")
	viper.BindPFlag("tmn.password", rootCmd.PersistentFlags().Lookup("tmn-password"))
	rootCmd.PersistentFlags().String("oauth-host", "", "Host for OAuth token server excluding https:// (or set environment HOST_OAUTH)")
	viper.BindPFlag("oauth.host", rootCmd.PersistentFlags().Lookup("oauth-host"))
	rootCmd.PersistentFlags().String("oauth-clientid", "", "Client ID for using OAuth (or set environment OAUTH_CLIENTID)")
	viper.BindPFlag("oauth.clientid", rootCmd.PersistentFlags().Lookup("oauth-clientid"))
	rootCmd.PersistentFlags().String("oauth-clientsecret", "", "Client Secret for using OAuth (or set environment OAUTH_CLIENTSECRET)")
	viper.BindPFlag("oauth.clientsecret", rootCmd.PersistentFlags().Lookup("oauth-clientsecret"))
	rootCmd.PersistentFlags().String("oauth-path", "", "Optional path for OAuth token server (default=/oauth/token), e.g /oauth2/api/v1/token for Neo (or set environment HOST_OAUTH_PATH)")
	viper.BindPFlag("oauth.clientsecret", rootCmd.PersistentFlags().Lookup("oauth-clientsecret"))

	rootCmd.PersistentFlags().StringP("debug-level", "d", "", "Debug level - FLASHPIPE, APACHE, ALL")
	// TODO - hide debug
	rootCmd.PersistentFlags().MarkHidden("debug")
	viper.BindPFlag("debug.level", rootCmd.PersistentFlags().Lookup("debug-level"))

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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
		viper.SetDefault("oauth.path", "/oauth/token")
		setOptionalVariable("oauth.path", "HOST_OAUTH_PATH")
	}

	viper.SetDefault("debug.flashpipe", "/tmp/log4j2-config/log4j2-debug-flashpipe.xml")
	viper.SetDefault("debug.apache", "/tmp/log4j2-config/log4j2-debug-apache.xml")
	viper.SetDefault("debug.all", "/tmp/log4j2-config/log4j2-debug-all.xml")

	debugLevel := viper.GetString("debug.level")
	if debugLevel != "" {
		log4jFile = viper.GetString(fmt.Sprintf("debug.%v", strings.ToLower(debugLevel)))
	}

	//for _, key := range viper.AllKeys() {
	//	fmt.Printf("%v = %v\n", key, viper.GetString(key))
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
