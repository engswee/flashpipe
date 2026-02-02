package deploy

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/engswee/flashpipe/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigLoader(t *testing.T) {
	loader := NewConfigLoader()
	assert.NotNil(t, loader)
	assert.Equal(t, SourceFile, loader.Source)
	assert.Equal(t, "*.y*ml", loader.FilePattern)
	assert.Equal(t, "bearer", loader.AuthType)
}

func TestDetectSource_File(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test file
	testFile := filepath.Join(tempDir, "config.yml")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	loader := NewConfigLoader()
	err = loader.DetectSource(testFile)
	require.NoError(t, err)

	assert.Equal(t, SourceFile, loader.Source)
	assert.Equal(t, testFile, loader.Path)
}

func TestDetectSource_Folder(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	loader := NewConfigLoader()
	err = loader.DetectSource(tempDir)
	require.NoError(t, err)

	assert.Equal(t, SourceFolder, loader.Source)
	assert.Equal(t, tempDir, loader.Path)
}

func TestDetectSource_URL(t *testing.T) {
	loader := NewConfigLoader()

	tests := []struct {
		name string
		url  string
	}{
		{"http", "http://example.com/config.yml"},
		{"https", "https://example.com/config.yml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := loader.DetectSource(tt.url)
			require.NoError(t, err)

			assert.Equal(t, SourceURL, loader.Source)
			assert.Equal(t, tt.url, loader.URL)
		})
	}
}

func TestDetectSource_NonExistent(t *testing.T) {
	loader := NewConfigLoader()
	err := loader.DetectSource("/nonexistent/path")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path does not exist")
}

func TestLoadSingleFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test config
	configFile := filepath.Join(tempDir, "test-config.yml")
	configContent := `
deploymentPrefix: TEST
packages:
  - integrationSuiteId: Package1
    displayName: Test Package 1
    artifacts:
      - artifactId: artifact1
        displayName: Artifact 1
        type: Integration
`
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	loader := NewConfigLoader()
	loader.Path = configFile
	loader.Source = SourceFile

	configs, err := loader.LoadConfigs()
	require.NoError(t, err)
	require.Len(t, configs, 1)

	assert.Equal(t, "TEST", configs[0].Config.DeploymentPrefix)
	assert.Len(t, configs[0].Config.Packages, 1)
	assert.Equal(t, "Package1", configs[0].Config.Packages[0].ID)
	assert.Equal(t, configFile, configs[0].Source)
	assert.Equal(t, "test-config.yml", configs[0].FileName)
}

func TestLoadFolder_SingleFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test config
	configFile := filepath.Join(tempDir, "config.yml")
	configContent := `
deploymentPrefix: TEST
packages:
  - integrationSuiteId: Package1
`
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	loader := NewConfigLoader()
	loader.Path = tempDir
	loader.Source = SourceFolder

	configs, err := loader.LoadConfigs()
	require.NoError(t, err)
	require.Len(t, configs, 1)

	assert.Equal(t, "TEST", configs[0].Config.DeploymentPrefix)
}

func TestLoadFolder_MultipleFiles(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create multiple test configs
	configs := map[string]string{
		"a-config.yml": `
deploymentPrefix: A
packages:
  - integrationSuiteId: PackageA
`,
		"b-config.yaml": `
deploymentPrefix: B
packages:
  - integrationSuiteId: PackageB
`,
		"c-config.yml": `
deploymentPrefix: C
packages:
  - integrationSuiteId: PackageC
`,
	}

	for filename, content := range configs {
		err = os.WriteFile(filepath.Join(tempDir, filename), []byte(content), 0644)
		require.NoError(t, err)
	}

	loader := NewConfigLoader()
	loader.Path = tempDir
	loader.Source = SourceFolder

	loadedConfigs, err := loader.LoadConfigs()
	require.NoError(t, err)
	require.Len(t, loadedConfigs, 3)

	// Verify alphabetical order
	assert.Equal(t, "a-config.yml", loadedConfigs[0].FileName)
	assert.Equal(t, "b-config.yaml", loadedConfigs[1].FileName)
	assert.Equal(t, "c-config.yml", loadedConfigs[2].FileName)

	// Verify order numbers
	assert.Equal(t, 0, loadedConfigs[0].Order)
	assert.Equal(t, 1, loadedConfigs[1].Order)
	assert.Equal(t, 2, loadedConfigs[2].Order)
}

func TestLoadFolder_Recursive(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create subdirectory structure
	subDir1 := filepath.Join(tempDir, "env1")
	subDir2 := filepath.Join(tempDir, "env2", "configs")
	err = os.MkdirAll(subDir1, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(subDir2, 0755)
	require.NoError(t, err)

	// Create configs in different directories
	configs := map[string]string{
		filepath.Join(tempDir, "root.yml"): "deploymentPrefix: ROOT\npackages: []",
		filepath.Join(subDir1, "env1.yml"): "deploymentPrefix: ENV1\npackages: []",
		filepath.Join(subDir2, "deep.yml"): "deploymentPrefix: DEEP\npackages: []",
	}

	for path, content := range configs {
		err = os.WriteFile(path, []byte(content), 0644)
		require.NoError(t, err)
	}

	loader := NewConfigLoader()
	loader.Path = tempDir
	loader.Source = SourceFolder

	loadedConfigs, err := loader.LoadConfigs()
	require.NoError(t, err)
	require.Len(t, loadedConfigs, 3)

	// Verify all files were found (alphabetically sorted by full path)
	foundPrefixes := make(map[string]bool)
	for _, cfg := range loadedConfigs {
		foundPrefixes[cfg.Config.DeploymentPrefix] = true
	}

	assert.True(t, foundPrefixes["ROOT"])
	assert.True(t, foundPrefixes["ENV1"])
	assert.True(t, foundPrefixes["DEEP"])
}

func TestLoadFolder_NoMatches(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create non-matching file
	err = os.WriteFile(filepath.Join(tempDir, "test.txt"), []byte("test"), 0644)
	require.NoError(t, err)

	loader := NewConfigLoader()
	loader.Path = tempDir
	loader.Source = SourceFolder

	_, err = loader.LoadConfigs()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no config files found")
}

func TestLoadFolder_CustomPattern(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create files with different extensions
	err = os.WriteFile(filepath.Join(tempDir, "config.yml"), []byte("deploymentPrefix: YML\npackages: []"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tempDir, "config.json"), []byte("deploymentPrefix: JSON\npackages: []"), 0644)
	require.NoError(t, err)

	loader := NewConfigLoader()
	loader.Path = tempDir
	loader.Source = SourceFolder
	loader.FilePattern = "*.json"

	configs, err := loader.LoadConfigs()
	require.NoError(t, err)
	// Should only load the .json file (not .yml)
	require.Len(t, configs, 1)
	assert.Equal(t, "JSON", configs[0].Config.DeploymentPrefix)
}

func TestLoadFolder_InvalidYAML(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create invalid YAML file
	err = os.WriteFile(filepath.Join(tempDir, "invalid.yml"), []byte("invalid: yaml: content:"), 0644)
	require.NoError(t, err)

	// Create valid YAML file
	err = os.WriteFile(filepath.Join(tempDir, "valid.yml"), []byte("deploymentPrefix: VALID\npackages: []"), 0644)
	require.NoError(t, err)

	loader := NewConfigLoader()
	loader.Path = tempDir
	loader.Source = SourceFolder

	configs, err := loader.LoadConfigs()
	require.NoError(t, err)
	// Should only load the valid file
	require.Len(t, configs, 1)
	assert.Equal(t, "VALID", configs[0].Config.DeploymentPrefix)
}

func TestLoadURL_Success(t *testing.T) {
	configContent := `
deploymentPrefix: REMOTE
packages:
  - integrationSuiteId: RemotePackage
`

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(configContent))
	}))
	defer server.Close()

	loader := NewConfigLoader()
	loader.URL = server.URL
	loader.Source = SourceURL

	configs, err := loader.LoadConfigs()
	require.NoError(t, err)
	require.Len(t, configs, 1)

	assert.Equal(t, "REMOTE", configs[0].Config.DeploymentPrefix)
	assert.Equal(t, server.URL, configs[0].Source)
}

func TestLoadURL_WithBearerAuth(t *testing.T) {
	expectedToken := "test-token-123"
	configContent := "deploymentPrefix: AUTH\npackages: []"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer "+expectedToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(configContent))
	}))
	defer server.Close()

	loader := NewConfigLoader()
	loader.URL = server.URL
	loader.Source = SourceURL
	loader.AuthToken = expectedToken
	loader.AuthType = "bearer"

	configs, err := loader.LoadConfigs()
	require.NoError(t, err)
	require.Len(t, configs, 1)
	assert.Equal(t, "AUTH", configs[0].Config.DeploymentPrefix)
}

func TestLoadURL_WithBasicAuth(t *testing.T) {
	expectedUser := "testuser"
	expectedPass := "testpass"
	configContent := "deploymentPrefix: BASIC\npackages: []"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != expectedUser || pass != expectedPass {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(configContent))
	}))
	defer server.Close()

	loader := NewConfigLoader()
	loader.URL = server.URL
	loader.Source = SourceURL
	loader.Username = expectedUser
	loader.Password = expectedPass

	configs, err := loader.LoadConfigs()
	require.NoError(t, err)
	require.Len(t, configs, 1)
	assert.Equal(t, "BASIC", configs[0].Config.DeploymentPrefix)
}

func TestLoadURL_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	loader := NewConfigLoader()
	loader.URL = server.URL
	loader.Source = SourceURL

	_, err := loader.LoadConfigs()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status 404")
}

func TestMergeConfigs_Single(t *testing.T) {
	configs := []*DeployConfigFile{
		{
			Config: &models.DeployConfig{
				DeploymentPrefix: "TEST",
				Packages: []models.Package{
					{ID: "Package1"},
				},
			},
			FileName: "test.yml",
		},
	}

	merged, err := MergeConfigs(configs)
	require.NoError(t, err)

	assert.Equal(t, "", merged.DeploymentPrefix)
	assert.Len(t, merged.Packages, 1)
	assert.Equal(t, "TESTPackage1", merged.Packages[0].ID)
}

func TestMergeConfigs_Multiple(t *testing.T) {
	configs := []*DeployConfigFile{
		{
			Config: &models.DeployConfig{
				DeploymentPrefix: "DEV",
				Packages: []models.Package{
					{ID: "Package1", DisplayName: "Pkg 1"},
				},
			},
			FileName: "dev.yml",
		},
		{
			Config: &models.DeployConfig{
				DeploymentPrefix: "QA",
				Packages: []models.Package{
					{ID: "Package2", DisplayName: "Pkg 2"},
				},
			},
			FileName: "qa.yml",
		},
	}

	merged, err := MergeConfigs(configs)
	require.NoError(t, err)

	assert.Equal(t, "", merged.DeploymentPrefix)
	assert.Len(t, merged.Packages, 2)

	// Verify prefixes applied
	assert.Equal(t, "DEVPackage1", merged.Packages[0].ID)
	assert.Equal(t, "DEV - Pkg 1", merged.Packages[0].DisplayName)

	assert.Equal(t, "QAPackage2", merged.Packages[1].ID)
	assert.Equal(t, "QA - Pkg 2", merged.Packages[1].DisplayName)
}

func TestMergeConfigs_NoPrefix(t *testing.T) {
	configs := []*DeployConfigFile{
		{
			Config: &models.DeployConfig{
				DeploymentPrefix: "",
				Packages: []models.Package{
					{ID: "Package1"},
				},
			},
			FileName: "test.yml",
		},
	}

	merged, err := MergeConfigs(configs)
	require.NoError(t, err)

	assert.Len(t, merged.Packages, 1)
	assert.Equal(t, "Package1", merged.Packages[0].ID) // No prefix applied
}

func TestMergeConfigs_DuplicateID(t *testing.T) {
	configs := []*DeployConfigFile{
		{
			Config: &models.DeployConfig{
				DeploymentPrefix: "ENV",
				Packages: []models.Package{
					{ID: "Package1"},
				},
			},
			FileName: "config1.yml",
		},
		{
			Config: &models.DeployConfig{
				DeploymentPrefix: "ENV",
				Packages: []models.Package{
					{ID: "Package1"}, // Same fully qualified ID
				},
			},
			FileName: "config2.yml",
		},
	}

	_, err := MergeConfigs(configs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate package ID")
	assert.Contains(t, err.Error(), "ENVPackage1")
}

func TestMergeConfigs_ArtifactPrefixing(t *testing.T) {
	configs := []*DeployConfigFile{
		{
			Config: &models.DeployConfig{
				DeploymentPrefix: "TEST",
				Packages: []models.Package{
					{
						ID: "Package1",
						Artifacts: []models.Artifact{
							{Id: "artifact1", Type: "Integration"},
							{Id: "artifact2", Type: "Integration"},
						},
					},
				},
			},
			FileName: "test.yml",
		},
	}

	merged, err := MergeConfigs(configs)
	require.NoError(t, err)

	require.Len(t, merged.Packages, 1)
	require.Len(t, merged.Packages[0].Artifacts, 2)

	// Verify artifact IDs are prefixed
	assert.Equal(t, "TEST_artifact1", merged.Packages[0].Artifacts[0].Id)
	assert.Equal(t, "TEST_artifact2", merged.Packages[0].Artifacts[1].Id)
}

func TestMergeConfigs_Empty(t *testing.T) {
	configs := []*DeployConfigFile{}

	_, err := MergeConfigs(configs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no configs to merge")
}

func TestMergeConfigs_DisplayNameGeneration(t *testing.T) {
	configs := []*DeployConfigFile{
		{
			Config: &models.DeployConfig{
				DeploymentPrefix: "PREFIX",
				Packages: []models.Package{
					{ID: "Package1"}, // No display name
				},
			},
			FileName: "test.yml",
		},
	}

	merged, err := MergeConfigs(configs)
	require.NoError(t, err)

	// Display name should be generated from prefix and ID
	assert.Equal(t, "PREFIX - Package1", merged.Packages[0].DisplayName)
}
