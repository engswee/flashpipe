package odata

import (
	"encoding/json"
	"fmt"
	"github.com/engswee/flashpipe/httpclnt"
	"net/http"
)

type Runtime struct {
	HttpExecuter *httpclnt.HTTPExecuter
	TmnHost      string
}

type runtimeArtifact struct {
	Root struct {
		Version string `json:"Version"`
		Status  string `json:"Status"`
	} `json:"d"`
}

type runtimeError struct {
	Parameter      []string `json:"parameter"`
	ChildInstances []struct {
		Parameter []string `json:"parameter"`
	} `json:"childInstances"`
}

func (r *Runtime) Get(id string) (resp *http.Response, err error) {
	url := fmt.Sprintf("https://%v/api/v1/IntegrationRuntimeArtifacts('%v')", r.TmnHost, id)

	headers := map[string]string{
		"Accept": "application/json",
	}
	return r.HttpExecuter.ExecGetRequest(url, headers)
}

func (r *Runtime) GetVersion(id string) (string, error) {
	resp, err := r.Get(id)
	if err != nil {
		return "", err
	}
	if resp.StatusCode == 404 { // artifact not deployed to runtime
		return "", nil
	} else if resp.StatusCode != 200 { // TODO - in Java > (code.startsWith('2'))
		return "", r.HttpExecuter.LogError(resp, "Get runtime artifact")
	} else {
		var jsonData *runtimeArtifact
		respBody, err := r.HttpExecuter.ReadRespBody(resp)
		err = json.Unmarshal(respBody, &jsonData)
		if err != nil {
			return "", err
		}
		if jsonData.Root.Status == "STARTED" {
			return jsonData.Root.Version, nil
		} else { // artifact runtime deployment failed or not complete
			return "", nil
		}
	}
}

func (r *Runtime) GetStatus(id string) (string, error) {
	resp, err := r.Get(id)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 { // TODO - in Java > (code.startsWith('2'))
		return "", r.HttpExecuter.LogError(resp, "Get runtime artifact")
	} else {
		var jsonData *runtimeArtifact
		respBody, err := r.HttpExecuter.ReadRespBody(resp)
		err = json.Unmarshal(respBody, &jsonData)
		if err != nil {
			return "", err
		}

		return jsonData.Root.Status, nil
	}
}

func (r *Runtime) GetErrorInfo(id string) (string, error) {
	url := fmt.Sprintf("https://%v/api/v1/IntegrationRuntimeArtifacts('%v')/ErrorInformation/$value", r.TmnHost, id)

	headers := map[string]string{
		"Accept": "application/json",
	}
	resp, err := r.HttpExecuter.ExecGetRequest(url, headers)
	if err != nil {
		return "", nil
	}
	if resp.StatusCode != 200 { // TODO - in Java > (code.startsWith('2'))
		return "", r.HttpExecuter.LogError(resp, "Get runtime artifact error information")
	} else {
		var jsonData *runtimeError
		respBody, err := r.HttpExecuter.ReadRespBody(resp)
		err = json.Unmarshal(respBody, &jsonData)
		if err != nil {
			return "", err
		}

		if len(jsonData.Parameter) > 0 {
			return jsonData.Parameter[0], nil
		} else {
			return jsonData.ChildInstances[0].Parameter[0], nil
		}
	}
}
