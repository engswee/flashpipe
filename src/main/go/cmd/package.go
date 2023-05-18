/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/engswee/flashpipe/runner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// packageCmd represents the package command
var packageCmd = &cobra.Command{
	Use:     "package",
	Aliases: []string{"pkg"},
	Short:   "Update integration package",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("[INFO] Executing update package command")

		setMandatoryVariable("package.file", "PACKAGE_FILE")
		setOptionalVariable("package.id.override", "PACKAGE_ID")
		setOptionalVariable("package.name.override", "PACKAGE_NAME")

		runner.JavaCmd("io.github.engswee.flashpipe.cpi.exec.UpdateIntegrationPackage", mavenRepoLocation, flashpipeLocation, log4jFile)
	},
}

func init() {
	updateCmd.AddCommand(packageCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// packageCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	packageCmd.Flags().StringP("package-file", "p", "", "Path to location of package file (or set environment PACKAGE_FILE)")
	viper.BindPFlag("package.file", packageCmd.Flags().Lookup("package-file"))
	packageCmd.Flags().String("package-id-override", "", "Override package ID from file (or set environment PACKAGE_ID)")
	viper.BindPFlag("package.id.override", packageCmd.Flags().Lookup("package-id-override"))
	packageCmd.Flags().String("package-name-override", "", "Override package name from file (or set environment PACKAGE_NAME)")
	viper.BindPFlag("package.name.override", packageCmd.Flags().Lookup("package-name-override"))
}
