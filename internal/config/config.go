package config

import (
	"os"

	"github.com/spf13/cobra"
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

func GetDirectory(cmd *cobra.Command, flagName string) string {
	val := os.ExpandEnv(GetString(cmd, flagName))
	return val
}

func GetDirectoryWithDefault(cmd *cobra.Command, flagName string, defaultValue string) string {
	val := GetDirectory(cmd, flagName)
	if val == "" {
		return defaultValue
	}
	return val
}
