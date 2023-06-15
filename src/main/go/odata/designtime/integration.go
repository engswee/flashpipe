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

type artifactData struct {
	Root struct {
		Version string `json:"Version"`
	} `json:"d"`
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
		var jsonData *artifactData
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
