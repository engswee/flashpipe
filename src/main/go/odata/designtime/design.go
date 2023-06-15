package designtime

import (
	"github.com/engswee/flashpipe/httpclnt"
	"net/http"
)

type DesigntimeArtifact interface {
	Deploy(id string) error
	Get(id string, version string) (*http.Response, error)
	GetVersion(id string, version string) (string, error)
}

type designtimeArtifactData struct {
	Root struct {
		Version string `json:"Version"`
	} `json:"d"`
}

func GetDesigntimeArtifactByType(artifactType string, exe *httpclnt.HTTPExecuter) DesigntimeArtifact {
	switch artifactType {
	case "MESSAGE_MAPPING":
		return NewMessageMapping(exe)
	case "SCRIPT_COLLECTION":
		return NewScriptCollection(exe)
	case "INTEGRATION_FLOW":
		return NewIntegration(exe)
	default:
		return nil
	}
}
