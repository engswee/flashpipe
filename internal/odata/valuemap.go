package odata

import (
	"github.com/engswee/flashpipe/internal/file"
	"github.com/engswee/flashpipe/internal/httpclnt"
	"github.com/rs/zerolog/log"
)

type ValueMapping struct {
	exe *httpclnt.HTTPExecuter
	typ string
}

// NewIntegration returns an initialised Integration instance.
func NewValueMapping(exe *httpclnt.HTTPExecuter) DesigntimeArtifact {
	i := new(ValueMapping)
	i.exe = exe
	i.typ = "ValueMapping"
	return i
}

func (vm *ValueMapping) Create(id string, name string, packageId string, artifactDir string) error {
	return create(id, name, packageId, artifactDir, vm.typ, vm.exe)
}
func (vm *ValueMapping) Update(id string, name string, packageId string, artifactDir string) error {
	log.Info().Msgf("Update of Value Mapping %v by executing delete followed by create", id)
	err := deleteCall(id, vm.typ, vm.exe)
	if err != nil {
		return err
	}
	return create(id, name, packageId, artifactDir, vm.typ, vm.exe)
}
func (vm *ValueMapping) Deploy(id string) error {
	return deploy(id, vm.typ, vm.exe)
}
func (vm *ValueMapping) Delete(id string) error {
	return deleteCall(id, vm.typ, vm.exe)
}
func (vm *ValueMapping) Get(id string, version string) (string, bool, error) {
	return get(id, version, vm.typ, vm.exe)
}
func (vm *ValueMapping) Download(targetFile string, id string) error {
	return download(targetFile, id, vm.typ, vm.exe)
}
func (vm *ValueMapping) CopyContent(srcDir string, tgtDir string) error {
	// Copy META-INF and value_mapping.xml separately so that other directories like QA, STG, PRD not copied
	err := file.ReplaceDir(srcDir+"/META-INF", tgtDir+"/META-INF")
	if err != nil {
		return err
	}
	err = file.CopyFile(srcDir+"/value_mapping.xml", tgtDir+"/value_mapping.xml")
	if err != nil {
		return err
	}
	return nil
}
func (vm *ValueMapping) CompareContent(srcDir string, tgtDir string, _ string, _ string) (bool, error) {
	// Diff directories
	log.Info().Msg("Checking for changes in META-INF directory")
	metaDiffer := file.DiffDirectories(srcDir+"/META-INF", tgtDir+"/META-INF")
	log.Info().Msg("Checking for changes in value_mapping.xml")
	xmlDiffer := file.DiffFile(srcDir+"/value_mapping.xml", tgtDir+"/value_mapping.xml")

	return metaDiffer || xmlDiffer, nil
}
