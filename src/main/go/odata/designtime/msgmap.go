package designtime

import (
	"encoding/json"
	"fmt"
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
	resp, err := get(id, version, mm.typ, mm.exe)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", mm.exe.LogError(resp, fmt.Sprintf("Get %v designtime artifact", mm.typ))
	} else {
		var jsonData *designtimeArtifactData
		respBody, err := mm.exe.ReadRespBody(resp)
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

func (mm *MessageMapping) Exists(id string, version string) (bool, error) {
	resp, err := get(id, version, mm.typ, mm.exe)
	if err != nil {
		return false, err
	}
	if resp.StatusCode == 200 {
		return true, nil
	} else if resp.StatusCode == 404 {
		return false, nil
	} else {
		return false, mm.exe.LogError(resp, fmt.Sprintf("Get %v designtime artifact", mm.typ))
	}
}

func (mm *MessageMapping) GetContent(id string, version string) ([]byte, error) {
	path := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts(Id='%v',Version='%v')/$value", mm.typ, id, version)

	resp, err := mm.exe.ExecGetRequest(path, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, mm.exe.LogError(resp, fmt.Sprintf("Download %v designtime artifact", mm.typ))
	} else {
		return mm.exe.ReadRespBody(resp)
	}
}
