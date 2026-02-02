package repo

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/engswee/flashpipe/internal/api"
	"github.com/rs/zerolog/log"
)

const (
	stringPropertiesFile = "String.properties"
	binaryDirName        = "Binary"
	metadataFileName     = "_metadata.json"
	defaultBinaryExt     = "bin"
)

// supportedContentTypes defines the valid content types that SAP CPI uses
// These are simple type strings (not MIME types)
var supportedContentTypes = map[string]bool{
	"xml":  true,
	"xsl":  true,
	"xsd":  true,
	"json": true,
	"txt":  true,
	"zip":  true,
	"gz":   true,
	"zlib": true,
	"crt":  true,
}

// PartnerDirectory handles Partner Directory file operations
type PartnerDirectory struct {
	ResourcesPath string
}

// NewPartnerDirectory creates a new Partner Directory repository
func NewPartnerDirectory(resourcesPath string) *PartnerDirectory {
	return &PartnerDirectory{
		ResourcesPath: resourcesPath,
	}
}

// GetLocalPIDs returns all PIDs that have local directories
func (pd *PartnerDirectory) GetLocalPIDs() ([]string, error) {
	entries, err := os.ReadDir(pd.ResourcesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read resources directory: %w", err)
	}

	var pids []string
	for _, entry := range entries {
		if entry.IsDir() {
			pids = append(pids, entry.Name())
		}
	}

	sort.Strings(pids)
	return pids, nil
}

// WriteStringParameters writes string parameters to a properties file
func (pd *PartnerDirectory) WriteStringParameters(pid string, params []api.StringParameter, replace bool) error {
	pidDir := filepath.Join(pd.ResourcesPath, pid)
	if err := os.MkdirAll(pidDir, 0755); err != nil {
		return fmt.Errorf("failed to create PID directory: %w", err)
	}

	propertiesFile := filepath.Join(pidDir, stringPropertiesFile)

	if replace || !fileExists(propertiesFile) {
		if err := writePropertiesFile(propertiesFile, params); err != nil {
			return err
		}
		log.Debug().Msgf("Created/Updated %s for PID %s", stringPropertiesFile, pid)
	} else {
		addedCount, err := mergePropertiesFile(propertiesFile, params)
		if err != nil {
			return err
		}
		log.Debug().Msgf("Merged %d new values into %s for PID %s", addedCount, stringPropertiesFile, pid)
	}

	return nil
}

// WriteBinaryParameters writes binary parameters to files
func (pd *PartnerDirectory) WriteBinaryParameters(pid string, params []api.BinaryParameter, replace bool) error {
	pidDir := filepath.Join(pd.ResourcesPath, pid)
	binaryDir := filepath.Join(pidDir, binaryDirName)

	if err := os.MkdirAll(binaryDir, 0755); err != nil {
		return fmt.Errorf("failed to create binary directory: %w", err)
	}

	for _, param := range params {
		filePath := filepath.Join(binaryDir, param.ID)

		// Check if file exists
		exists := fileExists(filePath)

		// Skip if not replacing and file exists
		if !replace && exists {
			log.Debug().Msgf("Skipping existing binary parameter %s/%s", pid, param.ID)
			continue
		}

		if err := saveBinaryParameterToFile(binaryDir, param); err != nil {
			return fmt.Errorf("failed to save binary parameter %s: %w", param.ID, err)
		}

		if err := updateMetadataFile(binaryDir, param.ID, param.ContentType); err != nil {
			return fmt.Errorf("failed to update metadata: %w", err)
		}
	}

	return nil
}

// ReadStringParameters reads string parameters from a properties file
func (pd *PartnerDirectory) ReadStringParameters(pid string) ([]api.StringParameter, error) {
	propertiesFile := filepath.Join(pd.ResourcesPath, pid, stringPropertiesFile)

	if !fileExists(propertiesFile) {
		return []api.StringParameter{}, nil
	}

	return readPropertiesFile(propertiesFile, pid)
}

// ReadBinaryParameters reads binary parameters from files
func (pd *PartnerDirectory) ReadBinaryParameters(pid string) ([]api.BinaryParameter, error) {
	binaryDir := filepath.Join(pd.ResourcesPath, pid, binaryDirName)

	if !dirExists(binaryDir) {
		return []api.BinaryParameter{}, nil
	}

	// Read metadata
	metadataPath := filepath.Join(binaryDir, metadataFileName)
	metadata := make(map[string]string)
	if fileExists(metadataPath) {
		data, err := os.ReadFile(metadataPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read metadata file: %w", err)
		}
		if err := json.Unmarshal(data, &metadata); err != nil {
			return nil, fmt.Errorf("failed to parse metadata: %w", err)
		}
	}

	// Read all binary files
	entries, err := os.ReadDir(binaryDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read binary directory: %w", err)
	}

	var params []api.BinaryParameter
	seenParams := make(map[string]bool)

	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == metadataFileName {
			continue
		}

		filePath := filepath.Join(binaryDir, entry.Name())

		// Use filename without extension as ID
		paramID := removeFileExtension(entry.Name())

		// Check for duplicates (same ID, different extension)
		if seenParams[paramID] {
			log.Warn().Msgf("Duplicate binary parameter %s/%s - skipping file %s", pid, paramID, entry.Name())
			continue
		}
		seenParams[paramID] = true

		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Warn().Msgf("Failed to read binary file %s: %v", entry.Name(), err)
			continue
		}

		// Encode to base64
		encoded := base64.StdEncoding.EncodeToString(data)

		// Get full content type from metadata (includes encoding if present)
		contentType := metadata[entry.Name()]
		if contentType == "" {
			// Infer from extension if not in metadata
			ext := strings.TrimPrefix(filepath.Ext(entry.Name()), ".")
			if ext == "" {
				ext = defaultBinaryExt
			}
			contentType = ext
		}

		log.Debug().Msgf("Loaded binary parameter %s/%s (%s, %d bytes)", pid, paramID, contentType, len(data))

		params = append(params, api.BinaryParameter{
			Pid:         pid,
			ID:          paramID,
			Value:       encoded,
			ContentType: contentType,
		})
	}

	return params, nil
}

// Helper functions

func writePropertiesFile(filePath string, params []api.StringParameter) error {
	// Sort by ID for consistent output
	sort.Slice(params, func(i, j int) bool {
		return params[i].ID < params[j].ID
	})

	var content strings.Builder
	for _, param := range params {
		content.WriteString(fmt.Sprintf("%s=%s\n", param.ID, escapePropertyValue(param.Value)))
	}

	if err := os.WriteFile(filePath, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to write properties file: %w", err)
	}

	return nil
}

func mergePropertiesFile(filePath string, newParams []api.StringParameter) (int, error) {
	// Read existing properties
	existing := make(map[string]string)
	if fileExists(filePath) {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return 0, fmt.Errorf("failed to read existing properties: %w", err)
		}

		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				existing[parts[0]] = parts[1]
			}
		}
	}

	// Add new parameters
	addedCount := 0
	for _, param := range newParams {
		if _, exists := existing[param.ID]; !exists {
			existing[param.ID] = escapePropertyValue(param.Value)
			addedCount++
		}
	}

	// Write back sorted
	keys := make([]string, 0, len(existing))
	for k := range existing {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var content strings.Builder
	for _, key := range keys {
		content.WriteString(fmt.Sprintf("%s=%s\n", key, existing[key]))
	}

	if err := os.WriteFile(filePath, []byte(content.String()), 0644); err != nil {
		return 0, fmt.Errorf("failed to write properties file: %w", err)
	}

	return addedCount, nil
}

func readPropertiesFile(filePath string, pid string) ([]api.StringParameter, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read properties file: %w", err)
	}

	var params []api.StringParameter
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			params = append(params, api.StringParameter{
				Pid:   pid,
				ID:    parts[0],
				Value: unescapePropertyValue(parts[1]),
			})
		}
	}

	return params, nil
}

func saveBinaryParameterToFile(binaryDir string, param api.BinaryParameter) error {
	// Decode base64
	data, err := base64.StdEncoding.DecodeString(param.Value)
	if err != nil {
		return fmt.Errorf("failed to decode base64: %w", err)
	}

	// Determine file extension from content type
	log.Debug().Msgf("Processing binary parameter %s with contentType: %s", param.ID, param.ContentType)
	ext := getFileExtension(param.ContentType)
	log.Debug().Msgf("Determined file extension: %s", ext)

	// Create filename: {ParamId}.{ext}
	filename := param.ID
	if ext != "" && !strings.HasSuffix(strings.ToLower(filename), "."+ext) {
		filename = fmt.Sprintf("%s.%s", param.ID, ext)
	}

	filePath := filepath.Join(binaryDir, filename)

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write binary file: %w", err)
	}

	log.Info().Msgf("Saved binary parameter: %s (%s, %d bytes)", filename, param.ContentType, len(data))
	return nil
}

func updateMetadataFile(binaryDir string, paramID string, contentType string) error {
	// Only store in metadata if contentType has encoding/parameters (contains semicolon)
	if !strings.Contains(contentType, ";") {
		return nil
	}

	metadataPath := filepath.Join(binaryDir, metadataFileName)

	metadata := make(map[string]string)
	if fileExists(metadataPath) {
		data, err := os.ReadFile(metadataPath)
		if err != nil {
			return fmt.Errorf("failed to read metadata: %w", err)
		}
		if err := json.Unmarshal(data, &metadata); err != nil {
			return fmt.Errorf("failed to parse metadata: %w", err)
		}
	}

	// Determine filename
	ext := getFileExtension(contentType)
	filename := paramID
	if ext != "" && !strings.HasSuffix(strings.ToLower(filename), "."+ext) {
		filename = fmt.Sprintf("%s.%s", paramID, ext)
	}

	// Store full content type (with encoding)
	metadata[filename] = contentType

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

func parseContentType(contentType string) (string, string) {
	// SAP CPI returns simple types like "xml", "json", "txt"
	// But may also include encoding like "xml; encoding=UTF-8"

	// Remove encoding/parameters (e.g., "xml; encoding=UTF-8" -> "xml")
	baseType := contentType
	if idx := strings.Index(contentType, ";"); idx > 0 {
		baseType = strings.TrimSpace(contentType[:idx])
	}

	// If it's a MIME type like "text/xml" or "application/json", extract the subtype
	if strings.Contains(baseType, "/") {
		parts := strings.Split(baseType, "/")
		if len(parts) == 2 {
			ext := parts[1]
			// Handle special cases like "application/octet-stream"
			if ext == "octet-stream" {
				return defaultBinaryExt, contentType
			}
			return ext, contentType
		}
	}

	// Otherwise it's already a simple type like "xml", "json", etc.
	return baseType, contentType
}

func escapePropertyValue(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\n", "\\n")
	value = strings.ReplaceAll(value, "\r", "\\r")
	return value
}

func unescapePropertyValue(value string) string {
	value = strings.ReplaceAll(value, "\\n", "\n")
	value = strings.ReplaceAll(value, "\\r", "\r")
	value = strings.ReplaceAll(value, "\\\\", "\\")
	return value
}

func getFileExtension(contentType string) string {
	ext, _ := parseContentType(contentType)
	// Use the extension if it's in our supported list or if it's reasonable
	if isValidContentType(ext) {
		return ext
	}
	// If not in supported list but looks valid (alphanumeric, 2-5 chars), still use it
	if ext != "" && len(ext) >= 2 && len(ext) <= 5 && isAlphanumeric(ext) {
		log.Debug().Msgf("Using non-standard extension: %s", ext)
		return ext
	}
	return defaultBinaryExt
}

func isAlphanumeric(s string) bool {
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}

func removeFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	if ext != "" {
		return strings.TrimSuffix(filename, ext)
	}
	return filename
}

func isValidContentType(ext string) bool {
	return supportedContentTypes[strings.ToLower(ext)]
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil && info.IsDir()
}
