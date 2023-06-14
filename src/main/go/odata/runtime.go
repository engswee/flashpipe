package odata

import (
	"encoding/json"
	"fmt"
	"github.com/engswee/flashpipe/httpclnt"
	"net/http"
)

type Runtime struct {
	exe *httpclnt.HTTPExecuter
}

type artifactData struct {
	Root struct {
		Version string `json:"Version"`
		Status  string `json:"Status"`
	} `json:"d"`
}

type artifactError struct {
	Parameter      []string `json:"parameter"`
	ChildInstances []struct {
		Parameter []string `json:"parameter"`
	} `json:"childInstances"`
}

// NewRuntime returns an initialised Runtime instance.
func NewRuntime(exe *httpclnt.HTTPExecuter) *Runtime {
	r := new(Runtime)
	r.exe = exe
	return r
}

func (r *Runtime) Get(id string) (resp *http.Response, err error) {
	url := fmt.Sprintf("/api/v1/IntegrationRuntimeArtifacts('%v')", id)

	headers := map[string]string{
		"Accept": "application/json",
	}
	return r.exe.ExecGetRequest(url, headers)
}

func (r *Runtime) GetVersion(id string) (string, error) {
	resp, err := r.Get(id)
	if err != nil {
		return "", err
	}
	if resp.StatusCode == 404 { // artifact not deployed to runtime
		return "", nil
	} else if resp.StatusCode != 200 { // TODO - in Java > (code.startsWith('2'))
		return "", r.exe.LogError(resp, "Get runtime artifact")
	} else {
		var jsonData *artifactData
		respBody, err := r.exe.ReadRespBody(resp)
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
		return "", r.exe.LogError(resp, "Get runtime artifact")
	} else {
		var jsonData *artifactData
		respBody, err := r.exe.ReadRespBody(resp)
		err = json.Unmarshal(respBody, &jsonData)
		if err != nil {
			return "", err
		}

		return jsonData.Root.Status, nil
	}
}

func (r *Runtime) GetErrorInfo(id string) (string, error) {
	url := fmt.Sprintf("/api/v1/IntegrationRuntimeArtifacts('%v')/ErrorInformation/$value", id)

	headers := map[string]string{
		"Accept": "application/json",
	}
	resp, err := r.exe.ExecGetRequest(url, headers)
	if err != nil {
		return "", nil
	}
	if resp.StatusCode != 200 { // TODO - in Java > (code.startsWith('2'))
		return "", r.exe.LogError(resp, "Get runtime artifact error information")
	} else {
		var jsonData *artifactError
		respBody, err := r.exe.ReadRespBody(resp)
		if err != nil {
			return "", err
		}
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
