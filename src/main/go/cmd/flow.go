package cmd

import (
	"fmt"
	"github.com/engswee/flashpipe/diff"
	"github.com/engswee/flashpipe/file"
	"github.com/engswee/flashpipe/logger"
	"github.com/engswee/flashpipe/runner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
)

var flowViper = viper.New()

// flowCmd represents the flow command
var flowCmd = &cobra.Command{
	Use:   "flow",
	Short: "Upload/update integration flow",
	Long: `Upload or update integration flows on the
SAP Integration Suite tenant.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Executing update flow command")

		setMandatoryVariable(flowViper, "iflow.id", "IFLOW_ID")
		setMandatoryVariable(flowViper, "iflow.name", "IFLOW_NAME")
		setMandatoryVariable(flowViper, "package.id", "PACKAGE_ID")
		setMandatoryVariable(flowViper, "package.name", "PACKAGE_NAME")
		setMandatoryVariable(flowViper, "dir.gitsrc", "GIT_SRC_DIR")
		setOptionalVariable(flowViper, "file.param", "PARAM_FILE")
		setOptionalVariable(flowViper, "dir.work", "WORK_DIR")
		setOptionalVariable(flowViper, "scriptmap", "SCRIPT_COLLECTION_MAP")

		iflowId := flowViper.GetString("iflow.id")
		iflowName := flowViper.GetString("iflow.name")
		gitSrcDir := flowViper.GetString("dir.gitsrc")
		defaultParamFile := gitSrcDir + "/src/main/resources/parameters.prop"
		flowViper.SetDefault("file.param", defaultParamFile)
		parametersFile := flowViper.GetString("file.param")
		workDir := flowViper.GetString("dir.work")
		scriptMap := flowViper.GetString("scriptmap")

		if parametersFile != defaultParamFile {
			logger.Info("Using", parametersFile, "as parameters.prop file")
			err := file.CopyFile(parametersFile, defaultParamFile)
			logger.ExitIfError(err)
		}
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
			downloadIFlow(zipFile)
			// 2 - Diff contents from tenant against Git
			changesFound, err := compareIFlowContents(workDir, zipFile, gitSrcDir, iflowId, iflowName, scriptMap)
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
	},
}

func init() {
	updateCmd.AddCommand(flowCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// flowCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	setStringFlagAndBind(flowViper, flowCmd, "iflow.id", "", "ID of Integration Flow [or set environment IFLOW_ID]")
	setStringFlagAndBind(flowViper, flowCmd, "iflow.name", "", "Name of Integration Flow [or set environment IFLOW_NAME]")
	setStringFlagAndBind(flowViper, flowCmd, "package.id", "", "ID of Integration Package [or set environment PACKAGE_ID]")
	setStringFlagAndBind(flowViper, flowCmd, "package.name", "", "Name of Integration Package [or set environment PACKAGE_NAME]")
	setStringFlagAndBind(flowViper, flowCmd, "dir.gitsrc", "", "Directory containing contents of Integration Flow [or set environment GIT_SRC_DIR]")
	setStringFlagAndBind(flowViper, flowCmd, "file.param", "", "Use to a different parameters.prop file instead of the default in src/main/resources/ [or set environment PARAM_FILE]")
	setStringFlagAndBind(flowViper, flowCmd, "dir.work", "/tmp", "Working directory for in-transit files [or set environment WORK_DIR]")
	setStringFlagAndBind(flowViper, flowCmd, "scriptmap", "", "Comma-separated source-target ID pairs for converting script collection references during upload/update [or set environment SCRIPT_COLLECTION_MAP]")
}

func prepareUploadDir(workDir string, gitSrcDir string) (err error) {
	// Clean up previous uploads
	iFlowDir := workDir + "/upload"
	err = os.RemoveAll(iFlowDir)
	if err != nil {
		return
	}

	//err = os.MkdirAll(iFlowDir+"/src/main", os.ModePerm)
	//if err != nil {
	//	return
	//}

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

func downloadIFlow(targetZipFile string) {
	logger.Info("Download existing IFlow from tenant for comparison")
	os.Setenv("OUTPUT_FILE", targetZipFile)
	os.Setenv("IFLOW_VER", "active")
	_, err := runner.JavaCmd("io.github.engswee.flashpipe.cpi.exec.DownloadDesignTimeArtifact", mavenRepoLocation, flashpipeLocation, log4jFile)
	logger.ExitIfErrorWithMsg(err, "Execution of java command failed")
}

func compareIFlowContents(workDir string, zipFile string, gitSrcDir string, iflowId string, iflowName string, scriptMap string) (changesFound bool, err error) {
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
