package cmd

import (
	"github.com/engswee/flashpipe/config"
	"github.com/engswee/flashpipe/logger"
	"github.com/engswee/flashpipe/runner"
	"github.com/spf13/cobra"
	"os"
)

func NewPackageCommand() *cobra.Command {

	packageCmd := &cobra.Command{
		Use:     "package",
		Aliases: []string{"pkg"},
		Short:   "Upload/update integration package",
		Long: `Upload or update integration package on the
SAP Integration Suite tenant.`,
		Run: func(cmd *cobra.Command, args []string) {
			runUpdatePackage(cmd)
		},
	}

	// Define cobra flags, the default value has the lowest (least significant) precedence
	packageCmd.Flags().String("package-file", "", "Path to location of package file [or set environment PACKAGE_FILE]")
	packageCmd.Flags().String("package-override-id", "", "Override package ID from file [or set environment PACKAGE_ID]")
	packageCmd.Flags().String("package-override-name", "", "Override package name from file [or set environment PACKAGE_NAME]")

	return packageCmd
}

func runUpdatePackage(cmd *cobra.Command) {
	logger.Info("Executing update package command")

	packageFile := config.GetMandatoryString(cmd, "package-file")
	packageId := config.GetString(cmd, "package-override-id")
	packageName := config.GetString(cmd, "package-override-name")

	// TODO - remove
	mavenRepoLocation := config.GetString(cmd, "location.mavenrepo")
	flashpipeLocation := config.GetString(cmd, "location.flashpipe")
	os.Setenv("HOST_TMN", config.GetMandatoryString(cmd, "tmn-host"))
	os.Setenv("HOST_OAUTH", config.GetMandatoryString(cmd, "oauth-host"))
	os.Setenv("OAUTH_CLIENTID", config.GetMandatoryString(cmd, "oauth-clientid"))
	os.Setenv("OAUTH_CLIENTSECRET", config.GetMandatoryString(cmd, "oauth-clientsecret"))
	os.Setenv("PACKAGE_FILE", packageFile)
	os.Setenv("PACKAGE_ID", packageId)
	os.Setenv("PACKAGE_NAME", packageName)

	_, err := runner.JavaCmd("io.github.engswee.flashpipe.cpi.exec.UpdateIntegrationPackage", mavenRepoLocation, flashpipeLocation, log4jFile)
	logger.ExitIfErrorWithMsg(err, "Execution of java command failed")
}
