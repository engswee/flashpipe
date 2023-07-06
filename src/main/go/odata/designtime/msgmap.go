package designtime

import (
	"encoding/json"
	"fmt"
	"github.com/engswee/flashpipe/httpclnt"
	"github.com/engswee/flashpipe/odata"
	"net/http"
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

func (mm *MessageMapping) Deploy(id string) (err error) {
	path := fmt.Sprintf("/api/v1/Deploy%vDesigntimeArtifact?Id='%s'&Version='active'", mm.typ, id)

	headers, cookies, err := odata.InitHeadersAndCookies(mm.exe)
	if err != nil {
		return
	}
	headers["Accept"] = "application/json"

	resp, err := mm.exe.ExecRequestWithCookies("POST", path, http.NoBody, headers, cookies)
	if err != nil {
		return
	}
	if resp.StatusCode != 202 {
		return mm.exe.LogError(resp, fmt.Sprintf("Deploy %v designtime artifact", mm.typ))
	}
	return nil
}

func (mm *MessageMapping) Create(id string, name string, packageId string, content string) (err error) {
	urlPath := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts", mm.typ)
	callType := fmt.Sprintf("Create %v designtime artifact", mm.typ)

	return Upsert(id, name, packageId, content, "POST", urlPath, 201, callType, mm.exe)
}

func (mm *MessageMapping) Update(id string, name string, packageId string, content string) (err error) {
	urlPath := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts(Id='%v',Version='active')", mm.typ, id)
	callType := fmt.Sprintf("Update %v designtime artifact", mm.typ)

	return Upsert(id, name, packageId, content, "PUT", urlPath, 200, callType, mm.exe)
}

func (mm *MessageMapping) Delete(id string) (err error) {
	urlPath := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts(Id='%v',Version='active')", mm.typ, id)
	callType := fmt.Sprintf("Delete %v designtime artifact", mm.typ)

	headers, cookies, err := odata.InitHeadersAndCookies(mm.exe)
	if err != nil {
		return
	}
	headers["Accept"] = "application/json"

	resp, err := mm.exe.ExecRequestWithCookies("DELETE", urlPath, http.NoBody, headers, cookies)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		return mm.exe.LogError(resp, callType)
	}
	return nil
}

func (mm *MessageMapping) Get(id string, version string) (resp *http.Response, err error) {
	path := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts(Id='%v',Version='%v')", mm.typ, id, version)

	headers := map[string]string{
		"Accept": "application/json",
	}
	return mm.exe.ExecGetRequest(path, headers)
}

func (mm *MessageMapping) GetVersion(id string, version string) (string, error) {
	resp, err := mm.Get(id, version)
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
	resp, err := mm.Get(id, version)
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
