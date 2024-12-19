package api

import (
	"bytes"
	"github.com/engswee/flashpipe/internal/config"
	"github.com/engswee/flashpipe/internal/httpclnt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"io"
	"net/http"
)

type ServiceDetails struct {
	Host              string
	Userid            string
	Password          string
	OauthHost         string
	OauthPath         string
	OauthClientId     string
	OauthClientSecret string
}

func GetServiceDetails(cmd *cobra.Command) *ServiceDetails {
	oauthHost := config.GetString(cmd, "oauth-host")
	if oauthHost == "" {
		return &ServiceDetails{
			Host:     config.GetString(cmd, "tmn-host"),
			Userid:   config.GetString(cmd, "tmn-userid"),
			Password: config.GetString(cmd, "tmn-password"),
		}
	} else {
		return &ServiceDetails{
			Host:              config.GetString(cmd, "tmn-host"),
			OauthHost:         oauthHost,
			OauthClientId:     config.GetString(cmd, "oauth-clientid"),
			OauthClientSecret: config.GetString(cmd, "oauth-clientsecret"),
			OauthPath:         config.GetString(cmd, "oauth-path"),
		}
	}
}

func InitHTTPExecuter(serviceDetails *ServiceDetails) *httpclnt.HTTPExecuter {
	return httpclnt.New(serviceDetails.OauthHost, serviceDetails.OauthPath, serviceDetails.OauthClientId, serviceDetails.OauthClientSecret, serviceDetails.Userid, serviceDetails.Password, serviceDetails.Host, "https", 443, true)
}

func modifyingCall(method string, urlPath string, content []byte, successCode int, callType string, exe *httpclnt.HTTPExecuter) error {
	return modifyingCallWithContentType(method, urlPath, content, "application/json", successCode, callType, exe)
}

func modifyingCallWithContentType(method string, urlPath string, content []byte, contentType string, successCode int, callType string, exe *httpclnt.HTTPExecuter) error {
	headers, cookies, err := InitHeadersAndCookies(exe)
	if err != nil {
		return err
	}

	headers["Accept"] = "application/json"
	var body io.Reader
	if len(content) > 0 {
		headers["Content-Type"] = contentType
		log.Debug().Msgf("Request body = %s", content)
		body = bytes.NewReader(content)
	} else {
		body = http.NoBody
	}

	resp, err := exe.ExecRequestWithCookies(method, urlPath, body, headers, cookies)
	if err != nil {
		return err
	}
	if resp.StatusCode != successCode {
		_, err = exe.LogError(resp, callType)
		return err
	}
	return nil
}

func readOnlyCall(urlPath string, callType string, exe *httpclnt.HTTPExecuter) (*http.Response, error) {
	return readOnlyCallWithBodyAndAcceptType(urlPath, nil, callType, "application/json", exe)
}

func readOnlyCallWithBody(urlPath string, content []byte, callType string, exe *httpclnt.HTTPExecuter) (*http.Response, error) {
	return readOnlyCallWithBodyAndAcceptType(urlPath, content, callType, "", exe)
}

func readOnlyCallWithBodyAndAcceptType(urlPath string, content []byte, callType string, acceptType string, exe *httpclnt.HTTPExecuter) (*http.Response, error) {
	headers := map[string]string{}
	if acceptType != "" {
		headers["Accept"] = acceptType
	}
	var body io.Reader
	if len(content) > 0 {
		log.Debug().Msgf("Request body = %s", content)
		body = bytes.NewReader(content)
	} else {
		body = http.NoBody
	}

	resp, err := exe.ExecRequestWithCookies(http.MethodGet, urlPath, body, headers, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		resBody, err := exe.LogError(resp, callType)
		resp.Body = io.NopCloser(bytes.NewReader(resBody))
		return resp, err
	}
	return resp, nil
}
