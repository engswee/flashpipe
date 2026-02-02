package cmd

import (
	"github.com/engswee/flashpipe/internal/config"
	"github.com/spf13/cobra"
)

// Helper functions for Partner Directory commands to support reading
// configuration from both command-line flags and nested config file keys
// These are thin wrappers around the config package functions for backward compatibility

// getConfigStringWithFallback reads a string value from command flag,
// falling back to a nested config key if the flag wasn't explicitly set
func getConfigStringWithFallback(cmd *cobra.Command, flagName, configKey string) string {
	return config.GetStringWithFallback(cmd, flagName, configKey)
}

// getConfigBoolWithFallback reads a bool value from command flag,
// falling back to a nested config key if the flag wasn't explicitly set
func getConfigBoolWithFallback(cmd *cobra.Command, flagName, configKey string) bool {
	return config.GetBoolWithFallback(cmd, flagName, configKey)
}

// getConfigStringSliceWithFallback reads a string slice value from command flag,
// falling back to a nested config key if the flag wasn't explicitly set
func getConfigStringSliceWithFallback(cmd *cobra.Command, flagName, configKey string) []string {
	return config.GetStringSliceWithFallback(cmd, flagName, configKey)
}

// contains checks if a string slice contains a specific string
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
