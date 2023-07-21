package odata

import (
	"encoding/json"
	"fmt"
	"github.com/engswee/flashpipe/file"
	"github.com/engswee/flashpipe/httpclnt"
	"github.com/rs/zerolog/log"
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
	DiffContent(firstDir string, secondDir string) bool
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
	artifact := &designtimeArtifactUpdateData{
		Name:            name,
		Id:              id,
		PackageId:       packageId,
		ArtifactContent: content,
	}
	// Update of Message Mapping fails as PackageId and Id are not allowed
	if method == "PUT" {
		artifact.Id = ""
		artifact.PackageId = ""
	}
	requestBody, err := json.Marshal(artifact)
	if err != nil {
		return nil, err
	}

	return requestBody, nil
}

func Download(targetFile string, id string, dt DesigntimeArtifact) error {
	log.Info().Msgf("Getting content of artifact %v from tenant for comparison", id)
	content, err := dt.GetContent(id, "active")
	if err != nil {
		return err
	}

	err = os.WriteFile(targetFile, content, os.ModePerm)
	if err != nil {
		return err
	}
	log.Info().Msgf("Content of artifact %v downloaded to %v", id, targetFile)
	return nil
}

func create(id string, name string, packageId string, artifactDir string, artifactType string, exe *httpclnt.HTTPExecuter) error {
	log.Info().Msgf("Creating %v designtime artifact %v", artifactType, id)
	urlPath := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts", artifactType)
	return upsert(id, name, packageId, artifactDir, "POST", urlPath, 201, artifactType, "Create", exe)
}

func update(id string, name string, packageId string, artifactDir string, artifactType string, exe *httpclnt.HTTPExecuter) error {
	log.Info().Msgf("Updating %v designtime artifact %v", artifactType, id)
	urlPath := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts(Id='%v',Version='active')", artifactType, id)
	return upsert(id, name, packageId, artifactDir, "PUT", urlPath, 200, artifactType, "Update", exe)
}

func deploy(id string, artifactType string, exe *httpclnt.HTTPExecuter) error {
	log.Info().Msgf("Deploying %v designtime artifact %v", artifactType, id)
	urlPath := fmt.Sprintf("/api/v1/Deploy%vDesigntimeArtifact?Id='%s'&Version='active'", artifactType, id)
	return modifyingCall("POST", urlPath, nil, 202, fmt.Sprintf("Deploy %v designtime artifact", artifactType), exe)
}

func deleteCall(id string, artifactType string, exe *httpclnt.HTTPExecuter) error {
	log.Info().Msgf("Deleting %v designtime artifact %v", artifactType, id)
	urlPath := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts(Id='%v',Version='active')", artifactType, id)
	return modifyingCall("DELETE", urlPath, nil, 200, fmt.Sprintf("Delete %v designtime artifact", artifactType), exe)
}

func upsert(id string, name string, packageId string, artifactDir string, method string, urlPath string, successCode int, artifactType string, callType string, exe *httpclnt.HTTPExecuter) error {
	// Zip directory and encode to base64
	encoded, err := file.ZipDirToBase64(artifactDir)
	if err != nil {
		return err
	}
	requestBody, err := constructUpdateBody(method, id, name, packageId, encoded)
	if err != nil {
		return err
	}

	return modifyingCall(method, urlPath, requestBody, successCode, fmt.Sprintf("%v %v designtime artifact", callType, artifactType), exe)
}

func getVersion(id string, version string, artifactType string, exe *httpclnt.HTTPExecuter) (string, error) {
	log.Info().Msgf("Getting version of %v designtime artifact %v", artifactType, id)
	urlPath := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts(Id='%v',Version='%v')", artifactType, id, version)

	callType := fmt.Sprintf("Get %v designtime artifact", artifactType)
	resp, err := readOnlyCall(urlPath, callType, exe)
	if err != nil {
		return "", err
	}
	// Process response to extract version
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

func exists(id string, version string, artifactType string, exe *httpclnt.HTTPExecuter) (bool, error) {
	log.Info().Msgf("Checking existence of %v designtime artifact %v", artifactType, id)
	urlPath := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts(Id='%v',Version='%v')", artifactType, id, version)

	callType := fmt.Sprintf("Get %v designtime artifact", artifactType)
	_, err := readOnlyCall(urlPath, callType, exe)
	if err != nil {
		if err.Error() == fmt.Sprintf("%v call failed with response code = 404", callType) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func getContent(id string, version string, artifactType string, exe *httpclnt.HTTPExecuter) ([]byte, error) {
	log.Info().Msgf("Getting content of %v designtime artifact %v", artifactType, id)
	urlPath := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts(Id='%v',Version='%v')/$value", artifactType, id, version)

	callType := fmt.Sprintf("Download %v designtime artifact", artifactType)
	resp, err := readOnlyCall(urlPath, callType, exe)
	if err != nil {
		return nil, err
	}
	return exe.ReadRespBody(resp)
}

func diffContent(firstDir string, secondDir string) bool {
	log.Info().Msg("Checking for changes in META-INF directory")
	metaDiffer := file.DiffDirectories(firstDir+"/META-INF", secondDir+"/META-INF")
	log.Info().Msg("Checking for changes in src/main/resources directory")
	resourcesDiffer := file.DiffDirectories(firstDir+"/src/main/resources", secondDir+"/src/main/resources")
	// TODO - to consider moving diff of parameters.prop here as it is only used in Sync but not Update
	if metaDiffer || resourcesDiffer {
		return true
	} else {
		return false
	}
}
