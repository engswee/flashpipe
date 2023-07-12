package designtime

import (
	"github.com/engswee/flashpipe/httpclnt"
	"github.com/engswee/flashpipe/odata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type ConfigurationSuite struct {
	suite.Suite
	serviceDetails *odata.ServiceDetails
	exe            *httpclnt.HTTPExecuter
}

func TestConfiguration_BasicAuth(t *testing.T) {
	suite.Run(t, &ConfigurationSuite{
		serviceDetails: &odata.ServiceDetails{
			Host:     os.Getenv("HOST_TMN"),
			Userid:   os.Getenv("BASIC_USERID"),
			Password: os.Getenv("BASIC_PASSWORD"),
		},
	})
}

func TestConfigurationOauth(t *testing.T) {
	suite.Run(t, &ConfigurationSuite{
		serviceDetails: &odata.ServiceDetails{
			Host:              os.Getenv("HOST_TMN"),
			OauthHost:         os.Getenv("HOST_OAUTH"),
			OauthPath:         os.Getenv("HOST_OAUTH_PATH"),
			OauthClientId:     os.Getenv("OAUTH_CLIENTID"),
			OauthClientSecret: os.Getenv("OAUTH_CLIENTSECRET"),
		},
	})
}

func (suite *ConfigurationSuite) SetupSuite() {
	println("Setting up suite")
	suite.exe = odata.InitHTTPExecuter(suite.serviceDetails)
}

func (suite *ConfigurationSuite) SetupTest() {
	println("Setting up test")
}

func (suite *ConfigurationSuite) TearDownTest() {
	println("Tearing down test")
}

func (suite *ConfigurationSuite) TearDownSuite() {
	println("Tearing down suite")
	c := NewConfiguration(suite.exe)

	err := c.Update("IFlow1", "active", "Endpoint", "/iflow1")
	if err != nil {
		suite.T().Fatalf("Update failed with error - %v", err)
	}
}

func (suite *ConfigurationSuite) TestConfiguration_Get() {
	c := NewConfiguration(suite.exe)

	parametersData, err := c.Get("IFlow1", "active")
	if err != nil {
		return
	}
	parameter := FindParameterByKey("Endpoint", parametersData.Root.Results)
	assert.Equal(suite.T(), "/iflow1", parameter.ParameterValue, "Parameter Endpoint should have value /iflow1")
}

func (suite *ConfigurationSuite) TestConfiguration_Update() {
	c := NewConfiguration(suite.exe)

	err := c.Update("IFlow1", "active", "Endpoint", "/flow_update")
	if err != nil {
		return
	}
	parametersData, err := c.Get("IFlow1", "active")
	if err != nil {
		return
	}
	parameter := FindParameterByKey("Endpoint", parametersData.Root.Results)
	assert.Equal(suite.T(), "/flow_update", parameter.ParameterValue, "Parameter Endpoint should have value /flow_update after update")
}
