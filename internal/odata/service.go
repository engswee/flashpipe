package odata

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
			Host:     config.GetMandatoryString(cmd, "tmn-host"),
			Userid:   config.GetMandatoryString(cmd, "tmn-userid"),
			Password: config.GetMandatoryString(cmd, "tmn-password"),
		}
	} else {
		return &ServiceDetails{
			Host:              config.GetMandatoryString(cmd, "tmn-host"),
			OauthHost:         oauthHost,
			OauthClientId:     config.GetMandatoryString(cmd, "oauth-clientid"),
			OauthClientSecret: config.GetMandatoryString(cmd, "oauth-clientsecret"),
			OauthPath:         config.GetString(cmd, "oauth-path"),
		}
	}
}

func InitHTTPExecuter(serviceDetails *ServiceDetails) *httpclnt.HTTPExecuter {
	return httpclnt.New(serviceDetails.OauthHost, serviceDetails.OauthPath, serviceDetails.OauthClientId, serviceDetails.OauthClientSecret, serviceDetails.Userid, serviceDetails.Password, serviceDetails.Host, "https", 443)
}

func modifyingCall(method string, urlPath string, content []byte, successCode int, callType string, exe *httpclnt.HTTPExecuter) error {
	headers, cookies, err := InitHeadersAndCookies(exe)
	if err != nil {
		return err
	}

	headers["Accept"] = "application/json"
	var body io.Reader
	if len(content) > 0 {
		headers["Content-Type"] = "application/json"
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
		return exe.LogError(resp, callType)
	}
	return nil
}

func readOnlyCall(urlPath string, callType string, exe *httpclnt.HTTPExecuter) (*http.Response, error) {
	headers := map[string]string{
		"Accept": "application/json",
	}

	resp, err := exe.ExecGetRequest(urlPath, headers)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return resp, exe.LogError(resp, callType)
	}
	return resp, nil
}
