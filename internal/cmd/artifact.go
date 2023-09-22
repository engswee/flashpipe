package cmd

import (
	"fmt"
	"github.com/engswee/flashpipe/internal/config"
	"github.com/engswee/flashpipe/internal/file"
	"github.com/engswee/flashpipe/internal/httpclnt"
	"github.com/engswee/flashpipe/internal/logger"
	"github.com/engswee/flashpipe/internal/odata"
	"github.com/magiconair/properties"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
)

func NewArtifactCommand() *cobra.Command {

	artifactCmd := &cobra.Command{
		Use:   "artifact",
		Short: "Create/update artifacts",
		Long: `Create or update artifacts on the
SAP Integration Suite tenant.`,
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
			runUpdateArtifact(cmd)
		},
	}

	// Define cobra flags, the default value has the lowest (least significant) precedence
	artifactCmd.Flags().String("artifact-id", "", "ID of artifact")
	artifactCmd.Flags().String("artifact-name", "", "Name of artifact")
	artifactCmd.Flags().String("package-id", "", "ID of Integration Package")
	artifactCmd.Flags().String("package-name", "", "Name of Integration Package")
	artifactCmd.Flags().String("dir-artifact", "", "Directory containing contents of designtime artifact")
	artifactCmd.Flags().String("file-param", "", "Use a different parameters.prop file instead of the default in src/main/resources/ ")
	artifactCmd.Flags().String("file-manifest", "", "Use a different MANIFEST.MF file instead of the default in META-INF/")
	artifactCmd.Flags().String("dir-work", "/tmp", "Working directory for in-transit files")
	artifactCmd.Flags().String("script-collection-map", "", "Comma-separated source-target ID pairs for converting script collection references during create/update")
	artifactCmd.Flags().String("artifact-type", "Integration", "Artifact type. Allowed values: Integration, MessageMapping, ScriptCollection, ValueMapping")
	// TODO - another flag for replacing value mapping in QAS?

	_ = artifactCmd.MarkFlagRequired("artifact-id")
	_ = artifactCmd.MarkFlagRequired("artifact-name")
	_ = artifactCmd.MarkFlagRequired("package-id")
	_ = artifactCmd.MarkFlagRequired("package-name")
	_ = artifactCmd.MarkFlagRequired("dir-artifact")

	return artifactCmd
}

func runUpdateArtifact(cmd *cobra.Command) {
	artifactType := config.GetString(cmd, "artifact-type")
	log.Info().Msgf("Executing update artifact %v command", artifactType)

	artifactId := config.GetString(cmd, "artifact-id")
	artifactName := config.GetString(cmd, "artifact-name")
	packageId := config.GetString(cmd, "package-id")
	packageName := config.GetString(cmd, "package-name")
	artifactDir := config.GetString(cmd, "dir-artifact")
	parametersFile := config.GetString(cmd, "file-param")
	manifestFile := config.GetString(cmd, "file-manifest")
	workDir := config.GetString(cmd, "dir-work")
	scriptMap := config.GetString(cmd, "script-collection-map")

	defaultParamFile := fmt.Sprintf("%v/src/main/resources/parameters.prop", artifactDir)
	if parametersFile == "" {
		parametersFile = defaultParamFile
	} else if parametersFile != defaultParamFile {
		log.Info().Msgf("Using %v as parameters.prop file", parametersFile)
		err := file.CopyFile(parametersFile, defaultParamFile)
		logger.ExitIfError(err)
	}

	defaultManifestFile := fmt.Sprintf("%v/META-INF/MANIFEST.MF", artifactDir)
	if manifestFile == "" {
		manifestFile = defaultManifestFile
	} else if manifestFile != defaultManifestFile {
		log.Info().Msgf("Using %v as MANIFEST.MF file", manifestFile)
		err := file.CopyFile(manifestFile, defaultManifestFile)
		logger.ExitIfError(err)
	}

	// Initialise HTTP executer
	serviceDetails := odata.GetServiceDetails(cmd)
	exe := odata.InitHTTPExecuter(serviceDetails)

	// Initialise designtime artifact
	dt := odata.NewDesigntimeArtifact(artifactType, exe)
	ip := odata.NewIntegrationPackage(exe)

	// Check if artifact already exist on tenant
	exists, err := artifactExists(artifactId, artifactType, packageId, dt, ip)
	logger.ExitIfError(err)
	if !exists {
		// Create artifact
		log.Info().Msgf("Artifact %v will be created", artifactId)
		// Create integration package first if required
		ip = odata.NewIntegrationPackage(exe)
		_, _, packageExists, err := ip.Get(packageId)
		logger.ExitIfError(err)
		if !packageExists {
			jsonData := new(odata.PackageSingleData)
			jsonData.Root.Id = packageId
			jsonData.Root.Name = packageName
			jsonData.Root.ShortText = packageId
			jsonData.Root.Version = "1.0.0"
			err := ip.Create(jsonData)
			logger.ExitIfError(err)
			log.Info().Msgf("Integration package %v created", packageId)
		}

		// TODO - manifest normalisation currently not in place as using workaround MANIFEST.MF replacement

		// Update the script collection in IFlow BPMN2 XML before upload
		if artifactType == "Integration" {
			err = file.UpdateBPMN(artifactDir, scriptMap)
			logger.ExitIfError(err)
		}

		err = prepareUploadDir(workDir, artifactDir, dt)
		logger.ExitIfError(err)

		err = createArtifact(artifactId, artifactName, packageId, workDir+"/upload", dt)
		logger.ExitIfError(err)

		log.Info().Msg("🏆 Designtime artifact created successfully")

	} else {
		// Update artifact
		log.Info().Msg("Checking if designtime artifact needs to be updated")
		// 1 - Download artifact content from tenant
		zipFile := fmt.Sprintf("%v/%v.zip", workDir, artifactId)
		err = dt.Download(zipFile, artifactId)
		logger.ExitIfError(err)
		// 2 - Diff contents from tenant against Git
		changesFound, err := compareArtifactContents(workDir, zipFile, artifactDir, scriptMap, dt)
		logger.ExitIfError(err)

		if changesFound == true {
			log.Info().Msg("Changes found in designtime artifact. Designtime artifact will be updated in CPI tenant")
			err = prepareUploadDir(workDir, artifactDir, dt)
			logger.ExitIfError(err)
			err = updateArtifact(artifactId, artifactName, packageId, workDir+"/upload", dt)
			logger.ExitIfError(err)

			// If runtime has the same version no, then undeploy it, otherwise it gets skipped during deployment
			designtimeVersion, _, err := dt.Get(artifactId, "active")
			logger.ExitIfError(err)
			r := odata.NewRuntime(exe)
			runtimeVersion, _, err := r.Get(artifactId)
			logger.ExitIfError(err)
			if runtimeVersion == designtimeVersion {
				log.Info().Msg("Undeploying existing runtime artifact with same version number due to changes in design")
				err = r.UnDeploy(artifactId)
				logger.ExitIfError(err)
			}

			log.Info().Msg("🏆 Designtime artifact updated successfully")
		} else {
			log.Info().Msg("🏆 No changes detected. Designtime artifact does not need to be updated")
		}

		// 4 - Update the configuration of the integration artifact based on parameters.prop file
		if artifactType == "Integration" && file.Exists(parametersFile) {
			log.Info().Msg("Updating configured parameter(s) of Integration designtime artifact where necessary")
			err = updateConfiguration(artifactId, parametersFile, exe)
			logger.ExitIfError(err)
		}
	}
}

func prepareUploadDir(workDir string, artifactDir string, dt odata.DesigntimeArtifact) (err error) {
	// Clean up previous uploads
	uploadDir := workDir + "/upload"
	err = os.RemoveAll(uploadDir)
	if err != nil {
		return
	}
	err = dt.CopyContent(artifactDir, uploadDir)
	return
}

func compareArtifactContents(workDir string, zipFile string, artifactDir string, scriptMap string, dt odata.DesigntimeArtifact) (bool, error) {
	tgtDir := fmt.Sprintf("%v/download", workDir)
	err := os.RemoveAll(tgtDir)
	if err != nil {
		return false, err
	}

	log.Info().Msgf("Unzipping downloaded designtime artifact %v to %v/download", zipFile, workDir)
	err = file.UnzipSource(zipFile, tgtDir)
	if err != nil {
		return false, err
	}

	return dt.CompareContent(artifactDir, tgtDir, scriptMap, "local")
}

func artifactExists(artifactId string, artifactType string, packageId string, dt odata.DesigntimeArtifact, ip *odata.IntegrationPackage) (bool, error) {
	_, exists, err := dt.Get(artifactId, "active")
	if err != nil {
		return false, err
	}
	if exists {
		log.Info().Msgf("Active version of artifact %v exists", artifactId)
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
		log.Info().Msgf("Active version of artifact %v does not exist", artifactId)
		return false, nil
	}
}

func createArtifact(artifactId string, artifactName string, packageId string, artifactDir string, dt odata.DesigntimeArtifact) error {
	err := dt.Create(artifactId, artifactName, packageId, artifactDir)
	if err != nil {
		return err
	}
	return nil
}

func updateArtifact(artifactId string, artifactName string, packageId string, artifactDir string, dt odata.DesigntimeArtifact) error {
	err := dt.Update(artifactId, artifactName, packageId, artifactDir)
	if err != nil {
		return err
	}
	return nil
}

func updateConfiguration(artifactId string, parametersFile string, exe *httpclnt.HTTPExecuter) error {
	// Get configured parameters from tenant
	c := odata.NewConfiguration(exe)
	tenantParameters, err := c.Get(artifactId, "active")
	if err != nil {
		return err
	}

	// Get parameters from parameters.prop file
	log.Info().Msgf("Getting parameters from %v file", parametersFile)
	fileParameters := properties.MustLoadFile(parametersFile, properties.UTF8)

	log.Info().Msg("Comparing parameters and updating where necessary")
	atLeastOneUpdated := false
	for _, result := range tenantParameters.Root.Results {
		if result.DataType != "custom:schedule" { // TODO - handle translation to Cron
			// Skip updating for schedulers which require translation to Cron values
			fileValue := fileParameters.GetString(result.ParameterKey, "")
			if fileValue != "" && fileValue != result.ParameterValue {
				log.Info().Msgf("Parameter %v to be updated from %v to %v", result.ParameterKey, result.ParameterValue, fileValue)
				err = c.Update(artifactId, "active", result.ParameterKey, fileValue)
				if err != nil {
					return err
				}
				atLeastOneUpdated = true
			}
		}
	}
	if atLeastOneUpdated {
		r := odata.NewRuntime(exe)
		version, _, err := r.Get(artifactId)
		if err != nil {
			return err
		}
		if version == "NOT_DEPLOYED" {
			log.Info().Msg("🏆 No existing runtime artifact deployed")
		} else {
			log.Info().Msg("🏆 Undeploying existing runtime artifact due to changes in configured parameters")
			err = r.UnDeploy(artifactId)
			if err != nil {
				return err
			}
		}
	} else {
		log.Info().Msg("🏆 No updates required for configured parameters")
	}
	return nil
}
