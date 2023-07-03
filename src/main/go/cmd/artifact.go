package cmd

import (
	"fmt"
	"github.com/engswee/flashpipe/config"
	"github.com/engswee/flashpipe/diff"
	"github.com/engswee/flashpipe/file"
	"github.com/engswee/flashpipe/logger"
	"github.com/engswee/flashpipe/runner"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

func NewArtifactCommand() *cobra.Command {

	artifactCmd := &cobra.Command{
		Use:   "artifact",
		Short: "Upload/update artifacts",
		Long: `Upload or update artifacts on the
SAP Integration Suite tenant.`,
		Run: func(cmd *cobra.Command, args []string) {
			runUpdateArtifact(cmd)
		},
	}

	// Define cobra flags, the default value has the lowest (least significant) precedence
	artifactCmd.Flags().String("iflow-id", "", "ID of Integration Flow [or set environment IFLOW_ID]")
	artifactCmd.Flags().String("iflow-name", "", "Name of Integration Flow [or set environment IFLOW_NAME]")
	artifactCmd.Flags().String("package-id", "", "ID of Integration Package [or set environment PACKAGE_ID]")
	artifactCmd.Flags().String("package-name", "", "Name of Integration Package [or set environment PACKAGE_NAME]")
	artifactCmd.Flags().String("dir-gitsrc", "", "Directory containing contents of Integration Flow [or set environment GIT_SRC_DIR]")
	artifactCmd.Flags().String("file-param", "", "Use to a different parameters.prop file instead of the default in src/main/resources/ [or set environment PARAM_FILE]")
	artifactCmd.Flags().String("dir-work", "/tmp", "Working directory for in-transit files [or set environment WORK_DIR]")
	artifactCmd.Flags().String("scriptmap", "", "Comma-separated source-target ID pairs for converting script collection references during upload/update [or set environment SCRIPT_COLLECTION_MAP]")

	return artifactCmd
}

func runUpdateArtifact(cmd *cobra.Command) {
	logger.Info("Executing update artifact command")

	iflowId := config.GetMandatoryString(cmd, "iflow-id")
	iflowName := config.GetMandatoryString(cmd, "iflow-name")
	packageId := config.GetMandatoryString(cmd, "package-id")
	packageName := config.GetMandatoryString(cmd, "package-name")
	gitSrcDir := config.GetMandatoryString(cmd, "dir-gitsrc")
	parametersFile := config.GetString(cmd, "file-param")
	workDir := config.GetString(cmd, "dir-work")
	scriptMap := config.GetString(cmd, "scriptmap")

	// TODO - remove
	mavenRepoLocation := config.GetString(cmd, "location.mavenrepo")
	flashpipeLocation := config.GetString(cmd, "location.flashpipe")
	log4jFile := config.GetString(cmd, "debug.flashpipe")
	os.Setenv("HOST_TMN", config.GetMandatoryString(cmd, "tmn-host"))
	os.Setenv("HOST_OAUTH", config.GetMandatoryString(cmd, "oauth-host"))
	os.Setenv("OAUTH_CLIENTID", config.GetMandatoryString(cmd, "oauth-clientid"))
	os.Setenv("OAUTH_CLIENTSECRET", config.GetMandatoryString(cmd, "oauth-clientsecret"))
	os.Setenv("IFLOW_ID", iflowId)
	os.Setenv("IFLOW_NAME", iflowName)
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

	// Check if IFlow already exist on tenant
	output, err := runner.JavaCmd("io.github.engswee.flashpipe.cpi.exec.QueryDesignTimeArtifact", mavenRepoLocation, flashpipeLocation, log4jFile)
	if strings.Contains(output, "Active version of IFlow does not exist") {
		// Upload IFlow
		logger.Info("IFlow will be uploaded to tenant")

		err = prepareUploadDir(workDir, gitSrcDir)
		logger.ExitIfError(err)

		_, err = runner.JavaCmd("io.github.engswee.flashpipe.cpi.exec.UploadDesignTimeArtifact", mavenRepoLocation, flashpipeLocation, log4jFile)
		logger.ExitIfErrorWithMsg(err, "Execution of java command failed")
		logger.Info("🏆 IFlow created successfully")

	} else if err == nil {
		// Update IFlow
		logger.Info("Checking if IFlow design needs to be updated")
		// 1 - Download IFlow from tenant
		zipFile := fmt.Sprintf("%v/%v.zip", workDir, iflowId)
		downloadIFlow(zipFile, mavenRepoLocation, flashpipeLocation, log4jFile)
		// 2 - Diff contents from tenant against Git
		changesFound, err := compareIFlowContents(workDir, zipFile, gitSrcDir, iflowId, iflowName, scriptMap, mavenRepoLocation, flashpipeLocation, log4jFile)
		logger.ExitIfError(err)

		if changesFound == true {
			logger.Info("Changes found in IFlow. IFlow design will be updated in CPI tenant")
			err = prepareUploadDir(workDir, gitSrcDir)
			logger.ExitIfError(err)

			_, err = runner.JavaCmd("io.github.engswee.flashpipe.cpi.exec.UpdateDesignTimeArtifact", mavenRepoLocation, flashpipeLocation, log4jFile)
			logger.ExitIfErrorWithMsg(err, "Execution of java command failed")
			logger.Info("🏆 IFlow design updated successfully")
		} else {
			logger.Info("🏆 No changes detected. IFlow design does not need to be updated")
		}

		// 4 - Update the configuration of the IFlow based on parameters.prop file
		logger.Info("Updating configured parameter(s) of IFlow where necessary")
		_, err = runner.JavaCmd("io.github.engswee.flashpipe.cpi.exec.UpdateConfiguration", mavenRepoLocation, flashpipeLocation, log4jFile)
		logger.ExitIfErrorWithMsg(err, "Execution of java command failed")

	} else {
		logger.ExitIfErrorWithMsg(err, "Execution of java command failed")
	}
}

func prepareUploadDir(workDir string, gitSrcDir string) (err error) {
	// Clean up previous uploads
	iFlowDir := workDir + "/upload"
	err = os.RemoveAll(iFlowDir)
	if err != nil {
		return
	}

	err = file.CopyDir(gitSrcDir+"/META-INF", iFlowDir+"/META-INF")
	if err != nil {
		return
	}

	err = file.CopyDir(gitSrcDir+"/src/main/resources", iFlowDir+"/src/main/resources")
	if err != nil {
		return
	}
	os.Setenv("IFLOW_DIR", iFlowDir) // TODO - remove when Java call no longer used
	return
}

func downloadIFlow(targetZipFile string, mavenRepoLocation string, flashpipeLocation string, log4jFile string) {
	logger.Info("Download existing IFlow from tenant for comparison")
	os.Setenv("OUTPUT_FILE", targetZipFile)
	os.Setenv("IFLOW_VER", "active")
	_, err := runner.JavaCmd("io.github.engswee.flashpipe.cpi.exec.DownloadDesignTimeArtifact", mavenRepoLocation, flashpipeLocation, log4jFile)
	logger.ExitIfErrorWithMsg(err, "Execution of java command failed")
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
	// Update the script collection in IFlow BPMN2 XML before diff comparison
	_, err = runner.JavaCmd("io.github.engswee.flashpipe.cpi.exec.BPMN2Handler", mavenRepoLocation, flashpipeLocation, log4jFile)
	logger.ExitIfErrorWithMsg(err, "Execution of java command failed")

	// Update the MANIFEST.MF file with script collection conversions
	_, err = runner.JavaCmdWithArgs(mavenRepoLocation, flashpipeLocation, log4jFile, "io.github.engswee.flashpipe.cpi.util.ManifestHandler", gitSrcDir+"/META-INF/MANIFEST.MF", iflowId, iflowName, scriptMap)
	logger.ExitIfErrorWithMsg(err, "Execution of java command failed")

	// Compare META-INF directory for any differences in the manifest file
	logger.Info("Checking for changes in META-INF directory")
	metaDirDiffer := diff.DiffDirectories(workDir+"/download/META-INF/", gitSrcDir+"/META-INF/")

	logger.Info("Checking for changes in src/main/resources directory")
	resourcesDirDiffer := diff.DiffDirectories(workDir+"/download/src/main/resources/", gitSrcDir+"/src/main/resources/")

	if metaDirDiffer == false && resourcesDirDiffer == false {
		changesFound = false
	} else {
		changesFound = true
	}
	return
}
