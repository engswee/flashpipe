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
	artifactCmd.Flags().String("artifact-id", "", "ID of artifact")
	artifactCmd.Flags().String("artifact-name", "", "Name of artifact. Defaults to artifact-id value when not provided")
	artifactCmd.Flags().String("package-id", "", "ID of Integration Package")
	artifactCmd.Flags().String("package-name", "", "Name of Integration Package. Defaults to package-id value when not provided")
	artifactCmd.Flags().String("dir-artifact", "", "Directory containing contents of designtime artifact")
	artifactCmd.Flags().String("file-param", "", "Use a different parameters.prop file instead of the default in src/main/resources/ ")
	artifactCmd.Flags().String("file-manifest", "", "Use a different MANIFEST.MF file instead of the default in META-INF/")
	artifactCmd.Flags().String("dir-work", "/tmp", "Working directory for in-transit files")
	artifactCmd.Flags().StringSlice("script-collection-map", nil, "Comma-separated source-target ID pairs for converting script collection references during create/update")
	artifactCmd.Flags().String("artifact-type", "Integration", "Artifact type. Allowed values: Integration, MessageMapping, ScriptCollection, ValueMapping")
	// TODO - another flag for replacing value mapping in QAS?

	_ = artifactCmd.MarkFlagRequired("artifact-id")
	_ = artifactCmd.MarkFlagRequired("package-id")
	_ = artifactCmd.MarkFlagRequired("dir-artifact")

	return artifactCmd
}

func runUpdateArtifact(cmd *cobra.Command) error {
	artifactType := config.GetString(cmd, "artifact-type")
	log.Info().Msgf("Executing update artifact %v command", artifactType)

	artifactId := config.GetString(cmd, "artifact-id")
	artifactName := config.GetString(cmd, "artifact-name")
	packageId := config.GetString(cmd, "package-id")
	packageName := config.GetString(cmd, "package-name")
	// Default package name to package ID if it is not provided
	if packageName == "" {
		log.Info().Msgf("Using package ID %v as package name", packageId)
		packageName = packageId
	}
	artifactDir, err := config.GetStringWithEnvExpand(cmd, "dir-artifact")
	if err != nil {
		return fmt.Errorf("security alert for --dir-artifact: %w", err)
	}
	parametersFile := config.GetString(cmd, "file-param")
	manifestFile := config.GetString(cmd, "file-manifest")
	workDir, err := config.GetStringWithEnvExpand(cmd, "dir-work")
	if err != nil {
		return fmt.Errorf("security alert for --dir-work: %w", err)
	}
	scriptMap := str.TrimSlice(config.GetStringSlice(cmd, "script-collection-map"))

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
