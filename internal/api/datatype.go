package api

import (
	"github.com/engswee/flashpipe/internal/httpclnt"
)

type DataType struct {
	exe *httpclnt.HTTPExecuter
	typ string
}

// NewDataType returns an initialised DataType instance.
func NewDataType(exe *httpclnt.HTTPExecuter) DesigntimeArtifact {
	dt := new(DataType)
	dt.exe = exe
	dt.typ = "DataType"
	return dt
}

func (dt *DataType) Create(id string, name string, packageId string, artifactDir string) error {
	return create(id, name, packageId, artifactDir, dt.typ, dt.exe)
}

func (dt *DataType) Update(id string, name string, packageId string, artifactDir string) error {
	return update(id, name, packageId, artifactDir, dt.typ, dt.exe)
}

func (dt *DataType) Deploy(id string) error {
	return deploy(id, dt.typ, dt.exe)
}

func (dt *DataType) Delete(id string) error {
	return deleteCall(id, dt.typ, dt.exe)
}

func (dt *DataType) Get(id string, version string) (string, string, bool, error) {
	return get(id, version, dt.typ, dt.exe)
}

func (dt *DataType) Download(targetFile string, id string) error {
	return download(targetFile, id, dt.typ, dt.exe)
}

func (dt *DataType) CopyContent(srcDir string, tgtDir string) error {
	return copyContent(srcDir, tgtDir)
}

func (dt *DataType) CompareContent(srcDir string, tgtDir string, _ []string, _ string) (bool, error) {
	return diffContent(srcDir, tgtDir), nil
}
