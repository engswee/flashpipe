package deploy

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// FileExists checks if a file exists
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil && !info.IsDir()
}

// DirExists checks if a directory exists
func DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// ValidateDeploymentPrefix validates that the deployment prefix only contains allowed characters
func ValidateDeploymentPrefix(prefix string) error {
	if prefix == "" {
		return nil // Empty prefix is valid
	}

	// Only allow alphanumeric and underscores
	matched, err := regexp.MatchString("^[a-zA-Z0-9_]+$", prefix)
	if err != nil {
		return fmt.Errorf("regex error: %w", err)
	}

	if !matched {
		return fmt.Errorf("deployment prefix can only contain alphanumeric characters (a-z, A-Z, 0-9) and underscores (_)")
	}

	return nil
}

// CopyDir recursively copies a directory
func CopyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		// Copy file
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(targetPath, data, info.Mode())
	})
}

// UpdateManifestBundleName updates the Bundle-Name and Bundle-SymbolicName in MANIFEST.MF
func UpdateManifestBundleName(manifestPath, bundleSymbolicName, bundleName, outputPath string) error {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read MANIFEST.MF: %w", err)
	}

	// Detect line ending style (CRLF or LF)
	lineEnding := "\n"
	if strings.Contains(string(data), "\r\n") {
		lineEnding = "\r\n"
	}

	// Split lines
	content := string(data)
	lines := strings.Split(content, lineEnding)

	var result []string
	bundleNameFound := false
	bundleSymbolicNameFound := false

	for _, line := range lines {
		trimmedLower := strings.ToLower(strings.TrimSpace(line))

		if strings.HasPrefix(trimmedLower, "bundle-name:") {
			result = append(result, fmt.Sprintf("Bundle-Name: %s", bundleName))
			bundleNameFound = true
		} else if strings.HasPrefix(trimmedLower, "bundle-symbolicname:") {
			result = append(result, fmt.Sprintf("Bundle-SymbolicName: %s", bundleSymbolicName))
			bundleSymbolicNameFound = true
		} else {
			result = append(result, line)
		}
	}

	// Add Bundle-Name if not found
	if !bundleNameFound {
		result = append(result, fmt.Sprintf("Bundle-Name: %s", bundleName))
	}

	// Add Bundle-SymbolicName if not found
	if !bundleSymbolicNameFound {
		result = append(result, fmt.Sprintf("Bundle-SymbolicName: %s", bundleSymbolicName))
	}

	// Write to output path with original line endings and ensure final newline
	finalContent := strings.Join(result, lineEnding)
	if !strings.HasSuffix(finalContent, lineEnding) {
		finalContent += lineEnding
	}

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(finalContent), 0644); err != nil {
		return fmt.Errorf("failed to write MANIFEST.MF: %w", err)
	}

	return nil
}

// MergeParametersFile reads parameters.prop, applies overrides, and writes to outputPath
func MergeParametersFile(paramsPath string, overrides map[string]interface{}, outputPath string) error {
	var lineEnding string = "\n"
	params := make(map[string]string)
	paramKeys := []string{} // Track order of keys

	// Read existing file if it exists
	if FileExists(paramsPath) {
		data, err := os.ReadFile(paramsPath)
		if err != nil {
			return fmt.Errorf("failed to read parameters.prop: %w", err)
		}

		// Detect line ending style
		content := string(data)
		if strings.Contains(content, "\r\n") {
			lineEnding = "\r\n"
		}

		// Split and process lines
		lines := strings.Split(content, lineEnding)

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)

			// Keep comments and empty lines as-is
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}

			// Parse key=value
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				params[key] = value
				paramKeys = append(paramKeys, key)
			}
		}
	}

	// Apply overrides
	for key, value := range overrides {
		valStr := fmt.Sprintf("%v", value)
		if _, exists := params[key]; !exists {
			// New key, add to order
			paramKeys = append(paramKeys, key)
		}
		params[key] = valStr
	}

	// Write back with preserved order
	var result []string
	for _, key := range paramKeys {
		result = append(result, fmt.Sprintf("%s=%s", key, params[key]))
	}

	// Join with original line endings and ensure final newline
	finalContent := strings.Join(result, lineEnding)
	if !strings.HasSuffix(finalContent, lineEnding) {
		finalContent += lineEnding
	}

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(finalContent), 0644); err != nil {
		return fmt.Errorf("failed to write parameters.prop: %w", err)
	}

	return nil
}

// FindParametersFile finds parameters.prop in various possible locations
func FindParametersFile(artifactDir string) string {
	possiblePaths := []string{
		filepath.Join(artifactDir, "src", "main", "resources", "parameters.prop"),
		filepath.Join(artifactDir, "src", "main", "resources", "script", "parameters.prop"),
		filepath.Join(artifactDir, "parameters.prop"),
	}

	for _, path := range possiblePaths {
		if FileExists(path) {
			return path
		}
	}

	// Return default path even if it doesn't exist
	return possiblePaths[0]
}

// GetManifestHeaders reads headers from MANIFEST.MF file
func GetManifestHeaders(manifestPath string) (map[string]string, error) {
	metadata := make(map[string]string)

	if !FileExists(manifestPath) {
		return metadata, nil
	}

	file, err := os.Open(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open MANIFEST.MF: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentKey string
	var currentValue strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.Contains(trimmed, ":") && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			// New key-value pair
			if currentKey != "" {
				metadata[currentKey] = strings.TrimSpace(currentValue.String())
			}
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				currentKey = strings.TrimSpace(parts[0])
				currentValue.Reset()
				currentValue.WriteString(strings.TrimSpace(parts[1]))
			}
		} else if currentKey != "" && (strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")) {
			// Continuation line
			currentValue.WriteString(" ")
			currentValue.WriteString(strings.TrimSpace(line))
		}
	}

	// Add the last entry
	if currentKey != "" {
		metadata[currentKey] = strings.TrimSpace(currentValue.String())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read MANIFEST.MF: %w", err)
	}

	return metadata, nil
}
