package api

import (
	"fmt"

	"github.com/engswee/flashpipe/internal/httpclnt"
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
	description, err := getDescriptionFromAdditionalAttributes(artifactDir)
	if err != nil {
		return err
	}

	log.Info().Msgf("Creating %v designtime artifact %v", mt.typ, id)
	urlPath := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts", mt.typ)
	return upsert(id, name, packageId, description, artifactDir, "POST", urlPath, 201, mt.typ, "Create", mt.exe)
}

func (mt *FaultMessageType) Update(id string, name string, packageId string, artifactDir string) error {
	description, err := getDescriptionFromAdditionalAttributes(artifactDir)
	if err != nil {
		return err
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
