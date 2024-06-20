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
