package httpclnt

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2/clientcredentials"
	"io"
	"net/http"
	"time"
)

type HTTPExecuter struct {
	basicUserId   string
	basicPassword string
	host          string
	scheme        string
	port          int
	httpClient    *http.Client
	AuthType      string
	showLogs      bool
}

// New returns an initialised HTTPExecuter instance.
func New(oauthHost string, oauthPath string, clientId string, clientSecret string, userId string, password string, host string, scheme string, port int, showLogs bool) *HTTPExecuter {
	e := new(HTTPExecuter)
	e.host = host
	e.scheme = scheme
	e.port = port
	e.showLogs = showLogs
	if oauthHost != "" {
		if showLogs {
			log.Debug().Msg("Initialising HTTP client with OAuth 2.0")
		}

		tokenURL := fmt.Sprintf("%v://%v:%d%v", scheme, oauthHost, port, oauthPath)
		if showLogs {
			log.Debug().Msgf("Setting up OAuth 2.0 client with token URL %v", tokenURL)
		}

		// Reference https://pkg.go.dev/golang.org/x/oauth2/clientcredentials#pkg-overview
		conf := &clientcredentials.Config{
			ClientID:     clientId,
			ClientSecret: clientSecret,
			TokenURL:     tokenURL,
		}

		ctx := context.Background()
		e.httpClient = conf.Client(ctx)
		e.AuthType = "OAUTH"
	} else {
		if showLogs {
			log.Debug().Msg("Initialising HTTP client with Basic Authentication")
		}
		e.httpClient = &http.Client{Timeout: 30 * time.Second}
		e.basicUserId = userId
		e.basicPassword = password
		e.AuthType = "BASIC"
	}
	return e
}

func (e *HTTPExecuter) ExecRequestWithCookies(method string, path string, body io.Reader, headers map[string]string, cookies []*http.Cookie) (resp *http.Response, err error) {

	url := fmt.Sprintf("%v://%v:%d%v", e.scheme, e.host, e.port, path)
	if e.showLogs {
		log.Debug().Msgf("Executing HTTP request: %v %v", method, url)
	}

	// Create new HTTP request
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return
	}

	// Set basic authentication if needed
	if e.basicUserId != "" {
		req.SetBasicAuth(e.basicUserId, e.basicPassword)
	}

	// Set HTTP headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Set cookies
	if len(cookies) > 0 {
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
	}

	// Execute HTTP request
	return e.httpClient.Do(req)
}

func (e *HTTPExecuter) ExecGetRequest(path string, headers map[string]string) (resp *http.Response, err error) {
	return e.ExecRequestWithCookies(http.MethodGet, path, http.NoBody, headers, nil)
}

func (e *HTTPExecuter) ReadRespBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (e *HTTPExecuter) LogError(resp *http.Response, callType string) (resBody []byte, err error) {
	resBody, err = e.ReadRespBody(resp)
	if err != nil {
		return
	}

	if len(resBody) != 0 && e.showLogs {
		log.Error().Msgf("Response body = %s", resBody)
	}

	return resBody, fmt.Errorf("%v call failed with response code = %d", callType, resp.StatusCode)
}
