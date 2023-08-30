package odata

import (
	"github.com/engswee/flashpipe/internal/httpclnt"
	"github.com/engswee/flashpipe/internal/logger"
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
			Host:     os.Getenv("FLASHPIPE_TMN_HOST"),
			Userid:   os.Getenv("FLASHPIPE_TMN_USERID"),
			Password: os.Getenv("FLASHPIPE_TMN_PASSWORD"),
		},
	})
}

func TestRuntimeOauth(t *testing.T) {
	suite.Run(t, &RuntimeSuite{
		serviceDetails: &ServiceDetails{
			Host:              os.Getenv("FLASHPIPE_TMN_HOST"),
			OauthHost:         os.Getenv("FLASHPIPE_OAUTH_HOST"),
			OauthPath:         os.Getenv("FLASHPIPE_OAUTH_PATH"),
			OauthClientId:     os.Getenv("FLASHPIPE_OAUTH_CLIENTID"),
			OauthClientSecret: os.Getenv("FLASHPIPE_OAUTH_CLIENTSECRET"),
		},
	})
}

func (suite *RuntimeSuite) SetupSuite() {
	println("========== Setting up suite - start ==========")
	suite.exe = InitHTTPExecuter(suite.serviceDetails)

	// Setup viper in case debug logs are required
	viper.SetEnvPrefix("FLASHPIPE")
	viper.AutomaticEnv()
	logger.InitConsoleLogger(viper.GetBool("debug"))

	setupPackage(suite.T(), "FlashPipeIntegrationTest", suite.exe)

	setupArtifact(suite.T(), "Integration_Test_IFlow", "FlashPipeIntegrationTest", "../../test/testdata/artifacts/create/Integration_Test_IFlow", "Integration", suite.exe)
	setupRuntime(suite.T(), "Integration_Test_IFlow", "Integration", suite.exe)

	setupArtifact(suite.T(), "Integration_Test_Message_Mapping", "FlashPipeIntegrationTest", "../../test/testdata/artifacts/create/Integration_Test_Message_Mapping", "MessageMapping", suite.exe)
	setupRuntime(suite.T(), "Integration_Test_Message_Mapping", "MessageMapping", suite.exe)
	println("========== Setting up suite - end ==========")
}

func (suite *RuntimeSuite) SetupTest() {
	println("---------- Setting up test - start ----------")
	println("---------- Setting up test - end ----------")
}

func (suite *RuntimeSuite) TearDownTest() {
	println("---------- Tearing down test - start ----------")
	println("---------- Tearing down test - end ----------")
}

func (suite *RuntimeSuite) TearDownSuite() {
	println("========== Tearing down suite - start ==========")

	tearDownPackage(suite.T(), "FlashPipeIntegrationTest", suite.exe)

	tearDownRuntime(suite.T(), "Integration_Test_IFlow", suite.exe)
	tearDownRuntime(suite.T(), "Integration_Test_Message_Mapping", suite.exe)
	println("========== Tearing down suite - end ==========")
}

func (suite *RuntimeSuite) TestRuntime_GetErrorInfo() {
	rt := NewRuntime(suite.exe)
	errorMessage, err := rt.GetErrorInfo("Integration_Test_Message_Mapping")
	if err != nil {
		suite.T().Fatalf("GetErrorInfo failed with error - %v", err)
	}
	assert.Contains(suite.T(), errorMessage, "Validation of the artifact failed", "errorMessage does not have validation error")
}

func (suite *RuntimeSuite) TestRuntime_Get() {
	rt := NewRuntime(suite.exe)
	version, status, err := rt.Get("Integration_Test_IFlow")
	if err != nil {
		suite.T().Fatalf("Get failed with error - %v", err)
	}
	if status == "STARTING" {
		time.Sleep(10 * time.Second)
		version, status, err = rt.Get("Integration_Test_IFlow")
		if err != nil {
			suite.T().Fatalf("Get failed with error - %v", err)
		}
	}
	assert.Equal(suite.T(), "STARTED", status, "Runtime status of Integration_Test_IFlow is not STARTED")
	assert.Equal(suite.T(), "1.0.0", version, "Runtime version of Integration_Test_IFlow is not 1.0.0")
}

func (suite *RuntimeSuite) TestRuntime_UnDeploy() {
	rt := NewRuntime(suite.exe)
	err := rt.UnDeploy("Integration_Test_IFlow")
	if err != nil {
		suite.T().Fatalf("UnDeploy failed with error - %v", err)
	}
}

func setupRuntime(t *testing.T, artifactId string, artifactType string, exe *httpclnt.HTTPExecuter) {
	r := NewRuntime(exe)

	_, status, err := r.Get(artifactId)
	if err != nil {
		t.Fatalf("Get failed with error - %v", err)
	}
	if status != "STARTED" {
		dt := NewDesigntimeArtifact(artifactType, exe)

		err = dt.Deploy(artifactId)
		if err != nil {
			t.Fatalf("Deploy failed with error - %v", err)
		}
		time.Sleep(10 * time.Second)
	}
}

func tearDownRuntime(t *testing.T, artifactId string, exe *httpclnt.HTTPExecuter) {
	r := NewRuntime(exe)

	version, _, err := r.Get(artifactId)
	if err != nil {
		t.Fatalf("get failed with error - %v", err)
	}
	if version != "NOT_DEPLOYED" {
		err = r.UnDeploy(artifactId)
		if err != nil {
			t.Fatalf("UnDeploy failed with error - %v", err)
		}
	}
}
