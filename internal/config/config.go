package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func GetString(cmd *cobra.Command, flagName string) string {
	val, _ := cmd.Flags().GetString(flagName)
	return val
}

func GetStringWithDefault(cmd *cobra.Command, flagName string, defaultValue string) string {
	val, _ := cmd.Flags().GetString(flagName)
	if val == "" {
		return defaultValue
	}
	return val
}

func GetStringSlice(cmd *cobra.Command, flagName string) []string {
	val, _ := cmd.Flags().GetStringSlice(flagName)
	return val
}

func GetInt(cmd *cobra.Command, flagName string) int {
	val, _ := cmd.Flags().GetInt(flagName)
	return val
}

func GetBool(cmd *cobra.Command, flagName string) bool {
	val, _ := cmd.Flags().GetBool(flagName)
	return val
}

func GetStringWithEnvExpand(cmd *cobra.Command, flagName string) (string, error) {
	val := os.ExpandEnv(GetString(cmd, flagName))

	isNoSensContFound, err := verifyNoSensitiveContent(val)
	if !isNoSensContFound {
		return "", fmt.Errorf("Sensitive content found in flag %v: %w", flagName, err)
	}

	return val, nil
}

func GetStringWithEnvExpandWithDefault(cmd *cobra.Command, flagName string, defaultValue string) (string, error) {
	val, err := GetStringWithEnvExpand(cmd, flagName)
	if err != nil {
		return "", fmt.Errorf("Sensitive content found in flag %v: %w", flagName, err)
	}

	if val == "" {
		return defaultValue, nil
	}

	return val, nil
}

func verifyNoSensitiveContent(input string) (bool, error) {
	sensContConfigParams := []string{
		"tmn-userid",
		"tmn-password",
		"oauth-clientid",
		"oauth-clientsecret",
	}

	for _, sensContConfigParam := range sensContConfigParams {
		if viper.IsSet(sensContConfigParam) && strings.Contains(input, viper.GetString(sensContConfigParam)) {
			return false, fmt.Errorf("Input contains sensitive content from configuration parameter %v", sensContConfigParam)
		}
	}

	return true, nil
}

// GetStringWithFallback reads a string value from command flag,
// falling back to a nested config key if the flag wasn't explicitly set
func GetStringWithFallback(cmd *cobra.Command, flagName, configKey string) string {
	// Check if flag was explicitly set on command line
	if cmd.Flags().Changed(flagName) {
		return GetString(cmd, flagName)
	}

	// Try to get from nested config key
	if viper.IsSet(configKey) {
		return viper.GetString(configKey)
	}

	// Fall back to flag default
	return GetString(cmd, flagName)
}

// GetBoolWithFallback reads a bool value from command flag,
// falling back to a nested config key if the flag wasn't explicitly set
func GetBoolWithFallback(cmd *cobra.Command, flagName, configKey string) bool {
	// Check if flag was explicitly set on command line
	if cmd.Flags().Changed(flagName) {
		return GetBool(cmd, flagName)
	}

	// Try to get from nested config key
	if viper.IsSet(configKey) {
		return viper.GetBool(configKey)
	}

	// Fall back to flag default
	return GetBool(cmd, flagName)
}

// GetIntWithFallback reads an int value from command flag,
// falling back to a nested config key if the flag wasn't explicitly set
func GetIntWithFallback(cmd *cobra.Command, flagName, configKey string) int {
	// Check if flag was explicitly set on command line
	if cmd.Flags().Changed(flagName) {
		return GetInt(cmd, flagName)
	}

	// Try to get from nested config key
	if viper.IsSet(configKey) {
		return viper.GetInt(configKey)
	}

	// Fall back to flag default
	return GetInt(cmd, flagName)
}

// GetStringSliceWithFallback reads a string slice value from command flag,
// falling back to a nested config key if the flag wasn't explicitly set
func GetStringSliceWithFallback(cmd *cobra.Command, flagName, configKey string) []string {
	// Check if flag was explicitly set on command line
	if cmd.Flags().Changed(flagName) {
		return GetStringSlice(cmd, flagName)
	}

	// Try to get from nested config key
	if viper.IsSet(configKey) {
		return viper.GetStringSlice(configKey)
	}

	// Fall back to flag default
	return GetStringSlice(cmd, flagName)
}

// GetStringWithEnvExpandAndFallback reads a string value with environment variable expansion,
// falling back to a nested config key if the flag wasn't explicitly set
func GetStringWithEnvExpandAndFallback(cmd *cobra.Command, flagName, configKey string) (string, error) {
	var val string

	// Check if flag was explicitly set on command line
	if cmd.Flags().Changed(flagName) {
		val = GetString(cmd, flagName)
	} else if viper.IsSet(configKey) {
		// Try to get from nested config key
		val = viper.GetString(configKey)
	} else {
		// Fall back to flag default
		val = GetString(cmd, flagName)
	}

	// Expand environment variables
	val = os.ExpandEnv(val)

	isNoSensContFound, err := verifyNoSensitiveContent(val)
	if !isNoSensContFound {
		return "", fmt.Errorf("Sensitive content found in flag %v: %w", flagName, err)
	}

	return val, nil
}
