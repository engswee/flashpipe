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

type runtimeData struct {
	Root struct {
		Version string `json:"Version"`
		Status  string `json:"Status"`
	} `json:"d"`
}

type runtimeError struct {
	Parameter []string `json:"parameter"`
}

// NewRuntime returns an initialised Runtime instance.
func NewRuntime(exe *httpclnt.HTTPExecuter) *Runtime {
	r := new(Runtime)
	r.exe = exe
	return r
}

func (r *Runtime) get(id string) (resp *http.Response, err error) {
	path := fmt.Sprintf("/api/v1/IntegrationRuntimeArtifacts('%v')", id)

	headers := map[string]string{
		"Accept": "application/json",
	}
	return r.exe.ExecGetRequest(path, headers)
}

func (r *Runtime) UnDeploy(id string) error {
	urlPath := fmt.Sprintf("/api/v1/IntegrationRuntimeArtifacts('%v')", id)

	return ModifyingCall("DELETE", urlPath, http.NoBody, 202, "", r.exe)
}

func (r *Runtime) GetVersion(id string) (string, error) {
	resp, err := r.get(id)
	if err != nil {
		return "", err
	}
	if resp.StatusCode == 404 { // artifact not deployed to runtime
		return "", nil
	} else if resp.StatusCode != 200 {
		return "", r.exe.LogError(resp, "Get runtime artifact")
	} else {
		var jsonData *runtimeData
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
	resp, err := r.get(id)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", r.exe.LogError(resp, "Get runtime artifact")
	} else {
		var jsonData *runtimeData
		respBody, err := r.exe.ReadRespBody(resp)
		err = json.Unmarshal(respBody, &jsonData)
		if err != nil {
			return "", err
		}

		return jsonData.Root.Status, nil
	}
}

func (r *Runtime) GetErrorInfo(id string) (string, error) {
	path := fmt.Sprintf("/api/v1/IntegrationRuntimeArtifacts('%v')/ErrorInformation/$value", id)

	headers := map[string]string{
		"Accept": "application/json",
	}
	resp, err := r.exe.ExecGetRequest(path, headers)
	if err != nil {
		return "", nil
	}
	if resp.StatusCode != 200 {
		return "", r.exe.LogError(resp, "Get runtime artifact error information")
	} else {
		var jsonData *runtimeError
		respBody, err := r.exe.ReadRespBody(resp)
		if err != nil {
			return "", err
		}
		err = json.Unmarshal(respBody, &jsonData)
		if err != nil {
			return "", err
		}
		return jsonData.Parameter[0], nil
	}
}
