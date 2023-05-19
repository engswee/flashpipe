package cmd

import (
	"fmt"
	"github.com/engswee/flashpipe/runner"
	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy integration flow to runtime",
	Long: `Deploy integration flow from design time to
runtime of SAP Integration Suite tenant.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("[INFO] Executing deploy command")

		setMandatoryVariable("iflowid", "IFLOW_ID")
		setOptionalVariable("delaylength", "DELAY_LENGTH")
		setOptionalVariable("maxchecklimit", "MAX_CHECK_LIMIT")
		setOptionalVariable("compareversions", "COMPARE_VERSIONS")

		runner.JavaCmd("io.github.engswee.flashpipe.cpi.exec.DeployDesignTimeArtifact", mavenRepoLocation, flashpipeLocation, log4jFile)

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
	setStringFlagAndBind(deployCmd, "iflowid", "", "Comma separated list of Integration Flow IDs [or set environment IFLOW_ID]")
	setIntFlagAndBind(deployCmd, "delaylength", 30, "Delay (in seconds) between each check of IFlow deployment status [or set environment DELAY_LENGTH]")
	setIntFlagAndBind(deployCmd, "maxchecklimit", 10, "Max number of times to check for IFlow deployment status [or set environment MAX_CHECK_LIMIT]")
	setBoolFlagAndBind(deployCmd, "compareversions", true, "Perform version comparison of design time against runtime before deployment [or set environment COMPARE_VERSIONS]")
}
