package cmd

import (
	"fmt"
	"github.com/engswee/flashpipe/file"
	"github.com/engswee/flashpipe/httpclnt"
	"github.com/engswee/flashpipe/logger"
	"github.com/engswee/flashpipe/odata"
	"github.com/engswee/flashpipe/odata/designtime"
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
		// Validate Normalise Manifest Action
		normaliseManifestAction := syncViper.GetString("normalize.manifest.action")
		switch normaliseManifestAction {
		case "NONE", "ADD_PREFIX", "ADD_SUFFIX", "DELETE_PREFIX", "DELETE_SUFFIX":
		default:
			return fmt.Errorf("invalid value for --normalize-manifest-action = %v", normaliseManifestAction)
		}
		// Validate Normalise Package Action
		normalisePackageAction := syncViper.GetString("normalize.package.action")
		switch normalisePackageAction {
		case "NONE", "ADD_PREFIX", "ADD_SUFFIX", "DELETE_PREFIX", "DELETE_SUFFIX":
		default:
			return fmt.Errorf("invalid value for --normalize-package-action = %v", normalisePackageAction)
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

		setMandatoryVariable(syncViper, "packageid", "PACKAGE_ID")
		setMandatoryVariable(syncViper, "dir.gitsrc", "GIT_SRC_DIR")
		setOptionalVariable(syncViper, "dir.work", "WORK_DIR")
		setOptionalVariable(syncViper, "dirnamingtype", "DIR_NAMING_TYPE")
		setOptionalVariable(syncViper, "drafthandling", "DRAFT_HANDLING")
		setOptionalVariable(syncViper, "ids.include", "INCLUDE_IDS")
		setOptionalVariable(syncViper, "ids.exclude", "EXCLUDE_IDS")
		setOptionalVariable(syncViper, "git.commitmsg", "COMMIT_MESSAGE")
		setOptionalVariable(syncViper, "scriptmap", "SCRIPT_COLLECTION_MAP")
		setOptionalVariable(syncViper, "normalize.manifest.action", "NORMALIZE_MANIFEST_ACTION")
		setOptionalVariable(syncViper, "normalize.manifest.prefixsuffix", "NORMALIZE_MANIFEST_PREFIX_SUFFIX")
		setOptionalVariable(syncViper, "syncpackagedetails", "SYNC_PACKAGE_LEVEL_DETAILS")
		setOptionalVariable(syncViper, "normalize.package.action", "NORMALIZE_PACKAGE_ACTION")
		setOptionalVariable(syncViper, "normalize.package.prefixsuffix.id", "NORMALIZE_PACKAGE_ID_PREFIX_SUFFIX")
		setOptionalVariable(syncViper, "normalize.package.prefixsuffix.name", "NORMALIZE_PACKAGE_NAME_PREFIX_SUFFIX")

		//_, err := runner.JavaCmd("io.github.engswee.flashpipe.cpi.exec.DownloadIntegrationPackageContent", mavenRepoLocation, flashpipeLocation, log4jFile)
		//logger.ExitIfErrorWithMsg(err, "Execution of java command failed")

		//if

		packageId := syncViper.GetString("packageid")
		gitSrcDir := syncViper.GetString("dir.gitsrc")
		workDir := syncViper.GetString("dir.work")
		dirNamingType := syncViper.GetString("dirnamingtype")
		draftHandling := syncViper.GetString("drafthandling")
		normaliseManifestAction := syncViper.GetString("normalize.manifest.action")
		//normalisePackageAction := syncViper.GetString("normalize.package.action")
		delimitedIdsInclude := syncViper.GetString("ids.include")
		delimitedIdsExclude := syncViper.GetString("ids.exclude")
		commitMsg := syncViper.GetString("git.commitmsg")
		normaliseManifestPrefixOrSuffix := syncViper.GetString("normalize.manifest.prefixsuffix")

		// Extract IDs from delimited values
		includedIds := str.ExtractDelimitedValues(delimitedIdsInclude, ",")
		excludedIds := str.ExtractDelimitedValues(delimitedIdsExclude, ",")
		syncArtifacts(packageId, workDir, gitSrcDir, includedIds, excludedIds, draftHandling, dirNamingType, normaliseManifestAction, normaliseManifestPrefixOrSuffix)

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

func syncArtifacts(packageId string, workDir string, gitSrcDir string, includedIds []string, excludedIds []string, draftHandling string, dirNamingType string, normaliseManifestAction string, normaliseManifestPrefixOrSuffix string) {

	// Initialise HTTP executer
	exe := httpclnt.New(oauthHost, oauthTokenPath, oauthClientId, oauthClientSecret, basicUserId, basicPassword, tmnHost, "https", 443)

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

	// Create temp directories in working dir
	err = os.MkdirAll(workDir+"/download", os.ModePerm)
	logger.ExitIfError(err)
	// TODO - collect error for handling
	//err = os.MkdirAll(workDir+"/from_git", os.ModePerm)
	//logger.ExitIfError(err)
	//err = os.MkdirAll(workDir+"/from_tenant", os.ModePerm)
	//logger.ExitIfError(err)

	filtered, err := filterArtifacts(artifacts, includedIds, excludedIds)
	logger.ExitIfError(err)

	// Process through the artifacts
	for _, artifact := range filtered {
		logger.Info("---------------------------------------------------------------------------------")
		logger.Info(fmt.Sprintf("📢 Begin processing for artifact %v", artifact.Id))
		// Check if artifact is in draft version
		if artifact.IsDraft {
			switch draftHandling {
			case "SKIP":
				logger.Warn(fmt.Sprintf("Artifact %v is in draft version, and will be skipped", artifact.Id))
				continue
			case "ADD":
				logger.Info(fmt.Sprintf("Artifact %v is in draft version, and will be added", artifact.Id))
			case "ERROR":
				logger.ExitIfError(fmt.Errorf("Artifact %v is in draft version. Save Version in Web UI first!", artifact.Id))
			}
		}
		// Download IFlow
		logger.Info(fmt.Sprintf("Downloading artifact %v from tenant for comparison", artifact.Id))
		dt := designtime.GetDesigntimeArtifactByType(artifact.ArtifactType, exe)
		bytes, err := dt.Download(artifact.Id, "active")
		logger.ExitIfError(err)
		targetDownloadFile := fmt.Sprintf("%v/download/%v.zip", workDir, artifact.Id)
		err = os.WriteFile(targetDownloadFile, bytes, os.ModePerm)
		logger.ExitIfError(err)
		logger.Info(fmt.Sprintf("Artifact %v downloaded to %v", artifact.Id, targetDownloadFile))

		// Normalise ID and Name
		normalisedId := str.Normalise(artifact.Id, normaliseManifestAction, normaliseManifestPrefixOrSuffix)
		normalisedName := str.Normalise(artifact.Name, normaliseManifestAction, normaliseManifestPrefixOrSuffix)
		logger.Debug(fmt.Sprintf("Normalized artifact ID - %v", normalisedId))
		logger.Debug(fmt.Sprintf("Normalized artifact name - %v", normalisedName))

		var directoryName string
		if dirNamingType == "NAME" {
			directoryName = normalisedName
		} else {
			directoryName = normalisedId
		}
		// Unzip artifact contents
		logger.Debug(fmt.Sprintf("Target artifact directory name - %v", directoryName))
		downloadedArtifactPath := fmt.Sprintf("%v/download/%v", workDir, directoryName)
		err = file.UnzipSource(targetDownloadFile, downloadedArtifactPath)
		logger.ExitIfError(err)
		logger.Info(fmt.Sprintf("Downloaded artifact unzipped to %v", downloadedArtifactPath))

		// Normalize MANIFEST.MF before sync to Git - TODO
		// https://github.com/gnewton/jargo/blob/master/jar.go
		//https://pkg.go.dev/github.com/quay/claircore/java/jar
		//https://github.com/quay/claircore/blob/v1.5.8/java/jar/jar.go
		//https://pkg.go.dev/net/textproto#Reader.ReadMIMEHeader

		//ScriptCollection scriptCollection = ScriptCollection.newInstance(scriptCollectionMap)
		//Map collections = scriptCollection.getCollections()
		//ManifestHandler.newInstance("${workDir}/download/${directoryName}/META-INF/MANIFEST.MF").normalizeAttributesInFile(normalizedIFlowID, normalizedIFlowName, scriptCollection.getTargetCollectionValues())

		// Normalize the script collection in IFlow BPMN2 XML before syncing to Git - TODO
		//if (collections.size()) {
		//	BPMN2Handler bpmn2Handler = new BPMN2Handler()
		//	bpmn2Handler.updateFiles(collections, "${workDir}/download/${directoryName}")
		//}

		gitArtifactPath := fmt.Sprintf("%v/%v", gitSrcDir, directoryName)
		if file.CheckFileExists(fmt.Sprintf("%v/META-INF/MANIFEST.MF", gitArtifactPath)) {
			// (1) If IFlow already exists in Git, then compare and update
			logger.Info("Comparing content from tenant against Git")

			// TODO - no longer required?
			// Copy to temp directory for diff comparison
			// Remove comments from parameters.prop before comparison only if it exists

			// Diff directories excluding parameters.prop
			dirDiffer := diffDirectories(downloadedArtifactPath, gitArtifactPath)
			// Diff parameters.prop ignoring commented lines
			downloadedParams := fmt.Sprintf("%v/src/main/resources/parameters.prop", downloadedArtifactPath)
			gitParams := fmt.Sprintf("%v/src/main/resources/parameters.prop", gitArtifactPath)
			var paramDiffer bool
			if file.CheckFileExists(downloadedParams) && file.CheckFileExists(gitParams) {
				paramDiffer = diffParams(downloadedParams, gitParams)
			} else if !file.CheckFileExists(downloadedParams) && !file.CheckFileExists(gitParams) {
				logger.Warn("Skipping diff of parameters.prop as it does not exist in both source and target")
			} else {
				paramDiffer = true
				logger.Info("Update required since parameters.prop does not exist in either source or target")
			}

			if dirDiffer || paramDiffer {
				logger.Info("🏆 Changes detected and will be updated to Git")
				// Update the changes into the Git directory
				err = file.ReplaceDir(downloadedArtifactPath, gitArtifactPath)
				logger.ExitIfError(err)
			} else {
				logger.Info("🏆 No changes detected. Update to Git not required")
			}

		} else { // (2) If IFlow does not exist in Git, then add it
			logger.Info(fmt.Sprintf("🏆 Artifact %v does not exist, and will be added to Git", artifact.Id))

			err = file.ReplaceDir(downloadedArtifactPath, gitArtifactPath)
			logger.ExitIfError(err)
		}
	}

	// TODO - write error wrapper - https://go.dev/blog/errors-are-values
	// Clean up working directory
	err = os.RemoveAll(workDir + "/download")
	logger.ExitIfError(err)
	//err = os.RemoveAll(workDir + "/from_git")
	//logger.ExitIfError(err)
	//err = os.RemoveAll(workDir + "/from_tenant")
	//logger.ExitIfError(err)

	logger.Info("---------------------------------------------------------------------------------")
	logger.Info(fmt.Sprintf("🏆 Completed processing of artifacts in integration package %v", packageId))
}

func filterArtifacts(artifacts []*odata.ArtifactDetails, includedIds []string, excludedIds []string) ([]*odata.ArtifactDetails, error) {
	var output []*odata.ArtifactDetails
	if len(includedIds) > 0 {
		for _, id := range includedIds {
			artifact := findArtifactById(id, artifacts)
			if artifact != nil {
				output = append(output, artifact)
			} else {
				return nil, fmt.Errorf("Artifact %v in INCLUDE_IDS does not exist", id)
			}
		}
		return output, nil
	} else if len(excludedIds) > 0 {
		for _, id := range excludedIds {
			artifact := findArtifactById(id, artifacts)
			if artifact == nil {
				return nil, fmt.Errorf("Artifact %v in EXCLUDE_IDS does not exist", id)
			}
		}
		for _, artifact := range artifacts {
			if !contains(artifact.Id, excludedIds) {
				output = append(output, artifact)
			}
		}
		return output, nil
	}
	return artifacts, nil
}

func findArtifactById(key string, list []*odata.ArtifactDetails) *odata.ArtifactDetails {
	for _, s := range list {
		if s.Id == key {
			return s
		}
	}
	return nil
}

func contains(key string, list []string) bool {
	for _, s := range list {
		if s == key {
			return true
		}
	}
	return false
}

func syncPackageDetails() {

}
