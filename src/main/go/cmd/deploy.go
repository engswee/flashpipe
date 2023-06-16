package cmd

import (
	"fmt"
	"github.com/engswee/flashpipe/httpclnt"
	"github.com/engswee/flashpipe/logger"
	"github.com/engswee/flashpipe/odata"
	"github.com/engswee/flashpipe/odata/designtime"
	"github.com/engswee/flashpipe/str"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"time"
)

var deployViper = viper.New()

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy designtime artifact to runtime",
	Long: `Deploy artifact from designtime to
runtime of SAP Integration Suite tenant.`,
	Args: func(cmd *cobra.Command, args []string) error {
		// Validate the artifact type
		artifactType := deployViper.GetString("artifact.type")
		switch artifactType {
		case "MESSAGE_MAPPING", "SCRIPT_COLLECTION", "INTEGRATION_FLOW":
		default:
			return fmt.Errorf("invalid value for --artifact-type = %v", artifactType)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		artifactType := deployViper.GetString("artifact.type")
		logger.Info(fmt.Sprintf("Executing deploy %v command", artifactType))

		artifactIds := setMandatoryVariable(deployViper, "ids", "IFLOW_ID")
		setOptionalVariable(deployViper, "delaylength", "DELAY_LENGTH")
		setOptionalVariable(deployViper, "maxchecklimit", "MAX_CHECK_LIMIT")
		setOptionalVariable(deployViper, "compareversions", "COMPARE_VERSIONS")

		deployArtifacts(artifactIds, artifactType)
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deployCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	setStringFlagAndBind(deployViper, deployCmd, "ids", "", "Comma separated list of artifact IDs [or set environment IFLOW_ID]")
	setIntFlagAndBind(deployViper, deployCmd, "delaylength", 30, "Delay (in seconds) between each check of artifact deployment status [or set environment DELAY_LENGTH]")
	setIntFlagAndBind(deployViper, deployCmd, "maxchecklimit", 10, "Max number of times to check for artifact deployment status [or set environment MAX_CHECK_LIMIT]")
	setBoolFlagAndBind(deployViper, deployCmd, "compareversions", true, "Perform version comparison of design time against runtime before deployment [or set environment COMPARE_VERSIONS]")
	setStringFlagAndBind(deployViper, deployCmd, "artifact.type", "INTEGRATION_FLOW", "Artifact type. Allowed values: INTEGRATION_FLOW, MESSAGE_MAPPING, SCRIPT_COLLECTION")
}

func deployArtifacts(delimitedIds string, artifactType string) {

	// Extract IDs from delimited values
	ids := str.ExtractDelimitedValues(delimitedIds, ",")

	delayLength := deployViper.GetInt("delaylength")
	maxCheckLimit := deployViper.GetInt("maxchecklimit")
	compareVersions := deployViper.GetBool("compareversions")

	// Initialise HTTP executer
	exe := httpclnt.New(oauthHost, oauthTokenPath, oauthClientId, oauthClientSecret, basicUserId, basicPassword, tmnHost)

	// Initialise designtime artifact
	dt := designtime.GetDesigntimeArtifactByType(artifactType, exe)

	// Initialised runtime artifact
	rt := odata.NewRuntime(exe)

	// Loop and deploy each artifact
	for i, id := range ids {
		logger.Info(fmt.Sprintf("Processing artifact %d - %v", i+1, id))
		err := deploySingle(dt, rt, id, compareVersions)
		logger.ExitIfError(err)
	}

	// Delay to allow deployment to start before checking the status
	// Only applicable if there is only 1 artifact, because if there are many, then there is an inherent delay already
	if len(ids) == 1 {
		time.Sleep(time.Duration(delayLength) * time.Second)
	}

	// Check deployment status of artifacts
	for i, id := range ids {
		err := checkDeploymentStatus(rt, delayLength, maxCheckLimit, id)
		logger.ExitIfError(err)

		logger.Info(fmt.Sprintf("Artifact %d - %v deployed successfully", i+1, id))
	}

	logger.Info("🏆 Artifact(s) deployment completed successfully")
}

func deploySingle(artifact designtime.DesigntimeArtifact, runtime *odata.Runtime, id string, compareVersions bool) error {
	logger.Info("Getting designtime version of artifact")
	designtimeVer, err := artifact.GetVersion(id, "active")
	if err != nil {
		return err
	}

	if compareVersions == true {
		logger.Info("Getting runtime version of artifact")
		runtimeVer, err := runtime.GetVersion(id)
		if err != nil {
			return err
		}

		// Compare designtime version with runtime version to determine if deployment is needed
		logger.Info("Comparing designtime version with runtime version")
		logger.Debug(fmt.Sprintf("Designtime version = %s. Runtime version = %s", designtimeVer, runtimeVer))
		if designtimeVer == runtimeVer {
			logger.Info(fmt.Sprintf("Artifact %v with version %v already deployed. Skipping runtime deployment", id, runtimeVer))
		} else {
			logger.Info(fmt.Sprintf("🚀 Artifact previously not deployed, or versions differ. Proceeding to deploy artifact %v with version %v", id, designtimeVer))
			err = artifact.Deploy(id)
			if err != nil {
				return err
			}
			logger.Info(fmt.Sprintf("Artifact %v deployment triggered", id))
		}
	} else {
		logger.Info(fmt.Sprintf("🚀 Proceeding to deploy artifact %v with version %v", id, designtimeVer))
		err = artifact.Deploy(id)
		if err != nil {
			return err
		}
		logger.Info(fmt.Sprintf("Artifact %v deployment triggered", id))
	}
	return nil
}

func checkDeploymentStatus(runtime *odata.Runtime, delayLength int, maxCheckLimit int, id string) error {
	logger.Info(fmt.Sprintf("Checking runtime status for artifact %v every %d seconds up to %d times", id, delayLength, maxCheckLimit))

	for i := 0; i < maxCheckLimit; i++ {
		status, err := runtime.GetStatus(id)
		if err != nil {
			return err
		}

		logger.Info(fmt.Sprintf("Check %d - Current artifact runtime status = %s", i, status))
		if status != "STARTING" {
			if status == "STARTED" {
				break
			} else {
				errorMessage, err := runtime.GetErrorInfo(id)
				if err != nil {
					return err
				}
				return fmt.Errorf("Artifact deployment unsuccessful, ended with status %s. Error message = %s", status, errorMessage)
			}
		}
		if i == (maxCheckLimit-1) && status != "STARTED" {
			return fmt.Errorf("Artifact status remained in %s after %d checks", status, maxCheckLimit)
		}
		time.Sleep(time.Duration(delayLength) * time.Second)
	}
	return nil
}
