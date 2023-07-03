package odata

import (
	"encoding/json"
	"fmt"
	"github.com/engswee/flashpipe/httpclnt"
	"github.com/engswee/flashpipe/logger"
	"net/http"
)

type IntegrationPackage struct {
	exe *httpclnt.HTTPExecuter
}

type packageSingleData struct {
	Root struct {
		Id                string `json:"Id"`
		Name              string `json:"Name"`
		Description       string `json:"Description"`
		ShortText         string `json:"ShortText"`
		Version           string `json:"Version"`
		Vendor            string `json:"Vendor"`
		Mode              string `json:"Mode"`
		SupportedPlatform string `json:"SupportedPlatform"`
		Products          string `json:"Products"`
		Keywords          string `json:"Keywords"`
		Countries         string `json:"Countries"`
		Industries        string `json:"Industries"`
		LineOfBusiness    string `json:"LineOfBusiness"`
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

func (ip *IntegrationPackage) Get(id string) (resp *http.Response, err error) {
	path := fmt.Sprintf("/api/v1/IntegrationPackages('%v')", id)

	headers := map[string]string{
		"Accept": "application/json",
	}
	return ip.exe.ExecGetRequest(path, headers)
}

func (ip *IntegrationPackage) IsReadOnly(id string) (bool, error) {
	logger.Info("Checking if package is marked as read only")
	resp, err := ip.Get(id)
	if err != nil {
		return false, err
	}
	if resp.StatusCode != 200 {
		return false, ip.exe.LogError(resp, "Get IntegrationPackages by ID")
	} else {
		var jsonData *packageSingleData
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

func (ip *IntegrationPackage) GetArtifactsByType(id string, artifactType string) (resp *http.Response, err error) {
	path := fmt.Sprintf("/api/v1/IntegrationPackages('%v')/%vDesigntimeArtifacts", id, artifactType)

	headers := map[string]string{
		"Accept": "application/json",
	}
	return ip.exe.ExecGetRequest(path, headers)
}

func (ip *IntegrationPackage) GetArtifactsData(id string, artifactType string) ([]*ArtifactDetails, error) {
	//logger.Info("Checking if package is marked as read only")
	resp, err := ip.GetArtifactsByType(id, artifactType)
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
			panic(err)
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
