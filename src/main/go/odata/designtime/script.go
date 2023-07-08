package designtime

import (
	"encoding/json"
	"fmt"
	"github.com/engswee/flashpipe/httpclnt"
)

type ScriptCollection struct {
	exe *httpclnt.HTTPExecuter
	typ string
}

// NewScriptCollection returns an initialised ScriptCollection instance.
func NewScriptCollection(exe *httpclnt.HTTPExecuter) DesigntimeArtifact {
	sc := new(ScriptCollection)
	sc.exe = exe
	sc.typ = "ScriptCollection"
	return sc
}

func (sc *ScriptCollection) Create(id string, name string, packageId string, artifactDir string) error {
	return create(id, name, packageId, artifactDir, sc.typ, sc.exe)
}
func (sc *ScriptCollection) Update(id string, name string, packageId string, artifactDir string) (err error) {
	return update(id, name, packageId, artifactDir, sc.typ, sc.exe)
}
func (sc *ScriptCollection) Deploy(id string) (err error) {
	return deploy(id, sc.typ, sc.exe)
}
func (sc *ScriptCollection) Delete(id string) (err error) {
	return deleteCall(id, sc.typ, sc.exe)
}

func (sc *ScriptCollection) GetVersion(id string, version string) (string, error) {
	resp, err := get(id, version, sc.typ, sc.exe)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", sc.exe.LogError(resp, fmt.Sprintf("Get %v designtime artifact", sc.typ))
	} else {
		var jsonData *designtimeArtifactData
		respBody, err := sc.exe.ReadRespBody(resp)
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

func (sc *ScriptCollection) Exists(id string, version string) (bool, error) {
	resp, err := get(id, version, sc.typ, sc.exe)
	if err != nil {
		return false, err
	}
	if resp.StatusCode == 200 {
		return true, nil
	} else if resp.StatusCode == 404 {
		return false, nil
	} else {
		return false, sc.exe.LogError(resp, fmt.Sprintf("Get %v designtime artifact", sc.typ))
	}
}

func (sc *ScriptCollection) GetContent(id string, version string) ([]byte, error) {
	path := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts(Id='%v',Version='%v')/$value", sc.typ, id, version)

	resp, err := sc.exe.ExecGetRequest(path, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, sc.exe.LogError(resp, fmt.Sprintf("Download %v designtime artifact", sc.typ))
	} else {
		return sc.exe.ReadRespBody(resp)
	}
}
