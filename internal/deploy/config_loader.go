package deploy

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/engswee/flashpipe/internal/models"
	"gopkg.in/yaml.v3"
)

// ConfigSource represents the type of configuration source
type ConfigSource string

const (
	SourceFile   ConfigSource = "file"
	SourceFolder ConfigSource = "folder"
	SourceURL    ConfigSource = "url"
)

// ConfigLoader handles loading deployment configurations from various sources
type ConfigLoader struct {
	Source      ConfigSource
	Path        string
	URL         string
	AuthToken   string
	AuthType    string // "bearer" or "basic"
	Username    string // for basic auth
	Password    string // for basic auth
	FilePattern string // pattern for config files in folders
	Debug       bool
}

// DeployConfigFile represents a loaded config file with metadata
type DeployConfigFile struct {
	Config   *models.DeployConfig
	Source   string // original source path/URL
	FileName string // base filename
	Order    int    // processing order
}

// NewConfigLoader creates a new config loader
func NewConfigLoader() *ConfigLoader {
	return &ConfigLoader{
		Source:      SourceFile,
		FilePattern: "*.y*ml", // default pattern matches .yml and .yaml
		AuthType:    "bearer",
	}
}

// DetectSource automatically detects the source type based on the path
func (cl *ConfigLoader) DetectSource(path string) error {
	// Check if it's a URL
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		cl.Source = SourceURL
		cl.URL = path
		return nil
	}

	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path does not exist: %s", path)
	}

	// Determine if it's a file or directory
	if info.IsDir() {
		cl.Source = SourceFolder
		cl.Path = path
	} else {
		cl.Source = SourceFile
		cl.Path = path
	}

	return nil
}

// LoadConfigs loads all configuration files based on the source type
func (cl *ConfigLoader) LoadConfigs() ([]*DeployConfigFile, error) {
	switch cl.Source {
	case SourceFile:
		return cl.loadSingleFile()
	case SourceFolder:
		return cl.loadFolder()
	case SourceURL:
		return cl.loadURL()
	default:
		return nil, fmt.Errorf("unsupported source type: %s", cl.Source)
	}
}

// loadSingleFile loads a single configuration file
func (cl *ConfigLoader) loadSingleFile() ([]*DeployConfigFile, error) {
	var config models.DeployConfig
	if err := readYAML(cl.Path, &config); err != nil {
		return nil, fmt.Errorf("failed to load config file %s: %w", cl.Path, err)
	}

	return []*DeployConfigFile{
		{
			Config:   &config,
			Source:   cl.Path,
			FileName: filepath.Base(cl.Path),
			Order:    0,
		},
	}, nil
}

// loadFolder loads all matching configuration files from a folder (including subdirectories recursively)
func (cl *ConfigLoader) loadFolder() ([]*DeployConfigFile, error) {
	var configFiles []*DeployConfigFile
	var files []string

	if cl.Debug {
		fmt.Printf("Scanning directory recursively: %s\n", cl.Path)
		fmt.Printf("File pattern: %s\n", cl.FilePattern)
	}

	// Walk through directory and all subdirectories recursively
	err := filepath.Walk(cl.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Log error but continue walking
			if cl.Debug {
				fmt.Printf("Warning: Error accessing path %s: %v\n", path, err)
			}
			return nil // Continue walking despite errors
		}

		// Skip directories (but continue walking into them)
		if info.IsDir() {
			if cl.Debug && path != cl.Path {
				fmt.Printf("Entering subdirectory: %s\n", path)
			}
			return nil
		}

		// Check if file matches pattern
		matched, err := filepath.Match(cl.FilePattern, filepath.Base(path))
		if err != nil {
			return fmt.Errorf("invalid file pattern: %w", err)
		}

		if matched {
			// Get relative path for better display
			relPath, _ := filepath.Rel(cl.Path, path)
			if cl.Debug {
				fmt.Printf("Found matching file: %s\n", relPath)
			}
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no config files found matching pattern '%s' in %s (searched recursively)", cl.FilePattern, cl.Path)
	}

	if cl.Debug {
		fmt.Printf("Found %d matching file(s)\n", len(files))
	}

	// Sort files alphabetically for consistent processing order
	sort.Strings(files)

	if cl.Debug {
		fmt.Println("Processing files in alphabetical order:")
		for i, f := range files {
			relPath, _ := filepath.Rel(cl.Path, f)
			fmt.Printf("  %d. %s\n", i+1, relPath)
		}
	}

	// Load each file
	successCount := 0
	for i, filePath := range files {
		var config models.DeployConfig
		if err := readYAML(filePath, &config); err != nil {
			relPath, _ := filepath.Rel(cl.Path, filePath)
			if cl.Debug {
				fmt.Printf("Warning: Failed to load config file %s: %v\n", relPath, err)
			}
			continue
		}

		// Get relative path from base directory for better display
		relPath, _ := filepath.Rel(cl.Path, filePath)

		configFiles = append(configFiles, &DeployConfigFile{
			Config:   &config,
			Source:   filePath,
			FileName: relPath,
			Order:    i,
		})

		successCount++
		if cl.Debug {
			fmt.Printf("✓ Loaded config file: %s (order: %d)\n", relPath, i)
		}
	}

	if len(configFiles) == 0 {
		return nil, fmt.Errorf("no valid config files found in %s (found %d file(s) but all failed to parse)", cl.Path, len(files))
	}

	if cl.Debug {
		fmt.Printf("\nSuccessfully loaded %d config file(s) out of %d found\n", successCount, len(files))
	}

	return configFiles, nil
}

// loadURL loads a configuration file from a remote URL
func (cl *ConfigLoader) loadURL() ([]*DeployConfigFile, error) {
	if cl.Debug {
		fmt.Printf("Fetching config from URL: %s\n", cl.URL)
	}

	// Create HTTP client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("GET", cl.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication if provided
	if cl.AuthToken != "" {
		if cl.AuthType == "bearer" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cl.AuthToken))
			if cl.Debug {
				fmt.Println("Using Bearer token authentication")
			}
		} else if cl.AuthType == "basic" {
			req.SetBasicAuth(cl.Username, cl.Password)
			if cl.Debug {
				fmt.Printf("Using Basic authentication with username: %s\n", cl.Username)
			}
		}
	} else if cl.Username != "" && cl.Password != "" {
		// Use basic auth if username/password provided without token
		req.SetBasicAuth(cl.Username, cl.Password)
		if cl.Debug {
			fmt.Printf("Using Basic authentication with username: %s\n", cl.Username)
		}
	}

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch URL: status %d", resp.StatusCode)
	}

	if cl.Debug {
		fmt.Printf("Successfully fetched config (status: %d)\n", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Save to temporary file for YAML parsing
	tempFile, err := os.CreateTemp("", "deploy-config-*.yml")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write(body); err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	tempFile.Close()

	// Parse YAML
	var config models.DeployConfig
	if err := readYAML(tempFile.Name(), &config); err != nil {
		return nil, fmt.Errorf("failed to parse config from URL: %w", err)
	}

	// Extract filename from URL
	urlParts := strings.Split(cl.URL, "/")
	fileName := urlParts[len(urlParts)-1]
	if fileName == "" {
		fileName = "remote-config.yml"
	}

	if cl.Debug {
		fmt.Printf("✓ Successfully parsed config from URL\n")
	}

	return []*DeployConfigFile{
		{
			Config:   &config,
			Source:   cl.URL,
			FileName: fileName,
			Order:    0,
		},
	}, nil
}

// MergeConfigs merges multiple deployment configs into a single config
func MergeConfigs(configs []*DeployConfigFile) (*models.DeployConfig, error) {
	if len(configs) == 0 {
		return nil, fmt.Errorf("no configs to merge")
	}

	// Merged config has NO deployment prefix since each package will have its own
	merged := &models.DeployConfig{
		DeploymentPrefix: "",
		Packages:         []models.Package{},
	}

	// Track fully qualified package IDs (with prefix) to detect true duplicates
	packageMap := make(map[string]string) // map[fullyQualifiedID]sourceFile

	// Merge packages from all configs
	for _, configFile := range configs {
		configPrefix := configFile.Config.DeploymentPrefix

		for _, pkg := range configFile.Config.Packages {
			// Create a copy of the package to avoid modifying the original
			mergedPkg := pkg

			// Calculate the fully qualified package ID
			fullyQualifiedID := pkg.ID
			if configPrefix != "" {
				fullyQualifiedID = configPrefix + "" + pkg.ID

				// Update the package ID and display name with prefix
				mergedPkg.ID = fullyQualifiedID

				// Update display name if it exists
				if mergedPkg.DisplayName != "" {
					mergedPkg.DisplayName = configPrefix + " - " + mergedPkg.DisplayName
				} else {
					mergedPkg.DisplayName = configPrefix + " - " + pkg.ID
				}
			}

			// Check for duplicate fully qualified IDs
			if existingSource, exists := packageMap[fullyQualifiedID]; exists {
				return nil, fmt.Errorf("duplicate package ID '%s' found in %s (already exists from %s)",
					fullyQualifiedID, configFile.FileName, existingSource)
			}

			// Apply prefix to all artifact IDs as well
			if configPrefix != "" {
				for i := range mergedPkg.Artifacts {
					mergedPkg.Artifacts[i].Id = configPrefix + "_" + mergedPkg.Artifacts[i].Id
				}
			}

			packageMap[fullyQualifiedID] = configFile.FileName
			merged.Packages = append(merged.Packages, mergedPkg)
		}
	}

	return merged, nil
}

// readYAML reads and unmarshals a YAML file
func readYAML(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if err := yaml.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	return nil
}
