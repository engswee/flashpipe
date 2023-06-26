package httpclnt

import (
	"fmt"
	"os"
	"testing"
)

func TestOauth(t *testing.T) {
	host := os.Getenv("HOST_TMN")
	oauthHost := os.Getenv("HOST_OAUTH")
	oauthPath := os.Getenv("HOST_OAUTH_PATH")
	clientId := os.Getenv("OAUTH_CLIENTID")
	clientSecret := os.Getenv("OAUTH_CLIENTSECRET")
	exe := New(oauthHost, oauthPath, clientId, clientSecret, "", "", host, "https", "443")
	path := fmt.Sprintf("/api/v1/IntegrationDesigntimeArtifacts(Id='%v',Version='Active')", "Hello")

	headers := map[string]string{
		"Accept": "application/json",
	}
	_, err := exe.ExecGetRequest(path, headers)
	if err != nil {
		t.Fatalf("HTTP call failed with error - %v", err)
	}
}

func TestBasicAuth(t *testing.T) {
	host := os.Getenv("HOST_TMN")
	userId := os.Getenv("BASIC_USERID")
	password := os.Getenv("BASIC_PASSWORD")
	exe := New("", "", "", "", userId, password, host, "https", "443")
	path := fmt.Sprintf("/api/v1/IntegrationDesigntimeArtifacts(Id='%v',Version='Active')", "Hello")

	headers := map[string]string{
		"Accept": "application/json",
	}
	_, err := exe.ExecGetRequest(path, headers)
	if err != nil {
		t.Fatalf("HTTP call failed with error - %v", err)
	}
}
