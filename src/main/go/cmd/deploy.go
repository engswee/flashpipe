package cmd

import (
	"github.com/engswee/flashpipe/logger"
	"github.com/engswee/flashpipe/runner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var deployViper = viper.New()

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy integration flow to runtime",
	Long: `Deploy integration flow from design time to
runtime of SAP Integration Suite tenant.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Executing deploy command")

		setMandatoryVariable(deployViper, "iflow.id", "IFLOW_ID")
		setOptionalVariable(deployViper, "delaylength", "DELAY_LENGTH")
		setOptionalVariable(deployViper, "maxchecklimit", "MAX_CHECK_LIMIT")
		setOptionalVariable(deployViper, "compareversions", "COMPARE_VERSIONS")

		_, err := runner.JavaCmd("io.github.engswee.flashpipe.cpi.exec.DeployDesignTimeArtifact", mavenRepoLocation, flashpipeLocation, log4jFile)
		if err != nil {
			logger.Error("Execution of java command failed")
		}
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deployCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	setStringFlagAndBind(deployViper, deployCmd, "iflow.id", "", "Comma separated list of Integration Flow IDs [or set environment IFLOW_ID]")
	setIntFlagAndBind(deployViper, deployCmd, "delaylength", 30, "Delay (in seconds) between each check of IFlow deployment status [or set environment DELAY_LENGTH]")
	setIntFlagAndBind(deployViper, deployCmd, "maxchecklimit", 10, "Max number of times to check for IFlow deployment status [or set environment MAX_CHECK_LIMIT]")
	setBoolFlagAndBind(deployViper, deployCmd, "compareversions", true, "Perform version comparison of design time against runtime before deployment [or set environment COMPARE_VERSIONS]")
}
