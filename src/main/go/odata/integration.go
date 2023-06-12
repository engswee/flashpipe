package odata

import (
	"encoding/json"
	"fmt"
	"github.com/engswee/flashpipe/httpclnt"
	"net/http"
)

type Integration struct {
	exe *httpclnt.HTTPExecuter
}

type designtimeArtifact struct {
	Root struct {
		Version string `json:"Version"`
	} `json:"d"`
}

// NewIntegration returns an initialised Integration instance.
func NewIntegration(exe *httpclnt.HTTPExecuter) DesignArtifact {
	i := new(Integration)
	i.exe = exe
	return i
}

func (int *Integration) Deploy(id string) (err error) {
	// TODO - csrf token

	url := fmt.Sprintf("/api/v1/DeployIntegrationDesigntimeArtifact?Id='%s'&Version='active'", id)
	//url := fmt.Sprintf("https://%v/api/v1/DeployMessageMappingDesigntimeArtifact?Id='%s'&Version='active'", int.host, id)

	headers := map[string]string{
		"Accept": "application/json",
	}
	resp, err := int.exe.ExecRequest("POST", url, http.NoBody, headers)
	if err != nil {
		return
	}
	if resp.StatusCode != 202 {
		return int.exe.LogError(resp, "Deploy designtime artifact")
	}
	return nil
}

func (int *Integration) Get(id string, version string) (resp *http.Response, err error) {
	url := fmt.Sprintf("/api/v1/IntegrationDesigntimeArtifacts(Id='%v',Version='%v')", id, version)
	//url := fmt.Sprintf("https://%v/api/v1/MessageMappingDesigntimeArtifacts(Id='%v',Version='%v')", int.host, id, version)

	headers := map[string]string{
		"Accept": "application/json",
	}
	return int.exe.ExecGetRequest(url, headers)
}

func (int *Integration) GetVersion(id string, version string) (string, error) {
	resp, err := int.Get(id, version)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", int.exe.LogError(resp, "Get designtime artifact")
	} else {
		var jsonData *designtimeArtifact
		respBody, err := int.exe.ReadRespBody(resp)
		err = json.Unmarshal(respBody, &jsonData)
		if err != nil {
			return "", err
		}
		return jsonData.Root.Version, nil
	}
}
