package repo

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/engswee/flashpipe/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseContentType_SimpleTypes(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		wantExt     string
		wantFull    string
	}{
		{
			name:        "simple xml",
			contentType: "xml",
			wantExt:     "xml",
			wantFull:    "xml",
		},
		{
			name:        "simple json",
			contentType: "json",
			wantExt:     "json",
			wantFull:    "json",
		},
		{
			name:        "simple txt",
			contentType: "txt",
			wantExt:     "txt",
			wantFull:    "txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext, full := parseContentType(tt.contentType)
			assert.Equal(t, tt.wantExt, ext)
			assert.Equal(t, tt.wantFull, full)
		})
	}
}

func TestParseContentType_WithEncoding(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		wantExt     string
		wantFull    string
	}{
		{
			name:        "xml with encoding",
			contentType: "xml; encoding=UTF-8",
			wantExt:     "xml",
			wantFull:    "xml; encoding=UTF-8",
		},
		{
			name:        "json with charset",
			contentType: "json; charset=utf-8",
			wantExt:     "json",
			wantFull:    "json; charset=utf-8",
		},
		{
			name:        "xml with multiple parameters",
			contentType: "xml; encoding=UTF-8; version=1.0",
			wantExt:     "xml",
			wantFull:    "xml; encoding=UTF-8; version=1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext, full := parseContentType(tt.contentType)
			assert.Equal(t, tt.wantExt, ext)
			assert.Equal(t, tt.wantFull, full)
		})
	}
}

func TestParseContentType_MIMETypes(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		wantExt     string
	}{
		{
			name:        "text/xml",
			contentType: "text/xml",
			wantExt:     "xml",
		},
		{
			name:        "application/json",
			contentType: "application/json",
			wantExt:     "json",
		},
		{
			name:        "application/xml",
			contentType: "application/xml",
			wantExt:     "xml",
		},
		{
			name:        "text/plain",
			contentType: "text/plain",
			wantExt:     "plain",
		},
		{
			name:        "application/octet-stream",
			contentType: "application/octet-stream",
			wantExt:     defaultBinaryExt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext, _ := parseContentType(tt.contentType)
			assert.Equal(t, tt.wantExt, ext)
		})
	}
}

func TestGetFileExtension_SupportedTypes(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		wantExt     string
	}{
		{"xml", "xml", "xml"},
		{"json", "json", "json"},
		{"xsl", "xsl", "xsl"},
		{"xsd", "xsd", "xsd"},
		{"txt", "txt", "txt"},
		{"zip", "zip", "zip"},
		{"crt", "crt", "crt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext := getFileExtension(tt.contentType)
			assert.Equal(t, tt.wantExt, ext)
		})
	}
}

func TestGetFileExtension_UnsupportedTypes(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		wantExt     string
	}{
		{
			name:        "unknown simple type",
			contentType: "unknown",
			wantExt:     defaultBinaryExt,
		},
		{
			name:        "empty",
			contentType: "",
			wantExt:     defaultBinaryExt,
		},
		{
			name:        "too long",
			contentType: "verylongextension",
			wantExt:     defaultBinaryExt,
		},
		{
			name:        "special characters",
			contentType: "xml$%",
			wantExt:     defaultBinaryExt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext := getFileExtension(tt.contentType)
			assert.Equal(t, tt.wantExt, ext)
		})
	}
}

func TestGetFileExtension_CustomValidTypes(t *testing.T) {
	// Non-standard but valid alphanumeric extensions (2-5 chars)
	tests := []struct {
		name        string
		contentType string
		wantExt     string
	}{
		{"pdf", "pdf", "pdf"},
		{"docx", "docx", "docx"},
		{"html", "html", "html"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext := getFileExtension(tt.contentType)
			assert.Equal(t, tt.wantExt, ext)
		})
	}
}

func TestEscapeUnescapePropertyValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple value",
			input:    "simple",
			expected: "simple",
		},
		{
			name:     "with newline",
			input:    "line1\nline2",
			expected: "line1\\nline2",
		},
		{
			name:     "with carriage return",
			input:    "line1\rline2",
			expected: "line1\\rline2",
		},
		{
			name:     "with backslash",
			input:    "path\\to\\file",
			expected: "path\\\\to\\\\file",
		},
		{
			name:     "with all special chars",
			input:    "line1\nline2\rline3\\backslash",
			expected: "line1\\nline2\\rline3\\\\backslash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+" escape", func(t *testing.T) {
			escaped := escapePropertyValue(tt.input)
			assert.Equal(t, tt.expected, escaped)
		})

		t.Run(tt.name+" unescape", func(t *testing.T) {
			unescaped := unescapePropertyValue(tt.expected)
			assert.Equal(t, tt.input, unescaped)
		})

		t.Run(tt.name+" roundtrip", func(t *testing.T) {
			roundtrip := unescapePropertyValue(escapePropertyValue(tt.input))
			assert.Equal(t, tt.input, roundtrip)
		})
	}
}

func TestRemoveFileExtension(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{"with extension", "file.xml", "file"},
		{"with multiple dots", "file.backup.xml", "file.backup"},
		{"no extension", "file", "file"},
		{"hidden file", ".gitignore", ""},
		{"multiple extensions", "archive.tar.gz", "archive.tar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeFileExtension(tt.filename)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestIsAlphanumeric(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"letters only", "xml", true},
		{"mixed case", "XmL", true},
		{"with numbers", "file123", true},
		{"with dash", "file-name", false},
		{"with underscore", "file_name", false},
		{"with dot", "file.ext", false},
		{"with space", "file name", false},
		{"empty", "", true},
		{"special chars", "file$", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAlphanumeric(tt.input)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestWriteAndReadStringParameters(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "pd-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	pd := NewPartnerDirectory(tempDir)
	pid := "TestPID"

	params := []api.StringParameter{
		{Pid: pid, ID: "param1", Value: "value1"},
		{Pid: pid, ID: "param2", Value: "value with\nnewline"},
		{Pid: pid, ID: "param3", Value: "value\\with\\backslash"},
	}

	// Write parameters
	err = pd.WriteStringParameters(pid, params, true)
	require.NoError(t, err)

	// Read parameters back
	readParams, err := pd.ReadStringParameters(pid)
	require.NoError(t, err)

	// Verify
	assert.Equal(t, len(params), len(readParams))
	for i, param := range params {
		assert.Equal(t, param.ID, readParams[i].ID)
		assert.Equal(t, param.Value, readParams[i].Value)
		assert.Equal(t, pid, readParams[i].Pid)
	}
}

func TestWriteStringParameters_MergeMode(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pd-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	pd := NewPartnerDirectory(tempDir)
	pid := "TestPID"

	// Write initial parameters
	initial := []api.StringParameter{
		{Pid: pid, ID: "param1", Value: "value1"},
		{Pid: pid, ID: "param2", Value: "value2"},
	}
	err = pd.WriteStringParameters(pid, initial, true)
	require.NoError(t, err)

	// Merge new parameters (replace=false)
	additional := []api.StringParameter{
		{Pid: pid, ID: "param3", Value: "value3"},
		{Pid: pid, ID: "param1", Value: "updated_value1"}, // Should be ignored
	}
	err = pd.WriteStringParameters(pid, additional, false)
	require.NoError(t, err)

	// Read back
	readParams, err := pd.ReadStringParameters(pid)
	require.NoError(t, err)

	// Verify merge behavior
	assert.Equal(t, 3, len(readParams))

	paramMap := make(map[string]string)
	for _, p := range readParams {
		paramMap[p.ID] = p.Value
	}

	assert.Equal(t, "value1", paramMap["param1"]) // Original should be preserved
	assert.Equal(t, "value2", paramMap["param2"])
	assert.Equal(t, "value3", paramMap["param3"]) // New param should be added
}

func TestWriteAndReadBinaryParameters(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pd-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	pd := NewPartnerDirectory(tempDir)
	pid := "TestPID"

	testData := []byte("<?xml version=\"1.0\"?><root>test</root>")
	encoded := base64.StdEncoding.EncodeToString(testData)

	params := []api.BinaryParameter{
		{Pid: pid, ID: "config", Value: encoded, ContentType: "xml"},
		{Pid: pid, ID: "schema", Value: encoded, ContentType: "xsd"},
	}

	// Write parameters
	err = pd.WriteBinaryParameters(pid, params, true)
	require.NoError(t, err)

	// Verify files exist
	configFile := filepath.Join(tempDir, pid, "Binary", "config.xml")
	schemaFile := filepath.Join(tempDir, pid, "Binary", "schema.xsd")
	assert.True(t, fileExists(configFile))
	assert.True(t, fileExists(schemaFile))

	// Read parameters back
	readParams, err := pd.ReadBinaryParameters(pid)
	require.NoError(t, err)

	// Verify
	assert.Equal(t, 2, len(readParams))

	paramMap := make(map[string]api.BinaryParameter)
	for _, p := range readParams {
		paramMap[p.ID] = p
	}

	assert.Equal(t, "xml", paramMap["config"].ContentType)
	assert.Equal(t, "xsd", paramMap["schema"].ContentType)
	assert.Equal(t, encoded, paramMap["config"].Value)
	assert.Equal(t, encoded, paramMap["schema"].Value)
}

func TestBinaryParameterWithEncoding(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pd-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	pd := NewPartnerDirectory(tempDir)
	pid := "TestPID"

	testData := []byte("<?xml version=\"1.0\"?><root>test</root>")
	encoded := base64.StdEncoding.EncodeToString(testData)

	params := []api.BinaryParameter{
		{Pid: pid, ID: "config", Value: encoded, ContentType: "xml; encoding=UTF-8"},
	}

	// Write parameter
	err = pd.WriteBinaryParameters(pid, params, true)
	require.NoError(t, err)

	// Verify metadata file was created
	metadataFile := filepath.Join(tempDir, pid, "Binary", metadataFileName)
	assert.True(t, fileExists(metadataFile))

	// Read metadata
	metadataBytes, err := os.ReadFile(metadataFile)
	require.NoError(t, err)

	var metadata map[string]string
	err = json.Unmarshal(metadataBytes, &metadata)
	require.NoError(t, err)

	// Verify metadata contains full content type
	assert.Equal(t, "xml; encoding=UTF-8", metadata["config.xml"])

	// Read parameter back
	readParams, err := pd.ReadBinaryParameters(pid)
	require.NoError(t, err)

	assert.Equal(t, 1, len(readParams))
	assert.Equal(t, "xml; encoding=UTF-8", readParams[0].ContentType)
	assert.Equal(t, encoded, readParams[0].Value)
}

func TestBinaryParameterWithoutEncoding_NoMetadata(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pd-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	pd := NewPartnerDirectory(tempDir)
	pid := "TestPID"

	testData := []byte("{\"key\": \"value\"}")
	encoded := base64.StdEncoding.EncodeToString(testData)

	params := []api.BinaryParameter{
		{Pid: pid, ID: "config", Value: encoded, ContentType: "json"},
	}

	// Write parameter
	err = pd.WriteBinaryParameters(pid, params, true)
	require.NoError(t, err)

	// Verify metadata file was NOT created (since no encoding)
	metadataFile := filepath.Join(tempDir, pid, "Binary", metadataFileName)
	assert.False(t, fileExists(metadataFile))
}

func TestGetLocalPIDs(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pd-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	pd := NewPartnerDirectory(tempDir)

	// Create some PID directories
	pids := []string{"PID001", "PID002", "ZZTEST"}
	for _, pid := range pids {
		err := os.MkdirAll(filepath.Join(tempDir, pid), 0755)
		require.NoError(t, err)
	}

	// Create a file (should be ignored)
	err = os.WriteFile(filepath.Join(tempDir, "notapid.txt"), []byte("test"), 0644)
	require.NoError(t, err)

	// Get local PIDs
	localPIDs, err := pd.GetLocalPIDs()
	require.NoError(t, err)

	// Verify PIDs are returned sorted
	assert.Equal(t, []string{"PID001", "PID002", "ZZTEST"}, localPIDs)
}

func TestGetLocalPIDs_EmptyDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pd-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	pd := NewPartnerDirectory(tempDir)

	localPIDs, err := pd.GetLocalPIDs()
	require.NoError(t, err)
	assert.Empty(t, localPIDs)
}

func TestGetLocalPIDs_NonExistentDirectory(t *testing.T) {
	pd := NewPartnerDirectory("/nonexistent/path")

	localPIDs, err := pd.GetLocalPIDs()
	require.NoError(t, err)
	assert.Empty(t, localPIDs)
}

func TestReadStringParameters_NonExistent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pd-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	pd := NewPartnerDirectory(tempDir)

	params, err := pd.ReadStringParameters("NonExistentPID")
	require.NoError(t, err)
	assert.Empty(t, params)
}

func TestReadBinaryParameters_NonExistent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pd-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	pd := NewPartnerDirectory(tempDir)

	params, err := pd.ReadBinaryParameters("NonExistentPID")
	require.NoError(t, err)
	assert.Empty(t, params)
}

func TestBinaryParameters_DuplicateHandling(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pd-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	pid := "TestPID"
	binaryDir := filepath.Join(tempDir, pid, "Binary")
	err = os.MkdirAll(binaryDir, 0755)
	require.NoError(t, err)

	// Create duplicate files with different extensions but same base name
	testData := []byte("test data")
	err = os.WriteFile(filepath.Join(binaryDir, "config.xml"), testData, 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(binaryDir, "config.txt"), testData, 0644)
	require.NoError(t, err)

	pd := NewPartnerDirectory(tempDir)

	// Read should handle duplicates (only return one)
	params, err := pd.ReadBinaryParameters(pid)
	require.NoError(t, err)

	// Should only get one parameter (the first one encountered)
	assert.Equal(t, 1, len(params))
	assert.Equal(t, "config", params[0].ID)
}

func TestWriteStringParameters_Sorted(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pd-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	pd := NewPartnerDirectory(tempDir)
	pid := "TestPID"

	// Write parameters in random order
	params := []api.StringParameter{
		{Pid: pid, ID: "zzz", Value: "last"},
		{Pid: pid, ID: "aaa", Value: "first"},
		{Pid: pid, ID: "mmm", Value: "middle"},
	}

	err = pd.WriteStringParameters(pid, params, true)
	require.NoError(t, err)

	// Read file content
	propertiesFile := filepath.Join(tempDir, pid, stringPropertiesFile)
	content, err := os.ReadFile(propertiesFile)
	require.NoError(t, err)

	// Verify alphabetical order
	lines := string(content)
	assert.Contains(t, lines, "aaa=first")
	assert.Contains(t, lines, "mmm=middle")
	assert.Contains(t, lines, "zzz=last")

	// First occurrence should be 'aaa'
	assert.True(t, func() bool {
		aaaIndex := -1
		mmmIndex := -1
		zzzIndex := -1
		for i, line := range []string{"aaa=first", "mmm=middle", "zzz=last"} {
			idx := indexOf(lines, line)
			if i == 0 {
				aaaIndex = idx
			} else if i == 1 {
				mmmIndex = idx
			} else {
				zzzIndex = idx
			}
		}
		return aaaIndex < mmmIndex && mmmIndex < zzzIndex
	}())
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestFileExists(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pd-test-*")
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

	assert.True(t, fileExists(testFile))
	assert.False(t, fileExists(testDir)) // Directory should return false
	assert.False(t, fileExists(filepath.Join(tempDir, "nonexistent.txt")))
}

func TestDirExists(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pd-test-*")
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

	assert.True(t, dirExists(testDir))
	assert.False(t, dirExists(testFile)) // File should return false
	assert.False(t, dirExists(filepath.Join(tempDir, "nonexistent")))
}

func TestIsValidContentType(t *testing.T) {
	tests := []struct {
		ext   string
		valid bool
	}{
		{"xml", true},
		{"json", true},
		{"xsl", true},
		{"xsd", true},
		{"txt", true},
		{"zip", true},
		{"gz", true},
		{"zlib", true},
		{"crt", true},
		{"unknown", false},
		{"pdf", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			result := isValidContentType(tt.ext)
			assert.Equal(t, tt.valid, result)
		})
	}
}
