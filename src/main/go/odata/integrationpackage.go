package odata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/engswee/flashpipe/httpclnt"
	"github.com/engswee/flashpipe/logger"
	"net/http"
)

type IntegrationPackage struct {
	exe *httpclnt.HTTPExecuter
}

type PackageSingleData struct {
	Root struct {
		Id                string `json:"Id"`
		Name              string `json:"Name"`
		Description       string `json:"Description"`
		ShortText         string `json:"ShortText"`
		Version           string `json:"Version"`
		Vendor            string `json:"Vendor,omitempty"`
		Mode              string `json:"Mode,omitempty"`
		SupportedPlatform string `json:"SupportedPlatform"`
		Products          string `json:"Products,omitempty"`
		Keywords          string `json:"Keywords,omitempty"`
		Countries         string `json:"Countries,omitempty"`
		Industries        string `json:"Industries,omitempty"`
		LineOfBusiness    string `json:"LineOfBusiness,omitempty"`
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
	logger.Info("Getting the list of IntegrationPackages")
	path := "/api/v1/IntegrationPackages"

	headers := map[string]string{
		"Accept": "application/json",
	}
	resp, err := ip.exe.ExecGetRequest(path, headers)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, ip.exe.LogError(resp, "Get IntegrationPackages list")
	} else {
		var jsonData *packageMultipleData
		respBody, err := ip.exe.ReadRespBody(resp)
		err = json.Unmarshal(respBody, &jsonData)
		if err != nil {
			return nil, err
		}
		var packageIds []string
		for _, result := range jsonData.Root.Results {
			packageIds = append(packageIds, result.Id)
		}
		return packageIds, nil
	}
}

func (ip *IntegrationPackage) get(id string) (resp *http.Response, err error) {
	path := fmt.Sprintf("/api/v1/IntegrationPackages('%v')", id)

	headers := map[string]string{
		"Accept": "application/json",
	}
	return ip.exe.ExecGetRequest(path, headers)
}

func (ip *IntegrationPackage) IsReadOnly(id string) (bool, error) {
	logger.Info("Checking if package is marked as read only")
	resp, err := ip.get(id)
	if err != nil {
		return false, err
	}
	if resp.StatusCode != 200 {
		return false, ip.exe.LogError(resp, "Get IntegrationPackages by ID")
	} else {
		var jsonData *PackageSingleData
		respBody, err := ip.exe.ReadRespBody(resp)
		err = json.Unmarshal(respBody, &jsonData)
		if err != nil {
			return false, err
		}
		if jsonData.Root.Mode == "READ_ONLY" {
			return true, nil
		} else {
			return false, nil
		}
	}
}

func (ip *IntegrationPackage) Exists(id string) (bool, error) {
	logger.Info(fmt.Sprintf("Checking existence of package %v", id))
	resp, err := ip.get(id)
	if err != nil {
		return false, err
	}
	if resp.StatusCode == 200 {
		return true, nil
	} else if resp.StatusCode == 404 {
		return false, nil
	} else {
		return false, ip.exe.LogError(resp, "Get IntegrationPackages by ID")
	}
}

func (ip *IntegrationPackage) getArtifactsByType(id string, artifactType string) (resp *http.Response, err error) {
	path := fmt.Sprintf("/api/v1/IntegrationPackages('%v')/%vDesigntimeArtifacts", id, artifactType)

	headers := map[string]string{
		"Accept": "application/json",
	}
	return ip.exe.ExecGetRequest(path, headers)
}

func (ip *IntegrationPackage) GetArtifactsData(id string, artifactType string) ([]*ArtifactDetails, error) {
	resp, err := ip.getArtifactsByType(id, artifactType)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, ip.exe.LogError(resp, fmt.Sprintf("Get %v designtime artifacts of IntegrationPackages", artifactType))
	} else {
		var jsonData *artifactData
		respBody, err := ip.exe.ReadRespBody(resp)
		err = json.Unmarshal(respBody, &jsonData)
		if err != nil {
			return nil, err
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
				ArtifactType: artifactType,
			})
		}
		return details, nil
	}
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

	return details, nil
}

func (ip *IntegrationPackage) Create(packageData *PackageSingleData) error {
	path := "/api/v1/IntegrationPackages"

	headers, cookies, err := InitHeadersAndCookies(ip.exe)
	if err != nil {
		return err
	}
	headers["Accept"] = "application/json"
	headers["Content-Type"] = "application/json"

	requestBody, err := ip.constructBody(packageData)
	if err != nil {
		return err
	}

	resp, err := ip.exe.ExecRequestWithCookies("POST", path, bytes.NewReader(requestBody), headers, cookies)
	if err != nil {
		return err
	}
	if resp.StatusCode != 201 {
		return ip.exe.LogError(resp, "Create integration package")
	}
	return nil
}

func (ip *IntegrationPackage) Update(packageData *PackageSingleData) error {
	packageId := packageData.Root.Id
	path := fmt.Sprintf("/api/v1/IntegrationPackages('%v')", packageId)

	headers, cookies, err := InitHeadersAndCookies(ip.exe)
	if err != nil {
		return err
	}
	headers["Accept"] = "application/json"
	headers["Content-Type"] = "application/json"

	requestBody, err := ip.constructBody(packageData)
	if err != nil {
		return err
	}

	resp, err := ip.exe.ExecRequestWithCookies("PUT", path, bytes.NewReader(requestBody), headers, cookies)
	if err != nil {
		return err
	}
	if resp.StatusCode != 202 {
		return ip.exe.LogError(resp, "Update integration package")
	}
	return nil
}

func (ip *IntegrationPackage) Delete(packageId string) error {
	path := fmt.Sprintf("/api/v1/IntegrationPackages('%v')", packageId)

	headers, cookies, err := InitHeadersAndCookies(ip.exe)
	if err != nil {
		return err
	}
	headers["Accept"] = "application/json"
	headers["Content-Type"] = "application/json"

	resp, err := ip.exe.ExecRequestWithCookies("DELETE", path, http.NoBody, headers, cookies)
	if err != nil {
		return err
	}
	if resp.StatusCode != 202 {
		return ip.exe.LogError(resp, "Delete integration package")
	}
	return nil
}

func (ip *IntegrationPackage) constructBody(packageData *PackageSingleData) ([]byte, error) {
	// Clear Mode field as it is not allowed in create/update
	packageData.Root.Mode = "" // TODO - need integration test for this

	jsonBody, err := json.Marshal(packageData)
	if err != nil {
		return nil, err
	}
	logger.Debug(fmt.Sprintf("Request body = %s", jsonBody))
	return jsonBody, nil
}

func FindArtifactById(key string, list []*ArtifactDetails) *ArtifactDetails {
	for _, s := range list {
		if s.Id == key {
			return s
		}
	}
	return nil
}
