package odata

import (
	"github.com/engswee/flashpipe/config"
	"github.com/engswee/flashpipe/httpclnt"
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

func ModifyingCall(method string, urlPath string, body io.Reader, successCode int, callType string, exe *httpclnt.HTTPExecuter) error {
	headers, cookies, err := InitHeadersAndCookies(exe)
	if err != nil {
		return err
	}

	headers["Accept"] = "application/json"
	if body != http.NoBody {
		headers["Content-Type"] = "application/json"
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
