package designtime

import (
	"github.com/engswee/flashpipe/httpclnt"
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
func (int *Integration) GetVersion(id string, version string) (string, error) {
	return getVersion(id, version, int.typ, int.exe)
}
func (int *Integration) Exists(id string, version string) (bool, error) {
	return exists(id, version, int.typ, int.exe)
}
func (int *Integration) GetContent(id string, version string) ([]byte, error) {
	return getContent(id, version, int.typ, int.exe)
}
