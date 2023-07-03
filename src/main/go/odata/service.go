package odata

import (
	"github.com/engswee/flashpipe/config"
	"github.com/engswee/flashpipe/httpclnt"
	"github.com/spf13/cobra"
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
	oauthHost := config.GetFlagAsString(cmd, "oauth.host")
	if oauthHost == "" {
		return &ServiceDetails{
			Host:     config.GetRequiredFlagAsString(cmd, "tmn.host"),
			Userid:   config.GetRequiredFlagAsString(cmd, "tmn.userid"),
			Password: config.GetRequiredFlagAsString(cmd, "tmn.password"),
		}
	} else {
		return &ServiceDetails{
			Host:              config.GetRequiredFlagAsString(cmd, "tmn.host"),
			OauthHost:         oauthHost,
			OauthClientId:     config.GetRequiredFlagAsString(cmd, "oauth.clientid"),
			OauthClientSecret: config.GetRequiredFlagAsString(cmd, "oauth.clientsecret"),
			OauthPath:         config.GetFlagAsString(cmd, "oauth.path"),
		}
	}
}

func InitHTTPExecuter(serviceDetails *ServiceDetails) *httpclnt.HTTPExecuter {
	return httpclnt.New(serviceDetails.OauthHost, serviceDetails.OauthPath, serviceDetails.OauthClientId, serviceDetails.OauthClientSecret, serviceDetails.Userid, serviceDetails.Password, serviceDetails.Host, "https", 443)
}
