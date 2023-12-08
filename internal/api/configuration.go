package api

import (
	"encoding/json"
	"fmt"
	"github.com/engswee/flashpipe/internal/httpclnt"
	"github.com/rs/zerolog/log"
	"net/url"
)

type Configuration struct {
	exe *httpclnt.HTTPExecuter
}

type ParametersData struct {
	Root struct {
		Results []*ParameterData `json:"results"`
	} `json:"d"`
}

type ParameterData struct {
	ParameterKey   string `json:"ParameterKey,omitempty"`
	ParameterValue string `json:"ParameterValue"`
	DataType       string `json:"DataType,omitempty"`
}

func NewConfiguration(exe *httpclnt.HTTPExecuter) *Configuration {
	c := new(Configuration)
	c.exe = exe
	return c
}

func (c *Configuration) Get(id string, version string) (*ParametersData, error) {
	log.Info().Msgf("Getting configuration parameters of Integration designtime artifact %v", id)
	urlPath := fmt.Sprintf("/api/v1/IntegrationDesigntimeArtifacts(Id='%v',Version='%v')/Configurations", id, version)

	callType := "Get configuration parameters"
	resp, err := readOnlyCall(urlPath, callType, c.exe)
	if err != nil {
		return nil, err
	}
	var jsonData *ParametersData
	respBody, err := c.exe.ReadRespBody(resp)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(respBody, &jsonData)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

func (c *Configuration) Update(id string, version string, key string, value string) error {
	log.Info().Msgf("Updating configuration parameter %v of Integration designtime artifact %v", key, id)
	// Spaces in key needs to be escaped
	encodedKey := url.PathEscape(key)
	urlPath := fmt.Sprintf("/api/v1/IntegrationDesigntimeArtifacts(Id='%v',Version='%v')/$links/Configurations('%v')", id, version, encodedKey)

	parameterData := &ParameterData{ParameterValue: value}
	requestBody, err := json.Marshal(parameterData)
	if err != nil {
		return err
	}

	return modifyingCall("PUT", urlPath, requestBody, 202, fmt.Sprintf("Update configuration parameter %v", key), c.exe)
}

func FindParameterByKey(key string, list []*ParameterData) *ParameterData {
	for _, s := range list {
		if s.ParameterKey == key {
			return s
		}
	}
	return nil
}
