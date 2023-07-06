package designtime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/engswee/flashpipe/httpclnt"
	"github.com/engswee/flashpipe/logger"
	"github.com/engswee/flashpipe/odata"
	"net/http"
	"os"
)

type DesigntimeArtifact interface {
	Deploy(id string) error
	Create(id string, name string, packageId string, content string) error
	Update(id string, name string, packageId string, content string) error
	Delete(id string) error
	Get(id string, version string) (*http.Response, error)
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
	bytes, err := dt.GetContent(id, "active")
	if err != nil {
		return err
	}

	err = os.WriteFile(targetFile, bytes, os.ModePerm)
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("Content of artifact %v downloaded to %v", id, targetFile))
	return nil
}

func Upsert(id string, name string, packageId string, content string, method string, urlPath string, successCode int, callType string, exe *httpclnt.HTTPExecuter) (err error) {
	headers, cookies, err := odata.InitHeadersAndCookies(exe)
	if err != nil {
		return
	}
	headers["Accept"] = "application/json"
	headers["Content-Type"] = "application/json"

	artifactData, err := constructUpdateBody(method, id, name, packageId, content)
	if err != nil {
		return
	}

	resp, err := exe.ExecRequestWithCookies(method, urlPath, bytes.NewReader(artifactData), headers, cookies)
	if err != nil {
		return
	}
	if resp.StatusCode != successCode {
		return exe.LogError(resp, callType)
	}
	return nil
}
