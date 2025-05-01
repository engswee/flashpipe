package api

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/engswee/flashpipe/internal/httpclnt"
	"github.com/go-errors/errors"
	"github.com/rs/zerolog/log"
)

type IntegrationPackage struct {
	exe *httpclnt.HTTPExecuter
}

type PackageSingleData struct {
	Root struct {
		Id             string `json:"Id"`
		Name           string `json:"Name"`
		Description    string `json:"Description"`
		ShortText      string `json:"ShortText"`
		Version        string `json:"Version"`
		Vendor         string `json:"Vendor,omitempty"`
		Mode           string `json:"Mode,omitempty"`
		Products       string `json:"Products,omitempty"`
		Keywords       string `json:"Keywords,omitempty"`
		Countries      string `json:"Countries,omitempty"`
		Industries     string `json:"Industries,omitempty"`
		LineOfBusiness string `json:"LineOfBusiness,omitempty"`
	} `json:"d"`
}

type artifactData struct {
	Root struct {
		Results []struct {
			Id      string `json:"Id"`
			Name    string `json:"Name"`
			Version string `json:"Version"`
		} `json:"results"`
	} `json:"d"`
}

type packageMultipleData struct {
	Root struct {
		Results []struct {
			Id string `json:"Id"`
		} `json:"results"`
	} `json:"d"`
}

type ArtifactDetails struct {
	Id           string
	Name         string
	IsDraft      bool
	Version      string
	ArtifactType string
}

// NewIntegrationPackage returns an initialised IntegrationPackage instance.
func NewIntegrationPackage(exe *httpclnt.HTTPExecuter) *IntegrationPackage {
	ip := new(IntegrationPackage)
	ip.exe = exe
	return ip
}

func (ip *IntegrationPackage) GetPackagesList() ([]string, error) {
	// Get the list of packages of the current tenant
	log.Info().Msg("Getting list of IntegrationPackages")
	urlPath := "/api/v1/IntegrationPackages"

	callType := "Get IntegrationPackages list"
	resp, err := readOnlyCall(urlPath, callType, ip.exe)
	if err != nil {
		return nil, err
	}
	// Process response to extract packages
	var jsonData *packageMultipleData
	respBody, err := ip.exe.ReadRespBody(resp)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(respBody, &jsonData)
	if err != nil {
		log.Error().Msgf("Error unmarshalling response as JSON. Response body = %s", respBody)
		return nil, errors.Wrap(err, 0)
	}
	var packageIds []string
	for _, result := range jsonData.Root.Results {
		packageIds = append(packageIds, result.Id)
	}
	return packageIds, nil
}

func (ip *IntegrationPackage) Get(id string) (packageData *PackageSingleData, readOnly bool, exists bool, err error) {
	log.Info().Msgf("Getting details of integration package %v", id)
	urlPath := fmt.Sprintf("/api/v1/IntegrationPackages('%v')", id)

	callType := "Get IntegrationPackages by ID"
	resp, err := readOnlyCall(urlPath, callType, ip.exe)
	if err != nil {
		if err.Error() == fmt.Sprintf("%v call failed with response code = 404", callType) {
			return nil, false, false, nil
		} else {
			return nil, false, false, err
		}
	}
	// Process response to extract details
	respBody, err := ip.exe.ReadRespBody(resp)
	if err != nil {
		return nil, false, false, err
	}
	err = json.Unmarshal(respBody, &packageData)
	if err != nil {
		log.Error().Msgf("Error unmarshalling response as JSON. Response body = %s", respBody)
		return nil, false, false, errors.Wrap(err, 0)
	}
	if packageData.Root.Mode == "READ_ONLY" {
		readOnly = true
	}
	return packageData, readOnly, true, nil
}

func (ip *IntegrationPackage) GetArtifactsData(id string, artifactType string) ([]*ArtifactDetails, error) {
	log.Info().Msgf("Getting %v designtime artifacts of package %v", artifactType, id)
	urlPath := fmt.Sprintf("/api/v1/IntegrationPackages('%v')/%vDesigntimeArtifacts", id, artifactType)

	callType := fmt.Sprintf("Get %v designtime artifacts of IntegrationPackages", artifactType)
	resp, err := readOnlyCall(urlPath, callType, ip.exe)
	if err != nil {
		return nil, err
	}
	// Process response to extract artifact details
	var jsonData *artifactData
	respBody, err := ip.exe.ReadRespBody(resp)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(respBody, &jsonData)
	if err != nil {
		log.Error().Msgf("Error unmarshalling response as JSON. Response body = %s", respBody)
		return nil, errors.Wrap(err, 0)
	}
	var details []*ArtifactDetails
	for _, result := range jsonData.Root.Results {
		var draft bool
		if result.Version == "Active" {
			draft = true
		} else {
			draft = false
		}
		details = append(details, &ArtifactDetails{
			Id:           result.Id,
			Name:         result.Name,
			IsDraft:      draft,
			Version:      result.Version,
			ArtifactType: artifactType,
		})
	}
	return details, nil
}

func (ip *IntegrationPackage) GetAllArtifacts(id string) ([]*ArtifactDetails, error) {
	var details []*ArtifactDetails
	integrations, err := ip.GetArtifactsData(id, "Integration")
	if err != nil {
		return nil, err
	}
	details = append(details, integrations...)
	mappings, err := ip.GetArtifactsData(id, "MessageMapping")
	if err != nil {
		return nil, err
	}
	details = append(details, mappings...)
	scripts, err := ip.GetArtifactsData(id, "ScriptCollection")
	if err != nil {
		return nil, err
	}
	details = append(details, scripts...)
	valmaps, err := ip.GetArtifactsData(id, "ValueMapping")
	if err != nil {
		return nil, err
	}
	details = append(details, valmaps...)

	return details, nil
}

func (ip *IntegrationPackage) Create(packageData *PackageSingleData) error {
	packageId := packageData.Root.Id
	log.Info().Msgf("Creating integration package %v", packageId)
	urlPath := "/api/v1/IntegrationPackages"

	requestBody, err := ip.constructBody(packageData)
	if err != nil {
		return err
	}

	return modifyingCall("POST", urlPath, requestBody, 201, "Create integration package", ip.exe)
}

func (ip *IntegrationPackage) Update(packageData *PackageSingleData) error {
	packageId := packageData.Root.Id
	log.Info().Msgf("Updating integration package %v", packageId)
	urlPath := fmt.Sprintf("/api/v1/IntegrationPackages('%v')", packageId)

	requestBody, err := ip.constructBody(packageData)
	if err != nil {
		return err
	}

	return modifyingCall("PUT", urlPath, requestBody, 202, "Update integration package", ip.exe)
}

func (ip *IntegrationPackage) Delete(packageId string) error {
	log.Info().Msgf("Deleting integration package %v", packageId)
	urlPath := fmt.Sprintf("/api/v1/IntegrationPackages('%v')", packageId)
	return modifyingCall("DELETE", urlPath, nil, 202, "Delete integration package", ip.exe)
}

func (ip *IntegrationPackage) constructBody(packageData *PackageSingleData) ([]byte, error) {
	// Clear Mode field as it is not allowed in create/update
	packageData.Root.Mode = ""

	requestBody, err := json.Marshal(packageData)
	if err != nil {
		return nil, err
	}
	return requestBody, nil
}

func FindArtifactById(key string, list []*ArtifactDetails) *ArtifactDetails {
	for _, s := range list {
		if s.Id == key {
			return s
		}
	}
	return nil
}

func GetPackageDetails(file string) (*PackageSingleData, error) {
	var jsonData *PackageSingleData

	fileContent, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(fileContent, &jsonData)
	if err != nil {
		log.Error().Msgf("Error unmarshalling file as JSON. Response body = %s", fileContent)
		return nil, errors.Wrap(err, 0)
	}
	return jsonData, nil
}
