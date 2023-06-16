package cmd

import (
	"fmt"
	"github.com/engswee/flashpipe/httpclnt"
	"github.com/engswee/flashpipe/logger"
	"github.com/engswee/flashpipe/odata"
	"github.com/engswee/flashpipe/repo"
	"github.com/engswee/flashpipe/str"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var syncViper = viper.New()

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync designtime artifacts from tenant to Git",
	Long: `Synchronise designtime artifacts from SAP Integration Suite
tenant to a Git repository.`,
	Args: func(cmd *cobra.Command, args []string) error {
		// Validate Directory Naming Type
		dirNamingType := syncViper.GetString("dirnamingtype")
		switch dirNamingType {
		case "ID", "NAME":
		default:
			return fmt.Errorf("invalid value for --dirnamingtype = %v", dirNamingType)
		}
		// Validate Draft Handling
		draftHandling := syncViper.GetString("drafthandling")
		switch draftHandling {
		case "SKIP", "ADD", "ERROR":
		default:
			return fmt.Errorf("invalid value for --drafthandling = %v", draftHandling)
		}
		// Validate Normalize Manifest Action
		normalizeManifestAction := syncViper.GetString("normalize.manifest.action")
		switch normalizeManifestAction {
		case "NONE", "ADD_PREFIX", "ADD_SUFFIX", "DELETE_PREFIX", "DELETE_SUFFIX":
		default:
			return fmt.Errorf("invalid value for --normalize-manifest-action = %v", normalizeManifestAction)
		}
		// Validate Normalize Package Action
		normalizePackageAction := syncViper.GetString("normalize.package.action")
		switch normalizePackageAction {
		case "NONE", "ADD_PREFIX", "ADD_SUFFIX", "DELETE_PREFIX", "DELETE_SUFFIX":
		default:
			return fmt.Errorf("invalid value for --normalize-package-action = %v", normalizePackageAction)
		}
		// Validate Include/Exclude IDs
		includedIds := syncViper.GetString("ids.include")
		excludedIds := syncViper.GetString("ids.exclude")
		if includedIds != "" && excludedIds != "" {
			return fmt.Errorf("--ids.include and --ids.exclude are mutually exclusive - use only one of them")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Executing sync command")

		packageId := setMandatoryVariable(syncViper, "packageid", "PACKAGE_ID")
		gitSrcDir := setMandatoryVariable(syncViper, "dir.gitsrc", "GIT_SRC_DIR")
		setOptionalVariable(syncViper, "dir.work", "WORK_DIR")
		setOptionalVariable(syncViper, "dirnamingtype", "DIR_NAMING_TYPE")
		draftHandling := setOptionalVariable(syncViper, "drafthandling", "DRAFT_HANDLING")
		delimitedIdsInclude := setOptionalVariable(syncViper, "ids.include", "INCLUDE_IDS")
		delimitedIdsExclude := setOptionalVariable(syncViper, "ids.exclude", "EXCLUDE_IDS")
		commitMsg := setOptionalVariable(syncViper, "git.commitmsg", "COMMIT_MESSAGE")
		setOptionalVariable(syncViper, "scriptmap", "SCRIPT_COLLECTION_MAP")
		setOptionalVariable(syncViper, "normalize.manifest.action", "NORMALIZE_MANIFEST_ACTION")
		setOptionalVariable(syncViper, "normalize.manifest.prefixsuffix", "NORMALIZE_MANIFEST_PREFIX_SUFFIX")
		setOptionalVariable(syncViper, "syncpackagedetails", "SYNC_PACKAGE_LEVEL_DETAILS")
		setOptionalVariable(syncViper, "normalize.package.action", "NORMALIZE_PACKAGE_ACTION")
		setOptionalVariable(syncViper, "normalize.package.prefixsuffix.id", "NORMALIZE_PACKAGE_ID_PREFIX_SUFFIX")
		setOptionalVariable(syncViper, "normalize.package.prefixsuffix.name", "NORMALIZE_PACKAGE_NAME_PREFIX_SUFFIX")

		//_, err := runner.JavaCmd("io.github.engswee.flashpipe.cpi.exec.DownloadIntegrationPackageContent", mavenRepoLocation, flashpipeLocation, log4jFile)
		//logger.ExitIfErrorWithMsg(err, "Execution of java command failed")

		syncPackage(packageId, draftHandling, delimitedIdsInclude, delimitedIdsExclude)

		err := repo.CommitToRepo(gitSrcDir, commitMsg)
		logger.ExitIfError(err)
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
	setStringFlagAndBind(syncViper, syncCmd, "packageid", "", "ID of Integration Package [or set environment PACKAGE_ID]")
	setStringFlagAndBind(syncViper, syncCmd, "dir.gitsrc", "", "Base directory containing contents of Integration Flow(s) [or set environment GIT_SRC_DIR]")
	setStringFlagAndBind(syncViper, syncCmd, "dir.work", "/tmp", "Working directory for in-transit files [or set environment WORK_DIR]")
	setStringFlagAndBind(syncViper, syncCmd, "dirnamingtype", "ID", "Name IFlow directories by ID or Name. Allowed values: ID, NAME [or set environment DIR_NAMING_TYPE]")
	setStringFlagAndBind(syncViper, syncCmd, "drafthandling", "SKIP", "Handling when IFlow is in draft version. Allowed values: SKIP, ADD, ERROR [or set environment DRAFT_HANDLING]")
	setStringFlagAndBind(syncViper, syncCmd, "ids.include", "", "List of included IFlow IDs [or set environment INCLUDE_IDS]")
	setStringFlagAndBind(syncViper, syncCmd, "ids.exclude", "", "List of excluded IFlow IDs [or set environment EXCLUDE_IDS]")
	setStringFlagAndBind(syncViper, syncCmd, "git.commitmsg", "Sync repo from tenant", "Message used in commit [or set environment COMMIT_MESSAGE]")
	setStringFlagAndBind(syncViper, syncCmd, "scriptmap", "", "Comma-separated source-target ID pairs for converting script collection references during sync [or set environment SCRIPT_COLLECTION_MAP]")
	setStringFlagAndBind(syncViper, syncCmd, "normalize.manifest.action", "NONE", "Action for normalizing IFlow ID & Name in MANIFEST.MF. Allowed values: NONE, ADD_PREFIX, ADD_SUFFIX, DELETE_PREFIX, DELETE_SUFFIX [or set environment NORMALIZE_MANIFEST_ACTION]")
	setStringFlagAndBind(syncViper, syncCmd, "normalize.manifest.prefixsuffix", "", "Prefix/suffix used for normalizing IFlow ID & Name in MANIFEST.MF [or set environment NORMALIZE_MANIFEST_PREFIX_SUFFIX]")
	setStringFlagAndBind(syncViper, syncCmd, "syncpackagedetails", "NO", "Sync details of Integration Package. Allowed values: NO, YES [or set environment SYNC_PACKAGE_LEVEL_DETAILS]")
	setStringFlagAndBind(syncViper, syncCmd, "normalize.package.action", "NONE", "Action for normalizing Package ID & Name package file. Allowed values: NONE, ADD_PREFIX, ADD_SUFFIX, DELETE_PREFIX, DELETE_SUFFIX [or set environment NORMALIZE_PACKAGE_ACTION]")
	setStringFlagAndBind(syncViper, syncCmd, "normalize.package.prefixsuffix.id", "", "Prefix/suffix used for normalizing Package ID [or set environment NORMALIZE_PACKAGE_ID_PREFIX_SUFFIX]")
	setStringFlagAndBind(syncViper, syncCmd, "normalize.package.prefixsuffix.name", "", "Prefix/suffix used for normalizing Package Name [or set environment NORMALIZE_PACKAGE_NAME_PREFIX_SUFFIX]")
}

func syncPackage(packageId string, draftHandling string, delimitedIdsInclude string, delimitedIdsExclude string) {
	// Extract IDs from delimited values
	includedIds := str.ExtractDelimitedValues(delimitedIdsInclude, ",")
	excludedIds := str.ExtractDelimitedValues(delimitedIdsExclude, ",")

	// TODO
	_ = includedIds
	_ = excludedIds

	// Initialise HTTP executer
	exe := httpclnt.New(oauthHost, oauthTokenPath, oauthClientId, oauthClientSecret, basicUserId, basicPassword, tmnHost)

	// Get all design time artifacts of package
	logger.Info(fmt.Sprintf("Getting artifacts in integration package %v", packageId))
	// Verify the package is downloadable
	ip := odata.NewIntegrationPackage(exe)
	readOnly, err := ip.IsReadOnly(packageId)
	logger.ExitIfError(err)
	if readOnly {
		logger.Warn(fmt.Sprintf("Skipping package %v as it is Configure-only and cannot be downloaded", packageId))
		return
	}
	artifacts, err := ip.GetAllArtifacts(packageId)
	logger.ExitIfError(err)

	for _, artifact := range artifacts {
		logger.Info("---------------------------------------------------------------------------------")
		logger.Info(fmt.Sprintf("📢 Begin processing for artifact %v", artifact.Id))
		if artifact.IsDraft {
			switch draftHandling {
			case "SKIP":
				logger.Warn(fmt.Sprintf("Artifact %v is in draft version, and will be skipped", artifact.Id))
				continue
			case "ADD":
				logger.Info(fmt.Sprintf("Artifact %v is in draft version, and will be added", artifact.Id))
			case "ERROR":
				logger.Error(fmt.Sprintf("Artifact %v is in draft version. Save Version in Web UI first!", artifact.Id))
				os.Exit(1)
			}
		}
		logger.Info(fmt.Sprintf("Downloading artifact %v from tenant for comparison", artifact.Id))
	}
	logger.Info("---------------------------------------------------------------------------------")
	logger.Info(fmt.Sprintf("🏆 Completed processing of artifacts in integration package %v", packageId))
}
