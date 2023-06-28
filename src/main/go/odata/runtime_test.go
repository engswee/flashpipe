package odata

import (
	"github.com/engswee/flashpipe/httpclnt"
	"os"
	"strings"
	"testing"
)

func TestRuntime_GetErrorInfo(t *testing.T) {
	host := os.Getenv("HOST_TMN")
	oauthHost := os.Getenv("HOST_OAUTH")
	oauthPath := os.Getenv("HOST_OAUTH_PATH")
	clientId := os.Getenv("OAUTH_CLIENTID")
	clientSecret := os.Getenv("OAUTH_CLIENTSECRET")
	exe := httpclnt.New(oauthHost, oauthPath, clientId, clientSecret, "", "", host, "https", 443)
	rt := NewRuntime(exe)
	errorMessage, err := rt.GetErrorInfo("Mapping1")
	if err != nil {
		t.Fatalf("HTTP call failed with error - %v", err)
	}
	if !strings.HasPrefix(errorMessage, "Validation of the artifact failed") {
		t.Fatalf("errorMessage does not have prefix")
	}
}
