package odata

import (
	"github.com/engswee/flashpipe/file"
	"github.com/engswee/flashpipe/httpclnt"
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
func (vm *ValueMapping) GetContent(id string, version string) ([]byte, error) {
	return getContent(id, version, vm.typ, vm.exe)
}
func (vm *ValueMapping) DiffContent(firstDir string, secondDir string) bool {
	log.Info().Msg("Checking for changes in META-INF directory")
	metaDiffer := file.DiffDirectories(firstDir+"/META-INF", secondDir+"/META-INF")
	log.Info().Msg("Checking for changes in value_mapping.xml")
	xmlDiffer := file.DiffParams(firstDir+"/value_mapping.xml", secondDir+"/value_mapping.xml")
	if metaDiffer || xmlDiffer {
		return true
	} else {
		return false
	}
}
func (vm *ValueMapping) CopyContent(srcDir string, tgtDir string) error {
	// Copy META-INF and value_mapping.xml separately so that other directories like QA, STG, PRD not copied
	err := file.CopyDir(srcDir+"/META-INF", tgtDir+"/META-INF")
	if err != nil {
		return err
	}
	err = file.CopyFile(srcDir+"/value_mapping.xml", tgtDir+"/value_mapping.xml")
	if err != nil {
		return err
	}
	return nil
}
