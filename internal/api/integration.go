package api

import (
	"github.com/engswee/flashpipe/internal/file"
	"github.com/engswee/flashpipe/internal/httpclnt"
)

type Integration struct {
	exe *httpclnt.HTTPExecuter
	typ string
}

// NewIntegration returns an initialised Integration instance.
func NewIntegration(exe *httpclnt.HTTPExecuter) DesigntimeArtifact {
	i := new(Integration)
	i.exe = exe
	i.typ = "Integration"
	return i
}

func (int *Integration) Create(id string, name string, packageId string, artifactDir string) error {
	return create(id, name, packageId, artifactDir, int.typ, int.exe)
}
func (int *Integration) Update(id string, name string, packageId string, artifactDir string) error {
	return update(id, name, packageId, artifactDir, int.typ, int.exe)
}
func (int *Integration) Deploy(id string) error {
	return deploy(id, int.typ, int.exe)
}
func (int *Integration) Delete(id string) error {
	return deleteCall(id, int.typ, int.exe)
}
func (int *Integration) Get(id string, version string) (string, string, bool, error) {
	return get(id, version, int.typ, int.exe)
}
func (int *Integration) Download(targetFile string, id string) error {
	return download(targetFile, id, int.typ, int.exe)
}
func (int *Integration) CopyContent(srcDir string, tgtDir string) error {
	return copyContent(srcDir, tgtDir)
}
func (int *Integration) CompareContent(srcDir string, tgtDir string, scriptMap []string, target string) (bool, error) {
	// Update the script collection in IFlow BPMN2 XML of source side before diff comparison
	err := file.UpdateBPMN(srcDir, scriptMap)
	if err != nil {
		return false, err
	}

	// Diff directories excluding parameters.prop
	dirDiffer := diffContent(srcDir, tgtDir)

	// Handling for parameters.prop differences
	// - Any configured value will remain in IFlow even if the IFlow is replaced and the parameter is no longer used
	// - Therefore diff of parameters.prop may come up with false differences
	if target == "git" {
		// When syncing (from tenant to Git), include diff of parameter.prop separately
		paramDiffer := DiffOptionalFile(srcDir, tgtDir, "src/main/resources/parameters.prop")
		return dirDiffer || paramDiffer, nil
	} else {
		// When uploading (from Git to tenant), API is used to update the configuration parameters separately
		return dirDiffer, nil
	}
}
