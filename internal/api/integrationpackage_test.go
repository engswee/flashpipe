package api

import (
	"github.com/engswee/flashpipe/internal/httpclnt"
	"github.com/engswee/flashpipe/internal/logger"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"slices"
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
			Host:     os.Getenv("FLASHPIPE_TMN_HOST"),
			Userid:   os.Getenv("FLASHPIPE_TMN_USERID"),
			Password: os.Getenv("FLASHPIPE_TMN_PASSWORD"),
		},
	})
}

func TestPackageOauth(t *testing.T) {
	suite.Run(t, &PackageSuite{
		serviceDetails: &ServiceDetails{
			Host:              os.Getenv("FLASHPIPE_TMN_HOST"),
			OauthHost:         os.Getenv("FLASHPIPE_OAUTH_HOST"),
			OauthPath:         os.Getenv("FLASHPIPE_OAUTH_PATH"),
			OauthClientId:     os.Getenv("FLASHPIPE_OAUTH_CLIENTID"),
			OauthClientSecret: os.Getenv("FLASHPIPE_OAUTH_CLIENTSECRET"),
		},
	})
}

func (suite *PackageSuite) SetupSuite() {
	println("========== Setting up suite - start ==========")
	suite.exe = InitHTTPExecuter(suite.serviceDetails)

	// Setup viper in case debug logs are required
	viper.SetEnvPrefix("FLASHPIPE")
	viper.AutomaticEnv()
	logger.InitConsoleLogger(viper.GetBool("debug"))

	setupPackage(suite.T(), "FlashPipeIntegrationTest", suite.exe)

	setupArtifact(suite.T(), "Integration_Test_IFlow", "FlashPipeIntegrationTest", "../../test/testdata/artifacts/create/Integration_Test_IFlow", "Integration", suite.exe)
	println("========== Setting up suite - end ==========")
}

func (suite *PackageSuite) SetupTest() {
	println("---------- Setting up test - start ----------")
	println("---------- Setting up test - end ----------")
}

func (suite *PackageSuite) TearDownTest() {
	println("---------- Tearing down test - start ----------")
	println("---------- Tearing down test - end ----------")
}

func (suite *PackageSuite) TearDownSuite() {
	println("========== Tearing down suite - start ==========")

	tearDownPackage(suite.T(), "FlashPipeIntegrationTest", suite.exe)
	tearDownPackage(suite.T(), "FlashPipeIntegrationTestCreate", suite.exe)
	println("========== Tearing down suite - end ==========")
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
		suite.T().Fatalf("Create failed with error - %v", err)
	}

	// Update
	jsonData.Root.Name = "FlashPipe Integration Test Update"
	jsonData.Root.ShortText = "FlashPipe Integration Test Update"
	jsonData.Root.Mode = "EDIT_ALLOWED"
	err = ip.Update(jsonData)
	if err != nil {
		suite.T().Fatalf("Update failed with error - %v", err)
	}

	// Get list
	packagesList, err := ip.GetPackagesList()
	if err != nil {
		suite.T().Fatalf("GetPackagesList failed with error - %v", err)
	}
	assert.Truef(suite.T(), slices.Contains(packagesList, packageId), "%v found in packagesList", packageId)

	// Check not read only
	_, readOnly, _, err := ip.Get(packageId)
	if err != nil {
		suite.T().Fatalf("IsReadOnly failed with error - %v", err)
	}
	assert.Falsef(suite.T(), readOnly, "%v is not read only", packageId)

	// Delete
	err = ip.Delete(packageId)
	if err != nil {
		suite.T().Fatalf("Delete failed with error - %v", err)
	}
}

func (suite *PackageSuite) TestIntegrationPackage_GetArtifacts() {
	ip := NewIntegrationPackage(suite.exe)

	artifacts, err := ip.GetAllArtifacts("FlashPipeIntegrationTest")
	if err != nil {
		suite.T().Fatalf("GetAllArtifacts failed with error - %v", err)
	}
	artifact := FindArtifactById("Integration_Test_IFlow", artifacts)
	assert.NotNil(suite.T(), artifact, "Integration_Test_IFlow found")
}

func setupPackage(t *testing.T, packageId string, exe *httpclnt.HTTPExecuter) {
	ip := NewIntegrationPackage(exe)

	_, _, packageExists, err := ip.Get(packageId)
	if err != nil {
		t.Logf("WARNING - Exists failed with error - %v", err)
	}
	if !packageExists {
		requestBody := new(PackageSingleData)
		requestBody.Root.Id = packageId
		requestBody.Root.Name = packageId
		requestBody.Root.ShortText = packageId

		err = ip.Create(requestBody)
		if err != nil {
			t.Logf("WARNING - Create failed with error - %v", err)
		}
	}
}

func tearDownPackage(t *testing.T, packageId string, exe *httpclnt.HTTPExecuter) {
	ip := NewIntegrationPackage(exe)

	_, _, packageExists, err := ip.Get(packageId)
	if err != nil {
		t.Logf("WARNING - Exists failed with error - %v", err)
	}
	if packageExists {
		err = ip.Delete(packageId)
		if err != nil {
			t.Logf("WARNING - Delete failed with error - %v", err)
		}
	}
}
