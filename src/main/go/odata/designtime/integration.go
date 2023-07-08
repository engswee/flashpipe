package designtime

import (
	"encoding/json"
	"fmt"
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
	resp, err := get(id, version, int.typ, int.exe)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", int.exe.LogError(resp, fmt.Sprintf("Get %v designtime artifact", int.typ))
	} else {
		var jsonData *designtimeArtifactData
		respBody, err := int.exe.ReadRespBody(resp)
		if err != nil {
			return "", err
		}
		err = json.Unmarshal(respBody, &jsonData)
		if err != nil {
			return "", err
		}
		return jsonData.Root.Version, nil
	}
}

func (int *Integration) Exists(id string, version string) (bool, error) {
	resp, err := get(id, version, int.typ, int.exe)
	if err != nil {
		return false, err
	}
	if resp.StatusCode == 200 {
		return true, nil
	} else if resp.StatusCode == 404 {
		return false, nil
	} else {
		return false, int.exe.LogError(resp, fmt.Sprintf("Get %v designtime artifact", int.typ))
	}
}

func (int *Integration) GetContent(id string, version string) ([]byte, error) {
	path := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts(Id='%v',Version='%v')/$value", int.typ, id, version)

	resp, err := int.exe.ExecGetRequest(path, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, int.exe.LogError(resp, fmt.Sprintf("Download %v designtime artifact", int.typ))
	} else {
		return int.exe.ReadRespBody(resp)
	}
}
