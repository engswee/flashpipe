package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/engswee/flashpipe/internal/file"
	"github.com/engswee/flashpipe/internal/httpclnt"
	"github.com/go-errors/errors"
	"github.com/rs/zerolog/log"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

type APIProxy struct {
	exe *httpclnt.HTTPExecuter
}

type APIProxyQuery struct {
	Root struct {
		Proxies struct {
			Entities []*APIEntity `json:"entities"`
		} `json:"apiproxies"`
	} `json:"selection"`
}

type APIEntity struct {
	Name string `json:"name"`
}

type apiProxyResponseData struct {
	Root struct {
		Results []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
			Status  string `json:"state"`
		} `json:"results"`
	} `json:"d"`
}

type APIProxyMetadata struct {
	Name    string
	Version string
	Status  string
}

func NewAPIProxy(exe *httpclnt.HTTPExecuter) *APIProxy {
	a := new(APIProxy)
	a.exe = exe
	return a
}

func (a *APIProxy) Download(apiName string, targetRootDir string) error {
	log.Info().Msgf("Downloading APIProxy %v", apiName)
	urlPath := fmt.Sprintf("/apiportal/api/1.0/ContentArchive.svc")

	// Construct query for payload body
	query := &APIProxyQuery{}
	query.Root.Proxies.Entities = append(query.Root.Proxies.Entities, &APIEntity{Name: apiName})
	requestBody, err := json.Marshal(query)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	callType := fmt.Sprintf("Get APIProxy")
	resp, err := readOnlyCallWithBody(urlPath, requestBody, callType, a.exe)
	if err != nil {
		return err
	}

	targetDir := fmt.Sprintf("%v/%v", targetRootDir, apiName)
	targetFile := fmt.Sprintf("%v/%v.zip", targetDir, apiName)
	// Create directory for target file if it doesn't exist yet
	err = os.MkdirAll(filepath.Dir(targetFile), os.ModePerm)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	content, err := a.exe.ReadRespBody(resp)
	err = os.WriteFile(targetFile, content, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	log.Info().Msgf("Unzipping contents to %v", targetDir)
	err = file.UnzipSource(targetFile, targetDir)
	if err != nil {
		return err
	}
	err = os.Remove(targetFile)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	return nil
}

func (a *APIProxy) Upload(sourceDir string, workDir string) error {
	targetZipFilePath := filepath.Clean(workDir) + string(os.PathSeparator) + filepath.Base(sourceDir) + ".zip"
	log.Debug().Msgf("Compressing contents of directory %v to file %v", sourceDir, targetZipFilePath)
	err := file.ZipDir(sourceDir, targetZipFilePath, false)
	if err != nil {
		return err
	}

	log.Info().Msgf("Uploading API content from %v", targetZipFilePath)
	// Construct multipart form body and content type
	body, cType, err := createFormDataFileRequest(nil, "file", targetZipFilePath)
	if err != nil {
		return err
	}

	urlPath := fmt.Sprintf("/apiportal/api/1.0/ContentArchive.svc")
	err = modifyingCallWithContentType("POST", urlPath, body.Bytes(), cType, 200, "Upload API ContentArchive", a.exe)
	if err != nil {
		return err
	}

	return nil
}

func (a *APIProxy) Get(id string) (bool, error) {
	log.Info().Msgf("Getting details of APIProxy %v", id)
	urlPath := fmt.Sprintf("/apiportal/api/1.0/Management.svc/APIProxies('%v')", id)

	callType := fmt.Sprintf("Get APIProxy")
	_, err := readOnlyCall(urlPath, callType, a.exe)
	if err != nil {
		if err.Error() == fmt.Sprintf("%v call failed with response code = 404", callType) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func (a *APIProxy) List() ([]*APIProxyMetadata, error) {
	log.Info().Msgf("Getting list of APIProxies")
	urlPath := fmt.Sprintf("/apiportal/api/1.0/Management.svc/APIProxies")

	callType := fmt.Sprintf("List APIProxies")
	resp, err := readOnlyCall(urlPath, callType, a.exe)
	// Process response to extract proxy details
	var jsonData *apiProxyResponseData
	respBody, err := a.exe.ReadRespBody(resp)
	err = json.Unmarshal(respBody, &jsonData)
	if err != nil {
		log.Warn().Msgf("⚠️ Please check that hostname and credentials for APIM are correct - do not use CPI values!")
		log.Error().Msgf("Error unmarshalling response as JSON. Response body = %s", respBody)
		return nil, errors.Wrap(err, 0)
	}
	var details []*APIProxyMetadata
	for _, result := range jsonData.Root.Results {
		details = append(details, &APIProxyMetadata{
			Name:    result.Name,
			Version: result.Version,
			Status:  result.Status,
		})
	}
	return details, nil
}

func (a *APIProxy) Delete(id string) error {
	log.Info().Msgf("Deleting APIProxy %v", id)

	urlPath := fmt.Sprintf("/apiportal/api/1.0/Management.svc/APIProxies('%v')", id)
	return modifyingCall("DELETE", urlPath, nil, 204, fmt.Sprintf("Delete APIProxy"), a.exe)
}

func createFormDataFileRequest(formDataParameters map[string]string, fileParameterName, inputFilePath string) (*bytes.Buffer, string, error) {
	inputFile, err := os.Open(inputFilePath)
	if err != nil {
		return nil, "", errors.Wrap(err, 0)
	}
	defer inputFile.Close()

	body := &bytes.Buffer{}
	multipartWriter := multipart.NewWriter(body)
	dstFileWriter, err := multipartWriter.CreateFormFile(fileParameterName, filepath.Base(inputFilePath))
	if err != nil {
		return nil, "", errors.Wrap(err, 0)
	}
	_, err = io.Copy(dstFileWriter, inputFile)
	if err != nil {
		return nil, "", errors.Wrap(err, 0)
	}

	for key, val := range formDataParameters {
		_ = multipartWriter.WriteField(key, val)
	}
	err = multipartWriter.Close()
	if err != nil {
		return nil, "", errors.Wrap(err, 0)
	}

	return body, multipartWriter.FormDataContentType(), err
}
