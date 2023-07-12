package designtime

import (
	"fmt"
	"github.com/engswee/flashpipe/httpclnt"
	"github.com/engswee/flashpipe/odata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type DesigntimeSuite struct {
	suite.Suite
	serviceDetails *odata.ServiceDetails
	exe            *httpclnt.HTTPExecuter
}

func TestBasicAuth(t *testing.T) {
	suite.Run(t, &DesigntimeSuite{
		serviceDetails: &odata.ServiceDetails{
			Host:     os.Getenv("HOST_TMN"),
			Userid:   os.Getenv("BASIC_USERID"),
			Password: os.Getenv("BASIC_PASSWORD"),
		},
	})
}

func TestOauth(t *testing.T) {
	suite.Run(t, &DesigntimeSuite{
		serviceDetails: &odata.ServiceDetails{
			Host:              os.Getenv("HOST_TMN"),
			OauthHost:         os.Getenv("HOST_OAUTH"),
			OauthPath:         os.Getenv("HOST_OAUTH_PATH"),
			OauthClientId:     os.Getenv("OAUTH_CLIENTID"),
			OauthClientSecret: os.Getenv("OAUTH_CLIENTSECRET"),
		},
	})
}

func (suite *DesigntimeSuite) SetupSuite() {
	println("Setting up suite")
	suite.exe = odata.InitHTTPExecuter(suite.serviceDetails)
}

func (suite *DesigntimeSuite) SetupTest() {
	println("Setting up test")
}

func (suite *DesigntimeSuite) TearDownTest() {
	println("Tearing down test")
}

func (suite *DesigntimeSuite) TearDownSuite() {
	println("Tearing down suite")
	cleanUpArtifact("Integration", "Integration_Test_IFlow", suite.exe, suite.T())
	cleanUpArtifact("MessageMapping", "Integration_Test_Message_Mapping", suite.exe, suite.T())
	cleanUpArtifact("ScriptCollection", "Integration_Test_Script_Collection", suite.exe, suite.T())
	cleanUpArtifact("ValueMapping", "Integration_Test_Value_Mapping", suite.exe, suite.T())
}

func cleanUpArtifact(artifactType string, artifactId string, exe *httpclnt.HTTPExecuter, t *testing.T) {
	//Check existence
	dt := NewDesigntimeArtifact(artifactType, exe)
	exists, err := dt.Exists(artifactId, "active")
	if err != nil {
		t.Fatalf("Exists failed with error - %v", err)
	}
	if exists {
		err := dt.Delete(artifactId)
		if err != nil {
			t.Fatalf("Delete failed with error - %v", err)
		}
	}
}

func (suite *DesigntimeSuite) TestIntegration_CreateUpdateDeployDelete() {
	dt := NewDesigntimeArtifact("Integration", suite.exe)
	createUpdateDeployDelete("Integration_Test_IFlow", "Integration Test IFlow", "FlashPipeIntegrationTest", dt, suite.T())
}

func (suite *DesigntimeSuite) TestMessageMapping_CreateUpdateDeployDelete() {
	dt := NewDesigntimeArtifact("MessageMapping", suite.exe)
	createUpdateDeployDelete("Integration_Test_Message_Mapping", "Integration Test Message Mapping", "FlashPipeIntegrationTest", dt, suite.T())
}

func (suite *DesigntimeSuite) TestScriptCollection_CreateUpdateDeployDelete() {
	dt := NewDesigntimeArtifact("ScriptCollection", suite.exe)
	createUpdateDeployDelete("Integration_Test_Script_Collection", "Integration Test Script Collection", "FlashPipeIntegrationTest", dt, suite.T())
}

func (suite *DesigntimeSuite) TestValueMapping_CreateUpdateDeployDelete() {
	dt := NewDesigntimeArtifact("ValueMapping", suite.exe)
	createUpdateDeployDelete("Integration_Test_Value_Mapping", "Integration Test Value Mapping", "FlashPipeIntegrationTest", dt, suite.T())
}

func createUpdateDeployDelete(id string, name string, packageId string, dt DesigntimeArtifact, t *testing.T) {
	// Create
	err := dt.Create(id, name, packageId, fmt.Sprintf("../../testdata/artifacts/create/%v", id))
	if err != nil {
		t.Fatalf("Create failed with error - %v", err)
	}
	// Check existence
	exists, err := dt.Exists(id, "active")
	if err != nil {
		t.Fatalf("Exists failed with error - %v", err)
	}
	if assert.True(t, exists, "Expected exists = true") {
		// Update
		err = dt.Update(id, name, packageId, fmt.Sprintf("../../testdata/artifacts/update/%v", id))
		if err != nil {
			t.Fatalf("Update failed with error - %v", err)
		}
		// Check version
		version, err := dt.GetVersion(id, "active")
		if err != nil {
			t.Fatalf("GetVersion failed with error - %v", err)
		}
		if assert.Equal(t, "1.0.1", version, "Expected version = 1.0.1") {
			// Deploy
			err = dt.Deploy(id)
			if err != nil {
				t.Fatalf("Deploy failed with error - %v", err)
			}
			// Get content
			content, err := dt.GetContent(id, "active")
			if err != nil {
				t.Fatalf("GetContent failed with error - %v", err)
			}
			assert.Greater(t, len(content), 0, "Expected len(content) > 0")
			// Delete
			err = dt.Delete(id)
			if err != nil {
				t.Fatalf("Delete failed with error - %v", err)
			}
		}
	}
}
