package cmd

import (
	"fmt"
	"github.com/engswee/flashpipe/repo"
	"github.com/engswee/flashpipe/runner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync integration flows from tenant to Git",
	Long: `Synchronise integration flows from SAP Integration Suite
tenant to a Git repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("[INFO] Executing sync command")

		setMandatoryVariable("packageid", "PACKAGE_ID")
		setMandatoryVariable("dir.gitsrc", "GIT_SRC_DIR")
		setOptionalVariable("dir.work", "WORK_DIR")
		setOptionalVariable("dirnamingtype", "DIR_NAMING_TYPE")
		setOptionalVariable("drafthandling", "DRAFT_HANDLING")
		setOptionalVariable("ids.include", "INCLUDE_IDS")
		setOptionalVariable("ids.exclude", "EXCLUDE_IDS")
		setOptionalVariable("git.commitmsg", "COMMIT_MESSAGE")
		setOptionalVariable("scriptmap", "SCRIPT_COLLECTION_MAP")
		setOptionalVariable("normalize.manifest.action", "NORMALIZE_MANIFEST_ACTION")
		setOptionalVariable("normalize.manifest.prefixsuffix", "NORMALIZE_MANIFEST_PREFIX_SUFFIX")
		setOptionalVariable("syncpackagedetails", "SYNC_PACKAGE_LEVEL_DETAILS")
		setOptionalVariable("normalize.package.action", "NORMALIZE_PACKAGE_ACTION")
		setOptionalVariable("normalize.package.prefixsuffix.id", "NORMALIZE_PACKAGE_ID_PREFIX_SUFFIX")
		setOptionalVariable("normalize.package.prefixsuffix.name", "NORMALIZE_PACKAGE_NAME_PREFIX_SUFFIX")

		runner.JavaCmd("io.github.engswee.flashpipe.cpi.exec.DownloadIntegrationPackageContent", mavenRepoLocation, flashpipeLocation, log4jFile)

		repo.CommitToRepo(viper.GetString("dir.gitsrc"), viper.GetString("git.commitmsg"))
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// syncCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	setStringFlagAndBind(syncCmd, "packageid", "", "ID of Integration Package [or set environment PACKAGE_ID]")
	setStringFlagAndBind(syncCmd, "dir.gitsrc", "", "Base directory containing contents of Integration Flow(s) [or set environment GIT_SRC_DIR]")
	setStringFlagAndBind(syncCmd, "dir.work", "/tmp", "Working directory for in-transit files [or set environment WORK_DIR]")
	setStringFlagAndBind(syncCmd, "dirnamingtype", "ID", "Name IFlow directories by ID or Name. Allowed values: ID, NAME [or set environment DIR_NAMING_TYPE]")
	setStringFlagAndBind(syncCmd, "drafthandling", "SKIP", "Handling when IFlow is in draft version. Allowed values: SKIP, ADD, ERROR [or set environment DRAFT_HANDLING]")
	setStringFlagAndBind(syncCmd, "ids.include", "", "List of included IFlow IDs [or set environment INCLUDE_IDS]")
	setStringFlagAndBind(syncCmd, "ids.exclude", "", "List of excluded IFlow IDs [or set environment EXCLUDE_IDS]")
	setStringFlagAndBind(syncCmd, "git.commitmsg", "Sync repo from tenant", "Message used in commit [or set environment COMMIT_MESSAGE]")
	setStringFlagAndBind(syncCmd, "scriptmap", "", "Comma-separated source-target ID pairs for converting script collection references during sync [or set environment SCRIPT_COLLECTION_MAP]")
	setStringFlagAndBind(syncCmd, "normalize.manifest.action", "NONE", "Action for normalizing IFlow ID & Name in MANIFEST.MF. Allowed values: NONE, ADD_PREFIX, ADD_SUFFIX, DELETE_PREFIX, DELETE_SUFFIX [or set environment NORMALIZE_MANIFEST_ACTION]")
	setStringFlagAndBind(syncCmd, "normalize.manifest.prefixsuffix", "", "Prefix/suffix used for normalizing IFlow ID & Name in MANIFEST.MF [or set environment NORMALIZE_MANIFEST_PREFIX_SUFFIX]")
	setStringFlagAndBind(syncCmd, "syncpackagedetails", "NO", "Sync details of Integration Package. Allowed values: NO, YES [or set environment SYNC_PACKAGE_LEVEL_DETAILS]")
	setStringFlagAndBind(syncCmd, "normalize.package.action", "NONE", "Action for normalizing Package ID & Name package file. Allowed values: NONE, ADD_PREFIX, ADD_SUFFIX, DELETE_PREFIX, DELETE_SUFFIX [or set environment NORMALIZE_PACKAGE_ACTION]")
	setStringFlagAndBind(syncCmd, "normalize.package.prefixsuffix.id", "", "Prefix/suffix used for normalizing Package ID [or set environment NORMALIZE_PACKAGE_ID_PREFIX_SUFFIX]")
	setStringFlagAndBind(syncCmd, "normalize.package.prefixsuffix.name", "", "Prefix/suffix used for normalizing Package Name [or set environment NORMALIZE_PACKAGE_NAME_PREFIX_SUFFIX]")
}
