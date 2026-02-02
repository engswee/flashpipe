package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/engswee/flashpipe/internal/api"
	"github.com/engswee/flashpipe/internal/config"
	"github.com/engswee/flashpipe/internal/deploy"
	"github.com/engswee/flashpipe/internal/models"
	flashpipeSync "github.com/engswee/flashpipe/internal/sync"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// OperationMode defines the orchestrator operation mode
type OperationMode string

const (
	ModeUpdateAndDeploy OperationMode = "update-and-deploy"
	ModeUpdateOnly      OperationMode = "update-only"
	ModeDeployOnly      OperationMode = "deploy-only"
)

// ProcessingStats tracks processing statistics
type ProcessingStats struct {
	PackagesUpdated           int
	PackagesDeployed          int
	PackagesFailed            int
	PackagesFiltered          int
	ArtifactsTotal            int
	ArtifactsDeployedSuccess  int
	ArtifactsDeployedFailed   int
	ArtifactsFiltered         int
	UpdateFailures            int
	DeployFailures            int
	SuccessfulPackageUpdates  map[string]bool
	SuccessfulArtifactUpdates map[string]bool
	SuccessfulArtifactDeploys map[string]bool
	FailedPackageUpdates      map[string]bool
	FailedArtifactUpdates     map[string]bool
	FailedArtifactDeploys     map[string]bool
}

// DeploymentTask represents an artifact ready for deployment
type DeploymentTask struct {
	ArtifactID   string
	ArtifactType string
	PackageID    string
	DisplayName  string
}

func NewFlashpipeOrchestratorCommand() *cobra.Command {
	var (
		packagesDir         string
		deployConfig        string
		deploymentPrefix    string
		packageFilter       string
		artifactFilter      string
		keepTemp            bool
		debugMode           bool
		configPattern       string
		mergeConfigs        bool
		updateMode          bool
		updateOnlyMode      bool
		deployOnlyMode      bool
		deployRetries       int
		deployDelaySeconds  int
		parallelDeployments int
	)

	orchestratorCmd := &cobra.Command{
		Use:          "orchestrator",
		Short:        "Orchestrate SAP CPI artifact updates and deployments",
		SilenceUsage: true, // Don't show usage on execution errors
		Long: `Orchestrate the complete deployment lifecycle for SAP CPI artifacts.

This command handles:
  - Updates artifacts in SAP CPI tenant with modified MANIFEST.MF and parameters
  - Deploys artifacts to make them active (in parallel for faster execution)
  - Supports deployment prefixes for multi-environment scenarios
  - Intelligent artifact grouping by type for efficient deployment
  - Filter by specific packages or artifacts
  - Load configs from files, folders, or remote URLs
  - Configure via YAML file for repeatable deployments

Configuration Sources:
  The --deploy-config flag accepts:
  - Single file:      ./001-deploy-config.yml
  - Folder:           ./configs (processes all matching files alphabetically)
  - Remote URL:       https://raw.githubusercontent.com/org/repo/main/config.yml

  Use --orchestrator-config to load all settings from a YAML file:
  - Sets all flags from YAML
  - CLI flags override YAML settings

Operation Modes:
  --update          Update and deploy artifacts (default)
  --update-only     Only update artifacts, don't deploy
  --deploy-only     Only deploy artifacts, don't update

Deployment Strategy:
  1. Update Phase: All packages and artifacts are updated first
  2. Deploy Phase: All artifacts are deployed in parallel
     - Deployments are triggered concurrently per package
     - Status is polled for all deployments simultaneously
     - Configurable parallelism and retry settings

Configuration:
  Settings can be loaded from the global config file (--config) under the
  'orchestrator' section. CLI flags override config file settings.`,
		Example: `  # Update and deploy with config from global flashpipe.yaml
  flashpipe orchestrator --update

  # Load specific config file
  flashpipe orchestrator --config ./my-config.yml --update

  # Override settings via CLI flags
  flashpipe orchestrator --config ./my-config.yml \
    --deployment-prefix DEV --parallel-deployments 5`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine operation mode
			mode := ModeUpdateAndDeploy
			if updateOnlyMode {
				mode = ModeUpdateOnly
			} else if deployOnlyMode {
				mode = ModeDeployOnly
			}

			// Load from viper config if available (CLI flags override config file)
			if !cmd.Flags().Changed("packages-dir") && viper.IsSet("orchestrator.packagesDir") {
				packagesDir = viper.GetString("orchestrator.packagesDir")
			}
			if !cmd.Flags().Changed("deploy-config") && viper.IsSet("orchestrator.deployConfig") {
				deployConfig = viper.GetString("orchestrator.deployConfig")
			}
			if !cmd.Flags().Changed("deployment-prefix") && viper.IsSet("orchestrator.deploymentPrefix") {
				deploymentPrefix = viper.GetString("orchestrator.deploymentPrefix")
			}
			if !cmd.Flags().Changed("package-filter") && viper.IsSet("orchestrator.packageFilter") {
				packageFilter = viper.GetString("orchestrator.packageFilter")
			}
			if !cmd.Flags().Changed("artifact-filter") && viper.IsSet("orchestrator.artifactFilter") {
				artifactFilter = viper.GetString("orchestrator.artifactFilter")
			}
			if !cmd.Flags().Changed("config-pattern") && viper.IsSet("orchestrator.configPattern") {
				configPattern = viper.GetString("orchestrator.configPattern")
			}
			if !cmd.Flags().Changed("merge-configs") && viper.IsSet("orchestrator.mergeConfigs") {
				mergeConfigs = viper.GetBool("orchestrator.mergeConfigs")
			}
			if !cmd.Flags().Changed("keep-temp") && viper.IsSet("orchestrator.keepTemp") {
				keepTemp = viper.GetBool("orchestrator.keepTemp")
			}
			if !updateMode && !updateOnlyMode && !deployOnlyMode && viper.IsSet("orchestrator.mode") {
				switch viper.GetString("orchestrator.mode") {
				case "update-and-deploy":
					mode = ModeUpdateAndDeploy
				case "update-only":
					mode = ModeUpdateOnly
				case "deploy-only":
					mode = ModeDeployOnly
				}
			}
			if !cmd.Flags().Changed("deploy-retries") && viper.IsSet("orchestrator.deployRetries") {
				deployRetries = viper.GetInt("orchestrator.deployRetries")
			}
			if !cmd.Flags().Changed("deploy-delay") && viper.IsSet("orchestrator.deployDelaySeconds") {
				deployDelaySeconds = viper.GetInt("orchestrator.deployDelaySeconds")
			}
			if !cmd.Flags().Changed("parallel-deployments") && viper.IsSet("orchestrator.parallelDeployments") {
				parallelDeployments = viper.GetInt("orchestrator.parallelDeployments")
			}

			// Validate required parameters
			if deployConfig == "" {
				return fmt.Errorf("--deploy-config is required (set via CLI flag or in config file under 'orchestrator.deployConfig')")
			}

			// Set defaults for deployment settings
			if deployRetries == 0 {
				deployRetries = 5
			}
			if deployDelaySeconds == 0 {
				deployDelaySeconds = 15
			}
			if parallelDeployments == 0 {
				parallelDeployments = 3
			}

			return runOrchestrator(cmd, mode, packagesDir, deployConfig,
				deploymentPrefix, packageFilter, artifactFilter, keepTemp, debugMode,
				configPattern, mergeConfigs, deployRetries, deployDelaySeconds, parallelDeployments)
		},
	}

	// Flags
	orchestratorCmd.Flags().StringVarP(&packagesDir, "packages-dir", "d", "", "Directory containing packages (config: orchestrator.packagesDir)")
	orchestratorCmd.Flags().StringVarP(&deployConfig, "deploy-config", "c", "", "Path to deployment config file/folder/URL (config: orchestrator.deployConfig)")
	orchestratorCmd.Flags().StringVarP(&deploymentPrefix, "deployment-prefix", "p", "", "Deployment prefix for package/artifact IDs (config: orchestrator.deploymentPrefix)")
	orchestratorCmd.Flags().StringVar(&packageFilter, "package-filter", "", "Comma-separated list of packages to include (config: orchestrator.packageFilter)")
	orchestratorCmd.Flags().StringVar(&artifactFilter, "artifact-filter", "", "Comma-separated list of artifacts to include (config: orchestrator.artifactFilter)")
	orchestratorCmd.Flags().BoolVar(&keepTemp, "keep-temp", false, "Keep temporary directory after execution (config: orchestrator.keepTemp)")
	orchestratorCmd.Flags().BoolVar(&debugMode, "debug", false, "Enable debug logging")
	orchestratorCmd.Flags().StringVar(&configPattern, "config-pattern", "*.y*ml", "File pattern for config files in folders (config: orchestrator.configPattern)")
	orchestratorCmd.Flags().BoolVar(&mergeConfigs, "merge-configs", false, "Merge multiple configs into single deployment (config: orchestrator.mergeConfigs)")
	orchestratorCmd.Flags().BoolVar(&updateMode, "update", false, "Update and deploy artifacts")
	orchestratorCmd.Flags().BoolVar(&updateOnlyMode, "update-only", false, "Only update artifacts, don't deploy")
	orchestratorCmd.Flags().BoolVar(&deployOnlyMode, "deploy-only", false, "Only deploy artifacts, don't update")
	orchestratorCmd.Flags().IntVar(&deployRetries, "deploy-retries", 0, "Number of retries for deployment status checks (config: orchestrator.deployRetries, default: 5)")
	orchestratorCmd.Flags().IntVar(&deployDelaySeconds, "deploy-delay", 0, "Delay in seconds between deployment status checks (config: orchestrator.deployDelaySeconds, default: 15)")
	orchestratorCmd.Flags().IntVar(&parallelDeployments, "parallel-deployments", 0, "Number of parallel deployments per package (config: orchestrator.parallelDeployments, default: 3)")

	return orchestratorCmd
}

// getServiceDetailsFromViperOrCmd reads service credentials from viper config or CLI flags
// This allows the orchestrator to use credentials from the global config file
func getServiceDetailsFromViperOrCmd(cmd *cobra.Command) *api.ServiceDetails {
	// Try to read from CLI flags first (via api.GetServiceDetails)
	serviceDetails := api.GetServiceDetails(cmd)

	// If host is empty, credentials weren't provided via CLI flags
	// Try to read from viper (global config file)
	if serviceDetails.Host == "" {
		tmnHost := viper.GetString("tmn-host")
		oauthHost := viper.GetString("oauth-host")

		if tmnHost == "" {
			log.Debug().Msg("No CPI credentials found in CLI flags or config file")
			return nil // No credentials found
		}

		log.Debug().Msg("Using CPI credentials from config file (viper)")
		log.Debug().Msgf("  tmn-host: %s", tmnHost)

		// Use OAuth if oauth-host is set
		if oauthHost != "" {
			log.Debug().Msgf("  oauth-host: %s", oauthHost)

			oauthPath := viper.GetString("oauth-path")
			if oauthPath == "" {
				oauthPath = "/oauth/token" // Default value
			}

			return &api.ServiceDetails{
				Host:              tmnHost,
				OauthHost:         oauthHost,
				OauthClientId:     viper.GetString("oauth-clientid"),
				OauthClientSecret: viper.GetString("oauth-clientsecret"),
				OauthPath:         oauthPath,
			}
		} else {
			log.Debug().Msg("  Using Basic Auth")
			return &api.ServiceDetails{
				Host:     tmnHost,
				Userid:   viper.GetString("tmn-userid"),
				Password: viper.GetString("tmn-password"),
			}
		}
	}

	log.Debug().Msg("Using CPI credentials from CLI flags")
	return serviceDetails
}

func runOrchestrator(cmd *cobra.Command, mode OperationMode, packagesDir, deployConfigPath,
	deploymentPrefix, packageFilterStr, artifactFilterStr string, keepTemp, debugMode bool,
	configPattern string, mergeConfigs bool, deployRetries, deployDelaySeconds, parallelDeployments int) error {

	log.Info().Msg("Starting flashpipe orchestrator")
	log.Info().Msgf("Deployment Strategy: Two-phase with parallel deployment")
	log.Info().Msgf("  Phase 1: Update all artifacts")
	log.Info().Msgf("  Phase 2: Deploy all artifacts in parallel (max %d concurrent)", parallelDeployments)

	// Validate deployment prefix
	if err := deploy.ValidateDeploymentPrefix(deploymentPrefix); err != nil {
		return err
	}

	// Parse filters
	packageFilter := parseFilter(packageFilterStr)
	artifactFilter := parseFilter(artifactFilterStr)

	// Initialize stats
	stats := ProcessingStats{
		SuccessfulArtifactUpdates: make(map[string]bool),
		SuccessfulPackageUpdates:  make(map[string]bool),
		SuccessfulArtifactDeploys: make(map[string]bool),
		FailedArtifactUpdates:     make(map[string]bool),
		FailedPackageUpdates:      make(map[string]bool),
		FailedArtifactDeploys:     make(map[string]bool),
	}

	// Setup config loader
	configLoader := deploy.NewConfigLoader()
	configLoader.Debug = debugMode
	configLoader.FilePattern = configPattern

	// Get auth settings from viper/config for remote URLs
	if viper.IsSet("host") {
		// Use CPI credentials from global config if deploying from URL
		configLoader.Username = config.GetString(cmd, "username")
		configLoader.Password = config.GetString(cmd, "password")
	}

	if err := configLoader.DetectSource(deployConfigPath); err != nil {
		return fmt.Errorf("failed to detect config source: %w", err)
	}

	log.Info().Msgf("Loading config from: %s (type: %s)", deployConfigPath, configLoader.Source)
	configFiles, err := configLoader.LoadConfigs()
	if err != nil {
		return fmt.Errorf("failed to load deployment config: %w", err)
	}

	log.Info().Msgf("Loaded %d config file(s)", len(configFiles))

	// Create temporary work directory if needed
	var workDir string
	if mode != ModeDeployOnly {
		tempDir, err := os.MkdirTemp("", "flashpipe-orchestrator-*")
		if err != nil {
			return fmt.Errorf("failed to create temp directory: %w", err)
		}
		workDir = tempDir

		if !keepTemp {
			defer os.RemoveAll(tempDir)
		} else {
			log.Info().Msgf("Temporary directory: %s", tempDir)
		}
	}

	log.Info().Msgf("Mode: %s", mode)
	log.Info().Msgf("Packages Directory: %s", packagesDir)

	if len(packageFilter) > 0 {
		log.Info().Msgf("Package filter: %s", strings.Join(packageFilter, ", "))
	}
	if len(artifactFilter) > 0 {
		log.Info().Msgf("Artifact filter: %s", strings.Join(artifactFilter, ", "))
	}

	// Get service details once (shared across all operations)
	// Read credentials from viper if not provided via CLI flags
	serviceDetails := getServiceDetailsFromViperOrCmd(cmd)
	if serviceDetails == nil {
		return fmt.Errorf("missing CPI credentials: provide via --config file or CLI flags (--tmn-host, --oauth-host, etc.)")
	}

	// Validate serviceDetails has required fields
	if serviceDetails.Host == "" {
		return fmt.Errorf("CPI host (tmn-host) is required but not provided")
	}

	log.Debug().Msg("CPI credentials successfully loaded:")
	log.Debug().Msgf("  Host: %s", serviceDetails.Host)
	if serviceDetails.OauthHost != "" {
		log.Debug().Msgf("  OAuth Host: %s", serviceDetails.OauthHost)
		log.Debug().Msg("  Auth Method: OAuth")
	} else {
		log.Debug().Msg("  Auth Method: Basic Auth")
	}

	// Collect all deployment tasks (will be executed in phase 2)
	var deploymentTasks []DeploymentTask

	// Process configs
	if mergeConfigs && len(configFiles) > 1 {
		log.Info().Msg("Merging multiple configs into single deployment")

		if deploymentPrefix != "" {
			log.Warn().Msg("Note: --deployment-prefix is ignored when merging configs with their own prefixes")
		}

		mergedConfig, err := deploy.MergeConfigs(configFiles)
		if err != nil {
			return fmt.Errorf("failed to merge configs: %w", err)
		}

		tasks, err := processPackages(mergedConfig, false, mode, packagesDir, workDir,
			packageFilter, artifactFilter, &stats, serviceDetails)
		if err != nil {
			return err
		}
		deploymentTasks = append(deploymentTasks, tasks...)
	} else {
		for _, configFile := range configFiles {
			if len(configFiles) > 1 {
				log.Info().Msgf("Processing Config: %s", configFile.FileName)
			}

			// Override deployment prefix if specified via CLI
			if deploymentPrefix != "" {
				configFile.Config.DeploymentPrefix = deploymentPrefix
			}

			log.Info().Msgf("Deployment Prefix: %s", configFile.Config.DeploymentPrefix)

			tasks, err := processPackages(configFile.Config, true, mode, packagesDir, workDir,
				packageFilter, artifactFilter, &stats, serviceDetails)
			if err != nil {
				log.Error().Msgf("Failed to process config %s: %v", configFile.FileName, err)
				continue
			}
			deploymentTasks = append(deploymentTasks, tasks...)
		}
	}

	// Phase 2: Deploy all artifacts in parallel (if not update-only mode)
	if mode != ModeUpdateOnly && len(deploymentTasks) > 0 {
		log.Info().Msg("")
		log.Info().Msg("═══════════════════════════════════════════════════════════════════════")
		log.Info().Msg("PHASE 2: DEPLOYING ALL ARTIFACTS IN PARALLEL")
		log.Info().Msg("═══════════════════════════════════════════════════════════════════════")
		log.Info().Msgf("Total artifacts to deploy: %d", len(deploymentTasks))
		log.Info().Msgf("Max concurrent deployments: %d", parallelDeployments)
		log.Info().Msg("")

		err := deployAllArtifactsParallel(deploymentTasks, parallelDeployments, deployRetries,
			deployDelaySeconds, &stats, serviceDetails)
		if err != nil {
			log.Error().Msgf("Deployment phase failed: %v", err)
		}
	}

	// Print summary
	printSummary(&stats)

	// Return error if there were failures
	if stats.PackagesFailed > 0 || stats.UpdateFailures > 0 || stats.DeployFailures > 0 {
		return fmt.Errorf("deployment completed with failures")
	}

	return nil
}

func processPackages(config *models.DeployConfig, applyPrefix bool, mode OperationMode,
	packagesDir, workDir string, packageFilter, artifactFilter []string,
	stats *ProcessingStats, serviceDetails *api.ServiceDetails) ([]DeploymentTask, error) {

	var deploymentTasks []DeploymentTask

	// Phase 1: Update all packages and artifacts
	if mode != ModeDeployOnly {
		log.Info().Msg("")
		log.Info().Msg("═══════════════════════════════════════════════════════════════════════")
		log.Info().Msg("PHASE 1: UPDATING ALL PACKAGES AND ARTIFACTS")
		log.Info().Msg("═══════════════════════════════════════════════════════════════════════")
		log.Info().Msg("")
	}

	for _, pkg := range config.Packages {
		// Apply package filter
		if !shouldInclude(pkg.ID, packageFilter) {
			log.Debug().Msgf("Skipping package %s (filtered)", pkg.ID)
			stats.PackagesFiltered++
			continue
		}

		if !pkg.Sync && !pkg.Deploy {
			log.Info().Msgf("Skipping package %s (sync=false, deploy=false)", pkg.ID)
			continue
		}

		log.Info().Msgf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		log.Info().Msgf("📦 Package: %s", pkg.ID)

		packageDir := filepath.Join(packagesDir, pkg.PackageDir)
		if !deploy.DirExists(packageDir) {
			log.Warn().Msgf("Package directory not found: %s", packageDir)
			continue
		}

		// Calculate final package ID and name
		finalPackageID := pkg.ID
		finalPackageName := pkg.DisplayName
		if finalPackageName == "" {
			finalPackageName = pkg.ID
		}

		// Apply prefix if needed
		if applyPrefix && config.DeploymentPrefix != "" {
			finalPackageID = config.DeploymentPrefix + "" + pkg.ID
			finalPackageName = config.DeploymentPrefix + " - " + finalPackageName
		}

		log.Info().Msgf("Package ID: %s", finalPackageID)
		log.Info().Msgf("Package Name: %s", finalPackageName)

		// Update package metadata
		if mode != ModeDeployOnly {
			err := updatePackage(&pkg, finalPackageID, finalPackageName, workDir, serviceDetails)
			if err != nil {
				log.Error().Msgf("Failed to update package %s: %v", pkg.ID, err)
				stats.FailedPackageUpdates[pkg.ID] = true
				stats.PackagesFailed++
				continue
			}
			stats.SuccessfulPackageUpdates[pkg.ID] = true
			stats.PackagesUpdated++
		}

		// Process artifacts for update
		if pkg.Sync && mode != ModeDeployOnly {
			if err := updateArtifacts(&pkg, packageDir, finalPackageID, finalPackageName,
				config.DeploymentPrefix, workDir, artifactFilter, stats, serviceDetails); err != nil {
				log.Error().Msgf("Failed to update artifacts for package %s: %v", pkg.ID, err)
				stats.UpdateFailures++
			}
		}

		// Collect deployment tasks (will be executed in phase 2)
		if pkg.Deploy && mode != ModeUpdateOnly {
			tasks := collectDeploymentTasks(&pkg, finalPackageID, config.DeploymentPrefix,
				artifactFilter, stats)
			deploymentTasks = append(deploymentTasks, tasks...)
		}
	}

	return deploymentTasks, nil
}

func updatePackage(pkg *models.Package, finalPackageID, finalPackageName, workDir string,
	serviceDetails *api.ServiceDetails) error {

	if serviceDetails == nil {
		return fmt.Errorf("serviceDetails is nil - cannot update package")
	}

	log.Info().Msg("Updating package in tenant...")

	description := pkg.Description
	if description == "" {
		description = finalPackageName
	}

	shortText := pkg.ShortText
	if shortText == "" {
		shortText = finalPackageName
	}

	// Create package JSON
	packageJSON := map[string]interface{}{
		"d": map[string]interface{}{
			"Id":          finalPackageID,
			"Name":        finalPackageName,
			"Description": description,
			"ShortText":   shortText,
		},
	}

	jsonData, err := json.MarshalIndent(packageJSON, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal package JSON: %w", err)
	}

	// Write to temporary file
	packageJSONPath := filepath.Join(workDir, "modified", fmt.Sprintf("package_%s.json", pkg.ID))
	if err := os.MkdirAll(filepath.Dir(packageJSONPath), 0755); err != nil {
		return fmt.Errorf("failed to create package JSON directory: %w", err)
	}

	if err := os.WriteFile(packageJSONPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write package JSON: %w", err)
	}

	// Use internal sync package update function
	exe := api.InitHTTPExecuter(serviceDetails)
	packageSynchroniser := flashpipeSync.NewSyncer("tenant", "CPIPackage", exe)

	err = packageSynchroniser.Exec(flashpipeSync.Request{PackageFile: packageJSONPath})
	if err != nil {
		log.Warn().Msgf("Package update warning (may not exist yet): %v", err)
		// Don't return error - package might not exist yet
		return nil
	}

	log.Info().Msg("  ✓ Package metadata updated")
	return nil
}

func updateArtifacts(pkg *models.Package, packageDir, finalPackageID, finalPackageName, prefix, workDir string,
	artifactFilter []string, stats *ProcessingStats, serviceDetails *api.ServiceDetails) error {

	updatedCount := 0
	log.Info().Msg("Updating artifacts...")

	if serviceDetails == nil {
		return fmt.Errorf("serviceDetails is nil - cannot initialize HTTP executer")
	}
	if serviceDetails.Host == "" {
		return fmt.Errorf("serviceDetails.Host is empty - check CPI credentials in config file")
	}

	log.Info().Msgf("DEBUG: ServiceDetails before InitHTTPExecuter:")
	log.Info().Msgf("  Host: %s", serviceDetails.Host)
	log.Info().Msgf("  OauthHost: %s", serviceDetails.OauthHost)
	log.Info().Msgf("  OauthClientId: %s", serviceDetails.OauthClientId)
	log.Info().Msgf("  OauthPath: %s", serviceDetails.OauthPath)
	log.Info().Msgf("  Userid: %s", serviceDetails.Userid)

	log.Debug().Msgf("Initializing HTTP executer with host: %s", serviceDetails.Host)
	exe := api.InitHTTPExecuter(serviceDetails)
	if exe == nil {
		return fmt.Errorf("failed to initialize HTTP executer")
	}

	log.Info().Msgf("DEBUG: exe after InitHTTPExecuter is NOT nil")

	synchroniser := flashpipeSync.New(exe)
	if synchroniser == nil {
		return fmt.Errorf("failed to initialize synchroniser")
	}

	log.Info().Msgf("DEBUG: synchroniser created successfully")

	for _, artifact := range pkg.Artifacts {
		// Apply artifact filter
		if !shouldInclude(artifact.Id, artifactFilter) {
			log.Debug().Msgf("Skipping artifact %s (filtered)", artifact.Id)
			stats.ArtifactsFiltered++
			continue
		}

		if !artifact.Sync {
			log.Debug().Msgf("Skipping artifact %s (sync=false)", artifact.DisplayName)
			continue
		}

		stats.ArtifactsTotal++

		artifactDir := filepath.Join(packageDir, artifact.ArtifactDir)
		if !deploy.DirExists(artifactDir) {
			log.Warn().Msgf("Artifact directory not found: %s", artifactDir)
			continue
		}

		// Calculate final artifact ID and name
		finalArtifactID := artifact.Id
		finalArtifactName := artifact.DisplayName
		if finalArtifactName == "" {
			finalArtifactName = artifact.Id
		}

		if prefix != "" {
			finalArtifactID = prefix + "_" + artifact.Id
		}

		log.Info().Msgf("  Updating: %s", finalArtifactID)

		// Map artifact type for synchroniser (uses simple type names)
		artifactType := mapArtifactTypeForSync(artifact.Type)

		// Create temp directory for this artifact
		tempArtifactDir := filepath.Join(workDir, artifact.Id)
		if err := deploy.CopyDir(artifactDir, tempArtifactDir); err != nil {
			log.Error().Msgf("Failed to copy artifact to temp: %v", err)
			stats.FailedArtifactUpdates[artifact.Id] = true
			continue
		}

		// Update MANIFEST.MF
		manifestPath := filepath.Join(tempArtifactDir, "META-INF", "MANIFEST.MF")
		modifiedManifestPath := filepath.Join(workDir, "modified", artifact.Id, "META-INF", "MANIFEST.MF")

		if deploy.FileExists(manifestPath) {
			if err := deploy.UpdateManifestBundleName(manifestPath, finalArtifactID, finalArtifactName, modifiedManifestPath); err != nil {
				log.Warn().Msgf("Failed to update MANIFEST.MF: %v", err)
			}
		}

		// Handle parameters.prop
		var modifiedParamsPath string
		paramsPath := deploy.FindParametersFile(tempArtifactDir)

		if paramsPath != "" && deploy.FileExists(paramsPath) {
			modifiedParamsPath = filepath.Join(workDir, "modified", artifact.Id, "parameters.prop")

			if len(artifact.ConfigOverrides) > 0 {
				if err := deploy.MergeParametersFile(paramsPath, artifact.ConfigOverrides, modifiedParamsPath); err != nil {
					log.Warn().Msgf("Failed to merge parameters: %v", err)
				} else {
					log.Debug().Msgf("Applied %d config overrides", len(artifact.ConfigOverrides))
				}
			} else {
				// No overrides, copy to modified location
				data, err := os.ReadFile(paramsPath)
				if err == nil {
					os.MkdirAll(filepath.Dir(modifiedParamsPath), 0755)
					os.WriteFile(modifiedParamsPath, data, 0644)
				}
			}
		}

		// Copy modified manifest to temp artifact dir for sync
		if deploy.FileExists(modifiedManifestPath) {
			targetManifestPath := filepath.Join(tempArtifactDir, "META-INF", "MANIFEST.MF")
			data, err := os.ReadFile(modifiedManifestPath)
			if err == nil {
				os.WriteFile(targetManifestPath, data, 0644)
			}
		}

		// Copy modified parameters if exists
		if modifiedParamsPath != "" && deploy.FileExists(modifiedParamsPath) {
			// Find the actual parameters location in the artifact
			actualParamsPath := deploy.FindParametersFile(tempArtifactDir)
			data, err := os.ReadFile(modifiedParamsPath)
			if err == nil {
				os.WriteFile(actualParamsPath, data, 0644)
			}
		}

		// Call internal sync function
		log.Debug().Msgf("DEBUG: About to call SingleArtifactToTenant for %s", finalArtifactID)
		log.Debug().Msgf("  synchroniser: %v", synchroniser)
		log.Debug().Msgf("  finalPackageID: %s", finalPackageID)
		log.Debug().Msgf("  artifactType: %s", artifactType)

		err := synchroniser.SingleArtifactToTenant(finalArtifactID, finalArtifactName, artifactType,
			finalPackageID, tempArtifactDir, workDir, "", nil)

		if err != nil {
			log.Error().Msgf("Update failed for %s: %v", finalArtifactName, err)
			stats.UpdateFailures++
			stats.FailedArtifactUpdates[artifact.Id] = true
			continue
		}

		log.Info().Msg("    ✓ Updated successfully")
		updatedCount++
		stats.SuccessfulArtifactUpdates[finalArtifactID] = true
	}

	if updatedCount > 0 {
		log.Info().Msgf("✓ Updated %d artifact(s) in package", updatedCount)
	}

	return nil
}

func collectDeploymentTasks(pkg *models.Package, finalPackageID, prefix string,
	artifactFilter []string, stats *ProcessingStats) []DeploymentTask {

	var tasks []DeploymentTask

	for _, artifact := range pkg.Artifacts {
		// Skip if update failed
		if stats.FailedArtifactUpdates[artifact.Id] {
			log.Debug().Msgf("Skipping artifact %s (due to failed update)", artifact.Id)
			continue
		}

		// Apply artifact filter
		if !shouldInclude(artifact.Id, artifactFilter) {
			log.Debug().Msgf("Skipping artifact %s (filtered)", artifact.Id)
			continue
		}

		if !artifact.Deploy {
			log.Debug().Msgf("Skipping artifact %s (deploy=false)", artifact.DisplayName)
			continue
		}

		finalArtifactID := artifact.Id
		if prefix != "" {
			finalArtifactID = prefix + "_" + artifact.Id
		}

		artifactType := artifact.Type
		if artifactType == "" {
			artifactType = "IntegrationFlow"
		}

		tasks = append(tasks, DeploymentTask{
			ArtifactID:   finalArtifactID,
			ArtifactType: artifactType,
			PackageID:    finalPackageID,
			DisplayName:  artifact.DisplayName,
		})
	}

	return tasks
}

func deployAllArtifactsParallel(tasks []DeploymentTask, maxConcurrent int,
	retries int, delaySeconds int, stats *ProcessingStats, serviceDetails *api.ServiceDetails) error {

	// Group tasks by package for better control
	tasksByPackage := make(map[string][]DeploymentTask)
	for _, task := range tasks {
		tasksByPackage[task.PackageID] = append(tasksByPackage[task.PackageID], task)
	}

	// Process each package's deployments
	for packageID, packageTasks := range tasksByPackage {
		log.Info().Msgf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		log.Info().Msgf("📦 Deploying %d artifacts for package: %s", len(packageTasks), packageID)

		// Deploy artifacts in parallel with semaphore
		var wg sync.WaitGroup
		semaphore := make(chan struct{}, maxConcurrent)
		resultChan := make(chan deployResult, len(packageTasks))

		for _, task := range packageTasks {
			wg.Add(1)
			go func(t DeploymentTask) {
				defer wg.Done()

				// Acquire semaphore
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				// Deploy artifact
				// Use mapArtifactTypeForSync because deployArtifacts calls api.NewDesigntimeArtifact
				flashpipeType := mapArtifactTypeForSync(t.ArtifactType)
				log.Info().Msgf("  → Deploying: %s (type: %s)", t.ArtifactID, t.ArtifactType)

				err := deployArtifacts([]string{t.ArtifactID}, flashpipeType, retries, delaySeconds, true, serviceDetails)

				resultChan <- deployResult{
					Task:  t,
					Error: err,
				}
			}(task)
		}

		// Wait for all deployments to complete
		wg.Wait()
		close(resultChan)

		// Process results
		successCount := 0
		failureCount := 0

		for result := range resultChan {
			if result.Error != nil {
				log.Error().Msgf("  ✗ Deploy failed: %s - %v", result.Task.ArtifactID, result.Error)
				stats.ArtifactsDeployedFailed++
				stats.DeployFailures++
				stats.FailedArtifactDeploys[result.Task.ArtifactID] = true
				failureCount++
			} else {
				log.Info().Msgf("  ✓ Deployed: %s", result.Task.ArtifactID)
				stats.ArtifactsDeployedSuccess++
				stats.SuccessfulArtifactDeploys[result.Task.ArtifactID] = true
				successCount++
			}
		}

		if failureCount == 0 {
			log.Info().Msgf("✓ All %d artifacts deployed successfully for package %s", successCount, packageID)
			stats.PackagesDeployed++
		} else {
			log.Warn().Msgf("⚠ Package %s: %d succeeded, %d failed", packageID, successCount, failureCount)
			stats.PackagesFailed++
		}
	}

	return nil
}

type deployResult struct {
	Task  DeploymentTask
	Error error
}

// mapArtifactType maps artifact types for deployment API calls
func mapArtifactType(artifactType string) string {
	switch strings.ToLower(artifactType) {
	case "integrationflow", "integration flow", "iflow":
		return "IntegrationDesigntimeArtifact"
	case "valuemapping", "value mapping":
		return "ValueMappingDesigntimeArtifact"
	case "messageMapping", "message mapping":
		return "MessageMappingDesigntimeArtifact"
	case "scriptcollection", "script collection":
		return "ScriptCollection"
	default:
		// Default to integration flow
		return "IntegrationDesigntimeArtifact"
	}
}

// mapArtifactTypeForSync maps artifact types for synchroniser (NewDesigntimeArtifact)
func mapArtifactTypeForSync(artifactType string) string {
	switch strings.ToLower(artifactType) {
	case "integrationflow", "integration flow", "iflow":
		return "Integration"
	case "valuemapping", "value mapping":
		return "ValueMapping"
	case "messagemapping", "message mapping":
		return "MessageMapping"
	case "scriptcollection", "script collection":
		return "ScriptCollection"
	default:
		// Default to integration flow
		return "Integration"
	}
}

func parseFilter(filterStr string) []string {
	if filterStr == "" {
		return nil
	}
	parts := strings.Split(filterStr, ",")
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func shouldInclude(id string, filter []string) bool {
	if len(filter) == 0 {
		return true
	}
	for _, f := range filter {
		if f == id {
			return true
		}
	}
	return false
}

func printSummary(stats *ProcessingStats) {
	log.Info().Msg("")
	log.Info().Msg("═══════════════════════════════════════════════════════════════════════")
	log.Info().Msg("📊 DEPLOYMENT SUMMARY")
	log.Info().Msg("═══════════════════════════════════════════════════════════════════════")
	log.Info().Msgf("Packages Updated:   %d", stats.PackagesUpdated)
	log.Info().Msgf("Packages Deployed:  %d", stats.PackagesDeployed)
	log.Info().Msgf("Packages Failed:    %d", stats.PackagesFailed)
	log.Info().Msgf("Packages Filtered:  %d", stats.PackagesFiltered)
	log.Info().Msg("───────────────────────────────────────────────────────────────────────")
	log.Info().Msgf("Artifacts Total:         %d", stats.ArtifactsTotal)
	log.Info().Msgf("Artifacts Updated:       %d", len(stats.SuccessfulArtifactUpdates))
	log.Info().Msgf("Artifacts Deployed OK:   %d", stats.ArtifactsDeployedSuccess)
	log.Info().Msgf("Artifacts Deployed Fail: %d", stats.ArtifactsDeployedFailed)
	log.Info().Msgf("Artifacts Filtered:      %d", stats.ArtifactsFiltered)
	log.Info().Msg("───────────────────────────────────────────────────────────────────────")

	if stats.UpdateFailures > 0 {
		log.Warn().Msgf("⚠ Update Failures: %d", stats.UpdateFailures)
		log.Info().Msg("Failed Artifact Updates:")
		for artifactID := range stats.FailedArtifactUpdates {
			log.Info().Msgf("  - %s", artifactID)
		}
	}

	if stats.DeployFailures > 0 {
		log.Warn().Msgf("⚠ Deploy Failures: %d", stats.DeployFailures)
		log.Info().Msg("Failed Artifact Deployments:")
		for artifactID := range stats.FailedArtifactDeploys {
			log.Info().Msgf("  - %s", artifactID)
		}
	}

	if stats.UpdateFailures == 0 && stats.DeployFailures == 0 {
		log.Info().Msg("✓ All operations completed successfully!")
	}

	log.Info().Msg("═══════════════════════════════════════════════════════════════════════")
}
