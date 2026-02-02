package deploy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileExists(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "utils-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	// Create a directory
	testDir := filepath.Join(tempDir, "testdir")
	err = os.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	tests := []struct {
		name string
		path string
		want bool
	}{
		{"existing file", testFile, true},
		{"directory (not a file)", testDir, false},
		{"non-existent", filepath.Join(tempDir, "nonexistent.txt"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FileExists(tt.path)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestDirExists(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "utils-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	// Create a directory
	testDir := filepath.Join(tempDir, "testdir")
	err = os.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	tests := []struct {
		name string
		path string
		want bool
	}{
		{"existing directory", testDir, true},
		{"file (not a directory)", testFile, false},
		{"non-existent", filepath.Join(tempDir, "nonexistent"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DirExists(tt.path)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestValidateDeploymentPrefix_Valid(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
	}{
		{"empty prefix", ""},
		{"alphanumeric", "Test123"},
		{"uppercase", "PRODUCTION"},
		{"lowercase", "development"},
		{"with underscores", "dev_environment_1"},
		{"numbers only", "123"},
		{"letters only", "abc"},
		{"single char", "A"},
		{"underscore only", "_"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDeploymentPrefix(tt.prefix)
			assert.NoError(t, err)
		})
	}
}

func TestValidateDeploymentPrefix_Invalid(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
	}{
		{"with dash", "dev-env"},
		{"with space", "dev env"},
		{"with dot", "dev.env"},
		{"with special chars", "dev@env"},
		{"with slash", "dev/env"},
		{"with brackets", "dev[env]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDeploymentPrefix(tt.prefix)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "deployment prefix can only contain")
		})
	}
}

func TestCopyDir(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "utils-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create source directory structure
	srcDir := filepath.Join(tempDir, "src")
	err = os.MkdirAll(filepath.Join(srcDir, "subdir"), 0755)
	require.NoError(t, err)

	// Create files
	err = os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("content1"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(srcDir, "subdir", "file2.txt"), []byte("content2"), 0644)
	require.NoError(t, err)

	// Copy directory
	dstDir := filepath.Join(tempDir, "dst")
	err = CopyDir(srcDir, dstDir)
	require.NoError(t, err)

	// Verify copied files
	content1, err := os.ReadFile(filepath.Join(dstDir, "file1.txt"))
	require.NoError(t, err)
	assert.Equal(t, "content1", string(content1))

	content2, err := os.ReadFile(filepath.Join(dstDir, "subdir", "file2.txt"))
	require.NoError(t, err)
	assert.Equal(t, "content2", string(content2))

	// Verify directory exists
	assert.True(t, DirExists(filepath.Join(dstDir, "subdir")))
}

func TestUpdateManifestBundleName_BothFieldsExist(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "utils-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	manifestContent := `Manifest-Version: 1.0
Bundle-Name: OldName
Bundle-SymbolicName: OldSymbolicName
Bundle-Version: 1.0.0
`
	manifestPath := filepath.Join(tempDir, "MANIFEST.MF")
	err = os.WriteFile(manifestPath, []byte(manifestContent), 0644)
	require.NoError(t, err)

	outputPath := filepath.Join(tempDir, "MANIFEST_OUT.MF")
	err = UpdateManifestBundleName(manifestPath, "NewSymbolicName", "NewName", outputPath)
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "Bundle-Name: NewName")
	assert.Contains(t, contentStr, "Bundle-SymbolicName: NewSymbolicName")
	assert.NotContains(t, contentStr, "OldName")
	assert.NotContains(t, contentStr, "OldSymbolicName")
	assert.Contains(t, contentStr, "Bundle-Version: 1.0.0")
}

func TestUpdateManifestBundleName_FieldsMissing(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "utils-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	manifestContent := `Manifest-Version: 1.0
Bundle-Version: 1.0.0
`
	manifestPath := filepath.Join(tempDir, "MANIFEST.MF")
	err = os.WriteFile(manifestPath, []byte(manifestContent), 0644)
	require.NoError(t, err)

	outputPath := filepath.Join(tempDir, "MANIFEST_OUT.MF")
	err = UpdateManifestBundleName(manifestPath, "NewSymbolicName", "NewName", outputPath)
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "Bundle-Name: NewName")
	assert.Contains(t, contentStr, "Bundle-SymbolicName: NewSymbolicName")
}

func TestUpdateManifestBundleName_PreservesLineEndings_LF(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "utils-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	manifestContent := "Manifest-Version: 1.0\nBundle-Name: OldName\n"
	manifestPath := filepath.Join(tempDir, "MANIFEST.MF")
	err = os.WriteFile(manifestPath, []byte(manifestContent), 0644)
	require.NoError(t, err)

	outputPath := filepath.Join(tempDir, "MANIFEST_OUT.MF")
	err = UpdateManifestBundleName(manifestPath, "NewSymbolicName", "NewName", outputPath)
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	// Should use LF
	assert.Contains(t, string(content), "\n")
	assert.NotContains(t, string(content), "\r\n")
}

func TestUpdateManifestBundleName_PreservesLineEndings_CRLF(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "utils-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	manifestContent := "Manifest-Version: 1.0\r\nBundle-Name: OldName\r\n"
	manifestPath := filepath.Join(tempDir, "MANIFEST.MF")
	err = os.WriteFile(manifestPath, []byte(manifestContent), 0644)
	require.NoError(t, err)

	outputPath := filepath.Join(tempDir, "MANIFEST_OUT.MF")
	err = UpdateManifestBundleName(manifestPath, "NewSymbolicName", "NewName", outputPath)
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	// Should preserve CRLF
	assert.Contains(t, string(content), "\r\n")
}

func TestUpdateManifestBundleName_CaseInsensitive(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "utils-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Mix case headers
	manifestContent := `bundle-name: OldName
BUNDLE-SYMBOLICNAME: OldSymbolicName
`
	manifestPath := filepath.Join(tempDir, "MANIFEST.MF")
	err = os.WriteFile(manifestPath, []byte(manifestContent), 0644)
	require.NoError(t, err)

	outputPath := filepath.Join(tempDir, "MANIFEST_OUT.MF")
	err = UpdateManifestBundleName(manifestPath, "NewSymbolicName", "NewName", outputPath)
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "Bundle-Name: NewName")
	assert.Contains(t, contentStr, "Bundle-SymbolicName: NewSymbolicName")
}

func TestMergeParametersFile_NewFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "utils-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	paramsPath := filepath.Join(tempDir, "parameters.prop")
	outputPath := filepath.Join(tempDir, "output.prop")

	overrides := map[string]interface{}{
		"param1": "value1",
		"param2": 123,
		"param3": true,
	}

	err = MergeParametersFile(paramsPath, overrides, outputPath)
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "param1=value1")
	assert.Contains(t, contentStr, "param2=123")
	assert.Contains(t, contentStr, "param3=true")
}

func TestMergeParametersFile_ExistingFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "utils-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create existing parameters file
	existingContent := `param1=oldvalue1
param2=oldvalue2
param3=oldvalue3
`
	paramsPath := filepath.Join(tempDir, "parameters.prop")
	err = os.WriteFile(paramsPath, []byte(existingContent), 0644)
	require.NoError(t, err)

	outputPath := filepath.Join(tempDir, "output.prop")

	overrides := map[string]interface{}{
		"param2": "newvalue2",
		"param4": "newvalue4",
	}

	err = MergeParametersFile(paramsPath, overrides, outputPath)
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "param1=oldvalue1") // Unchanged
	assert.Contains(t, contentStr, "param2=newvalue2") // Overridden
	assert.Contains(t, contentStr, "param3=oldvalue3") // Unchanged
	assert.Contains(t, contentStr, "param4=newvalue4") // New
}

func TestMergeParametersFile_PreservesOrder(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "utils-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	existingContent := `aaa=value1
zzz=value2
mmm=value3
`
	paramsPath := filepath.Join(tempDir, "parameters.prop")
	err = os.WriteFile(paramsPath, []byte(existingContent), 0644)
	require.NoError(t, err)

	outputPath := filepath.Join(tempDir, "output.prop")

	overrides := map[string]interface{}{
		"bbb": "newvalue",
	}

	err = MergeParametersFile(paramsPath, overrides, outputPath)
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	lines := strings.Split(string(content), "\n")
	var paramLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			paramLines = append(paramLines, line)
		}
	}

	// Original order should be preserved, new param added at end
	assert.Equal(t, "aaa=value1", paramLines[0])
	assert.Equal(t, "zzz=value2", paramLines[1])
	assert.Equal(t, "mmm=value3", paramLines[2])
	assert.Equal(t, "bbb=newvalue", paramLines[3])
}

func TestMergeParametersFile_PreservesLineEndings_LF(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "utils-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	existingContent := "param1=value1\nparam2=value2\n"
	paramsPath := filepath.Join(tempDir, "parameters.prop")
	err = os.WriteFile(paramsPath, []byte(existingContent), 0644)
	require.NoError(t, err)

	outputPath := filepath.Join(tempDir, "output.prop")

	err = MergeParametersFile(paramsPath, map[string]interface{}{}, outputPath)
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	assert.Contains(t, string(content), "\n")
	assert.NotContains(t, string(content), "\r\n")
}

func TestMergeParametersFile_PreservesLineEndings_CRLF(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "utils-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	existingContent := "param1=value1\r\nparam2=value2\r\n"
	paramsPath := filepath.Join(tempDir, "parameters.prop")
	err = os.WriteFile(paramsPath, []byte(existingContent), 0644)
	require.NoError(t, err)

	outputPath := filepath.Join(tempDir, "output.prop")

	err = MergeParametersFile(paramsPath, map[string]interface{}{}, outputPath)
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	assert.Contains(t, string(content), "\r\n")
}

func TestFindParametersFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "utils-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name         string
		setupFunc    func(string) error
		expectedPath string
	}{
		{
			name: "in src/main/resources",
			setupFunc: func(dir string) error {
				path := filepath.Join(dir, "src", "main", "resources")
				if err := os.MkdirAll(path, 0755); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(path, "parameters.prop"), []byte("test"), 0644)
			},
			expectedPath: "src/main/resources/parameters.prop",
		},
		{
			name: "in src/main/resources/script",
			setupFunc: func(dir string) error {
				path := filepath.Join(dir, "src", "main", "resources", "script")
				if err := os.MkdirAll(path, 0755); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(path, "parameters.prop"), []byte("test"), 0644)
			},
			expectedPath: "src/main/resources/script/parameters.prop",
		},
		{
			name: "in root",
			setupFunc: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "parameters.prop"), []byte("test"), 0644)
			},
			expectedPath: "parameters.prop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir, err := os.MkdirTemp(tempDir, "find-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(testDir)

			err = tt.setupFunc(testDir)
			require.NoError(t, err)

			result := FindParametersFile(testDir)
			expected := filepath.Join(testDir, filepath.FromSlash(tt.expectedPath))
			assert.Equal(t, expected, result)
			assert.True(t, FileExists(result))
		})
	}
}

func TestFindParametersFile_NotFound(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "utils-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	result := FindParametersFile(tempDir)
	// Should return default path even if it doesn't exist
	expected := filepath.Join(tempDir, "src", "main", "resources", "parameters.prop")
	assert.Equal(t, expected, result)
}

func TestGetManifestHeaders(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "utils-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	manifestContent := `Manifest-Version: 1.0
Bundle-Name: Test Bundle
Bundle-SymbolicName: com.test.bundle
Bundle-Version: 1.0.0
Import-Package: javax.xml.bind,
 javax.xml.stream
Export-Package: com.test.api
`
	manifestPath := filepath.Join(tempDir, "MANIFEST.MF")
	err = os.WriteFile(manifestPath, []byte(manifestContent), 0644)
	require.NoError(t, err)

	headers, err := GetManifestHeaders(manifestPath)
	require.NoError(t, err)

	assert.Equal(t, "1.0", headers["Manifest-Version"])
	assert.Equal(t, "Test Bundle", headers["Bundle-Name"])
	assert.Equal(t, "com.test.bundle", headers["Bundle-SymbolicName"])
	assert.Equal(t, "1.0.0", headers["Bundle-Version"])
	assert.Equal(t, "javax.xml.bind, javax.xml.stream", headers["Import-Package"])
	assert.Equal(t, "com.test.api", headers["Export-Package"])
}

func TestGetManifestHeaders_MultilineContinuation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "utils-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	manifestContent := `Manifest-Version: 1.0
Import-Package: javax.xml.bind,
 javax.xml.stream,
 javax.xml.transform
Bundle-Name: Test
`
	manifestPath := filepath.Join(tempDir, "MANIFEST.MF")
	err = os.WriteFile(manifestPath, []byte(manifestContent), 0644)
	require.NoError(t, err)

	headers, err := GetManifestHeaders(manifestPath)
	require.NoError(t, err)

	// Continuation lines should be merged with spaces
	expected := "javax.xml.bind, javax.xml.stream, javax.xml.transform"
	assert.Equal(t, expected, headers["Import-Package"])
}

func TestGetManifestHeaders_NonExistent(t *testing.T) {
	headers, err := GetManifestHeaders("/nonexistent/MANIFEST.MF")
	require.NoError(t, err)
	assert.Empty(t, headers)
}

func TestGetManifestHeaders_Empty(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "utils-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	manifestPath := filepath.Join(tempDir, "MANIFEST.MF")
	err = os.WriteFile(manifestPath, []byte(""), 0644)
	require.NoError(t, err)

	headers, err := GetManifestHeaders(manifestPath)
	require.NoError(t, err)
	assert.Empty(t, headers)
}
