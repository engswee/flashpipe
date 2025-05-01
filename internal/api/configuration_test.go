package api

import (
	"os"
	"testing"

	"github.com/engswee/flashpipe/internal/httpclnt"
	"github.com/engswee/flashpipe/internal/logger"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ConfigurationSuite struct {
	suite.Suite
	serviceDetails *ServiceDetails
	exe            *httpclnt.HTTPExecuter
}

func TestConfigurationBasicAuth(t *testing.T) {
	suite.Run(t, &ConfigurationSuite{
		serviceDetails: &ServiceDetails{
			Host:     os.Getenv("FLASHPIPE_TMN_HOST"),
			Userid:   os.Getenv("FLASHPIPE_TMN_USERID"),
			Password: os.Getenv("FLASHPIPE_TMN_PASSWORD"),
		},
	})
}

func TestConfigurationOauth(t *testing.T) {
	suite.Run(t, &ConfigurationSuite{
		serviceDetails: &ServiceDetails{
			Host:              os.Getenv("FLASHPIPE_TMN_HOST"),
			OauthHost:         os.Getenv("FLASHPIPE_OAUTH_HOST"),
			OauthPath:         os.Getenv("FLASHPIPE_OAUTH_PATH"),
			OauthClientId:     os.Getenv("FLASHPIPE_OAUTH_CLIENTID"),
			OauthClientSecret: os.Getenv("FLASHPIPE_OAUTH_CLIENTSECRET"),
		},
	})
}

func (suite *ConfigurationSuite) SetupSuite() {
	println("========== Setting up suite - start ==========")
	suite.exe = InitHTTPExecuter(suite.serviceDetails)

	// Setup viper in case debug logs are required
	viper.SetEnvPrefix("FLASHPIPE")
	viper.AutomaticEnv()
	logger.InitConsoleLogger(viper.GetBool("debug"))

	setupPackage(suite.T(), "FlashPipeIntegrationTest", suite.exe)

	setupArtifact(suite.T(), "Integration_Test_IFlow", "FlashPipeIntegrationTest", "../../test/testdata/artifacts/update/Integration_Test_IFlow", "Integration", suite.exe)
	println("========== Setting up suite - end ==========")
}

func (suite *ConfigurationSuite) SetupTest() {
	println("---------- Setting up test - start ----------")
	println("---------- Setting up test - end ----------")
}

func (suite *ConfigurationSuite) TearDownTest() {
	println("---------- Tearing down test - start ----------")
	println("---------- Tearing down test - end ----------")
}

func (suite *ConfigurationSuite) TearDownSuite() {
	println("========== Tearing down suite - start ==========")

	tearDownPackage(suite.T(), "FlashPipeIntegrationTest", suite.exe)
	println("========== Tearing down suite - end ==========")
}

func (suite *ConfigurationSuite) TestConfiguration_Get() {
	c := NewConfiguration(suite.exe)

	parametersData, err := c.Get("Integration_Test_IFlow", "active")
	if err != nil {
		suite.T().Fatalf("Get failed with error - %v", err)
	}
	parameterSenderEndpoint := FindParameterByKey("Sender Endpoint", parametersData.Root.Results)
	assert.Equal(suite.T(), "/flow", parameterSenderEndpoint.ParameterValue, "Parameter Sender Endpoint should have value /flow")
	parameter1 := FindParameterByKey("Parameter 1", parametersData.Root.Results)
	assert.Equal(suite.T(), "Value1", parameter1.ParameterValue, "Parameter 1 should have value Value1")
	parameter2 := FindParameterByKey("Parameter 2", parametersData.Root.Results)
	assert.Equal(suite.T(), "Value 2 plus ${property.Parameter1}", parameter2.ParameterValue, "Parameter 2 should have value Value 2 plus ${property.Parameter1}")
}

func (suite *ConfigurationSuite) TestConfiguration_Update() {
	c := NewConfiguration(suite.exe)

	err := c.Update("Integration_Test_IFlow", "active", "Sender Endpoint", "/flow_update")
	if err != nil {
		suite.T().Fatalf("Update failed with error - %v", err)
	}
	err = c.Update("Integration_Test_IFlow", "active", "Parameter 1", "Value 1 updated")
	if err != nil {
		suite.T().Fatalf("Update failed with error - %v", err)
	}
	err = c.Update("Integration_Test_IFlow", "active", "Parameter 2", "Value 2 with ${header.Parameter1}")
	if err != nil {
		suite.T().Fatalf("Update failed with error - %v", err)
	}

	parametersData, err := c.Get("Integration_Test_IFlow", "active")
	if err != nil {
		suite.T().Fatalf("Get failed with error - %v", err)
	}
	parameterSenderEndpoint := FindParameterByKey("Sender Endpoint", parametersData.Root.Results)
	assert.Equal(suite.T(), "/flow_update", parameterSenderEndpoint.ParameterValue, "Parameter Sender Endpoint should have value /flow_update after update")
	parameter1 := FindParameterByKey("Parameter 1", parametersData.Root.Results)
	assert.Equal(suite.T(), "Value 1 updated", parameter1.ParameterValue, "Parameter 1 should have value Value 1 updated after update")
	parameter2 := FindParameterByKey("Parameter 2", parametersData.Root.Results)
	assert.Equal(suite.T(), "Value 2 with ${header.Parameter1}", parameter2.ParameterValue, "Parameter 2 should have value Value 2 with ${header.Parameter1} after update")
}
