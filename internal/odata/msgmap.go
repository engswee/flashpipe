package odata

import (
	"github.com/engswee/flashpipe/internal/httpclnt"
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
func (mm *MessageMapping) Get(id string, version string) (string, bool, error) {
	return get(id, version, mm.typ, mm.exe)
}
func (mm *MessageMapping) GetContent(id string, version string) ([]byte, error) {
	return getContent(id, version, mm.typ, mm.exe)
}
func (mm *MessageMapping) CopyContent(srcDir string, tgtDir string) error {
	return copyContent(srcDir, tgtDir)
}
func (mm *MessageMapping) CompareContent(srcDir string, tgtDir string, _ string, _ string) (bool, error) {
	// Diff directories
	return diffContent(srcDir, tgtDir), nil
}
