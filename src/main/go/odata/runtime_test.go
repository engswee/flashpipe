package odata

import (
	"github.com/engswee/flashpipe/httpclnt"
	"github.com/engswee/flashpipe/logger"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
	"time"
)

type RuntimeSuite struct {
	suite.Suite
	serviceDetails *ServiceDetails
	exe            *httpclnt.HTTPExecuter
}

func TestRuntimeBasicAuth(t *testing.T) {
	suite.Run(t, &RuntimeSuite{
		serviceDetails: &ServiceDetails{
			Host:     os.Getenv("HOST_TMN"),
			Userid:   os.Getenv("BASIC_USERID"),
			Password: os.Getenv("BASIC_PASSWORD"),
		},
	})
}

func TestRuntimeOauth(t *testing.T) {
	suite.Run(t, &RuntimeSuite{
		serviceDetails: &ServiceDetails{
			Host:              os.Getenv("HOST_TMN"),
			OauthHost:         os.Getenv("HOST_OAUTH"),
			OauthPath:         os.Getenv("HOST_OAUTH_PATH"),
			OauthClientId:     os.Getenv("OAUTH_CLIENTID"),
			OauthClientSecret: os.Getenv("OAUTH_CLIENTSECRET"),
		},
	})
}

func (suite *RuntimeSuite) SetupSuite() {
	println("========== Setting up suite ==========")
	suite.exe = InitHTTPExecuter(suite.serviceDetails)

	// Setup viper in case debug logs are required
	viper.SetEnvPrefix("FLASHPIPE")
	viper.AutomaticEnv()
	logger.InitConsoleLogger(viper.GetBool("debug"))

	setupPackage(suite.T(), "FlashPipeIntegrationTest", suite.exe)

	setupArtifact(suite.T(), "IFlow1", "FlashPipeIntegrationTest", "../testdata/artifacts/setup/IFlow1", "Integration", suite.exe)
	setupRuntime(suite.T(), "IFlow1", "Integration", suite.exe)
	time.Sleep(5 * time.Second)

	setupArtifact(suite.T(), "Mapping1", "FlashPipeIntegrationTest", "../testdata/artifacts/setup/Mapping1", "MessageMapping", suite.exe)
	setupRuntime(suite.T(), "Mapping1", "MessageMapping", suite.exe)
	time.Sleep(5 * time.Second)
}

func (suite *RuntimeSuite) SetupTest() {
	println("---------- Setting up test ----------")
}

func (suite *RuntimeSuite) TearDownTest() {
	println("---------- Tearing down test ----------")
}

func (suite *RuntimeSuite) TearDownSuite() {
	println("========== Tearing down suite ==========")

	tearDownPackage(suite.T(), "FlashPipeIntegrationTest", suite.exe)

	tearDownRuntime(suite.T(), "IFlow1", suite.exe)
	tearDownRuntime(suite.T(), "Mapping1", suite.exe)
}

func (suite *RuntimeSuite) TestRuntime_GetErrorInfo() {
	rt := NewRuntime(suite.exe)
	errorMessage, err := rt.GetErrorInfo("Mapping1")
	if err != nil {
		suite.T().Fatalf("GetErrorInfo failed with error - %v", err)
	}
	assert.Contains(suite.T(), errorMessage, "Validation of the artifact failed", "errorMessage does not have validation error")
}

func (suite *RuntimeSuite) TestRuntime_GetStatusVersion() {
	rt := NewRuntime(suite.exe)
	status, err := rt.GetStatus("IFlow1")
	if err != nil {
		suite.T().Fatalf("GetStatus failed with error - %v", err)
	}
	assert.Equal(suite.T(), "STARTED", status, "Runtime status of IFlow1 is not STARTED")
	version, err := rt.GetVersion("IFlow1")
	if err != nil {
		suite.T().Fatalf("GetVersion failed with error - %v", err)
	}
	assert.Equal(suite.T(), "1.0.1", version, "Runtime version of IFlow1 is not 1.0.1")
}

func (suite *RuntimeSuite) TestRuntime_UnDeploy() {
	rt := NewRuntime(suite.exe)
	err := rt.UnDeploy("IFlow1")
	if err != nil {
		suite.T().Fatalf("UnDeploy failed with error - %v", err)
	}
}

func setupRuntime(t *testing.T, artifactId string, artifactType string, exe *httpclnt.HTTPExecuter) {
	r := NewRuntime(exe)

	log.Info().Msgf("Checking if runtime artifact %v exists for testing", artifactId)
	version, err := r.GetVersion(artifactId)
	if err != nil {
		t.Fatalf("GetVersion failed with error - %v", err)
	}
	if version == "NOT_DEPLOYED" {
		dt := NewDesigntimeArtifact(artifactType, exe)

		log.Info().Msgf("Setting up runtime artifact %v for testing", artifactId)
		err = dt.Deploy(artifactId)
		if err != nil {
			t.Fatalf("Deploy failed with error - %v", err)
		}
	}
}

func tearDownRuntime(t *testing.T, artifactId string, exe *httpclnt.HTTPExecuter) {
	r := NewRuntime(exe)

	log.Info().Msgf("Checking if artifact %v still exists", artifactId)
	version, err := r.GetVersion(artifactId)
	if err != nil {
		t.Fatalf("get failed with error - %v", err)
	}
	if version != "NOT_DEPLOYED" {
		log.Info().Msgf("Tearing down runtime artifact %v", artifactId)
		err = r.UnDeploy(artifactId)
		if err != nil {
			t.Fatalf("UnDeploy failed with error - %v", err)
		}
	}
}
