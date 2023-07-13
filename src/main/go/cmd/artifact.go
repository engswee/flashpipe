package cmd

import (
	"fmt"
	"github.com/engswee/flashpipe/config"
	"github.com/engswee/flashpipe/file"
	"github.com/engswee/flashpipe/logger"
	"github.com/engswee/flashpipe/odata"
	"github.com/engswee/flashpipe/runner"
	"github.com/spf13/cobra"
	"os"
)

func NewArtifactCommand() *cobra.Command {

	artifactCmd := &cobra.Command{
		Use:   "artifact",
		Short: "Create/update artifacts",
		Long: `Create or update artifacts on the
SAP Integration Suite tenant.`,
		Args: func(cmd *cobra.Command, args []string) error {
			//  TODO - Flags are not bind to Viper at this point ??
			// Validate the artifact type
			artifactType := config.GetString(cmd, "artifact-type")
			switch artifactType {
			case "MessageMapping", "ScriptCollection", "Integration", "ValueMapping":
			default:
				return fmt.Errorf("invalid value for --artifact-type = %v", artifactType)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			runUpdateArtifact(cmd)
		},
	}

	// Define cobra flags, the default value has the lowest (least significant) precedence
	artifactCmd.Flags().String("artifact-id", "", "ID of artifact [or set environment ARTIFACT_ID]")
	artifactCmd.Flags().String("artifact-name", "", "Name of artifact [or set environment ARTIFACT_NAME]")
	artifactCmd.Flags().String("package-id", "", "ID of Integration Package [or set environment PACKAGE_ID]")
	artifactCmd.Flags().String("package-name", "", "Name of Integration Package [or set environment PACKAGE_NAME]")
	artifactCmd.Flags().String("dir-gitsrc", "", "Directory containing contents of Integration Flow [or set environment GIT_SRC_DIR]")
	artifactCmd.Flags().String("file-param", "", "Use to a different parameters.prop file instead of the default in src/main/resources/ [or set environment PARAM_FILE]")
	artifactCmd.Flags().String("dir-work", "/tmp", "Working directory for in-transit files [or set environment WORK_DIR]")
	artifactCmd.Flags().String("scriptmap", "", "Comma-separated source-target ID pairs for converting script collection references during create/update [or set environment SCRIPT_COLLECTION_MAP]")
	artifactCmd.Flags().String("artifact-type", "Integration", "Artifact type. Allowed values: Integration, MessageMapping, ScriptCollection, ValueMapping")

	return artifactCmd
}

func runUpdateArtifact(cmd *cobra.Command) {
	artifactType := config.GetString(cmd, "artifact-type")
	logger.Info(fmt.Sprintf("Executing update artifact %v command", artifactType))

	artifactId := config.GetMandatoryString(cmd, "artifact-id")
	artifactName := config.GetMandatoryString(cmd, "artifact-name")
	packageId := config.GetMandatoryString(cmd, "package-id")
	packageName := config.GetMandatoryString(cmd, "package-name")
	gitSrcDir := config.GetMandatoryString(cmd, "dir-gitsrc")
	parametersFile := config.GetString(cmd, "file-param")
	workDir := config.GetString(cmd, "dir-work")
	scriptMap := config.GetString(cmd, "scriptmap")

	// TODO - ID and package name from file rather than parameter

	// TODO - remove
	mavenRepoLocation := config.GetString(cmd, "location.mavenrepo")
	flashpipeLocation := config.GetString(cmd, "location.flashpipe")
	log4jFile := config.GetString(cmd, "debug.flashpipe")
	os.Setenv("HOST_TMN", config.GetMandatoryString(cmd, "tmn-host"))
	os.Setenv("HOST_OAUTH", config.GetMandatoryString(cmd, "oauth-host"))
	os.Setenv("OAUTH_CLIENTID", config.GetMandatoryString(cmd, "oauth-clientid"))
	os.Setenv("OAUTH_CLIENTSECRET", config.GetMandatoryString(cmd, "oauth-clientsecret"))
	os.Setenv("IFLOW_ID", artifactId)
	os.Setenv("IFLOW_NAME", artifactName)
	os.Setenv("PACKAGE_ID", packageId)
	os.Setenv("PACKAGE_NAME", packageName)

	defaultParamFile := fmt.Sprintf("%v/src/main/resources/parameters.prop", gitSrcDir)
	if parametersFile == "" {
		parametersFile = defaultParamFile
	} else if parametersFile != defaultParamFile {
		logger.Info("Using", parametersFile, "as parameters.prop file")
		err := file.CopyFile(parametersFile, defaultParamFile)
		logger.ExitIfError(err)
	}
	// TODO - used to pass to Java class UpdateConfiguration. to be removed when Java deprecated
	os.Setenv("PARAM_FILE", parametersFile)

	// Initialise HTTP executer
	serviceDetails := odata.GetServiceDetails(cmd)
	exe := odata.InitHTTPExecuter(serviceDetails)

	// Initialise designtime artifact
	dt := odata.NewDesigntimeArtifact(artifactType, exe)
	ip := odata.NewIntegrationPackage(exe)

	// Check if IFlow already exist on tenant
	exists, err := artifactExists(artifactId, artifactType, packageId, dt, ip)
	logger.ExitIfError(err)
	if !exists {
		// Create artifact
		logger.Info(fmt.Sprintf("Artifact %v will be created", artifactId))

		err = prepareUploadDir(workDir, gitSrcDir, artifactType)
		logger.ExitIfError(err)

		err = createArtifact(artifactId, artifactName, packageId, workDir+"/upload", scriptMap, dt)
		logger.ExitIfError(err)
		//_, err = runner.JavaCmd("io.github.engswee.flashpipe.cpi.exec.UploadDesignTimeArtifact", mavenRepoLocation, flashpipeLocation, log4jFile)
		//logger.ExitIfErrorWithMsg(err, "Execution of java command failed")

		logger.Info("🏆 Artifact created successfully")

	} else {
		// Update IFlow
		logger.Info("Checking if designtime artifact needs to be updated")
		// 1 - Download artifact content from tenant
		zipFile := fmt.Sprintf("%v/%v.zip", workDir, artifactId)
		err = odata.Download(zipFile, artifactId, dt)
		logger.ExitIfError(err)
		// 2 - Diff contents from tenant against Git
		// TODO - refactor and combine with implementation used in synchroniser
		changesFound, err := compareIFlowContents(workDir, zipFile, gitSrcDir, artifactId, artifactName, scriptMap, mavenRepoLocation, flashpipeLocation, log4jFile)
		logger.ExitIfError(err)

		if changesFound == true {
			logger.Info("Changes found in IFlow. IFlow design will be updated in CPI tenant")
			err = prepareUploadDir(workDir, gitSrcDir, artifactType)
			logger.ExitIfError(err)
			err = updateArtifact(artifactId, artifactName, packageId, workDir+"/upload", scriptMap, dt)
			logger.ExitIfError(err)

			//// If runtime has the same version no, then undeploy it, otherwise it gets skipped during deployment
			//def designtimeVersion = designTimeArtifact.getVersion(this.iFlowId, 'active', false)
			//RuntimeArtifact runtimeArtifact = new RuntimeArtifact(this.httpExecuter)
			//def runtimeVersion = runtimeArtifact.getVersion(this.iFlowId)
			//
			//if (runtimeVersion == designtimeVersion) {
			//	logger.info('Undeploying existing runtime artifact with same version number due to changes in design')
			//	runtimeArtifact.undeploy(this.iFlowId, csrfToken)
			//}

			//_, err = runner.JavaCmd("io.github.engswee.flashpipe.cpi.exec.UpdateDesignTimeArtifact", mavenRepoLocation, flashpipeLocation, log4jFile)
			//logger.ExitIfErrorWithMsg(err, "Execution of java command failed")

			logger.Info("🏆 IFlow design updated successfully")
		} else {
			logger.Info("🏆 No changes detected. IFlow design does not need to be updated")
		}

		// TODO - only applicable for Integration
		// 4 - Update the configuration of the IFlow based on parameters.prop file
		logger.Info("Updating configured parameter(s) of IFlow where necessary")
		_, err = runner.JavaCmd("io.github.engswee.flashpipe.cpi.exec.UpdateConfiguration", mavenRepoLocation, flashpipeLocation, log4jFile)
		logger.ExitIfErrorWithMsg(err, "Execution of java command failed")

	}
}

func prepareUploadDir(workDir string, gitSrcDir string, artifactType string) (err error) {
	// Clean up previous uploads
	iFlowDir := workDir + "/upload"
	err = os.RemoveAll(iFlowDir)
	if err != nil {
		return
	}
	// TODO - Copy META-INF and resources separately so that other directories like QA, STG, PRD not copied
	err = file.CopyDir(gitSrcDir+"/META-INF", iFlowDir+"/META-INF")
	if err != nil {
		return
	}
	// TODO - for value mapping it only has value_mapping.xml file
	if artifactType == "ValueMapping" {
		file.CopyFile(gitSrcDir+"/value_mapping.xml", iFlowDir+"/value_mapping.xml")
	} else {
		err = file.CopyDir(gitSrcDir+"/src/main/resources", iFlowDir+"/src/main/resources")
		if err != nil {
			return
		}
	}
	os.Setenv("IFLOW_DIR", iFlowDir) // TODO - remove when Java call no longer used
	return
}

func compareIFlowContents(workDir string, zipFile string, gitSrcDir string, iflowId string, iflowName string, scriptMap string, mavenRepoLocation string, flashpipeLocation string, log4jFile string) (changesFound bool, err error) {
	err = os.RemoveAll(workDir + "/download")
	if err != nil {
		return
	}

	logger.Info("Unzipping downloaded IFlow artifact", zipFile, "to", workDir+"/download")
	err = file.UnzipSource(zipFile, workDir+"/download")
	if err != nil {
		return
	}
	// TODO - Update the script collection in IFlow BPMN2 XML before diff comparison
	_, err = runner.JavaCmd("io.github.engswee.flashpipe.cpi.exec.BPMN2Handler", mavenRepoLocation, flashpipeLocation, log4jFile)
	logger.ExitIfErrorWithMsg(err, "Execution of java command failed")

	// TODO - Update the MANIFEST.MF file with script collection conversions
	_, err = runner.JavaCmdWithArgs(mavenRepoLocation, flashpipeLocation, log4jFile, "io.github.engswee.flashpipe.cpi.util.ManifestHandler", gitSrcDir+"/META-INF/MANIFEST.MF", iflowId, iflowName, scriptMap)
	logger.ExitIfErrorWithMsg(err, "Execution of java command failed")

	//// Compare META-INF directory for any differences in the manifest file
	//logger.Info("Checking for changes in META-INF directory")
	//metaDirDiffer := file.DiffDirectories(workDir+"/download/META-INF/", gitSrcDir+"/META-INF/")
	//
	//logger.Info("Checking for changes in src/main/resources directory")
	//resourcesDirDiffer := file.DiffDirectories(workDir+"/download/src/main/resources/", gitSrcDir+"/src/main/resources/")
	//
	//if metaDirDiffer == false && resourcesDirDiffer == false {
	//	changesFound = false
	//} else {
	//	changesFound = true
	//}
	// Diff directories excluding parameters.prop
	dirDiffer := file.DiffDirectories(workDir+"/download", gitSrcDir)
	// Diff parameters.prop ignoring commented lines
	downloadedParams := fmt.Sprintf("%v/download/src/main/resources/parameters.prop", workDir)
	gitParams := fmt.Sprintf("%v/src/main/resources/parameters.prop", gitSrcDir)
	var paramDiffer bool
	if file.CheckFileExists(downloadedParams) && file.CheckFileExists(gitParams) {
		paramDiffer = file.DiffParams(downloadedParams, gitParams)
	} else if !file.CheckFileExists(downloadedParams) && !file.CheckFileExists(gitParams) {
		logger.Warn("Skipping diff of parameters.prop as it does not exist in both source and target")
	} else {
		paramDiffer = true
		logger.Info("Update required since parameters.prop does not exist in either source or target")
	}

	if dirDiffer || paramDiffer {
		changesFound = true
	} else {
		changesFound = false
	}
	return
}

func artifactExists(artifactId string, artifactType string, packageId string, dt odata.DesigntimeArtifact, ip *odata.IntegrationPackage) (bool, error) {
	logger.Info(fmt.Sprintf("Checking if %v exists", artifactId))
	exists, err := dt.Exists(artifactId, "active")
	if err != nil {
		return false, err
	}
	if exists {
		logger.Info(fmt.Sprintf("Active version of artifact %v exists", artifactId))
		//  Check if version is in draft mode
		var details []*odata.ArtifactDetails
		details, err = ip.GetArtifactsData(packageId, artifactType)
		if err != nil {
			return false, err
		}
		artifact := odata.FindArtifactById(artifactId, details)
		if artifact == nil {
			return false, fmt.Errorf("Artifact %v not found in package %v", artifactId, packageId)
		}
		if artifact.IsDraft {
			return false, fmt.Errorf("Artifact %v is in Draft state. Save Version of artifact in Web UI first!", artifactId)
		}
		return true, nil
	} else {
		logger.Info(fmt.Sprintf("Active version of artifact %v does not exist", artifactId))
		return false, nil
	}
}

func createArtifact(artifactId string, artifactName string, packageId string, artifactDir string, scriptCollectionMap string, dt odata.DesigntimeArtifact) error {
	err := dt.Create(artifactId, artifactName, packageId, artifactDir)
	if err != nil {
		return err
	}
	return nil
}

func updateArtifact(artifactId string, artifactName string, packageId string, artifactDir string, scriptCollectionMap string, dt odata.DesigntimeArtifact) error {
	err := dt.Update(artifactId, artifactName, packageId, artifactDir)
	if err != nil {
		return err
	}
	return nil
}
