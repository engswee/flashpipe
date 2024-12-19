package httpclnt

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMockOauth(t *testing.T) {
	// Set credentials details
	const clientId = "dummyid"
	const clientSecret = "dummysecret"
	const token = "token123"

	// Set up local server with mock HTTP responses
	mux := http.NewServeMux()
	// Handler for OAuth token
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		encoded := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v:%v", clientId, clientSecret)))
		if auth != fmt.Sprintf("Basic %v", encoded) {
			http.Error(w, "Invalid credentials for token URL authorization", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{ "access_token": "%v" }`, token)))
	})
	// Handler for OData endpoint using OAuth token
	mux.HandleFunc("/api/v1/IntegrationDesigntimeArtifacts(Id='Dummy',Version='Active')", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != fmt.Sprintf("Bearer %v", token) {
			http.Error(w, "Invalid token for endpoint authorization", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{ "d": { "Id": "Dummy" } }`)))
	})
	svr := httptest.NewServer(mux)

	defer svr.Close()

	// Initialise HTTP executer
	host, port := GetHostPort(svr.URL)
	exe := New(host, "/oauth/token", clientId, clientSecret, "", "", host, "http", port, true)

	headers := map[string]string{
		"Accept": "application/json",
	}
	// Execute HTTP request
	resp, err := exe.ExecGetRequest("/api/v1/IntegrationDesigntimeArtifacts(Id='Dummy',Version='Active')", headers)

	// Verify HTTP response
	if err != nil {
		t.Fatalf("HTTP call failed with error - %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("HTTP call failed with response code - %v", resp.StatusCode)
	}
}

func TestMockBasicAuth(t *testing.T) {
	// Set credentials details
	const userId = "dummyuser"
	const password = "dummypassword"

	// Set up local server with mock HTTP responses
	mux := http.NewServeMux()
	// Handler for OData endpoint using basic authentication
	mux.HandleFunc("/api/v1/IntegrationDesigntimeArtifacts(Id='Dummy',Version='Active')", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		encoded := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v:%v", userId, password)))
		if auth != fmt.Sprintf("Basic %v", encoded) {
			http.Error(w, "Invalid credentials for basic authentication", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{ "d": { "Id": "Dummy" } }`)))
	})
	svr := httptest.NewServer(mux)

	defer svr.Close()

	// Initialise HTTP executer
	host, port := GetHostPort(svr.URL)
	exe := New("", "", "", "", userId, password, host, "http", port, true)

	headers := map[string]string{
		"Accept": "application/json",
	}
	// Execute HTTP request
	resp, err := exe.ExecGetRequest("/api/v1/IntegrationDesigntimeArtifacts(Id='Dummy',Version='Active')", headers)
	if err != nil {
		t.Fatalf("HTTP call failed with error - %v", err)
	}
	// Verify HTTP response
	if resp.StatusCode != 200 {
		t.Fatalf("HTTP call failed with response code - %v", resp.StatusCode)
	}
}

func TestMockBasicAuthIDNotFound(t *testing.T) {
	// Set credentials details
	const userId = "dummyuser"
	const password = "dummypassword"

	// Set up local server with mock HTTP responses
	mux := http.NewServeMux()
	// Handler for OData endpoint using basic authentication
	mux.HandleFunc("/api/v1/IntegrationDesigntimeArtifacts(Id='Dummy',Version='Active')", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		encoded := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v:%v", userId, password)))
		if auth != fmt.Sprintf("Basic %v", encoded) {
			http.Error(w, "Invalid credentials for basic authentication", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf(`{ "error": { "code": "Not Found" } }`)))
	})
	svr := httptest.NewServer(mux)

	defer svr.Close()

	// Initialise HTTP executer
	host, port := GetHostPort(svr.URL)
	exe := New("", "", "", "", userId, password, host, "http", port, true)

	headers := map[string]string{
		"Accept": "application/json",
	}
	// Execute HTTP request
	resp, err := exe.ExecGetRequest("/api/v1/IntegrationDesigntimeArtifacts(Id='Dummy',Version='Active')", headers)
	if err != nil {
		t.Fatalf("HTTP call failed with error - %v", err)
	}
	// Verify HTTP response
	if resp.StatusCode == http.StatusNotFound {
		_, err = exe.LogError(resp, "Get Integration designtime")
		errMsg := err.Error()
		if errMsg != "Get Integration designtime call failed with response code = 404" {
			t.Fatalf("Actual error returned = %s", errMsg)
		}
	} else {
		t.Fatalf("HTTP call failed with response code - %v", resp.StatusCode)
	}
}

func TestOauth(t *testing.T) {
	host := os.Getenv("FLASHPIPE_TMN_HOST")
	oauthHost := os.Getenv("FLASHPIPE_OAUTH_HOST")
	oauthPath := os.Getenv("FLASHPIPE_OAUTH_PATH")
	clientId := os.Getenv("FLASHPIPE_OAUTH_CLIENTID")
	clientSecret := os.Getenv("FLASHPIPE_OAUTH_CLIENTSECRET")
	exe := New(oauthHost, oauthPath, clientId, clientSecret, "", "", host, "https", 443, true)

	headers := map[string]string{
		"Accept": "application/json",
	}
	resp, err := exe.ExecGetRequest("/api/v1/", headers)
	if err != nil {
		t.Fatalf("HTTP call failed with error - %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("HTTP call failed with response code - %v", resp.StatusCode)
	}
}

func TestBasicAuth(t *testing.T) {
	host := os.Getenv("FLASHPIPE_TMN_HOST")
	userId := os.Getenv("FLASHPIPE_TMN_USERID")
	password := os.Getenv("FLASHPIPE_TMN_PASSWORD")
	exe := New("", "", "", "", userId, password, host, "https", 443, true)

	headers := map[string]string{
		"Accept": "application/json",
	}
	resp, err := exe.ExecGetRequest("/api/v1/", headers)
	if err != nil {
		t.Fatalf("HTTP call failed with error - %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("HTTP call failed with response code - %v", resp.StatusCode)
	}
}
