package odata

import (
	"github.com/engswee/flashpipe/httpclnt"
)

type MessageMapping struct {
	exe *httpclnt.HTTPExecuter
	typ string
}

// NewMessageMapping returns an initialised MessageMapping instance.
func NewMessageMapping(exe *httpclnt.HTTPExecuter) DesigntimeArtifact {
	mm := new(MessageMapping)
	mm.exe = exe
	mm.typ = "MessageMapping"
	return mm
}

func (mm *MessageMapping) Create(id string, name string, packageId string, artifactDir string) error {
	return create(id, name, packageId, artifactDir, mm.typ, mm.exe)
}
func (mm *MessageMapping) Update(id string, name string, packageId string, artifactDir string) (err error) {
	return update(id, name, packageId, artifactDir, mm.typ, mm.exe)
}
func (mm *MessageMapping) Deploy(id string) (err error) {
	return deploy(id, mm.typ, mm.exe)
}
func (mm *MessageMapping) Delete(id string) (err error) {
	return deleteCall(id, mm.typ, mm.exe)
}
func (mm *MessageMapping) GetVersion(id string, version string) (string, error) {
	return getVersion(id, version, mm.typ, mm.exe)
}
func (mm *MessageMapping) Exists(id string, version string) (bool, error) {
	return exists(id, version, mm.typ, mm.exe)
}
func (mm *MessageMapping) GetContent(id string, version string) ([]byte, error) {
	return getContent(id, version, mm.typ, mm.exe)
}
func (mm *MessageMapping) DiffContent(firstDir string, secondDir string) bool {
	return diffContent(firstDir, secondDir)
}
