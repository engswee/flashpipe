package api

import (
	"encoding/json"
	"fmt"
	"github.com/engswee/flashpipe/internal/httpclnt"
	"github.com/go-errors/errors"
	"github.com/rs/zerolog/log"
	"io"
	"strings"
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
	log.Info().Msgf("Undeploying runtime artifact %v", id)
	urlPath := fmt.Sprintf("/api/v1/IntegrationRuntimeArtifacts('%v')", id)

	return modifyingCall("DELETE", urlPath, nil, 202, "", r.exe)
}

func (r *Runtime) Get(id string) (version string, status string, err error) {
	log.Info().Msgf("Getting details of runtime artifact %v", id)
	urlPath := fmt.Sprintf("/api/v1/IntegrationRuntimeArtifacts('%v')", id)

	callType := "Get runtime artifact"
	resp, err := readOnlyCall(urlPath, callType, r.exe)
	if err != nil {
		if err.Error() == fmt.Sprintf("%v call failed with response code = 404", callType) { // artifact not deployed to runtime
			return "NOT_DEPLOYED", "", nil
		} else {
			bytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return "", "", err
			}
			respBody := string(bytes[:])
			if strings.Contains(respBody, "Requested entity could not be found") { // artifact not deployed to runtime
				return "NOT_DEPLOYED", "", nil
			}
			return "", "", err
		}
	}
	// Process response to extract version and status
	var jsonData *runtimeData
	respBody, err := r.exe.ReadRespBody(resp)
	err = json.Unmarshal(respBody, &jsonData)
	if err != nil {
		log.Error().Msgf("Error unmarshalling response as JSON. Response body = %s", respBody)
		return "", "", errors.Wrap(err, 0)
	}
	if jsonData.Root.Status == "STARTED" {
		return jsonData.Root.Version, "STARTED", nil
	} else { // artifact runtime deployment failed or not complete
		return "", jsonData.Root.Status, nil
	}
}

func (r *Runtime) GetErrorInfo(id string) (string, error) {
	log.Info().Msgf("Getting error info of runtime artifact %v", id)
	urlPath := fmt.Sprintf("/api/v1/IntegrationRuntimeArtifacts('%v')/ErrorInformation/$value", id)

	callType := "Get runtime artifact error information"
	resp, err := readOnlyCall(urlPath, callType, r.exe)
	// TODO - sometimes the error information is only available after some time, so the API returns 204 No content (instead of 200) in the meantime
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
		log.Error().Msgf("Error unmarshalling response as JSON. Response body = %s", respBody)
		return "", errors.Wrap(err, 0)
	}
	return jsonData.Parameter[0], nil
}
