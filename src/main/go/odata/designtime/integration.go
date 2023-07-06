package designtime

import (
	"encoding/json"
	"fmt"
	"github.com/engswee/flashpipe/httpclnt"
	"github.com/engswee/flashpipe/odata"
	"net/http"
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

func (int *Integration) Deploy(id string) (err error) {
	path := fmt.Sprintf("/api/v1/Deploy%vDesigntimeArtifact?Id='%s'&Version='active'", int.typ, id)

	headers, cookies, err := odata.InitHeadersAndCookies(int.exe)
	if err != nil {
		return
	}
	headers["Accept"] = "application/json"

	resp, err := int.exe.ExecRequestWithCookies("POST", path, http.NoBody, headers, cookies)
	if err != nil {
		return
	}
	if resp.StatusCode != 202 {
		return int.exe.LogError(resp, fmt.Sprintf("Deploy %v designtime artifact", int.typ))
	}
	return nil
}

func (int *Integration) Create(id string, name string, packageId string, content string) (err error) {
	urlPath := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts", int.typ)
	callType := fmt.Sprintf("Create %v designtime artifact", int.typ)

	return Upsert(id, name, packageId, content, "POST", urlPath, 201, callType, int.exe)
}

func (int *Integration) Update(id string, name string, packageId string, content string) (err error) {
	urlPath := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts(Id='%v',Version='active')", int.typ, id)
	callType := fmt.Sprintf("Update %v designtime artifact", int.typ)

	return Upsert(id, name, packageId, content, "PUT", urlPath, 200, callType, int.exe)
}

func (int *Integration) Delete(id string) (err error) {
	urlPath := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts(Id='%v',Version='active')", int.typ, id)
	callType := fmt.Sprintf("Delete %v designtime artifact", int.typ)

	headers, cookies, err := odata.InitHeadersAndCookies(int.exe)
	if err != nil {
		return
	}
	headers["Accept"] = "application/json"

	resp, err := int.exe.ExecRequestWithCookies("DELETE", urlPath, http.NoBody, headers, cookies)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		return int.exe.LogError(resp, callType)
	}
	return nil
}

func (int *Integration) Get(id string, version string) (resp *http.Response, err error) {
	path := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts(Id='%v',Version='%v')", int.typ, id, version)

	headers := map[string]string{
		"Accept": "application/json",
	}
	return int.exe.ExecGetRequest(path, headers)
}

func (int *Integration) GetVersion(id string, version string) (string, error) {
	resp, err := int.Get(id, version)
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
	resp, err := int.Get(id, version)
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
