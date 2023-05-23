package cmd

import (
	"fmt"
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

		parametersFile := flowViper.GetString("file.param")
		defaultParamFile := fmt.Sprint(flowViper.GetString("dir.gitsrc"), "/src/main/resources/parameters.prop")
		if parametersFile != "" && parametersFile != defaultParamFile {
			logger.Info("Using", parametersFile, "as parameters.prop file")
			_, err := file.Copy(parametersFile, defaultParamFile)
			logger.CheckIfError(err)
		}
		// Check if IFlow already exist on tenant
		output, err := runner.JavaCmd("io.github.engswee.flashpipe.cpi.exec.QueryDesignTimeArtifact", mavenRepoLocation, flashpipeLocation, log4jFile)
		if strings.Contains(output, "Active version of IFlow does not exist") {
			// Upload IFlow
			uploadIFlow()
		} else if err == nil {
			// Update IFlow
			logger.Info("Checking if IFlow design needs to be updated")
		} else {
			logger.Error("Execution of java command failed")
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

func uploadIFlow() {
	logger.Info("IFlow will be uploaded to tenant")

	// Clean up previous uploads
	iFlowDir := flowViper.GetString("dir.work") + "/upload"
	err := os.RemoveAll(iFlowDir)
	logger.CheckIfError(err)

	err = os.MkdirAll(iFlowDir+"/src/main", os.ModePerm)
	logger.CheckIfError(err)

	err = file.CopyDir(flowViper.GetString("dir.gitsrc")+"/META-INF", iFlowDir+"/META-INF")
	logger.CheckIfError(err)

	err = file.CopyDir(flowViper.GetString("dir.gitsrc")+"/src/main/resources", iFlowDir+"/src/main/resources")
	logger.CheckIfError(err)
	os.Setenv("IFLOW_DIR", iFlowDir)

	_, err = runner.JavaCmd("io.github.engswee.flashpipe.cpi.exec.UploadDesignTimeArtifact", mavenRepoLocation, flashpipeLocation, log4jFile)
	logger.CheckIfErrorWithMsg(err, "Execution of java command failed")
	logger.Info("🏆 IFlow created successfully")
}
