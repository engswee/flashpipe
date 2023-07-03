package cmd

import (
	"github.com/engswee/flashpipe/config"
	"github.com/engswee/flashpipe/logger"
	"github.com/engswee/flashpipe/repo"
	"github.com/engswee/flashpipe/runner"
	"github.com/spf13/cobra"
	"os"
	"time"
)

func NewSnapshotCommand() *cobra.Command {

	snapshotCmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Snapshot integration packages from tenant to Git",
		Long: `Snapshot all editable integration packages from SAP Integration Suite
tenant to a Git repository.`,
		Run: func(cmd *cobra.Command, args []string) {
			runSnapshot(cmd)
		},
	}

	// Define cobra flags, the default value has the lowest (least significant) precedence
	snapshotCmd.Flags().String("dir-gitsrc", "", "Base directory containing contents of artifacts (grouped into packages) [or set environment GIT_SRC_DIR]")
	snapshotCmd.Flags().String("dir-work", "/tmp", "Working directory for in-transit files [or set environment WORK_DIR]")
	snapshotCmd.Flags().String("drafthandling", "SKIP", "Handling when IFlow is in draft version. Allowed values: SKIP, ADD, ERROR [or set environment DRAFT_HANDLING]")
	snapshotCmd.Flags().String("git-commitmsg", "Tenant snapshot of "+time.Now().Format(time.UnixDate), "Message used in commit [or set environment COMMIT_MESSAGE]")
	snapshotCmd.Flags().Bool("syncpackagedetails", false, "Sync details of Integration Packages [or set environment SYNC_PACKAGE_LEVEL_DETAILS]")

	return snapshotCmd
}

func runSnapshot(cmd *cobra.Command) {
	logger.Info("Executing snapshot command")

	gitSrcDir := config.GetMandatoryString(cmd, "dir-gitsrc")
	workDir := config.GetString(cmd, "dir-work")
	draftHandling := config.GetString(cmd, "drafthandling")
	commitMsg := config.GetString(cmd, "git-commitmsg")
	syncPackageLevelDetails := config.GetBool(cmd, "syncpackagedetails")

	// TODO - remove
	mavenRepoLocation := config.GetString(cmd, "location.mavenrepo")
	flashpipeLocation := config.GetString(cmd, "location.flashpipe")
	log4jFile := config.GetString(cmd, "debug.flashpipe")
	os.Setenv("HOST_TMN", config.GetMandatoryString(cmd, "tmn-host"))
	os.Setenv("HOST_OAUTH", config.GetMandatoryString(cmd, "oauth-host"))
	os.Setenv("OAUTH_CLIENTID", config.GetMandatoryString(cmd, "oauth-clientid"))
	os.Setenv("OAUTH_CLIENTSECRET", config.GetMandatoryString(cmd, "oauth-clientsecret"))
	os.Setenv("GIT_SRC_DIR", gitSrcDir)
	os.Setenv("WORK_DIR", workDir)
	_ = draftHandling
	_ = syncPackageLevelDetails

	_, err := runner.JavaCmd("io.github.engswee.flashpipe.cpi.exec.GetTenantSnapshot", mavenRepoLocation, flashpipeLocation, log4jFile)
	logger.ExitIfErrorWithMsg(err, "Execution of java command failed")

	err = repo.CommitToRepo(gitSrcDir, commitMsg)
	logger.ExitIfError(err)
}
