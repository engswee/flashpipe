package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/engswee/flashpipe/internal/analytics"
	"github.com/engswee/flashpipe/internal/config"
	"github.com/engswee/flashpipe/internal/file"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func NewConfigGenerateCommand() *cobra.Command {

	configCmd := &cobra.Command{
		Use:   "config-generate",
		Short: "Generate or update deployment configuration",
		Long: `Generate or update deployment configuration from package directory structure.

This command scans the packages directory and generates/updates a deployment configuration
file (001-deploy-config.yml) with all discovered packages and artifacts.

Features:
  - Extracts package metadata from {PackageName}.json files
  - Extracts artifact display names from MANIFEST.MF (Bundle-Name)
  - Extracts artifact types from MANIFEST.MF (SAP-BundleType)
  - Preserves existing configuration settings (sync/deploy flags, config overrides)
  - Smart merging of new and existing configurations
  - Filter by specific packages or artifacts`,
		Example: `  # Generate config with defaults
  flashpipe config-generate

  # Specify custom directories
  flashpipe config-generate --packages-dir ./my-packages --output ./my-config.yml

  # Generate config for specific packages only
  flashpipe config-generate --package-filter "DeviceManagement,GenericPipeline"

  # Generate config for specific artifacts only
  flashpipe config-generate --artifact-filter "MDMEquipmentMutationOutbound,GenericBroadcaster"

  # Combine package and artifact filters
  flashpipe config-generate --package-filter "DeviceManagement" --artifact-filter "MDMEquipmentMutationOutbound"`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			startTime := time.Now()
			if err = runConfigGenerate(cmd); err != nil {
				cmd.SilenceUsage = true
			}
			analytics.Log(cmd, err, startTime)
			return
		},
	}

	configCmd.Flags().String("packages-dir", "./packages",
		"Path to packages directory")
	configCmd.Flags().String("output", "./001-deploy-config.yml",
		"Path to output configuration file")
	configCmd.Flags().StringSlice("package-filter", nil,
		"Comma separated list of packages to include (e.g., 'Package1,Package2')")
	configCmd.Flags().StringSlice("artifact-filter", nil,
		"Comma separated list of artifacts to include (e.g., 'Artifact1,Artifact2')")

	return configCmd
}

func runConfigGenerate(cmd *cobra.Command) error {
	packagesDir := config.GetString(cmd, "packages-dir")
	outputFile := config.GetString(cmd, "output")
	packageFilter := config.GetStringSlice(cmd, "package-filter")
	artifactFilter := config.GetStringSlice(cmd, "artifact-filter")

	generator := NewConfigGenerator(packagesDir, outputFile, packageFilter, artifactFilter)

	if err := generator.Generate(); err != nil {
		return err
	}

	return nil
}

// ConfigGenerator handles configuration generation
type ConfigGenerator struct {
	PackagesDir    string
	OutputFile     string
	PackageFilter  []string
	ArtifactFilter []string
	ExistingConfig *DeployConfig
	Stats          GenerationStats
}

// GenerationStats tracks generation statistics
type GenerationStats struct {
	PackagesPreserved          int
	PackagesAdded              int
	PackagesRemoved            int
	PackagesFiltered           int
	PackagePropertiesExtracted int
	PackagePropertiesPreserved int
	ArtifactsPreserved         int
	ArtifactsAdded             int
	ArtifactsRemoved           int
	ArtifactsFiltered          int
	ArtifactsNameExtracted     int
	ArtifactsNamePreserved     int
	ArtifactsTypeExtracted     int
	ArtifactsTypePreserved     int
}

// DeployConfig represents the complete deployment configuration
type DeployConfig struct {
	DeploymentPrefix string    `yaml:"deploymentPrefix,omitempty"`
	Packages         []Package `yaml:"packages"`
}

// Package represents a SAP CPI package
type Package struct {
	ID          string     `yaml:"integrationSuiteId"`
	PackageDir  string     `yaml:"packageDir,omitempty"`
	DisplayName string     `yaml:"displayName,omitempty"`
	Description string     `yaml:"description,omitempty"`
	ShortText   string     `yaml:"short_text,omitempty"`
	Sync        bool       `yaml:"sync"`
	Deploy      bool       `yaml:"deploy"`
	Artifacts   []Artifact `yaml:"artifacts"`
}

// Artifact represents a SAP CPI artifact
type Artifact struct {
	Id              string                 `yaml:"artifactId"`
	ArtifactDir     string                 `yaml:"artifactDir"`
	DisplayName     string                 `yaml:"displayName,omitempty"`
	Type            string                 `yaml:"type"`
	Sync            bool                   `yaml:"sync"`
	Deploy          bool                   `yaml:"deploy"`
	ConfigOverrides map[string]interface{} `yaml:"configOverrides,omitempty"`
}

// PackageMetadata represents metadata from package JSON
type PackageMetadata struct {
	ID          string `json:"Id"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
	ShortText   string `json:"ShortText"`
}

// NewConfigGenerator creates a new configuration generator
func NewConfigGenerator(packagesDir, outputFile string, packageFilter, artifactFilter []string) *ConfigGenerator {
	return &ConfigGenerator{
		PackagesDir:    packagesDir,
		OutputFile:     outputFile,
		PackageFilter:  packageFilter,
		ArtifactFilter: artifactFilter,
	}
}

// shouldIncludePackage checks if a package should be included based on filter
func (g *ConfigGenerator) shouldIncludePackage(packageName string) bool {
	if len(g.PackageFilter) == 0 {
		return true
	}
	for _, filterPkg := range g.PackageFilter {
		if filterPkg == packageName {
			return true
		}
	}
	return false
}

// shouldIncludeArtifact checks if an artifact should be included based on filter
func (g *ConfigGenerator) shouldIncludeArtifact(artifactName string) bool {
	if len(g.ArtifactFilter) == 0 {
		return true
	}
	for _, filterArt := range g.ArtifactFilter {
		if filterArt == artifactName {
			return true
		}
	}
	return false
}

// Generate generates or updates the deployment configuration
func (g *ConfigGenerator) Generate() error {
	log.Info().Msg("Generating/Updating Configuration")
	log.Info().Msgf("Packages directory: %s", g.PackagesDir)
	log.Info().Msgf("Config file: %s", g.OutputFile)

	if len(g.PackageFilter) > 0 {
		log.Info().Msgf("Package filter: %s", strings.Join(g.PackageFilter, ", "))
	}
	if len(g.ArtifactFilter) > 0 {
		log.Info().Msgf("Artifact filter: %s", strings.Join(g.ArtifactFilter, ", "))
	}

	// Check if packages directory exists
	if _, err := os.Stat(g.PackagesDir); os.IsNotExist(err) {
		return fmt.Errorf("packages directory '%s' not found", g.PackagesDir)
	}

	// Load existing config if it exists
	if _, err := os.Stat(g.OutputFile); err == nil {
		log.Info().Msg("Loading existing configuration...")
		data, err := os.ReadFile(g.OutputFile)
		if err != nil {
			return fmt.Errorf("failed to read existing config: %w", err)
		}
		var existingConfig DeployConfig
		if err := yaml.Unmarshal(data, &existingConfig); err != nil {
			return fmt.Errorf("failed to parse existing config: %w", err)
		}
		g.ExistingConfig = &existingConfig
	}

	// Create new config structure
	newConfig := DeployConfig{
		DeploymentPrefix: "",
		Packages:         []Package{},
	}

	// Preserve deployment prefix if exists
	if g.ExistingConfig != nil {
		newConfig.DeploymentPrefix = g.ExistingConfig.DeploymentPrefix
	}

	// Build map of existing packages and artifacts for quick lookup
	existingPackages := make(map[string]Package)
	existingArtifacts := make(map[string]map[string]Artifact)

	if g.ExistingConfig != nil {
		for _, pkg := range g.ExistingConfig.Packages {
			existingPackages[pkg.ID] = pkg
			existingArtifacts[pkg.ID] = make(map[string]Artifact)
			for _, art := range pkg.Artifacts {
				existingArtifacts[pkg.ID][art.Id] = art
			}
		}
	}

	// Scan packages directory
	entries, err := os.ReadDir(g.PackagesDir)
	if err != nil {
		return fmt.Errorf("failed to read packages directory: %w", err)
	}

	processedPackages := make(map[string]bool)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		packageName := entry.Name()

		// Apply package filter
		if !g.shouldIncludePackage(packageName) {
			g.Stats.PackagesFiltered++
			continue
		}

		packageDir := filepath.Join(g.PackagesDir, packageName)

		log.Debug().Msgf("Processing package: %s", packageName)

		processedPackages[packageName] = true

		// Extract package metadata
		metadata := g.extractPackageMetadata(packageDir, packageName)

		// Check if package exists in old config
		var pkg Package
		if existingPkg, exists := existingPackages[packageName]; exists {
			pkg = existingPkg
			g.Stats.PackagesPreserved++

			if metadata != nil {
				if pkg.PackageDir == "" || pkg.DisplayName == "" {
					g.Stats.PackagePropertiesExtracted++
				} else {
					g.Stats.PackagePropertiesPreserved++
				}

				if pkg.PackageDir == "" {
					pkg.PackageDir = metadata.ID
				}
				if pkg.DisplayName == "" {
					pkg.DisplayName = metadata.Name
				}
				if pkg.Description == "" {
					pkg.Description = metadata.Description
				}
				if pkg.ShortText == "" {
					pkg.ShortText = metadata.ShortText
				}
			}
		} else {
			pkg = Package{
				ID:     packageName,
				Sync:   true,
				Deploy: true,
			}

			if metadata != nil {
				pkg.PackageDir = metadata.ID
				pkg.DisplayName = metadata.Name
				pkg.Description = metadata.Description
				pkg.ShortText = metadata.ShortText
				g.Stats.PackagePropertiesExtracted++
			}

			g.Stats.PackagesAdded++
		}

		// Reset artifacts slice
		pkg.Artifacts = []Artifact{}

		// Scan artifacts
		artifactEntries, err := os.ReadDir(packageDir)
		if err != nil {
			log.Warn().Msgf("Failed to read package directory: %v", err)
			continue
		}

		processedArtifacts := make(map[string]bool)

		for _, artEntry := range artifactEntries {
			if !artEntry.IsDir() {
				continue
			}

			artifactName := artEntry.Name()

			// Apply artifact filter
			if !g.shouldIncludeArtifact(artifactName) {
				g.Stats.ArtifactsFiltered++
				continue
			}

			artifactDir := filepath.Join(packageDir, artifactName)

			processedArtifacts[artifactName] = true

			// Extract artifact metadata from MANIFEST.MF
			bundleName, artifactType := g.extractManifestMetadata(artifactDir)

			// Check if artifact exists in old config
			var artifact Artifact
			if existingArtMap, pkgExists := existingArtifacts[packageName]; pkgExists {
				if existingArt, artExists := existingArtMap[artifactName]; artExists {
					artifact = existingArt
					g.Stats.ArtifactsPreserved++

					if bundleName != "" {
						if artifact.DisplayName == "" {
							g.Stats.ArtifactsNameExtracted++
							artifact.DisplayName = bundleName
						} else {
							g.Stats.ArtifactsNamePreserved++
						}
					}

					if artifactType != "" {
						if artifact.Type == "" {
							g.Stats.ArtifactsTypeExtracted++
							artifact.Type = artifactType
						} else {
							g.Stats.ArtifactsTypePreserved++
						}
					}

					if artifact.ArtifactDir == "" {
						artifact.ArtifactDir = artifactName
					}
				} else {
					artifact = Artifact{
						Id:              artifactName,
						ArtifactDir:     artifactName,
						DisplayName:     bundleName,
						Type:            artifactType,
						Sync:            true,
						Deploy:          true,
						ConfigOverrides: make(map[string]interface{}),
					}

					if bundleName != "" {
						g.Stats.ArtifactsNameExtracted++
					}
					if artifactType != "" {
						g.Stats.ArtifactsTypeExtracted++
					}

					g.Stats.ArtifactsAdded++
				}
			} else {
				artifact = Artifact{
					Id:              artifactName,
					ArtifactDir:     artifactName,
					DisplayName:     bundleName,
					Type:            artifactType,
					Sync:            true,
					Deploy:          true,
					ConfigOverrides: make(map[string]interface{}),
				}

				if bundleName != "" {
					g.Stats.ArtifactsNameExtracted++
				}
				if artifactType != "" {
					g.Stats.ArtifactsTypeExtracted++
				}

				g.Stats.ArtifactsAdded++
			}

			pkg.Artifacts = append(pkg.Artifacts, artifact)
		}

		// Count removed artifacts
		if existingArtMap, pkgExists := existingArtifacts[packageName]; pkgExists {
			for artName := range existingArtMap {
				if !processedArtifacts[artName] {
					g.Stats.ArtifactsRemoved++
				}
			}
		}

		// Only add package if it has artifacts (when artifact filter is used)
		if len(g.ArtifactFilter) > 0 && len(pkg.Artifacts) == 0 {
			continue
		}

		newConfig.Packages = append(newConfig.Packages, pkg)
	}

	// Count removed packages
	if g.ExistingConfig != nil {
		for _, pkg := range g.ExistingConfig.Packages {
			if !processedPackages[pkg.ID] {
				g.Stats.PackagesRemoved++
			}
		}
	}

	// Sort packages by ID for consistency
	sort.Slice(newConfig.Packages, func(i, j int) bool {
		return newConfig.Packages[i].ID < newConfig.Packages[j].ID
	})

	// Write config file
	if err := g.writeConfigFile(g.OutputFile, &newConfig); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	g.printSummary()

	return nil
}

func (g *ConfigGenerator) extractPackageMetadata(packageDir, packageName string) *PackageMetadata {
	jsonFile := filepath.Join(packageDir, packageName+".json")
	if _, err := os.Stat(jsonFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(jsonFile)
	if err != nil {
		log.Warn().Msgf("Failed to read package JSON: %v", err)
		return nil
	}

	var wrapper struct {
		D PackageMetadata `json:"d"`
	}

	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		log.Warn().Msgf("Failed to parse package JSON: %v", err)
		return nil
	}

	return &wrapper.D
}

func (g *ConfigGenerator) extractManifestMetadata(artifactDir string) (bundleName, artifactType string) {
	manifestPath := filepath.Join(artifactDir, "META-INF", "MANIFEST.MF")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return "", ""
	}

	manifestData, err := file.ReadManifest(manifestPath)
	if err != nil {
		log.Warn().Msgf("Failed to read manifest: %v", err)
		return "", ""
	}

	bundleName = manifestData["Bundle-Name"]
	artifactType = manifestData["SAP-BundleType"]

	return bundleName, artifactType
}

func (g *ConfigGenerator) writeConfigFile(outputPath string, cfg *DeployConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	header := `# SAP CPI Deployment Configuration
# Generated by: flashpipe config-generate
#
# ============================================================================
# FIELD DESCRIPTIONS
# ============================================================================
#
# PACKAGE FIELDS:
#   integrationSuiteId (required): Unique ID of the integration package in SAP CPI
#   packageDir (required): Local directory containing the package artifacts
#   displayName (optional): Override the package's display name in SAP CPI
#   description (optional): Override the package description
#   short_text (optional): Override the package short text
#   sync (default: true): Whether to update/sync this package to the tenant
#   deploy (default: true): Whether to deploy this package
#   artifacts: List of artifacts within this package
#
# ARTIFACT FIELDS:
#   artifactId (required): Unique ID of the artifact (IFlow, Script Collection, etc.)
#   artifactDir (required): Local directory path to the artifact
#   displayName (optional): Override the artifact's display name in SAP CPI
#   type (required): Artifact type (Integration, ScriptCollection, MessageMapping, ValueMapping)
#   sync (default: true): Whether to update/sync this artifact to the tenant
#   deploy (default: true): Whether to deploy/activate this artifact
#   configOverrides (optional): Override parameter values from parameters.prop
#
# ============================================================================

`

	return os.WriteFile(outputPath, []byte(header+string(data)), 0644)
}

func (g *ConfigGenerator) printSummary() {
	log.Info().Msgf("Configuration saved to: %s", g.OutputFile)
	log.Info().Msg("Summary of Changes:")
	log.Info().Msg("  Packages:")
	log.Info().Msgf("    - Preserved: %d", g.Stats.PackagesPreserved)
	log.Info().Msgf("    - Added:     %d", g.Stats.PackagesAdded)
	log.Info().Msgf("    - Removed:   %d", g.Stats.PackagesRemoved)
	if g.Stats.PackagesFiltered > 0 {
		log.Info().Msgf("    - Filtered:  %d", g.Stats.PackagesFiltered)
	}
	log.Info().Msg("  Package Properties (from {PackageName}.json):")
	log.Info().Msgf("    - Extracted: %d", g.Stats.PackagePropertiesExtracted)
	log.Info().Msgf("    - Preserved: %d", g.Stats.PackagePropertiesPreserved)
	log.Info().Msg("  Artifacts:")
	log.Info().Msgf("    - Preserved: %d (settings kept)", g.Stats.ArtifactsPreserved)
	log.Info().Msgf("    - Added:     %d (defaults applied)", g.Stats.ArtifactsAdded)
	log.Info().Msgf("    - Removed:   %d (deleted from config)", g.Stats.ArtifactsRemoved)
	if g.Stats.ArtifactsFiltered > 0 {
		log.Info().Msgf("    - Filtered:  %d", g.Stats.ArtifactsFiltered)
	}
	log.Info().Msg("  Artifact Display Names (Bundle-Name from MANIFEST.MF):")
	log.Info().Msgf("    - Extracted: %d", g.Stats.ArtifactsNameExtracted)
	log.Info().Msgf("    - Preserved: %d", g.Stats.ArtifactsNamePreserved)
	log.Info().Msg("  Artifact Types (SAP-BundleType from MANIFEST.MF):")
	log.Info().Msgf("    - Extracted: %d", g.Stats.ArtifactsTypeExtracted)
	log.Info().Msgf("    - Preserved: %d", g.Stats.ArtifactsTypePreserved)
}
