package models

// OrchestratorConfig represents orchestrator-specific settings
type OrchestratorConfig struct {
	PackagesDir      string `yaml:"packagesDir"`
	DeployConfig     string `yaml:"deployConfig"`
	DeploymentPrefix string `yaml:"deploymentPrefix,omitempty"`
	PackageFilter    string `yaml:"packageFilter,omitempty"`
	ArtifactFilter   string `yaml:"artifactFilter,omitempty"`
	ConfigPattern    string `yaml:"configPattern,omitempty"`
	MergeConfigs     bool   `yaml:"mergeConfigs,omitempty"`
	KeepTemp         bool   `yaml:"keepTemp,omitempty"`
	Mode             string `yaml:"mode,omitempty"` // "update-and-deploy", "update-only", "deploy-only"
	// Deployment settings
	DeployRetries       int `yaml:"deployRetries,omitempty"`
	DeployDelaySeconds  int `yaml:"deployDelaySeconds,omitempty"`
	ParallelDeployments int `yaml:"parallelDeployments,omitempty"`
}

// DeployConfig represents the complete deployment configuration
type DeployConfig struct {
	DeploymentPrefix string              `yaml:"deploymentPrefix"`
	Packages         []Package           `yaml:"packages"`
	Orchestrator     *OrchestratorConfig `yaml:"orchestrator,omitempty"`
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

func (p *Package) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Set defaults
	type rawPackage Package
	raw := rawPackage{
		Sync:   true,
		Deploy: true,
	}

	if err := unmarshal(&raw); err != nil {
		return err
	}

	*p = Package(raw)
	return nil
}

// Artifact represents a SAP CPI artifact (Integration Flow, Script Collection, etc.)
type Artifact struct {
	Id              string                 `yaml:"artifactId"`
	ArtifactDir     string                 `yaml:"artifactDir"`
	DisplayName     string                 `yaml:"displayName"`
	Type            string                 `yaml:"type"`
	Sync            bool                   `yaml:"sync"`
	Deploy          bool                   `yaml:"deploy"`
	ConfigOverrides map[string]interface{} `yaml:"configOverrides"`
}

func (a *Artifact) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Set defaults
	type rawArtifact Artifact
	raw := rawArtifact{
		Sync:   true,
		Deploy: true,
	}

	if err := unmarshal(&raw); err != nil {
		return err
	}

	*a = Artifact(raw)
	return nil
}

// PackageMetadata represents metadata extracted from {PackageName}.json
type PackageMetadata struct {
	ID          string `json:"Id"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
	ShortText   string `json:"ShortText"`
}

// PackageJSON represents the structure of {PackageName}.json files
type PackageJSON struct {
	D PackageMetadata `json:"d"`
}
