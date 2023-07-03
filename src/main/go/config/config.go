package config

import (
	"fmt"
	"github.com/engswee/flashpipe/logger"
	"github.com/spf13/cobra"
	"os"
)

func GetRequiredFlagAsString(cmd *cobra.Command, flagName string) string {
	val := GetFlagAsString(cmd, flagName)
	if val == "" {
		logger.Error(fmt.Sprintf("Mandatory parameter %v is not populated", flagName))
		os.Exit(1)
	}
	return val
}

func GetFlagAsString(cmd *cobra.Command, flagName string) string {
	val, _ := cmd.Flags().GetString(flagName)
	return val
}

func GetFlagAsInt(cmd *cobra.Command, flagName string) int {
	val, _ := cmd.Flags().GetInt(flagName)
	return val
}

func GetFlagAsBool(cmd *cobra.Command, flagName string) bool {
	val, _ := cmd.Flags().GetBool(flagName)
	return val
}
