package odata

import (
	"encoding/json"
	"fmt"
	"github.com/engswee/flashpipe/httpclnt"
	"net/http"
)

type Integration struct {
	HttpExecuter *httpclnt.HTTPExecuter
	TmnHost      string
}

type designtimeArtifact struct {
	Root struct {
		Version string `json:"Version"`
	} `json:"d"`
}

func (int *Integration) Deploy(id string) (err error) {
	// TODO - csrf token

	url := fmt.Sprintf("https://%v/api/v1/DeployIntegrationDesigntimeArtifact?Id='%s'&Version='active'", int.TmnHost, id)
	//url := fmt.Sprintf("https://%v/api/v1/DeployMessageMappingDesigntimeArtifact?Id='%s'&Version='active'", int.TmnHost, id)

	headers := map[string]string{
		"Accept": "application/json",
	}
	resp, err := int.HttpExecuter.ExecRequest("POST", url, http.NoBody, headers)
	if err != nil {
		return
	}
	if resp.StatusCode != 202 {
		return int.HttpExecuter.LogError(resp, "Deploy designtime artifact")
	}
	return nil
}

func (int *Integration) Get(id string, version string) (resp *http.Response, err error) {
	url := fmt.Sprintf("https://%v/api/v1/IntegrationDesigntimeArtifacts(Id='%v',Version='%v')", int.TmnHost, id, version)
	//url := fmt.Sprintf("https://%v/api/v1/MessageMappingDesigntimeArtifacts(Id='%v',Version='%v')", int.TmnHost, id, version)

	headers := map[string]string{
		"Accept": "application/json",
	}
	return int.HttpExecuter.ExecGetRequest(url, headers)
}

func (int *Integration) GetVersion(id string, version string) (string, error) {
	resp, err := int.Get(id, version)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", int.HttpExecuter.LogError(resp, "Get designtime artifact")
	} else {
		var jsonData *designtimeArtifact
		respBody, err := int.HttpExecuter.ReadRespBody(resp)
		err = json.Unmarshal(respBody, &jsonData)
		if err != nil {
			return "", err
		}
		return jsonData.Root.Version, nil
	}
}
