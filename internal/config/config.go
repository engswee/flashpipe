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

	isNoSecretsFound, err := validateInputContainsNoSecrets(val)
	if !isNoSecretsFound {
		return "", fmt.Errorf("Secrets found in flag %v: %w", flagName, err)
	}

	return val, nil
}

func GetStringWithEnvExpandWithDefault(cmd *cobra.Command, flagName string, defaultValue string) (string, error) {
	val, err := GetStringWithEnvExpand(cmd, flagName)
	if err != nil {
		return "", fmt.Errorf("Secrets found in flag %v: %w", flagName, err)
	}

	if val == "" {
		return defaultValue, nil
	}

	return val, nil
}

func validateInputContainsNoSecrets(input string) (bool, error) {
	secretsConfigParams := []string{
		"tmn-userid",
		"tmn-password",
		"oauth-clientid",
		"oauth-clientsecret",
	}

	for _, secretsConfigParam := range secretsConfigParams {
		if viper.IsSet(secretsConfigParam) && strings.Contains(input, viper.GetString(secretsConfigParam)) {
			return false, fmt.Errorf("Input contains value of secret configuration parameter %v", secretsConfigParam)
		}
	}

	return true, nil
}
