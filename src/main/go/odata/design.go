package odata

import "net/http"

type DesignArtifact interface {
	Deploy(id string) error
	Get(id string, version string) (*http.Response, error)
	GetVersion(id string, version string) (string, error)
}
