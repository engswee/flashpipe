package odata

import (
	"github.com/engswee/flashpipe/httpclnt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"strings"
	"testing"
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
	println("Setting up suite")
	suite.exe = InitHTTPExecuter(suite.serviceDetails)
}

func (suite *RuntimeSuite) SetupTest() {
	println("Setting up test")
}

func (suite *RuntimeSuite) TearDownTest() {
	println("Tearing down test")
}

func (suite *RuntimeSuite) TearDownSuite() {
	println("Tearing down suite")

}

func (suite *RuntimeSuite) TestRuntime_GetStatusVersion() {
	rt := NewRuntime(suite.exe)
	status, err := rt.GetStatus("IFlow1")
	if err != nil {
		suite.T().Fatalf("HTTP call failed with error - %v", err)
	}
	assert.Equal(suite.T(), "STARTED", status, "Runtime status of IFlow1 is not STARTED")
	version, err := rt.GetVersion("IFlow1")
	if err != nil {
		suite.T().Fatalf("HTTP call failed with error - %v", err)
	}
	assert.Equal(suite.T(), "1.0.1", version, "Runtime version of IFlow1 is not 1.0.1")
}

func (suite *RuntimeSuite) TestRuntime_GetErrorInfo() {
	rt := NewRuntime(suite.exe)
	errorMessage, err := rt.GetErrorInfo("Mapping1")
	if err != nil {
		suite.T().Fatalf("HTTP call failed with error - %v", err)
	}
	if !strings.HasPrefix(errorMessage, "Validation of the artifact failed") {
		suite.T().Fatalf("errorMessage does not have prefix")
	}
}

//func (suite *RuntimeSuite) TestRuntime_UnDeploy() {
//	rt := NewRuntime(suite.exe)
//	err := rt.UnDeploy("Hello")
//	if err != nil {
//		suite.T().Fatalf("HTTP call failed with error - %v", err)
//	}
//}
