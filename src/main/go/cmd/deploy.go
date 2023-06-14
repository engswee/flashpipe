package cmd

import (
	"errors"
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
	Short: "Deploy integration flow to runtime",
	Long: `Deploy integration flow from design time to
runtime of SAP Integration Suite tenant.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Executing deploy command")

		iFlows := setMandatoryVariable(deployViper, "iflow.id", "IFLOW_ID")
		setOptionalVariable(deployViper, "delaylength", "DELAY_LENGTH")
		setOptionalVariable(deployViper, "maxchecklimit", "MAX_CHECK_LIMIT")
		setOptionalVariable(deployViper, "compareversions", "COMPARE_VERSIONS")

		deployArtifacts(iFlows)
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
	setStringFlagAndBind(deployViper, deployCmd, "iflow.id", "", "Comma separated list of Integration Flow IDs [or set environment IFLOW_ID]")
	setIntFlagAndBind(deployViper, deployCmd, "delaylength", 30, "Delay (in seconds) between each check of IFlow deployment status [or set environment DELAY_LENGTH]")
	setIntFlagAndBind(deployViper, deployCmd, "maxchecklimit", 10, "Max number of times to check for IFlow deployment status [or set environment MAX_CHECK_LIMIT]")
	setBoolFlagAndBind(deployViper, deployCmd, "compareversions", true, "Perform version comparison of design time against runtime before deployment [or set environment COMPARE_VERSIONS]")
}

func deployArtifacts(iFlows string) {
	//https://www.digitalocean.com/community/tutorials/how-to-use-json-in-go
	//https://www.digitalocean.com/community/tutorials/how-to-make-http-requests-in-go
	//https://pkg.go.dev/net/http
	//https://www.alexedwards.net/blog/basic-authentication-in-go
	//https://github.com/golang/oauth2
	//https://www.sohamkamani.com/golang/oauth/

	// Extract IDs from delimited values
	ids := str.ExtractDelimitedValues(iFlows, ",")

	delayLength := deployViper.GetInt("delaylength")
	maxCheckLimit := deployViper.GetInt("maxchecklimit")
	compareVersions := deployViper.GetBool("compareversions")

	// Initialise HTTP executer
	exe := httpclnt.New(oauthHost, oauthTokenPath, oauthClientId, oauthClientSecret, basicUserId, basicPassword, tmnHost)

	// Initialise designtime artifact
	dt := designtime.NewIntegration(exe)

	// Initialised runtime artifact
	rt := odata.NewRuntime(exe)

	// Loop and deploy each IFlow
	for i, id := range ids {
		logger.Info(fmt.Sprintf("Processing IFlow %d - %v", i+1, id))
		err := deploySingle(dt, rt, id, compareVersions)
		logger.ExitIfError(err)
	}

	// Check deployment status of IFlows
	for i, id := range ids {
		err := checkDeploymentStatus(rt, delayLength, maxCheckLimit, id)
		logger.ExitIfError(err)

		logger.Info(fmt.Sprintf("IFlow %d - %v deployed successfully", i+1, id))
	}

	logger.Info("🏆 IFlow(s) deployment completed successfully")
}

func deploySingle(artifact designtime.DesigntimeArtifact, runtime *odata.Runtime, id string, compareVersions bool) error {
	logger.Info("Getting designtime version of IFlow")
	designtimeVer, err := artifact.GetVersion(id, "active")
	if err != nil {
		return err
	}

	if compareVersions == true {
		logger.Info("Getting runtime version of IFlow")
		runtimeVer, err := runtime.GetVersion(id)
		if err != nil {
			return err
		}

		// Compare designtime version with runtime version to determine if deployment is needed
		logger.Info("Comparing designtime version with runtime version")
		logger.Debug(fmt.Sprintf("Designtime version = %s. Runtime version = %s", designtimeVer, runtimeVer))
		if designtimeVer == runtimeVer {
			logger.Info(fmt.Sprintf("IFlow %v with version %v already deployed. Skipping runtime deployment", id, runtimeVer))
		} else {
			logger.Info(fmt.Sprintf("🚀 IFlow previously not deployed, or versions differ. Proceeding to deploy IFlow %v with version %v", id, designtimeVer))
			err = artifact.Deploy(id)
			if err != nil {
				return err
			}
			logger.Info(fmt.Sprintf("IFlow %v deployment triggered", id))
		}
	} else {
		logger.Info(fmt.Sprintf("🚀 Proceeding to deploy IFlow %v with version %v", id, designtimeVer))
		err = artifact.Deploy(id)
		if err != nil {
			return err
		}
		logger.Info(fmt.Sprintf("IFlow %v deployment triggered", id))
	}
	return nil
}

func checkDeploymentStatus(runtime *odata.Runtime, delayLength int, maxCheckLimit int, id string) error {
	logger.Info(fmt.Sprintf("Checking deployment status for IFlow %v every %d seconds up to %d times", id, delayLength, maxCheckLimit))

	for i := 0; i < maxCheckLimit; i++ {
		// Delay to allow deployment to start before checking the status
		time.Sleep(time.Duration(delayLength) * time.Second)

		status, err := runtime.GetStatus(id)
		if err != nil {
			return err
		}

		logger.Info(fmt.Sprintf("Check %d - Current IFlow status = %s", i, status))
		if status != "STARTING" {
			if status == "STARTED" {
				break
			} else {
				errorMessage, err := runtime.GetErrorInfo(id)
				if err != nil {
					return err
				}
				return errors.New(fmt.Sprintf("IFlow deployment unsuccessful, ended with status %s. Error message = %s", status, errorMessage))
			}
		}
		if i == (maxCheckLimit-1) && status != "STARTED" {
			return errors.New(fmt.Sprintf("IFlow status remained in %s after %d checks", status, maxCheckLimit))
		}
	}
	return nil
}
