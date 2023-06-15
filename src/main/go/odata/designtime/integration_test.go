package designtime

import (
	"github.com/engswee/flashpipe/httpclnt"
	"os"
	"testing"
)

func TestIntegrationDeployOauth(t *testing.T) {
	host := os.Getenv("HOST_TMN")
	oauthHost := os.Getenv("HOST_OAUTH")
	oauthPath := os.Getenv("HOST_OAUTH_PATH")
	clientId := os.Getenv("OAUTH_CLIENTID")
	clientSecret := os.Getenv("OAUTH_CLIENTSECRET")
	exe := httpclnt.New(oauthHost, oauthPath, clientId, clientSecret, "", "", host)
	dt := NewIntegration(exe)

	err := dt.Deploy("Hello")
	if err != nil {
		t.Fatalf("Deployment failed with error - %v", err)
	}
}

func TestIntegrationDeployBasicAuth(t *testing.T) {
	host := os.Getenv("HOST_TMN")
	userId := os.Getenv("BASIC_USERID")
	password := os.Getenv("BASIC_PASSWORD")
	exe := httpclnt.New("", "", "", "", userId, password, host)
	dt := NewIntegration(exe)

	err := dt.Deploy("Hello")
	if err != nil {
		t.Fatalf("Deployment failed with error - %v", err)
	}
}
