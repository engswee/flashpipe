package designtime

import (
	"fmt"
	"github.com/engswee/flashpipe/httpclnt"
	"github.com/engswee/flashpipe/logger"
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
	logger.Info(fmt.Sprintf("Update of Value Mapping %v by executing delete followed by create", id))
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
func (vm *ValueMapping) GetVersion(id string, version string) (string, error) {
	return getVersion(id, version, vm.typ, vm.exe)
}
func (vm *ValueMapping) Exists(id string, version string) (bool, error) {
	return exists(id, version, vm.typ, vm.exe)
}
func (vm *ValueMapping) GetContent(id string, version string) ([]byte, error) {
	return getContent(id, version, vm.typ, vm.exe)
}
