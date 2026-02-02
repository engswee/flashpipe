package cmd

import (
	"fmt"
	"time"

	"github.com/engswee/flashpipe/internal/analytics"
	"github.com/engswee/flashpipe/internal/api"
	"github.com/engswee/flashpipe/internal/config"
	"github.com/engswee/flashpipe/internal/file"
	"github.com/engswee/flashpipe/internal/httpclnt"
	"github.com/engswee/flashpipe/internal/str"
	"github.com/engswee/flashpipe/internal/sync"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func NewArtifactCommand() *cobra.Command {
	artifactCmd := &cobra.Command{
		Use:   "artifact",
		Short: "Create/update artifacts",
		Long: `Create or update artifacts on the
SAP Integration Suite tenant.

Configuration:
  Settings can be loaded from the global config file (--config) under the
  'update.artifact' section. CLI flags override config file settings.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Validate the artifact type
			artifactType := config.GetStringWithFallback(cmd, "artifact-type", "update.artifact.artifactType")
			switch artifactType {
			case "MessageMapping", "ScriptCollection", "Integration", "ValueMapping":
			default:
				return fmt.Errorf("invalid value for --artifact-type = %v", artifactType)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			startTime := time.Now()
			if err = runUpdateArtifact(cmd); err != nil {
				cmd.SilenceUsage = true
			}
			analytics.Log(cmd, err, startTime)
			return
		},
	}

	// Define cobra flags, the default value has the lowest (least significant) precedence
	// Note: These can be set in config file under 'update.artifact' key
	artifactCmd.Flags().String("artifact-id", "", "ID of artifact (config: update.artifact.artifactId)")
	artifactCmd.Flags().String("artifact-name", "", "Name of artifact. Defaults to artifact-id value when not provided (config: update.artifact.artifactName)")
	artifactCmd.Flags().String("package-id", "", "ID of Integration Package (config: update.artifact.packageId)")
	artifactCmd.Flags().String("package-name", "", "Name of Integration Package. Defaults to package-id value when not provided (config: update.artifact.packageName)")
	artifactCmd.Flags().String("dir-artifact", "", "Directory containing contents of designtime artifact (config: update.artifact.dirArtifact)")
	artifactCmd.Flags().String("file-param", "", "Use a different parameters.prop file instead of the default in src/main/resources/ (config: update.artifact.fileParam)")
	artifactCmd.Flags().String("file-manifest", "", "Use a different MANIFEST.MF file instead of the default in META-INF/ (config: update.artifact.fileManifest)")
	artifactCmd.Flags().String("dir-work", "/tmp", "Working directory for in-transit files (config: update.artifact.dirWork)")
	artifactCmd.Flags().StringSlice("script-collection-map", nil, "Comma-separated source-target ID pairs for converting script collection references during create/update (config: update.artifact.scriptCollectionMap)")
	artifactCmd.Flags().String("artifact-type", "Integration", "Artifact type. Allowed values: Integration, MessageMapping, ScriptCollection, ValueMapping (config: update.artifact.artifactType)")
	// TODO - another flag for replacing value mapping in QAS?

	_ = artifactCmd.MarkFlagRequired("artifact-id")
	_ = artifactCmd.MarkFlagRequired("package-id")
	_ = artifactCmd.MarkFlagRequired("dir-artifact")

	return artifactCmd
}

func runUpdateArtifact(cmd *cobra.Command) error {
	// Support reading from config file under 'update.artifact' key
	artifactType := config.GetStringWithFallback(cmd, "artifact-type", "update.artifact.artifactType")
	log.Info().Msgf("Executing update artifact %v command", artifactType)

	artifactId := config.GetStringWithFallback(cmd, "artifact-id", "update.artifact.artifactId")
	artifactName := config.GetStringWithFallback(cmd, "artifact-name", "update.artifact.artifactName")
	packageId := config.GetStringWithFallback(cmd, "package-id", "update.artifact.packageId")
	packageName := config.GetStringWithFallback(cmd, "package-name", "update.artifact.packageName")
	// Default package name to package ID if it is not provided
	if packageName == "" {
		log.Info().Msgf("Using package ID %v as package name", packageId)
		packageName = packageId
	}
	artifactDir, err := config.GetStringWithEnvExpandAndFallback(cmd, "dir-artifact", "update.artifact.dirArtifact")
	if err != nil {
		return fmt.Errorf("security alert for --dir-artifact: %w", err)
	}
	parametersFile := config.GetStringWithFallback(cmd, "file-param", "update.artifact.fileParam")
	manifestFile := config.GetStringWithFallback(cmd, "file-manifest", "update.artifact.fileManifest")
	workDir, err := config.GetStringWithEnvExpandAndFallback(cmd, "dir-work", "update.artifact.dirWork")
	if err != nil {
		return fmt.Errorf("security alert for --dir-work: %w", err)
	}
	scriptMap := str.TrimSlice(config.GetStringSliceWithFallback(cmd, "script-collection-map", "update.artifact.scriptCollectionMap"))

	defaultParamFile := fmt.Sprintf("%v/src/main/resources/parameters.prop", artifactDir)
	if parametersFile == "" {
		parametersFile = defaultParamFile
	} else if parametersFile != defaultParamFile {
		log.Info().Msgf("Using %v as parameters.prop file", parametersFile)
		err := file.CopyFile(parametersFile, defaultParamFile)
		if err != nil {
			return err
		}
	}

	defaultManifestFile := fmt.Sprintf("%v/META-INF/MANIFEST.MF", artifactDir)
	if manifestFile == "" {
		manifestFile = defaultManifestFile
	} else if manifestFile != defaultManifestFile {
		log.Info().Msgf("Using %v as MANIFEST.MF file", manifestFile)
		err := file.CopyFile(manifestFile, defaultManifestFile)
		if err != nil {
			return err
		}
	}

	// Default artifact name from Manifest file or artifact ID
	if artifactName == "" {
		headers, err := sync.GetManifestHeaders(manifestFile)
		if err != nil {
			return err
		}
		bundleName := headers.Get("Bundle-Name")
		// remove spaces due to length of bundle name exceeding MANIFEST.MF width
		bundleName = str.TrimManifestField(bundleName, 72)
		if bundleName != "" {
			log.Info().Msgf("Using %v from Bundle-Name in MANIFEST.MF as artifact name", bundleName)
			artifactName = bundleName
		} else {
			log.Info().Msgf("Using artifact ID %v as artifact name", artifactId)
			artifactName = artifactId
		}
	}

	// Initialise HTTP executer
	serviceDetails := api.GetServiceDetails(cmd)
	exe := api.InitHTTPExecuter(serviceDetails)

	// Create integration package first if required
	err = createPackage(packageId, packageName, exe)
	if err != nil {
		return err
	}

	synchroniser := sync.New(exe)

	err = synchroniser.SingleArtifactToTenant(artifactId, artifactName, artifactType, packageId, artifactDir, workDir, parametersFile, scriptMap)
	if err != nil {
		return err
	}
	return nil
}

func createPackage(packageId string, packageName string, exe *httpclnt.HTTPExecuter) error {
	// Check if integration package exists
	ip := api.NewIntegrationPackage(exe)
	_, _, packageExists, err := ip.Get(packageId)
	if err != nil {
		return err
	}

	if !packageExists {
		jsonData := new(api.PackageSingleData)
		jsonData.Root.Id = packageId
		jsonData.Root.Name = packageName
		jsonData.Root.ShortText = packageId
		jsonData.Root.Version = "1.0.0"
		err = ip.Create(jsonData)
		if err != nil {
			return err
		}
		log.Info().Msgf("Integration package %v created", packageId)
	}
	return nil
}
