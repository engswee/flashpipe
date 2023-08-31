package cmd

import (
	"fmt"
	"github.com/engswee/flashpipe/internal/config"
	"github.com/engswee/flashpipe/internal/logger"
	"github.com/engswee/flashpipe/internal/odata"
	"github.com/engswee/flashpipe/internal/str"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"time"
)

func NewDeployCommand() *cobra.Command {

	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy designtime artifact to runtime",
		Long: `Deploy artifact from designtime to
runtime of SAP Integration Suite tenant.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
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
			runDeploy(cmd)
		},
	}

	// Define cobra flags, the default value has the lowest (least significant) precedence
	deployCmd.Flags().String("artifact-ids", "", "Comma separated list of artifact IDs")
	deployCmd.Flags().Int("delay-length", 30, "Delay (in seconds) between each check of artifact deployment status")
	deployCmd.Flags().Int("max-check-limit", 10, "Max number of times to check for artifact deployment status")
	// To set to false, use --compare-versions=false
	deployCmd.Flags().Bool("compare-versions", true, "Perform version comparison of design time against runtime before deployment")
	deployCmd.Flags().String("artifact-type", "Integration", "Artifact type. Allowed values: Integration, MessageMapping, ScriptCollection, ValueMapping")

	return deployCmd
}

func runDeploy(cmd *cobra.Command) {
	serviceDetails := odata.GetServiceDetails(cmd)

	artifactType := config.GetString(cmd, "artifact-type")
	log.Info().Msgf("Executing deploy %v command", artifactType)

	artifactIds := config.GetMandatoryString(cmd, "artifact-ids")
	delayLength := config.GetInt(cmd, "delay-length")
	maxCheckLimit := config.GetInt(cmd, "max-check-limit")
	compareVersions := config.GetBool(cmd, "compare-versions")

	deployArtifacts(artifactIds, artifactType, delayLength, maxCheckLimit, compareVersions, serviceDetails)
}

func deployArtifacts(delimitedIds string, artifactType string, delayLength int, maxCheckLimit int, compareVersions bool, serviceDetails *odata.ServiceDetails) {

	// Extract IDs from delimited values
	ids := str.ExtractDelimitedValues(delimitedIds, ",")

	// Initialise HTTP executer
	exe := odata.InitHTTPExecuter(serviceDetails)

	// Initialise designtime artifact
	dt := odata.NewDesigntimeArtifact(artifactType, exe)

	// Initialised runtime artifact
	rt := odata.NewRuntime(exe)

	// Loop and deploy each artifact
	for i, id := range ids {
		log.Info().Msgf("Processing artifact %d - %v", i+1, id)
		err := deploySingle(dt, rt, id, compareVersions)
		// TODO - PRIO1 write error wrapper - https://go.dev/blog/errors-are-values
		logger.ExitIfError(err)
	}

	// Check deployment status of artifacts
	for i, id := range ids {
		err := checkDeploymentStatus(rt, delayLength, maxCheckLimit, id)
		logger.ExitIfError(err)
		// TODO - PRIO1 write error wrapper - https://go.dev/blog/errors-are-values

		log.Info().Msgf("Artifact %d - %v deployed successfully", i+1, id)
	}

	log.Info().Msg("🏆 Artifact(s) deployment completed successfully")
}

func deploySingle(artifact odata.DesigntimeArtifact, runtime *odata.Runtime, id string, compareVersions bool) error {
	designtimeVer, exists, err := artifact.Get(id, "active")
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("Designtime artifact %v does not exist", id)
	}

	if compareVersions == true {
		runtimeVer, _, err := runtime.Get(id)
		if err != nil {
			return err
		}

		// Compare designtime version with runtime version to determine if deployment is needed
		log.Info().Msg("Comparing designtime version with runtime version")
		log.Debug().Msgf("Designtime version = %s. Runtime version = %s", designtimeVer, runtimeVer)
		if designtimeVer == runtimeVer {
			log.Info().Msgf("Artifact %v with version %v already deployed. Skipping runtime deployment", id, runtimeVer)
		} else {
			log.Info().Msgf("🚀 Artifact previously not deployed, or versions differ. Proceeding to deploy artifact %v with version %v", id, designtimeVer)
			err = artifact.Deploy(id)
			if err != nil {
				return err
			}
			log.Info().Msgf("Artifact %v deployment triggered", id)
		}
	} else {
		log.Info().Msgf("🚀 Proceeding to deploy artifact %v with version %v", id, designtimeVer)
		err = artifact.Deploy(id)
		if err != nil {
			return err
		}
		log.Info().Msgf("Artifact %v deployment triggered", id)
	}
	return nil
}

func checkDeploymentStatus(runtime *odata.Runtime, delayLength int, maxCheckLimit int, id string) error {
	log.Info().Msgf("Checking runtime status for artifact %v every %d seconds up to %d times", id, delayLength, maxCheckLimit)

	for i := 0; i < maxCheckLimit; i++ {
		version, status, err := runtime.Get(id)
		if err != nil {
			return err
		}
		log.Info().Msgf("Check %d - Current artifact runtime status = %s", i+1, status)
		if version == "NOT_DEPLOYED" {
			time.Sleep(time.Duration(delayLength) * time.Second)
			continue
		}
		if status == "STARTED" {
			return nil
		} else if status != "STARTING" {
			errorMessage, err := runtime.GetErrorInfo(id)
			if err != nil {
				return err
			}
			return fmt.Errorf("Artifact deployment unsuccessful, ended with status %s. Error message = %s", status, errorMessage)
		}
		if i == (maxCheckLimit-1) && status != "STARTED" {
			return fmt.Errorf("Artifact status remained in %s after %d checks", status, maxCheckLimit)
		}
		time.Sleep(time.Duration(delayLength) * time.Second)
	}
	return nil
}
