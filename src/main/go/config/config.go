package config

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
)

func GetMandatoryString(cmd *cobra.Command, flagName string) string {
	val := GetString(cmd, flagName)
	if val == "" {
		log.Error().Msgf("Mandatory parameter %v is not populated", flagName)
		os.Exit(1)
	}
	return val
}

func GetString(cmd *cobra.Command, flagName string) string {
	val, _ := cmd.Flags().GetString(flagName)
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
