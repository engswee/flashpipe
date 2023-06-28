package designtime

import (
	"github.com/engswee/flashpipe/httpclnt"
	"net/http"
)

type DesigntimeArtifact interface {
	Deploy(id string) error
	Get(id string, version string) (*http.Response, error)
	GetVersion(id string, version string) (string, error)
	Download(id string, version string) ([]byte, error)
}

type designtimeArtifactData struct {
	Root struct {
		Version string `json:"Version"`
	} `json:"d"`
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
