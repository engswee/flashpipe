package api

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/engswee/flashpipe/internal/file"
	"github.com/engswee/flashpipe/internal/httpclnt"
	"github.com/go-errors/errors"
	"github.com/rs/zerolog/log"
)

type FaultMessageType struct {
	exe *httpclnt.HTTPExecuter
	typ string
}

// NewFaultMessageType returns an initialised FaultMessageType instance.
func NewFaultMessageType(exe *httpclnt.HTTPExecuter) DesigntimeArtifact {
	dt := new(FaultMessageType)
	dt.exe = exe
	dt.typ = "FaultMessageType"
	return dt
}

func (dt *FaultMessageType) Create(id string, name string, packageId string, artifactDir string) error {
	// For FaultMessageType create, the API requires the Description field to be included in the request body.
	// The description is stored in the additionalAttributes.json file in the artifact directory.
	var description string
	attrFile := artifactDir + "/src/main/resources/additionalAttributes.json"
	if file.Exists(attrFile) {
		fileContent, err := os.ReadFile(attrFile)
		if err != nil {
			return err
		}
		var jsonData *artifactAdditionalAttributes

		err = json.Unmarshal(fileContent, &jsonData)
		if err != nil {
			log.Error().Msgf("Error unmarshalling file as JSON. Response body = %s", fileContent)
			return errors.Wrap(err, 0)
		}
		log.Info().Msgf("additionalAttributes.json file found. Description = %s", jsonData.Description)
		description = jsonData.Description
	} else {
		log.Info().Msgf("additionalAttributes.json file not found. Description will be unchanged")
	}
	log.Info().Msgf("Creating %v designtime artifact %v", dt.typ, id)
	urlPath := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts", dt.typ)
	return upsert(id, name, packageId, description, artifactDir, "POST", urlPath, 201, dt.typ, "Create", dt.exe)
}

func (dt *FaultMessageType) Update(id string, name string, packageId string, artifactDir string) error {
	// For FaultMessageType update, the API requires the Description field to be included in the request body.
	// The description is stored in the additionalAttributes.json file in the artifact directory.
	var description string
	attrFile := artifactDir + "/src/main/resources/additionalAttributes.json"
	if file.Exists(attrFile) {
		fileContent, err := os.ReadFile(attrFile)
		if err != nil {
			return err
		}
		var jsonData *artifactAdditionalAttributes

		err = json.Unmarshal(fileContent, &jsonData)
		if err != nil {
			log.Error().Msgf("Error unmarshalling file as JSON. Response body = %s", fileContent)
			return errors.Wrap(err, 0)
		}
		log.Info().Msgf("additionalAttributes.json file found. Description = %s", jsonData.Description)
		description = jsonData.Description
	} else {
		log.Info().Msgf("additionalAttributes.json file not found. Description will be unchanged")
	}

	log.Info().Msgf("Updating %v designtime artifact %v", dt.typ, id)
	urlPath := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts(Id='%v',Version='active')", dt.typ, id)
	return upsert(id, name, packageId, description, artifactDir, "PUT", urlPath, 200, dt.typ, "Update", dt.exe)
}

func (dt *FaultMessageType) Deploy(id string) error {
	log.Warn().Msgf("Deployment of FaultMessageType designtime artifact not supported. Skipping deployment of %v", id)
	return nil
}

func (dt *FaultMessageType) Delete(id string) error {
	return deleteCall(id, dt.typ, dt.exe)
}

func (dt *FaultMessageType) Get(id string, version string) (string, string, bool, error) {
	return get(id, version, dt.typ, dt.exe)
}

func (dt *FaultMessageType) Download(targetFile string, id string) error {
	return download(targetFile, id, dt.typ, dt.exe)
}

func (dt *FaultMessageType) CopyContent(srcDir string, tgtDir string) error {
	return copyContent(srcDir, tgtDir)
}

func (dt *FaultMessageType) CompareContent(srcDir string, tgtDir string, _ []string, _ string) (bool, error) {
	return diffContent(srcDir, tgtDir), nil
}
