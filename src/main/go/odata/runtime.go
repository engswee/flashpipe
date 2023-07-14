package odata

import (
	"encoding/json"
	"fmt"
	"github.com/engswee/flashpipe/httpclnt"
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

func (r *Runtime) UnDeploy(id string) error {
	urlPath := fmt.Sprintf("/api/v1/IntegrationRuntimeArtifacts('%v')", id)

	return modifyingCall("DELETE", urlPath, nil, 202, "", r.exe)
}

func (r *Runtime) GetVersion(id string) (string, error) {
	urlPath := fmt.Sprintf("/api/v1/IntegrationRuntimeArtifacts('%v')", id)

	callType := "Get runtime artifact"
	resp, err := readOnlyCall(urlPath, callType, r.exe)
	if err != nil {
		if err.Error() == fmt.Sprintf("%v call failed with response code = 404", callType) { // artifact not deployed to runtime
			return "NOT_DEPLOYED", nil
		} else {
			return "", err
		}
	}
	// Process response to extract version
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

func (r *Runtime) GetStatus(id string) (string, error) {
	urlPath := fmt.Sprintf("/api/v1/IntegrationRuntimeArtifacts('%v')", id)

	callType := "Get runtime artifact"
	resp, err := readOnlyCall(urlPath, callType, r.exe)
	if err != nil {
		return "", err
	}
	// Process response to extract status
	var jsonData *runtimeData
	respBody, err := r.exe.ReadRespBody(resp)
	err = json.Unmarshal(respBody, &jsonData)
	if err != nil {
		return "", err
	}

	return jsonData.Root.Status, nil
}

func (r *Runtime) GetErrorInfo(id string) (string, error) {
	urlPath := fmt.Sprintf("/api/v1/IntegrationRuntimeArtifacts('%v')/ErrorInformation/$value", id)

	callType := "Get runtime artifact error information"
	resp, err := readOnlyCall(urlPath, callType, r.exe)
	if err != nil {
		return "", err
	}
	// Process response to extract error info
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
