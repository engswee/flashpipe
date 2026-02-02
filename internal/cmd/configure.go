package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/engswee/flashpipe/internal/api"
	"github.com/engswee/flashpipe/internal/deploy"
	"github.com/engswee/flashpipe/internal/httpclnt"
	"github.com/engswee/flashpipe/internal/models"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// ConfigureStats tracks configuration processing statistics
type ConfigureStats struct {
	PackagesProcessed         int
	PackagesWithErrors        int
	ArtifactsProcessed        int
	ArtifactsConfigured       int
	ArtifactsDeployed         int
	ArtifactsFailed           int
	ParametersUpdated         int
	ParametersFailed          int
	BatchRequestsExecuted     int
	IndividualRequestsUsed    int
	DeploymentTasksQueued     int
	DeploymentTasksSuccessful int
	DeploymentTasksFailed     int
}

// ConfigurationTask represents a configuration update task
type ConfigurationTask struct {
	PackageID   string
	ArtifactID  string
	Version     string
	Parameters  []models.ConfigurationParameter
	UseBatch    bool
	BatchSize   int
	DisplayName string
}

func NewConfigureCommand() *cobra.Command {
	var (
		configPath          string
		deploymentPrefix    string
		packageFilter       string
		artifactFilter      string
		dryRun              bool
		deployRetries       int
		deployDelaySeconds  int
		parallelDeployments int
		batchSize           int
		disableBatch        bool
	)

	configureCmd := &cobra.Command{
		Use:   "configure",
		Short: "Configure SAP CPI artifact parameters",
		Long: `Configure parameters for SAP CPI artifacts using YAML configuration files.

This command:
  - Updates configuration parameters for Integration artifacts
  - Supports batch operations for efficient parameter updates
  - Optionally deploys artifacts after configuration
  - Two-phase operation: Configure all artifacts, then deploy if requested
  - Supports deployment prefixes for multi-environment scenarios

Configuration File Structure:
  The YAML file should define packages and artifacts with their parameters:

  deploymentPrefix: "DEV_"  # Optional
  packages:
    - integrationSuiteId: "MyPackage"
      displayName: "My Integration Package"
      deploy: false  # Deploy all artifacts in this package after configuration
      artifacts:
        - artifactId: "MyFlow"
          displayName: "My Integration Flow"
          type: "Integration"
          version: "active"  # Optional, defaults to "active"
          deploy: true       # Deploy this specific artifact after configuration
          parameters:
            - key: "DatabaseURL"
              value: "jdbc:mysql://localhost:3306/mydb"
            - key: "MaxRetries"
              value: "5"
          batch:
            enabled: true    # Use batch operations (default: true)
            batchSize: 90    # Parameters per batch (default: 90)

Operation Modes:
  1. Configure Only: Updates parameters without deployment (default)
  2. Configure + Deploy: Updates parameters then deploys artifacts (when deploy: true)

Batch Processing:
  - By default, uses OData $batch for efficient parameter updates
  - Configurable batch size (default: 90 parameters per request)
  - Falls back to individual requests if batch fails
  - Can be disabled globally with --disable-batch flag

Configuration:
  Settings can be loaded from the global config file (--config) under the
  'configure' section. CLI flags override config file settings.`,
		Example: `  # Configure artifacts from a config file
  flashpipe configure --config-path ./config/dev-config.yml

  # Configure and deploy
  flashpipe configure --config-path ./config/prod-config.yml

  # Dry run to see what would be changed
  flashpipe configure --config-path ./config.yml --dry-run

  # Apply deployment prefix
  flashpipe configure --config-path ./config.yml --deployment-prefix DEV_

  # Disable batch processing
  flashpipe configure --config-path ./config.yml --disable-batch`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load from viper config if available (CLI flags override config file)
			if !cmd.Flags().Changed("config-path") && viper.IsSet("configure.configPath") {
				configPath = viper.GetString("configure.configPath")
			}
			if !cmd.Flags().Changed("deployment-prefix") && viper.IsSet("configure.deploymentPrefix") {
				deploymentPrefix = viper.GetString("configure.deploymentPrefix")
			}
			if !cmd.Flags().Changed("package-filter") && viper.IsSet("configure.packageFilter") {
				packageFilter = viper.GetString("configure.packageFilter")
			}
			if !cmd.Flags().Changed("artifact-filter") && viper.IsSet("configure.artifactFilter") {
				artifactFilter = viper.GetString("configure.artifactFilter")
			}
			if !cmd.Flags().Changed("dry-run") && viper.IsSet("configure.dryRun") {
				dryRun = viper.GetBool("configure.dryRun")
			}
			if !cmd.Flags().Changed("deploy-retries") && viper.IsSet("configure.deployRetries") {
				deployRetries = viper.GetInt("configure.deployRetries")
			}
			if !cmd.Flags().Changed("deploy-delay") && viper.IsSet("configure.deployDelaySeconds") {
				deployDelaySeconds = viper.GetInt("configure.deployDelaySeconds")
			}
			if !cmd.Flags().Changed("parallel-deployments") && viper.IsSet("configure.parallelDeployments") {
				parallelDeployments = viper.GetInt("configure.parallelDeployments")
			}
			if !cmd.Flags().Changed("batch-size") && viper.IsSet("configure.batchSize") {
				batchSize = viper.GetInt("configure.batchSize")
			}
			if !cmd.Flags().Changed("disable-batch") && viper.IsSet("configure.disableBatch") {
				disableBatch = viper.GetBool("configure.disableBatch")
			}

			// Validate required parameters
			if configPath == "" {
				return fmt.Errorf("--config-path is required (set via CLI flag or in config file under 'configure.configPath')")
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
			if batchSize == 0 {
				batchSize = httpclnt.DefaultBatchSize
			}

			return runConfigure(cmd, configPath, deploymentPrefix, packageFilter, artifactFilter,
				dryRun, deployRetries, deployDelaySeconds, parallelDeployments, batchSize, disableBatch)
		},
	}

	// Flags
	configureCmd.Flags().StringVarP(&configPath, "config-path", "c", "", "Path to configuration YAML file (config: configure.configPath)")
	configureCmd.Flags().StringVarP(&deploymentPrefix, "deployment-prefix", "p", "", "Deployment prefix for artifact IDs (config: configure.deploymentPrefix)")
	configureCmd.Flags().StringVar(&packageFilter, "package-filter", "", "Comma-separated list of packages to include (config: configure.packageFilter)")
	configureCmd.Flags().StringVar(&artifactFilter, "artifact-filter", "", "Comma-separated list of artifacts to include (config: configure.artifactFilter)")
	configureCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without making changes (config: configure.dryRun)")
	configureCmd.Flags().IntVar(&deployRetries, "deploy-retries", 0, "Number of retries for deployment status checks (config: configure.deployRetries, default: 5)")
	configureCmd.Flags().IntVar(&deployDelaySeconds, "deploy-delay", 0, "Delay in seconds between deployment status checks (config: configure.deployDelaySeconds, default: 15)")
	configureCmd.Flags().IntVar(&parallelDeployments, "parallel-deployments", 0, "Number of parallel deployments (config: configure.parallelDeployments, default: 3)")
	configureCmd.Flags().IntVar(&batchSize, "batch-size", 0, "Number of parameters per batch request (config: configure.batchSize, default: 90)")
	configureCmd.Flags().BoolVar(&disableBatch, "disable-batch", false, "Disable batch processing, use individual requests (config: configure.disableBatch)")

	return configureCmd
}

func runConfigure(cmd *cobra.Command, configPath, deploymentPrefix, packageFilterStr, artifactFilterStr string,
	dryRun bool, deployRetries, deployDelaySeconds, parallelDeployments, batchSize int, disableBatch bool) error {

	log.Info().Msg("Starting artifact configuration")

	// Validate deployment prefix
	if deploymentPrefix != "" {
		if err := deploy.ValidateDeploymentPrefix(deploymentPrefix); err != nil {
			return err
		}
	}

	// Parse filters
	packageFilter := parseFilter(packageFilterStr)
	artifactFilter := parseFilter(artifactFilterStr)

	// Load configuration from file or folder
	log.Info().Msgf("Loading configuration from: %s", configPath)
	configFiles, err := loadConfigureConfigs(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	log.Info().Msgf("Loaded %d configuration file(s)", len(configFiles))
	log.Info().Msgf("Deployment prefix: %s", deploymentPrefix)
	log.Info().Msgf("Dry run: %v", dryRun)
	log.Info().Msgf("Batch processing: %v (size: %d)", !disableBatch, batchSize)

	// Merge all configurations
	configData := mergeConfigureConfigs(configFiles, deploymentPrefix)

	// Apply deployment prefix if specified
	if deploymentPrefix != "" {
		configData.DeploymentPrefix = deploymentPrefix
	}

	// Initialize stats
	stats := &ConfigureStats{}

	// Get service details
	serviceDetails := getServiceDetailsFromViperOrCmd(cmd)
	exe := api.InitHTTPExecuter(serviceDetails)

	// Phase 1: Configure all artifacts
	log.Info().Msg("")
	log.Info().Msg("═══════════════════════════════════════════════════════════════════════")
	log.Info().Msg("PHASE 1: CONFIGURING ARTIFACTS")
	log.Info().Msg("═══════════════════════════════════════════════════════════════════════")

	deploymentTasks, err := configureAllArtifacts(exe, configData, packageFilter, artifactFilter,
		stats, dryRun, batchSize, disableBatch)
	if err != nil {
		return err
	}

	// Phase 2: Deploy artifacts if requested
	if len(deploymentTasks) > 0 && !dryRun {
		log.Info().Msg("")
		log.Info().Msg("═══════════════════════════════════════════════════════════════════════")
		log.Info().Msg("PHASE 2: DEPLOYING CONFIGURED ARTIFACTS")
		log.Info().Msg("═══════════════════════════════════════════════════════════════════════")
		log.Info().Msgf("Deploying %d artifacts with max %d parallel deployments per package",
			len(deploymentTasks), parallelDeployments)

		err := deployConfiguredArtifacts(exe, deploymentTasks, deployRetries, deployDelaySeconds,
			parallelDeployments, stats)
		if err != nil {
			log.Error().Msgf("Deployment phase failed: %v", err)
		}
	}

	// Print summary
	printConfigureSummary(stats, dryRun)

	// Return error if there were failures
	if stats.ArtifactsFailed > 0 || stats.DeploymentTasksFailed > 0 {
		return fmt.Errorf("configuration/deployment completed with errors")
	}

	return nil
}

// ConfigureConfigFile represents a loaded config file with metadata
type ConfigureConfigFile struct {
	Config   *models.ConfigureConfig
	Source   string
	FileName string
}

func loadConfigureConfigs(path string) ([]*ConfigureConfigFile, error) {
	// Check if path is a file or directory
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to access path: %w", err)
	}

	if info.IsDir() {
		return loadConfigureConfigsFromFolder(path)
	}
	return loadConfigureConfigFromFile(path)
}

func loadConfigureConfigFromFile(path string) ([]*ConfigureConfigFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var cfg models.ConfigureConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return []*ConfigureConfigFile{
		{
			Config:   &cfg,
			Source:   path,
			FileName: filepath.Base(path),
		},
	}, nil
}

func loadConfigureConfigsFromFolder(folderPath string) ([]*ConfigureConfigFile, error) {
	var configFiles []*ConfigureConfigFile

	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Match YAML files (*.yml, *.yaml)
		name := entry.Name()
		if !strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yaml") {
			continue
		}

		filePath := filepath.Join(folderPath, name)
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Warn().Msgf("Failed to read config file %s: %v", name, err)
			continue
		}

		var cfg models.ConfigureConfig
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			log.Warn().Msgf("Failed to parse config file %s: %v", name, err)
			continue
		}

		configFiles = append(configFiles, &ConfigureConfigFile{
			Config:   &cfg,
			Source:   filePath,
			FileName: name,
		})
	}

	if len(configFiles) == 0 {
		return nil, fmt.Errorf("no valid configuration files found in folder: %s", folderPath)
	}

	log.Info().Msgf("Loaded %d configuration file(s) from folder", len(configFiles))
	return configFiles, nil
}

func mergeConfigureConfigs(configFiles []*ConfigureConfigFile, overridePrefix string) *models.ConfigureConfig {
	merged := &models.ConfigureConfig{
		Packages: []models.ConfigurePackage{},
	}

	// Use override prefix if provided, otherwise use first config's prefix
	if overridePrefix != "" {
		merged.DeploymentPrefix = overridePrefix
	} else if len(configFiles) > 0 && configFiles[0].Config.DeploymentPrefix != "" {
		merged.DeploymentPrefix = configFiles[0].Config.DeploymentPrefix
	}

	// Merge all packages from all config files
	for _, configFile := range configFiles {
		log.Info().Msgf("  Merging packages from: %s", configFile.FileName)
		merged.Packages = append(merged.Packages, configFile.Config.Packages...)
	}

	return merged
}

func configureAllArtifacts(exe *httpclnt.HTTPExecuter, cfg *models.ConfigureConfig,
	packageFilter, artifactFilter []string, stats *ConfigureStats, dryRun bool,
	batchSize int, disableBatch bool) ([]DeploymentTask, error) {

	var deploymentTasks []DeploymentTask
	configuration := api.NewConfiguration(exe)

	for _, pkg := range cfg.Packages {
		stats.PackagesProcessed++

		// Apply deployment prefix to package ID
		packageID := pkg.ID
		if cfg.DeploymentPrefix != "" {
			packageID = cfg.DeploymentPrefix + packageID
		}

		// Apply package filter
		if len(packageFilter) > 0 && !shouldInclude(pkg.ID, packageFilter) {
			log.Info().Msgf("Skipping package %s (filtered out)", packageID)
			continue
		}

		log.Info().Msg("")
		log.Info().Msgf("📦 Processing package: %s", packageID)
		if pkg.DisplayName != "" {
			log.Info().Msgf("   Display Name: %s", pkg.DisplayName)
		}

		packageHasError := false

		for _, artifact := range pkg.Artifacts {
			stats.ArtifactsProcessed++

			// Apply deployment prefix to artifact ID
			artifactID := artifact.ID
			if cfg.DeploymentPrefix != "" {
				artifactID = cfg.DeploymentPrefix + artifactID
			}

			// Apply artifact filter
			if len(artifactFilter) > 0 && !shouldInclude(artifact.ID, artifactFilter) {
				log.Info().Msgf("   Skipping artifact %s (filtered out)", artifactID)
				continue
			}

			log.Info().Msg("")
			log.Info().Msgf("   🔧 Configuring artifact: %s", artifactID)
			if artifact.DisplayName != "" {
				log.Info().Msgf("      Display Name: %s", artifact.DisplayName)
			}
			log.Info().Msgf("      Type: %s", artifact.Type)
			log.Info().Msgf("      Version: %s", artifact.Version)
			log.Info().Msgf("      Parameters: %d", len(artifact.Parameters))

			if dryRun {
				log.Info().Msg("      [DRY RUN] Would update the following parameters:")
				for _, param := range artifact.Parameters {
					log.Info().Msgf("        - %s = %s", param.Key, param.Value)
				}
				stats.ArtifactsConfigured++
				stats.ParametersUpdated += len(artifact.Parameters)

				// Queue for deployment if requested
				if artifact.Deploy || pkg.Deploy {
					stats.DeploymentTasksQueued++
					log.Info().Msgf("      [DRY RUN] Would deploy after configuration")
				}
				continue
			}

			// Determine batch settings
			useBatch := !disableBatch
			effectiveBatchSize := batchSize

			if artifact.Batch != nil {
				useBatch = artifact.Batch.Enabled && !disableBatch
				if artifact.Batch.BatchSize > 0 {
					effectiveBatchSize = artifact.Batch.BatchSize
				}
			}

			// Update configuration parameters
			var configErr error
			if useBatch && len(artifact.Parameters) > 0 {
				configErr = updateParametersBatch(exe, configuration, artifactID, artifact.Version,
					artifact.Parameters, effectiveBatchSize, stats)
			} else {
				configErr = updateParametersIndividual(configuration, artifactID, artifact.Version,
					artifact.Parameters, stats)
			}

			if configErr != nil {
				log.Error().Msgf("      ❌ Failed to configure artifact: %v", configErr)
				stats.ArtifactsFailed++
				packageHasError = true
				continue
			}

			stats.ArtifactsConfigured++
			log.Info().Msgf("      ✅ Successfully configured %d parameters", len(artifact.Parameters))

			// Queue for deployment if requested
			if artifact.Deploy || pkg.Deploy {
				deploymentTasks = append(deploymentTasks, DeploymentTask{
					ArtifactID:   artifactID,
					ArtifactType: artifact.Type,
					PackageID:    packageID,
					DisplayName:  artifact.DisplayName,
				})
				stats.DeploymentTasksQueued++
				log.Info().Msgf("      📋 Queued for deployment")
			}
		}

		if packageHasError {
			stats.PackagesWithErrors++
		}
	}

	return deploymentTasks, nil
}

func updateParametersBatch(exe *httpclnt.HTTPExecuter, configuration *api.Configuration,
	artifactID, version string, parameters []models.ConfigurationParameter,
	batchSize int, stats *ConfigureStats) error {

	log.Info().Msgf("      Using batch operations (batch size: %d)", batchSize)

	// Get current configuration to verify parameters exist
	currentConfig, err := configuration.Get(artifactID, version)
	if err != nil {
		return fmt.Errorf("failed to get current configuration: %w", err)
	}

	// Build batch request
	batch := exe.NewBatchRequest()
	validParams := 0

	for _, param := range parameters {
		// Verify parameter exists
		existingParam := api.FindParameterByKey(param.Key, currentConfig.Root.Results)
		if existingParam == nil {
			log.Warn().Msgf("      ⚠️  Parameter %s not found in artifact, skipping", param.Key)
			stats.ParametersFailed++
			continue
		}

		// Add to batch
		requestBody := fmt.Sprintf(`{"ParameterValue":"%s"}`, escapeJSON(param.Value))
		urlPath := fmt.Sprintf("/api/v1/IntegrationDesigntimeArtifacts(Id='%s',Version='%s')/$links/Configurations('%s')",
			artifactID, version, param.Key)

		batch.AddOperation(httpclnt.BatchOperation{
			Method:    "PUT",
			Path:      urlPath,
			Body:      []byte(requestBody),
			ContentID: fmt.Sprintf("param_%d", validParams),
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		})
		validParams++
	}

	if validParams == 0 {
		return fmt.Errorf("no valid parameters to update")
	}

	// Execute batch in chunks
	resp, err := batch.ExecuteInBatches(batchSize)
	if err != nil {
		log.Warn().Msgf("      ⚠️  Batch operation failed: %v, falling back to individual requests", err)
		return updateParametersIndividual(configuration, artifactID, version, parameters, stats)
	}

	stats.BatchRequestsExecuted++

	// Process batch results
	successCount := 0
	failCount := 0

	for _, opResp := range resp.Operations {
		if opResp.Error != nil {
			failCount++
			stats.ParametersFailed++
		} else if opResp.StatusCode >= 200 && opResp.StatusCode < 300 {
			successCount++
			stats.ParametersUpdated++
		} else {
			failCount++
			stats.ParametersFailed++
		}
	}

	if failCount > 0 {
		return fmt.Errorf("%d parameters failed to update in batch", failCount)
	}

	return nil
}

func updateParametersIndividual(configuration *api.Configuration, artifactID, version string,
	parameters []models.ConfigurationParameter, stats *ConfigureStats) error {

	log.Info().Msgf("      Using individual requests")

	failCount := 0
	successCount := 0

	for _, param := range parameters {
		err := configuration.Update(artifactID, version, param.Key, param.Value)
		if err != nil {
			log.Error().Msgf("      ❌ Failed to update parameter %s: %v", param.Key, err)
			stats.ParametersFailed++
			failCount++
		} else {
			stats.ParametersUpdated++
			stats.IndividualRequestsUsed++
			successCount++
		}
	}

	if failCount > 0 {
		return fmt.Errorf("%d parameters failed to update", failCount)
	}

	return nil
}

func deployConfiguredArtifacts(exe *httpclnt.HTTPExecuter, tasks []DeploymentTask,
	deployRetries, deployDelaySeconds, parallelDeployments int, stats *ConfigureStats) error {

	// Group tasks by package
	packageTasks := make(map[string][]DeploymentTask)
	for _, task := range tasks {
		packageTasks[task.PackageID] = append(packageTasks[task.PackageID], task)
	}

	log.Info().Msgf("Deploying artifacts across %d packages", len(packageTasks))

	var wg sync.WaitGroup
	resultsChan := make(chan deployResult, len(tasks))

	// Deploy all artifacts in parallel
	for packageID, pkgTasks := range packageTasks {
		log.Info().Msgf("Package %s: deploying %d artifacts", packageID, len(pkgTasks))

		// Process artifacts in this package with controlled parallelism
		semaphore := make(chan struct{}, parallelDeployments)

		for _, task := range pkgTasks {
			wg.Add(1)
			go func(t DeploymentTask) {
				defer wg.Done()
				semaphore <- struct{}{}        // Acquire
				defer func() { <-semaphore }() // Release

				log.Info().Msgf("  Deploying %s (type: %s)", t.ArtifactID, t.ArtifactType)

				deployErr := deployArtifact(exe, t, deployRetries, deployDelaySeconds)
				resultsChan <- deployResult{Task: t, Error: deployErr}
			}(task)
		}
	}

	// Wait for all deployments
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	for result := range resultsChan {
		if result.Error != nil {
			log.Error().Msgf("  ❌ Failed to deploy %s: %v", result.Task.ArtifactID, result.Error)
			stats.DeploymentTasksFailed++
		} else {
			log.Info().Msgf("  ✅ Successfully deployed %s", result.Task.ArtifactID)
			stats.DeploymentTasksSuccessful++
			stats.ArtifactsDeployed++
		}
	}

	return nil
}

func deployArtifact(exe *httpclnt.HTTPExecuter, task DeploymentTask,
	maxRetries, delaySeconds int) error {

	// Initialize designtime artifact based on type
	dt := api.NewDesigntimeArtifact(task.ArtifactType, exe)

	// Initialize runtime artifact for status checking
	rt := api.NewRuntime(exe)

	// Deploy the artifact
	log.Info().Msgf("    Deploying %s (type: %s)", task.ArtifactID, task.ArtifactType)
	err := dt.Deploy(task.ArtifactID)
	if err != nil {
		return fmt.Errorf("failed to initiate deployment: %w", err)
	}

	log.Info().Msgf("    Deployment triggered for %s", task.ArtifactID)

	// Poll for deployment status
	for i := 0; i < maxRetries; i++ {
		time.Sleep(time.Duration(delaySeconds) * time.Second)

		version, status, err := rt.Get(task.ArtifactID)
		if err != nil {
			log.Warn().Msgf("    Failed to get deployment status (attempt %d/%d): %v",
				i+1, maxRetries, err)
			continue
		}

		log.Info().Msgf("    Check %d/%d - Status: %s, Version: %s", i+1, maxRetries, status, version)

		if version == "NOT_DEPLOYED" {
			continue
		}

		if status == "STARTED" {
			return nil
		} else if status != "STARTING" {
			// Get error details
			time.Sleep(time.Duration(delaySeconds) * time.Second)
			errorMessage, err := rt.GetErrorInfo(task.ArtifactID)
			if err != nil {
				return fmt.Errorf("deployment failed with status %s: %w", status, err)
			}
			return fmt.Errorf("deployment failed with status %s: %s", status, errorMessage)
		}
	}

	return fmt.Errorf("deployment status check timed out after %d attempts", maxRetries)
}

func printConfigureSummary(stats *ConfigureStats, dryRun bool) {
	log.Info().Msg("")
	log.Info().Msg("═══════════════════════════════════════════════════════════════════════")
	if dryRun {
		log.Info().Msg("DRY RUN SUMMARY")
	} else {
		log.Info().Msg("CONFIGURATION SUMMARY")
	}
	log.Info().Msg("═══════════════════════════════════════════════════════════════════════")
	log.Info().Msgf("Packages processed:          %d", stats.PackagesProcessed)
	log.Info().Msgf("Packages with errors:        %d", stats.PackagesWithErrors)
	log.Info().Msgf("Artifacts processed:         %d", stats.ArtifactsProcessed)
	log.Info().Msgf("Artifacts configured:        %d", stats.ArtifactsConfigured)
	log.Info().Msgf("Artifacts failed:            %d", stats.ArtifactsFailed)
	log.Info().Msgf("Parameters updated:          %d", stats.ParametersUpdated)
	log.Info().Msgf("Parameters failed:           %d", stats.ParametersFailed)

	if !dryRun {
		log.Info().Msg("")
		log.Info().Msg("Performance:")
		log.Info().Msgf("Batch requests executed:     %d", stats.BatchRequestsExecuted)
		log.Info().Msgf("Individual requests used:    %d", stats.IndividualRequestsUsed)
	}

	if stats.DeploymentTasksQueued > 0 {
		log.Info().Msg("")
		log.Info().Msg("Deployment:")
		log.Info().Msgf("Deployment tasks queued:     %d", stats.DeploymentTasksQueued)
		if !dryRun {
			log.Info().Msgf("Deployments successful:      %d", stats.DeploymentTasksSuccessful)
			log.Info().Msgf("Deployments failed:          %d", stats.DeploymentTasksFailed)
			log.Info().Msgf("Artifacts deployed:          %d", stats.ArtifactsDeployed)
		}
	}

	log.Info().Msg("═══════════════════════════════════════════════════════════════════════")

	if stats.ArtifactsFailed > 0 || stats.DeploymentTasksFailed > 0 {
		log.Error().Msg("❌ Configuration/Deployment completed with errors")
	} else if dryRun {
		log.Info().Msg("✅ Dry run completed successfully")
	} else {
		log.Info().Msg("✅ Configuration/Deployment completed successfully")
	}
}

func escapeJSON(s string) string {
	// Simple JSON string escaping
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}
