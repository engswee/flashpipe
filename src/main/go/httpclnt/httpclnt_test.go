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
	exe := New(host, "/oauth/token", clientId, clientSecret, "", "", host, "http", port)

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
	exe := New("", "", "", "", userId, password, host, "http", port)

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
	exe := New("", "", "", "", userId, password, host, "http", port)

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
		err := exe.LogError(resp, "Get Integration designtime")
		errMsg := err.Error()
		if errMsg != "Get Integration designtime call failed with response code = 404" {
			t.Fatalf("Actual error returned = %s", errMsg)
		}
	} else {
		t.Fatalf("HTTP call failed with response code - %v", resp.StatusCode)
	}
}

func TestOauth(t *testing.T) {
	host := os.Getenv("HOST_TMN")
	oauthHost := os.Getenv("HOST_OAUTH")
	oauthPath := os.Getenv("HOST_OAUTH_PATH")
	clientId := os.Getenv("OAUTH_CLIENTID")
	clientSecret := os.Getenv("OAUTH_CLIENTSECRET")
	exe := New(oauthHost, oauthPath, clientId, clientSecret, "", "", host, "https", 443)
	path := fmt.Sprintf("/api/v1/IntegrationDesigntimeArtifacts(Id='%v',Version='Active')", "Hello")

	headers := map[string]string{
		"Accept": "application/json",
	}
	resp, err := exe.ExecGetRequest(path, headers)
	if err != nil {
		t.Fatalf("HTTP call failed with error - %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("HTTP call failed with response code - %v", resp.StatusCode)
	}
}

func TestBasicAuth(t *testing.T) {
	host := os.Getenv("HOST_TMN")
	userId := os.Getenv("BASIC_USERID")
	password := os.Getenv("BASIC_PASSWORD")
	exe := New("", "", "", "", userId, password, host, "https", 443)
	path := fmt.Sprintf("/api/v1/IntegrationDesigntimeArtifacts(Id='%v',Version='Active')", "Hello")

	headers := map[string]string{
		"Accept": "application/json",
	}
	resp, err := exe.ExecGetRequest(path, headers)
	if err != nil {
		t.Fatalf("HTTP call failed with error - %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("HTTP call failed with response code - %v", resp.StatusCode)
	}
}
