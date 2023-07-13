package odata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/engswee/flashpipe/file"
	"github.com/engswee/flashpipe/httpclnt"
	"github.com/engswee/flashpipe/logger"
	"net/http"
	"os"
)

type DesigntimeArtifact interface {
	Create(id string, name string, packageId string, artifactDir string) error
	Update(id string, name string, packageId string, artifactDir string) error
	Deploy(id string) error
	Delete(id string) error
	GetVersion(id string, version string) (string, error)
	Exists(id string, version string) (bool, error)
	GetContent(id string, version string) ([]byte, error)
}

type designtimeArtifactData struct {
	Root struct {
		Version string `json:"Version"`
	} `json:"d"`
}

type designtimeArtifactUpdateData struct {
	Name            string `json:"Name"`
	Id              string `json:"Id,omitempty"`
	PackageId       string `json:"PackageId,omitempty"`
	ArtifactContent string `json:"ArtifactContent"`
}

func NewDesigntimeArtifact(artifactType string, exe *httpclnt.HTTPExecuter) DesigntimeArtifact {
	switch artifactType {
	case "MessageMapping":
		return NewMessageMapping(exe)
	case "ScriptCollection":
		return NewScriptCollection(exe)
	case "Integration":
		return NewIntegration(exe)
	case "ValueMapping":
		return NewValueMapping(exe)
	default:
		return nil
	}
}

func constructUpdateBody(method string, id string, name string, packageId string, content string) ([]byte, error) {
	artifactData := &designtimeArtifactUpdateData{
		Name:            name,
		Id:              id,
		PackageId:       packageId,
		ArtifactContent: content,
	}
	// Update of Message Mapping fails as PackageId and Id are not allowed
	if method == "PUT" {
		artifactData.Id = ""
		artifactData.PackageId = ""
	}
	jsonBody, err := json.Marshal(artifactData)
	if err != nil {
		return nil, err
	}
	logger.Debug(fmt.Sprintf("Request body = %s", jsonBody))

	return jsonBody, nil
}

func Download(targetFile string, id string, dt DesigntimeArtifact) error {
	logger.Info(fmt.Sprintf("Getting content of artifact %v from tenant for comparison", id))
	content, err := dt.GetContent(id, "active")
	if err != nil {
		return err
	}

	err = os.WriteFile(targetFile, content, os.ModePerm)
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("Content of artifact %v downloaded to %v", id, targetFile))
	return nil
}

func create(id string, name string, packageId string, artifactDir string, artifactType string, exe *httpclnt.HTTPExecuter) error {
	urlPath := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts", artifactType)
	return upsert(id, name, packageId, artifactDir, "POST", urlPath, 201, artifactType, "Create", exe)
}

func update(id string, name string, packageId string, artifactDir string, artifactType string, exe *httpclnt.HTTPExecuter) error {
	urlPath := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts(Id='%v',Version='active')", artifactType, id)
	return upsert(id, name, packageId, artifactDir, "PUT", urlPath, 200, artifactType, "Update", exe)
}

func deploy(id string, artifactType string, exe *httpclnt.HTTPExecuter) error {
	urlPath := fmt.Sprintf("/api/v1/Deploy%vDesigntimeArtifact?Id='%s'&Version='active'", artifactType, id)
	return ModifyingCall("POST", urlPath, http.NoBody, 202, fmt.Sprintf("Deploy %v designtime artifact", artifactType), exe)
}

func deleteCall(id string, artifactType string, exe *httpclnt.HTTPExecuter) error {
	urlPath := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts(Id='%v',Version='active')", artifactType, id)
	return ModifyingCall("DELETE", urlPath, http.NoBody, 200, fmt.Sprintf("Delete %v designtime artifact", artifactType), exe)
}

func upsert(id string, name string, packageId string, artifactDir string, method string, urlPath string, successCode int, artifactType string, callType string, exe *httpclnt.HTTPExecuter) error {
	// Zip directory and encode to base64
	encoded, err := file.ZipDirToBase64(artifactDir)
	if err != nil {
		return err
	}
	artifactData, err := constructUpdateBody(method, id, name, packageId, encoded)
	if err != nil {
		return err
	}

	return ModifyingCall(method, urlPath, bytes.NewReader(artifactData), successCode, fmt.Sprintf("%v %v designtime artifact", callType, artifactType), exe)
}

func get(id string, version string, artifactType string, exe *httpclnt.HTTPExecuter) (*http.Response, error) {
	path := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts(Id='%v',Version='%v')", artifactType, id, version)

	headers := map[string]string{
		"Accept": "application/json",
	}
	return exe.ExecGetRequest(path, headers)
}

func getVersion(id string, version string, artifactType string, exe *httpclnt.HTTPExecuter) (string, error) {
	resp, err := get(id, version, artifactType, exe)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", exe.LogError(resp, fmt.Sprintf("Get %v designtime artifact", artifactType))
	} else {
		var jsonData *designtimeArtifactData
		respBody, err := exe.ReadRespBody(resp)
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

func exists(id string, version string, artifactType string, exe *httpclnt.HTTPExecuter) (bool, error) {
	resp, err := get(id, version, artifactType, exe)
	if err != nil {
		return false, err
	}
	if resp.StatusCode == 200 {
		return true, nil
	} else if resp.StatusCode == 404 {
		return false, nil
	} else {
		return false, exe.LogError(resp, fmt.Sprintf("Get %v designtime artifact", artifactType))
	}
}

func getContent(id string, version string, artifactType string, exe *httpclnt.HTTPExecuter) ([]byte, error) {
	path := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts(Id='%v',Version='%v')/$value", artifactType, id, version)

	resp, err := exe.ExecGetRequest(path, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, exe.LogError(resp, fmt.Sprintf("Download %v designtime artifact", artifactType))
	} else {
		return exe.ReadRespBody(resp)
	}
}
