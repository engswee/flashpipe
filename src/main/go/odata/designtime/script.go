package designtime

import (
	"encoding/json"
	"fmt"
	"github.com/engswee/flashpipe/httpclnt"
	"github.com/engswee/flashpipe/odata"
	"net/http"
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

func (sc *ScriptCollection) Deploy(id string) (err error) {
	path := fmt.Sprintf("/api/v1/Deploy%vDesigntimeArtifact?Id='%s'&Version='active'", sc.typ, id)

	headers, cookies, err := odata.InitHeadersAndCookies(sc.exe)
	if err != nil {
		return
	}
	headers["Accept"] = "application/json"

	resp, err := sc.exe.ExecRequestWithCookies("POST", path, http.NoBody, headers, cookies)
	if err != nil {
		return
	}
	if resp.StatusCode != 202 {
		return sc.exe.LogError(resp, fmt.Sprintf("Deploy %v designtime artifact", sc.typ))
	}
	return nil
}

func (sc *ScriptCollection) Create(id string, name string, packageId string, content string) (err error) {
	urlPath := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts", sc.typ)
	callType := fmt.Sprintf("Create %v designtime artifact", sc.typ)

	return Upsert(id, name, packageId, content, "POST", urlPath, 201, callType, sc.exe)
}

func (sc *ScriptCollection) Update(id string, name string, packageId string, content string) (err error) {
	urlPath := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts(Id='%v',Version='active')", sc.typ, id)
	callType := fmt.Sprintf("Update %v designtime artifact", sc.typ)

	return Upsert(id, name, packageId, content, "PUT", urlPath, 200, callType, sc.exe)
}

func (sc *ScriptCollection) Delete(id string) (err error) {
	urlPath := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts(Id='%v',Version='active')", sc.typ, id)
	callType := fmt.Sprintf("Delete %v designtime artifact", sc.typ)

	headers, cookies, err := odata.InitHeadersAndCookies(sc.exe)
	if err != nil {
		return
	}
	headers["Accept"] = "application/json"

	resp, err := sc.exe.ExecRequestWithCookies("DELETE", urlPath, http.NoBody, headers, cookies)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		return sc.exe.LogError(resp, callType)
	}
	return nil
}

func (sc *ScriptCollection) Get(id string, version string) (resp *http.Response, err error) {
	path := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts(Id='%v',Version='%v')", sc.typ, id, version)

	headers := map[string]string{
		"Accept": "application/json",
	}
	return sc.exe.ExecGetRequest(path, headers)
}

func (sc *ScriptCollection) GetVersion(id string, version string) (string, error) {
	resp, err := sc.Get(id, version)
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
	resp, err := sc.Get(id, version)
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
