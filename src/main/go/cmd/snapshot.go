package cmd

import (
	"github.com/engswee/flashpipe/logger"
	"github.com/engswee/flashpipe/repo"
	"github.com/engswee/flashpipe/runner"
	"github.com/spf13/viper"
	"time"

	"github.com/spf13/cobra"
)

var snapshotViper = viper.New()

// snapshotCmd represents the snapshot command
var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Snapshot integration packages from tenant to Git",
	Long: `Snapshot all editable integration packages from SAP Integration Suite
tenant to a Git repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Executing snapshot command")

		gitSrcDir := setMandatoryVariable(snapshotViper, "dir.gitsrc", "GIT_SRC_DIR")
		setOptionalVariable(snapshotViper, "dir.work", "WORK_DIR")
		setOptionalVariable(snapshotViper, "drafthandling", "DRAFT_HANDLING")
		commitMsg := setOptionalVariable(snapshotViper, "git.commitmsg", "COMMIT_MESSAGE")
		setOptionalVariable(snapshotViper, "syncpackagedetails", "SYNC_PACKAGE_LEVEL_DETAILS")

		_, err := runner.JavaCmd("io.github.engswee.flashpipe.cpi.exec.GetTenantSnapshot", mavenRepoLocation, flashpipeLocation, log4jFile)
		logger.ExitIfErrorWithMsg(err, "Execution of java command failed")

		err = repo.CommitToRepo(gitSrcDir, commitMsg)
		logger.ExitIfError(err)
	},
}

func init() {
	rootCmd.AddCommand(snapshotCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// snapshotCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	setStringFlagAndBind(snapshotViper, snapshotCmd, "dir.gitsrc", "", "Base directory containing contents of artifacts (grouped into packages) [or set environment GIT_SRC_DIR]")
	setStringFlagAndBind(snapshotViper, snapshotCmd, "dir.work", "/tmp", "Working directory for in-transit files [or set environment WORK_DIR]")
	setStringFlagAndBind(snapshotViper, snapshotCmd, "drafthandling", "SKIP", "Handling when IFlow is in draft version. Allowed values: SKIP, ADD, ERROR [or set environment DRAFT_HANDLING]")
	setStringFlagAndBind(snapshotViper, snapshotCmd, "git.commitmsg", "Tenant snapshot of "+time.Now().Format(time.UnixDate), "Message used in commit [or set environment COMMIT_MESSAGE]")
	setStringFlagAndBind(snapshotViper, snapshotCmd, "syncpackagedetails", "NO", "Sync details of Integration Packages. Allowed values: NO, YES [or set environment SYNC_PACKAGE_LEVEL_DETAILS]")
}
