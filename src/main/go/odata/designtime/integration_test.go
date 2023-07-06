package designtime

import (
	"fmt"
	"github.com/engswee/flashpipe/httpclnt"
	"net/http"
	"net/http/httptest"
	"testing"
)

//func TestIntegration_CreateUpdateDeployDeleteOauth(t *testing.T) {
//	host := os.Getenv("HOST_TMN")
//	oauthHost := os.Getenv("HOST_OAUTH")
//	oauthPath := os.Getenv("HOST_OAUTH_PATH")
//	clientId := os.Getenv("OAUTH_CLIENTID")
//	clientSecret := os.Getenv("OAUTH_CLIENTSECRET")
//	exe := httpclnt.New(oauthHost, oauthPath, clientId, clientSecret, "", "", host, "https", 443)
//	dt := NewDesigntimeArtifact("Integration", exe)
//
//	createUpdateDeployDelete("Integration_Test_IFlow", "Integration Test IFlow", "FlashPipeIntegrationTest", dt, t)
//
//}
//
//func TestIntegration_CreateUpdateDeployDeleteBasicAuth(t *testing.T) {
//	host := os.Getenv("HOST_TMN")
//	userId := os.Getenv("BASIC_USERID")
//	password := os.Getenv("BASIC_PASSWORD")
//	exe := httpclnt.New("", "", "", "", userId, password, host, "https", 443)
//	dt := NewDesigntimeArtifact("Integration", exe)
//
//	createUpdateDeployDelete("Integration_Test_IFlow", "Integration Test IFlow", "FlashPipeIntegrationTest", dt, t)
//
//}

func TestMockDeployBasic(t *testing.T) {
	const csrfToken = "mytoken"
	const artifactId = "IFlow1"

	// Set up local server with mock HTTP responses
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-csrf-token", csrfToken)
	})
	mux.HandleFunc("/api/v1/DeployIntegrationDesigntimeArtifact", func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("x-csrf-token")
		if token != csrfToken {
			http.Error(w, "Invalid value for x-csrf-token", http.StatusForbidden)
			return
		}
		values := r.URL.Query()
		if values.Get("Id") != fmt.Sprintf("'%v'", artifactId) {
			http.Error(w, "Incorrect value for query parameter Id", http.StatusBadRequest)
			return
		}
		if values.Get("Version") != "'active'" {
			http.Error(w, "Incorrect value for query parameter Version", http.StatusBadRequest)
			return
		}
		w.WriteHeader(202)
		w.Write([]byte("74d45405-68cf-4e3d-7701-2507f804178c"))
	})
	svr := httptest.NewServer(mux)

	defer svr.Close()

	host, port := httpclnt.GetHostPort(svr.URL)
	exe := httpclnt.New("", "", "", "", "dummy", "dummy", host, "http", port)
	dt := NewIntegration(exe)

	err := dt.Deploy(artifactId)
	if err != nil {
		t.Fatalf("Deployment failed with error - %v", err)
	}
}

func TestMockDeployOauth(t *testing.T) {
	const oauthToken = "token123"
	const artifactId = "IFlow1"

	// Set up local server with mock HTTP responses
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{ "access_token": "%v" }`, oauthToken)))
	})
	mux.HandleFunc("/api/v1/DeployIntegrationDesigntimeArtifact", func(w http.ResponseWriter, r *http.Request) {
		values := r.URL.Query()
		if values.Get("Id") != fmt.Sprintf("'%v'", artifactId) {
			http.Error(w, "Incorrect value for query parameter Id", http.StatusBadRequest)
			return
		}
		if values.Get("Version") != "'active'" {
			http.Error(w, "Incorrect value for query parameter Version", http.StatusBadRequest)
			return
		}
		w.WriteHeader(202)
		w.Write([]byte("74d45405-68cf-4e3d-7701-2507f804178c"))
	})
	svr := httptest.NewServer(mux)

	defer svr.Close()

	host, port := httpclnt.GetHostPort(svr.URL)
	exe := httpclnt.New(host, "/oauth/token", "dummy", "dummy", "", "", host, "http", port)
	dt := NewIntegration(exe)

	err := dt.Deploy(artifactId)
	if err != nil {
		t.Fatalf("Deployment failed with error - %v", err)
	}
}
