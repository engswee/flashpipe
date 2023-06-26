package designtime

import (
	"fmt"
	"github.com/engswee/flashpipe/httpclnt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestIntegrationDeployOauth(t *testing.T) {
	host := os.Getenv("HOST_TMN")
	oauthHost := os.Getenv("HOST_OAUTH")
	oauthPath := os.Getenv("HOST_OAUTH_PATH")
	clientId := os.Getenv("OAUTH_CLIENTID")
	clientSecret := os.Getenv("OAUTH_CLIENTSECRET")
	exe := httpclnt.New(oauthHost, oauthPath, clientId, clientSecret, "", "", host, "https", "443")
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
	exe := httpclnt.New("", "", "", "", userId, password, host, "https", "443")
	dt := NewIntegration(exe)

	err := dt.Deploy("Hello")
	if err != nil {
		t.Fatalf("Deployment failed with error - %v", err)
	}
}

func TestMockDeploy(t *testing.T) {

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/", func(w http.ResponseWriter, r *http.Request) {
		//auth := r.Header.Get("Auth")
		//if auth != secret {
		//	http.Error(w, "Auth header was incorrect", http.StatusUnauthorized)
		//	return
		//}
		// Header was good, tell 'em what day it is
		w.Header().Set("x-csrf-token", "mytoken")
		//w.Write([]byte(`{ "day": "Sunday" }`))
	})
	mux.HandleFunc("/api/v1/DeployIntegrationDesigntimeArtifact", func(w http.ResponseWriter, r *http.Request) {
		values := r.URL.Query()
		if values.Get("Id") == "" {
			http.Error(w, "Missing query parameter Id", http.StatusBadRequest)
			return
		}
		if values.Get("Version") == "" {
			http.Error(w, "Missing query parameter Version", http.StatusBadRequest)
			return
		}
		//w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(202)
		response := fmt.Sprintf("74d45405-68cf-4e3d-7701-2507f804178c")
		w.Write([]byte(response))
	})
	svr := httptest.NewServer(mux)

	defer svr.Close()
	//host := os.Getenv("HOST_TMN")
	userId := os.Getenv("BASIC_USERID")
	password := os.Getenv("BASIC_PASSWORD")
	urlParts := strings.Split(strings.TrimPrefix(svr.URL, "http://"), ":")
	exe := httpclnt.New("", "", "", "", userId, password, urlParts[0], "http", urlParts[1])
	dt := NewIntegration(exe)

	err := dt.Deploy("Hello")
	if err != nil {
		t.Fatalf("Deployment failed with error - %v", err)
	}
}
