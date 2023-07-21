package cmd

import (
	"fmt"
	"github.com/engswee/flashpipe/config"
	"github.com/engswee/flashpipe/file"
	"github.com/engswee/flashpipe/httpclnt"
	"github.com/engswee/flashpipe/logger"
	"github.com/engswee/flashpipe/odata"
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
	artifactCmd.Flags().String("file-manifest", "", "Use to a different parameters.prop file instead of the default in src/main/resources/ [or set environment MANIFEST_FILE]")
	artifactCmd.Flags().String("dir-work", "/tmp", "Working directory for in-transit files [or set environment WORK_DIR]")
	artifactCmd.Flags().String("scriptmap", "", "Comma-separated source-target ID pairs for converting script collection references during create/update [or set environment SCRIPT_COLLECTION_MAP]")
	artifactCmd.Flags().String("artifact-type", "Integration", "Artifact type. Allowed values: Integration, MessageMapping, ScriptCollection, ValueMapping")

	return artifactCmd
}

func runUpdateArtifact(cmd *cobra.Command) {
	artifactType := config.GetString(cmd, "artifact-type")
	log.Info().Msgf("Executing update artifact %v command", artifactType)

	artifactId := config.GetMandatoryString(cmd, "artifact-id")
	artifactName := config.GetMandatoryString(cmd, "artifact-name")
	packageId := config.GetMandatoryString(cmd, "package-id")
	packageName := config.GetMandatoryString(cmd, "package-name")
	gitSrcDir := config.GetMandatoryString(cmd, "dir-gitsrc")
	parametersFile := config.GetString(cmd, "file-param")
	manifestFile := config.GetString(cmd, "file-manifest")
	workDir := config.GetString(cmd, "dir-work")
	scriptMap := config.GetString(cmd, "scriptmap")

	// TODO - ID and package name from file rather than parameter

	defaultParamFile := fmt.Sprintf("%v/src/main/resources/parameters.prop", gitSrcDir)
	if parametersFile == "" {
		parametersFile = defaultParamFile
	} else if parametersFile != defaultParamFile {
		log.Info().Msgf("Using %v as parameters.prop file", parametersFile)
		err := file.CopyFile(parametersFile, defaultParamFile)
		logger.ExitIfError(err)
	}

	defaultManifestFile := fmt.Sprintf("%v/META-INF/MANIFEST.MF", gitSrcDir)
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

	// Check if IFlow already exist on tenant
	exists, err := artifactExists(artifactId, artifactType, packageId, dt, ip)
	logger.ExitIfError(err)
	if !exists {
		// Create artifact
		log.Info().Msgf("Artifact %v will be created", artifactId)

		err = prepareUploadDir(workDir, gitSrcDir, artifactType)
		logger.ExitIfError(err)

		ip = odata.NewIntegrationPackage(exe)
		packageExists, err := ip.Exists(packageId)
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

		err = createArtifact(artifactId, artifactName, packageId, workDir+"/upload", scriptMap, dt)
		logger.ExitIfError(err)

		// TODO - manifest normalisation currently not in place as using workaround MANIFEST.MF replacement

		// Update the script collection in IFlow BPMN2 XML before upload
		err = file.UpdateBPMN(gitSrcDir, scriptMap)
		logger.ExitIfError(err)

		log.Info().Msg("🏆 Artifact created successfully")

	} else {
		// Update IFlow
		log.Info().Msg("Checking if designtime artifact needs to be updated")
		// 1 - Download artifact content from tenant
		zipFile := fmt.Sprintf("%v/%v.zip", workDir, artifactId)
		err = odata.Download(zipFile, artifactId, dt)
		logger.ExitIfError(err)
		// 2 - Diff contents from tenant against Git
		// TODO - refactor and combine with implementation used in synchroniser
		changesFound, err := compareIFlowContents(workDir, zipFile, gitSrcDir, artifactId, artifactName, scriptMap, dt)
		logger.ExitIfError(err)

		if changesFound == true {
			log.Info().Msg("Changes found in IFlow. IFlow design will be updated in CPI tenant")
			err = prepareUploadDir(workDir, gitSrcDir, artifactType)
			logger.ExitIfError(err)
			err = updateArtifact(artifactId, artifactName, packageId, workDir+"/upload", scriptMap, dt)
			logger.ExitIfError(err)

			// If runtime has the same version no, then undeploy it, otherwise it gets skipped during deployment
			designtimeVersion, err := dt.GetVersion(artifactId, "active")
			logger.ExitIfError(err)
			r := odata.NewRuntime(exe)
			runtimeVersion, err := r.GetVersion(artifactId)
			logger.ExitIfError(err)
			if runtimeVersion == designtimeVersion {
				log.Info().Msg("Undeploying existing runtime artifact with same version number due to changes in design")
				err = r.UnDeploy(artifactId)
				logger.ExitIfError(err)
			}

			log.Info().Msg("🏆 IFlow design updated successfully")
		} else {
			log.Info().Msg("🏆 No changes detected. IFlow design does not need to be updated")
		}

		// 4 - Update the configuration of the integration artifact based on parameters.prop file
		if artifactType == "Integration" {
			log.Info().Msg("Updating configured parameter(s) of Integration designtime artifact where necessary")
			err = updateConfiguration(artifactId, parametersFile, exe)
			logger.ExitIfError(err)
		}
	}
}

func prepareUploadDir(workDir string, gitSrcDir string, artifactType string) (err error) {
	// Clean up previous uploads
	iFlowDir := workDir + "/upload"
	err = os.RemoveAll(iFlowDir)
	if err != nil {
		return
	}

	// Copy META-INF and resources separately so that other directories like QA, STG, PRD not copied
	err = file.CopyDir(gitSrcDir+"/META-INF", iFlowDir+"/META-INF")
	if err != nil {
		return
	}
	// TODO - for value mapping it only has value_mapping.xml file - implement as method of each type
	if artifactType == "ValueMapping" {
		file.CopyFile(gitSrcDir+"/value_mapping.xml", iFlowDir+"/value_mapping.xml")
	} else {
		err = file.CopyDir(gitSrcDir+"/src/main/resources", iFlowDir+"/src/main/resources")
		if err != nil {
			return
		}
	}
	return
}

func compareIFlowContents(workDir string, zipFile string, gitSrcDir string, iflowId string, iflowName string, scriptMap string, dt odata.DesigntimeArtifact) (bool, error) {
	err := os.RemoveAll(workDir + "/download")
	if err != nil {
		return false, err
	}

	log.Info().Msgf("Unzipping downloaded IFlow artifact %v to %v/download", zipFile, workDir)
	err = file.UnzipSource(zipFile, workDir+"/download")
	if err != nil {
		return false, err
	}
	// Update the script collection in IFlow BPMN2 XML before diff comparison
	err = file.UpdateBPMN(gitSrcDir, scriptMap)
	if err != nil {
		return false, err
	}

	// TODO - Update the MANIFEST.MF file with script collection conversions

	// Diff directories excluding parameters.prop
	// - Any configured value will remain in IFlow even if the IFlow is replaced and the parameter is no longer used
	// - Therefore diff of parameters.prop may come up with false differences
	dirDiffer := dt.DiffContent(workDir+"/download", gitSrcDir)
	if dirDiffer {
		return true, nil
	} else {
		return false, nil
	}
}

func artifactExists(artifactId string, artifactType string, packageId string, dt odata.DesigntimeArtifact, ip *odata.IntegrationPackage) (bool, error) {
	exists, err := dt.Exists(artifactId, "active")
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
		version, err := r.GetVersion(artifactId)
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
