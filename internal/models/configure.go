package models

// ConfigureConfig represents the complete configuration file structure
type ConfigureConfig struct {
	DeploymentPrefix string             `yaml:"deploymentPrefix,omitempty"`
	Packages         []ConfigurePackage `yaml:"packages"`
}

// ConfigurePackage represents a package containing artifacts to configure
type ConfigurePackage struct {
	ID          string              `yaml:"integrationSuiteId"`
	DisplayName string              `yaml:"displayName,omitempty"`
	Deploy      bool                `yaml:"deploy"` // Deploy all artifacts in package after configuration
	Artifacts   []ConfigureArtifact `yaml:"artifacts"`
}

func (p *ConfigurePackage) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Set defaults
	type rawPackage ConfigurePackage
	raw := rawPackage{
		Deploy: false, // By default, don't deploy unless explicitly requested
	}

	if err := unmarshal(&raw); err != nil {
		return err
	}

	*p = ConfigurePackage(raw)
	return nil
}

// ConfigureArtifact represents an artifact with its configuration parameters
type ConfigureArtifact struct {
	ID          string                   `yaml:"artifactId"`
	DisplayName string                   `yaml:"displayName,omitempty"`
	Type        string                   `yaml:"type"`                 // Integration, MessageMapping, ScriptCollection, ValueMapping
	Version     string                   `yaml:"version,omitempty"`    // Artifact version, defaults to "active"
	Deploy      bool                     `yaml:"deploy"`               // Deploy this specific artifact after configuration
	Parameters  []ConfigurationParameter `yaml:"parameters,omitempty"` // List of configuration parameters to update
	Batch       *BatchSettings           `yaml:"batch,omitempty"`      // Optional batch processing settings
}

func (a *ConfigureArtifact) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Set defaults
	type rawArtifact ConfigureArtifact
	raw := rawArtifact{
		Version: "active",
		Deploy:  false, // By default, don't deploy unless explicitly requested
	}

	if err := unmarshal(&raw); err != nil {
		return err
	}

	*a = ConfigureArtifact(raw)
	return nil
}

// ConfigurationParameter represents a single configuration parameter to update
type ConfigurationParameter struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

// BatchSettings allows per-artifact batch configuration
type BatchSettings struct {
	Enabled   bool `yaml:"enabled"`             // Enable batch processing for this artifact
	BatchSize int  `yaml:"batchSize,omitempty"` // Number of parameters per batch request
}

func (b *BatchSettings) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Set defaults
	type rawBatch BatchSettings
	raw := rawBatch{
		Enabled:   true,
		BatchSize: 90, // Default batch size from batch.go
	}

	if err := unmarshal(&raw); err != nil {
		return err
	}

	*b = BatchSettings(raw)
	return nil
}
