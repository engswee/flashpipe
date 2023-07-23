package cmd

import (
	"fmt"
	"github.com/engswee/flashpipe/config"
	"github.com/engswee/flashpipe/logger"
	"github.com/engswee/flashpipe/odata"
	"github.com/engswee/flashpipe/repo"
	"github.com/engswee/flashpipe/str"
	"github.com/engswee/flashpipe/sync"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func NewSyncCommand() *cobra.Command {

	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync designtime artifacts from tenant to Git",
		Long: `Synchronise designtime artifacts from SAP Integration Suite
tenant to a Git repository.`,
		Args: func(cmd *cobra.Command, args []string) error {
			// Validate Directory Naming Type
			dirNamingType := config.GetString(cmd, "dirnamingtype")
			switch dirNamingType {
			case "ID", "NAME":
			default:
				return fmt.Errorf("invalid value for --dirnamingtype = %v", dirNamingType)
			}
			// Validate Draft Handling
			draftHandling := config.GetString(cmd, "drafthandling")
			switch draftHandling {
			case "SKIP", "ADD", "ERROR":
			default:
				return fmt.Errorf("invalid value for --drafthandling = %v", draftHandling)
			}
			// Validate Normalise Manifest Action
			//normaliseManifestAction := config.GetString(cmd, "normalise-manifest-action")
			//switch normaliseManifestAction {
			//case "NONE", "ADD_PREFIX", "ADD_SUFFIX", "DELETE_PREFIX", "DELETE_SUFFIX":
			//default:
			//	return fmt.Errorf("invalid value for --normalise-manifest-action = %v", normaliseManifestAction)
			//}
			// Validate Normalise Package Action
			normalisePackageAction := config.GetString(cmd, "normalise-package-action")
			switch normalisePackageAction {
			case "NONE", "ADD_PREFIX", "ADD_SUFFIX", "DELETE_PREFIX", "DELETE_SUFFIX":
			default:
				return fmt.Errorf("invalid value for --normalise-package-action = %v", normalisePackageAction)
			}
			// Validate Include/Exclude IDs
			includedIds := config.GetString(cmd, "ids.include")
			excludedIds := config.GetString(cmd, "ids.exclude")
			if includedIds != "" && excludedIds != "" {
				return fmt.Errorf("--ids.include and --ids.exclude are mutually exclusive - use only one of them")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			runSync(cmd)
		},
	}

	// Define cobra flags, the default value has the lowest (least significant) precedence
	syncCmd.Flags().String("packageid", "", "ID of Integration Package [or set environment PACKAGE_ID]")
	syncCmd.Flags().String("dir-gitsrc", "", "Base directory containing contents of artifacts [or set environment GIT_SRC_DIR]")
	syncCmd.Flags().String("dir-work", "/tmp", "Working directory for in-transit files [or set environment WORK_DIR]")
	syncCmd.Flags().String("dirnamingtype", "ID", "Name artifact directory by ID or Name. Allowed values: ID, NAME [or set environment DIR_NAMING_TYPE]")
	syncCmd.Flags().String("drafthandling", "SKIP", "Handling when artifact is in draft version. Allowed values: SKIP, ADD, ERROR [or set environment DRAFT_HANDLING]")
	syncCmd.Flags().String("ids-include", "", "List of included artifact IDs [or set environment INCLUDE_IDS]")
	syncCmd.Flags().String("ids-exclude", "", "List of excluded artifact IDs [or set environment EXCLUDE_IDS]")
	syncCmd.Flags().String("git-commitmsg", "Sync repo from tenant", "Message used in commit [or set environment COMMIT_MESSAGE]")
	syncCmd.Flags().String("scriptmap", "", "Comma-separated source-target ID pairs for converting script collection references during sync [or set environment SCRIPT_COLLECTION_MAP]")
	//syncCmd.Flags().String("normalise-manifest-action", "NONE", "Action for normalising artifact ID & Name in MANIFEST.MF. Allowed values: NONE, ADD_PREFIX, ADD_SUFFIX, DELETE_PREFIX, DELETE_SUFFIX [or set environment NORMALISE_MANIFEST_ACTION]")
	//syncCmd.Flags().String("normalise-manifest-prefixsuffix", "", "Prefix/suffix used for normalising artifact ID & Name in MANIFEST.MF [or set environment NORMALISE_MANIFEST_PREFIX_SUFFIX]")
	syncCmd.Flags().Bool("syncpackagedetails", false, "Sync details of Integration Package [or set environment SYNC_PACKAGE_LEVEL_DETAILS]")
	syncCmd.Flags().String("normalise-package-action", "NONE", "Action for normalising Package ID & Name package file. Allowed values: NONE, ADD_PREFIX, ADD_SUFFIX, DELETE_PREFIX, DELETE_SUFFIX [or set environment NORMALISE_PACKAGE_ACTION]")
	syncCmd.Flags().String("normalise-package-prefixsuffix-id", "", "Prefix/suffix used for normalising Package ID [or set environment NORMALISE_PACKAGE_ID_PREFIX_SUFFIX]")
	syncCmd.Flags().String("normalise-package-prefixsuffix-name", "", "Prefix/suffix used for normalising Package Name [or set environment NORMALISE_PACKAGE_NAME_PREFIX_SUFFIX]")

	return syncCmd
}

func runSync(cmd *cobra.Command) {
	log.Info().Msg("Executing sync command")

	packageId := config.GetMandatoryString(cmd, "packageid")
	gitSrcDir := config.GetMandatoryString(cmd, "dir-gitsrc")
	workDir := config.GetString(cmd, "dir-work")
	dirNamingType := config.GetString(cmd, "dirnamingtype")
	draftHandling := config.GetString(cmd, "drafthandling")
	delimitedIdsInclude := config.GetString(cmd, "ids-include")
	delimitedIdsExclude := config.GetString(cmd, "ids-exclude")
	commitMsg := config.GetString(cmd, "git-commitmsg")
	scriptCollectionMap := config.GetString(cmd, "scriptmap")
	//normaliseManifestAction := config.GetString(cmd, "normalise-manifest-action")
	//normaliseManifestPrefixOrSuffix := config.GetString(cmd, "normalise-manifest-prefixsuffix")
	syncPackageLevelDetails := config.GetBool(cmd, "syncpackagedetails")
	normalisePackageAction := config.GetString(cmd, "normalise-package-action")
	normalisePackageIDPrefixOrSuffix := config.GetString(cmd, "normalise-package-prefixsuffix-id")
	normalisePackageNamePrefixOrSuffix := config.GetString(cmd, "normalise-package-prefixsuffix-name")

	// TODO - implement normalisation
	//_ = scriptCollectionMap
	_ = normalisePackageAction
	_ = normalisePackageIDPrefixOrSuffix
	_ = normalisePackageNamePrefixOrSuffix

	serviceDetails := odata.GetServiceDetails(cmd)

	// Initialise HTTP executer
	exe := odata.InitHTTPExecuter(serviceDetails)
	synchroniser := sync.New(exe)

	if syncPackageLevelDetails {
		err := synchroniser.SyncPackageDetails(packageId, workDir, gitSrcDir)
		logger.ExitIfError(err)
	}

	// Extract IDs from delimited values
	includedIds := str.ExtractDelimitedValues(delimitedIdsInclude, ",")
	excludedIds := str.ExtractDelimitedValues(delimitedIdsExclude, ",")
	synchroniser.SyncArtifacts(packageId, workDir, gitSrcDir, includedIds, excludedIds, draftHandling, dirNamingType, scriptCollectionMap)

	err := repo.CommitToRepo(gitSrcDir, commitMsg)
	logger.ExitIfError(err)
}
