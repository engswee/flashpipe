package api

import (
	"fmt"

	"github.com/engswee/flashpipe/internal/httpclnt"
	"github.com/rs/zerolog/log"
)

type MessageType struct {
	exe *httpclnt.HTTPExecuter
	typ string
}

// NewMessageType returns an initialised MessageType instance.
func NewMessageType(exe *httpclnt.HTTPExecuter) DesigntimeArtifact {
	mt := new(MessageType)
	mt.exe = exe
	mt.typ = "MessageType"
	return mt
}

func (mt *MessageType) Create(id string, name string, packageId string, artifactDir string) error {
	description, err := getDescriptionFromAdditionalAttributes(artifactDir)
	if err != nil {
		return err
	}

	log.Info().Msgf("Creating %v designtime artifact %v", mt.typ, id)
	urlPath := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts", mt.typ)
	return upsert(id, name, packageId, description, artifactDir, "POST", urlPath, 201, mt.typ, "Create", mt.exe)
}

func (mt *MessageType) Update(id string, name string, packageId string, artifactDir string) error {
	description, err := getDescriptionFromAdditionalAttributes(artifactDir)
	if err != nil {
		return err
	}

	log.Info().Msgf("Updating %v designtime artifact %v", mt.typ, id)
	urlPath := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts(Id='%v',Version='active')", mt.typ, id)
	return upsert(id, name, packageId, description, artifactDir, "PUT", urlPath, 200, mt.typ, "Update", mt.exe)
}

func (mt *MessageType) Deploy(id string) error {
	log.Warn().Msgf("Deployment of MessageType designtime artifact not supported. Skipping deployment of %v", id)
	return nil
}

func (mt *MessageType) Delete(id string) error {
	return deleteCall(id, mt.typ, mt.exe)
}

func (mt *MessageType) Get(id string, version string) (string, string, bool, error) {
	return get(id, version, mt.typ, mt.exe)
}

func (mt *MessageType) Download(targetFile string, id string) error {
	return download(targetFile, id, mt.typ, mt.exe)
}

func (mt *MessageType) CopyContent(srcDir string, tgtDir string) error {
	return copyContent(srcDir, tgtDir)
}

func (mt *MessageType) CompareContent(srcDir string, tgtDir string, _ []string, _ string) (bool, error) {
	return diffContent(srcDir, tgtDir), nil
}
