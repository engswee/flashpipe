package designtime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/engswee/flashpipe/httpclnt"
	"github.com/engswee/flashpipe/logger"
	"github.com/engswee/flashpipe/odata"
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
	path := fmt.Sprintf("/api/v1/IntegrationDesigntimeArtifacts(Id='%v',Version='%v')/Configurations", id, version)
	headers := map[string]string{
		"Accept": "application/json",
	}
	resp, err := c.exe.ExecGetRequest(path, headers)

	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, c.exe.LogError(resp, fmt.Sprintf("Get configuration parameters"))
	} else {
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
}

func (c *Configuration) Update(id string, version string, key string, value string) error {
	path := fmt.Sprintf("/api/v1/IntegrationDesigntimeArtifacts(Id='%v',Version='%v')/$links/Configurations('%v')", id, version, key)
	headers, cookies, err := odata.InitHeadersAndCookies(c.exe)
	if err != nil {
		return err
	}
	headers["Accept"] = "application/json"
	headers["Content-Type"] = "application/json"

	parameterData := &ParameterData{
		ParameterValue: value,
	}
	jsonBody, err := json.Marshal(parameterData)
	if err != nil {
		return err
	}
	logger.Debug(fmt.Sprintf("Request body = %s", jsonBody))

	resp, err := c.exe.ExecRequestWithCookies("PUT", path, bytes.NewReader(jsonBody), headers, cookies)
	if err != nil {
		return err
	}
	if resp.StatusCode != 202 {
		return c.exe.LogError(resp, fmt.Sprintf("Update configuration parameter %v", key))
	}
	return nil
}

func FindParameterByKey(key string, list []*ParameterData) *ParameterData {
	for _, s := range list {
		if s.ParameterKey == key {
			return s
		}
	}
	return nil
}
