package odata

import (
	"fmt"
	"github.com/engswee/flashpipe/internal/file"
	"github.com/engswee/flashpipe/internal/httpclnt"
	"github.com/rs/zerolog/log"
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
func (int *Integration) Get(id string, version string) (string, bool, error) {
	return get(id, version, int.typ, int.exe)
}
func (int *Integration) Download(targetFile string, id string) error {
	return download(targetFile, id, int.typ, int.exe)
}
func (int *Integration) CopyContent(srcDir string, tgtDir string) error {
	return copyContent(srcDir, tgtDir)
}
func (int *Integration) CompareContent(srcDir string, tgtDir string, scriptMap string, source string) (bool, error) {
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
	if source == "remote" {
		// When syncing (from tenant to Git), include diff of parameter.prop separately
		paramDiffer := int.diffParam(srcDir, tgtDir)
		return dirDiffer || paramDiffer, nil
	} else {
		// When uploading (from Git to tenant), API is used to update the configuration parameters separately
		return dirDiffer, nil
	}
}
func (int *Integration) diffParam(srcDir string, tgtDir string) bool {
	downloadedParams := fmt.Sprintf("%v/src/main/resources/parameters.prop", srcDir)
	gitParams := fmt.Sprintf("%v/src/main/resources/parameters.prop", tgtDir)
	if file.Exists(downloadedParams) && file.Exists(gitParams) {
		return file.DiffFile(downloadedParams, gitParams)
	} else if !file.Exists(downloadedParams) && !file.Exists(gitParams) {
		log.Warn().Msg("Skipping diff of parameters.prop as it does not exist in both source and target")
		return false
	}
	log.Info().Msg("File parameters.prop does not exist in either source or target")
	return true
}
