package designtime

import "net/http"

type DesigntimeArtifact interface {
	Deploy(id string) error
	Get(id string, version string) (*http.Response, error)
	GetVersion(id string, version string) (string, error)
}
