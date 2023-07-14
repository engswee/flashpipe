package odata

import (
	"github.com/engswee/flashpipe/httpclnt"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type ConfigurationSuite struct {
	suite.Suite
	serviceDetails *ServiceDetails
	exe            *httpclnt.HTTPExecuter
}

func TestConfiguration_BasicAuth(t *testing.T) {
	suite.Run(t, &ConfigurationSuite{
		serviceDetails: &ServiceDetails{
			Host:     os.Getenv("HOST_TMN"),
			Userid:   os.Getenv("BASIC_USERID"),
			Password: os.Getenv("BASIC_PASSWORD"),
		},
	})
}

func TestConfigurationOauth(t *testing.T) {
	suite.Run(t, &ConfigurationSuite{
		serviceDetails: &ServiceDetails{
			Host:              os.Getenv("HOST_TMN"),
			OauthHost:         os.Getenv("HOST_OAUTH"),
			OauthPath:         os.Getenv("HOST_OAUTH_PATH"),
			OauthClientId:     os.Getenv("OAUTH_CLIENTID"),
			OauthClientSecret: os.Getenv("OAUTH_CLIENTSECRET"),
		},
	})
}

func (suite *ConfigurationSuite) SetupSuite() {
	println("========== Setting up suite ==========")
	suite.exe = InitHTTPExecuter(suite.serviceDetails)
	// Setup viper in case debug logs are required
	viper.SetEnvPrefix("FLASHPIPE")
	viper.AutomaticEnv()

	setupPackage(suite.T(), "FlashPipeIntegrationTest", suite.exe)

	setupArtifact(suite.T(), "IFlow1", "FlashPipeIntegrationTest", "../testdata/artifacts/setup/IFlow1", "Integration", suite.exe)
}

func (suite *ConfigurationSuite) SetupTest() {
	println("---------- Setting up test ----------")
}

func (suite *ConfigurationSuite) TearDownTest() {
	println("---------- Tearing down test ----------")
}

func (suite *ConfigurationSuite) TearDownSuite() {
	println("========== Tearing down suite ==========")

	tearDownPackage(suite.T(), "FlashPipeIntegrationTest", suite.exe)
}

func (suite *ConfigurationSuite) TestConfiguration_Get() {
	c := NewConfiguration(suite.exe)

	parametersData, err := c.Get("IFlow1", "active")
	if err != nil {
		suite.T().Fatalf("Get failed with error - %v", err)
	}
	parameter := FindParameterByKey("Endpoint", parametersData.Root.Results)
	assert.Equal(suite.T(), "/flow1", parameter.ParameterValue, "Parameter Endpoint should have value /flow1")
}

func (suite *ConfigurationSuite) TestConfiguration_Update() {
	c := NewConfiguration(suite.exe)

	err := c.Update("IFlow1", "active", "Endpoint", "/flow_update")
	if err != nil {
		suite.T().Fatalf("Update failed with error - %v", err)
	}
	parametersData, err := c.Get("IFlow1", "active")
	if err != nil {
		suite.T().Fatalf("Get failed with error - %v", err)
	}
	parameter := FindParameterByKey("Endpoint", parametersData.Root.Results)
	assert.Equal(suite.T(), "/flow_update", parameter.ParameterValue, "Parameter Endpoint should have value /flow_update after update")
}
