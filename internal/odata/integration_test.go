package odata

import (
	"fmt"
	"github.com/engswee/flashpipe/internal/httpclnt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIntegration_DeployMockBasic(t *testing.T) {
	const csrfToken = "dummycsrfToken"
	const artifactId = "DummyIFlow"

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

func TestIntegration_DeployMockOauth(t *testing.T) {
	const oauthToken = "dummyoauthToken"
	const artifactId = "DummyIFlow"

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

func TestIntegration_diffParam(t *testing.T) {
	exe := httpclnt.New("", "", "", "", "", "", "localhost", "http", 8081)
	dt := NewDesigntimeArtifact("Integration", exe)

	dirDiffer, err := dt.CompareContent("../../test/testdata/artifacts/collection/IFlow1", "../../test/testdata/artifacts/update/Integration_Test_IFlow", "", "remote")
	if err != nil {
		t.Fatalf("CompareContent failed with error - %v", err)
	}
	assert.True(t, dirDiffer, "Directory contents do not differ")
}
