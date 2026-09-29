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
	mt := new(FaultMessageType)
	mt.exe = exe
	mt.typ = "FaultMessageType"
	return mt
}

func (mt *FaultMessageType) Create(id string, name string, packageId string, artifactDir string) error {
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
	log.Info().Msgf("Creating %v designtime artifact %v", mt.typ, id)
	urlPath := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts", mt.typ)
	return upsert(id, name, packageId, description, artifactDir, "POST", urlPath, 201, mt.typ, "Create", mt.exe)
}

func (mt *FaultMessageType) Update(id string, name string, packageId string, artifactDir string) error {
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

	log.Info().Msgf("Updating %v designtime artifact %v", mt.typ, id)
	urlPath := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts(Id='%v',Version='active')", mt.typ, id)
	return upsert(id, name, packageId, description, artifactDir, "PUT", urlPath, 200, mt.typ, "Update", mt.exe)
}

func (mt *FaultMessageType) Deploy(id string) error {
	log.Warn().Msgf("Deployment of FaultMessageType designtime artifact not supported. Skipping deployment of %v", id)
	return nil
}

func (mt *FaultMessageType) Delete(id string) error {
	return deleteCall(id, mt.typ, mt.exe)
}

func (mt *FaultMessageType) Get(id string, version string) (string, string, bool, error) {
	return get(id, version, mt.typ, mt.exe)
}

func (mt *FaultMessageType) Download(targetFile string, id string) error {
	return download(targetFile, id, mt.typ, mt.exe)
}

func (mt *FaultMessageType) CopyContent(srcDir string, tgtDir string) error {
	return copyContent(srcDir, tgtDir)
}

func (mt *FaultMessageType) CompareContent(srcDir string, tgtDir string, _ []string, _ string) (bool, error) {
	return diffContent(srcDir, tgtDir), nil
}
