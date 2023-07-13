package odata

import (
	"github.com/engswee/flashpipe/httpclnt"
	"github.com/engswee/flashpipe/str"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type PackageSuite struct {
	suite.Suite
	serviceDetails *ServiceDetails
	exe            *httpclnt.HTTPExecuter
}

func TestPackageBasicAuth(t *testing.T) {
	suite.Run(t, &PackageSuite{
		serviceDetails: &ServiceDetails{
			Host:     os.Getenv("HOST_TMN"),
			Userid:   os.Getenv("BASIC_USERID"),
			Password: os.Getenv("BASIC_PASSWORD"),
		},
	})
}

func TestPackageOauth(t *testing.T) {
	suite.Run(t, &PackageSuite{
		serviceDetails: &ServiceDetails{
			Host:              os.Getenv("HOST_TMN"),
			OauthHost:         os.Getenv("HOST_OAUTH"),
			OauthPath:         os.Getenv("HOST_OAUTH_PATH"),
			OauthClientId:     os.Getenv("OAUTH_CLIENTID"),
			OauthClientSecret: os.Getenv("OAUTH_CLIENTSECRET"),
		},
	})
}

func (suite *PackageSuite) SetupSuite() {
	println("Setting up suite")
	suite.exe = InitHTTPExecuter(suite.serviceDetails)
	ip := NewIntegrationPackage(suite.exe)
	const packageId = "FlashPipeIntegrationTest"
	exists, err := ip.Exists(packageId)
	if err != nil {
		suite.T().Fatalf("Exists failed with error - %v", err)
	}
	if !exists {
		requestBody := new(PackageSingleData)
		requestBody.Root.Id = packageId
		requestBody.Root.Name = packageId
		requestBody.Root.ShortText = packageId

		// Create
		err = ip.Create(requestBody)
		if err != nil {
			suite.T().Fatalf("Create package failed with error - %v", err)
		}
	}
	//const artifactId = "IFlow1"
	//dt := designtime.NewDesigntimeArtifact("Integration", suite.exe)
	//exists, err = dt.Exists(artifactId, "active")
	//if err != nil {
	//	suite.T().Fatalf("Exists failed with error - %v", err)
	//}
	//if !exists {
	//	err := dt.Create(artifactId, artifactId, packageId, fmt.Sprintf("../../testdata/artifacts/setup/%v", artifactId))
	//	if err != nil {
	//		suite.T().Fatalf("Create integration designtime artifact failed with error - %v", err)
	//	}
	//}
}

func (suite *PackageSuite) SetupTest() {
	println("Setting up test")
}

func (suite *PackageSuite) TearDownTest() {
	println("Tearing down test")
}

func (suite *PackageSuite) TearDownSuite() {
	println("Tearing down suite")
	const packageId = "FlashPipeIntegrationTestCreate"
	ip := NewIntegrationPackage(suite.exe)
	exists, err := ip.Exists(packageId)
	if err != nil {
		suite.T().Fatalf("Exists failed with error - %v", err)
	}
	if exists {
		err := ip.Delete(packageId)
		if err != nil {
			suite.T().Fatalf("Delete failed with error - %v", err)
		}
	}
}

func (suite *PackageSuite) TestIntegrationPackage_CreateUpdateDelete() {
	const packageId = "FlashPipeIntegrationTestCreate"
	ip := NewIntegrationPackage(suite.exe)

	jsonData := new(PackageSingleData)
	jsonData.Root.Id = packageId
	jsonData.Root.Name = "FlashPipe Integration Test Create"
	jsonData.Root.ShortText = "FlashPipe Integration Test Create"
	jsonData.Root.Mode = "EDIT_ALLOWED"
	// Create
	err := ip.Create(jsonData)
	if err != nil {
		suite.T().Fatalf("Create package failed with error - %v", err)
	}

	// Update
	jsonData.Root.Name = "FlashPipe Integration Test Update"
	jsonData.Root.Name = "FlashPipe Integration Test Update"
	jsonData.Root.Mode = "EDIT_ALLOWED"
	err = ip.Update(jsonData)
	if err != nil {
		suite.T().Fatalf("Update package failed with error - %v", err)
	}

	// Get list
	packagesList, err := ip.GetPackagesList()
	if err != nil {
		suite.T().Fatalf("Get packages failed with error - %v", err)
	}
	assert.Truef(suite.T(), str.Contains(packageId, packagesList), "%v found in packagesList", packageId)

	// Check not read only
	readOnly, err := ip.IsReadOnly(packageId)
	if err != nil {
		return
	}
	assert.Falsef(suite.T(), readOnly, "%v is not read only", packageId)

	// Delete
	err = ip.Delete(packageId)
	if err != nil {
		suite.T().Fatalf("Delete package failed with error - %v", err)
	}
}

func (suite *PackageSuite) TestIntegrationPackage_GetArtifacts() {
	ip := NewIntegrationPackage(suite.exe)

	artifacts, err := ip.GetAllArtifacts("FlashPipeIntegrationTest")
	if err != nil {
		return
	}
	artifact := FindArtifactById("IFlow1", artifacts)
	assert.NotNil(suite.T(), artifact, "IFlow1 found")
}
