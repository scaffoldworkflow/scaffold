package apply

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"scaffold/client/auth"
	"scaffold/client/logger"

	"gopkg.in/yaml.v3"
)

func DoApply(profile, fileName string) {
	p := auth.ReadProfile(profile)

	uri := fmt.Sprintf("%s://%s:%s", p.Protocol, p.Host, p.Port)

	var yamlData map[string]interface{}

	fileData, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(fileData, &yamlData)
	if err != nil {
		panic(err)
	}

	exists := true

	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("%s/api/v1/cascade/%s", uri, yamlData["name"].(string))
	req, _ := http.NewRequest("GET", requestURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", p.APIToken))
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Fatalf("", "Encountered error checking for cascade existence: %v", err)
	}
	if resp.StatusCode >= 400 {
		logger.Debugf("", "Got status code %d, cascade exists", resp.StatusCode)
		exists = false
	}

	postBody, _ := json.Marshal(yamlData)
	postBodyBuffer := bytes.NewBuffer(postBody)

	if exists {
		logger.Debugf("", "Already exists!")

		httpClient := &http.Client{}
		requestURL := fmt.Sprintf("%s/api/v1/cascade/%s", uri, yamlData["name"].(string))
		req, _ := http.NewRequest("PUT", requestURL, postBodyBuffer)
		req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", p.APIToken))
		req.Header.Set("Content-Type", "application/json")
		resp, err := httpClient.Do(req)
		if err != nil {
			logger.Fatalf("", "Put request failed with error: %s", err.Error())
		}
		if resp.StatusCode >= 400 {
			logger.Fatalf("", "Put request failed with status code %v", resp.StatusCode)
		}
	} else {
		logger.Debugf("", "Doesn't exist")
		httpClient := &http.Client{}
		requestURL := fmt.Sprintf("%s/api/v1/cascade", uri)
		req, _ := http.NewRequest("POST", requestURL, postBodyBuffer)
		req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", p.APIToken))
		req.Header.Set("Content-Type", "application/json")
		resp, err := httpClient.Do(req)
		if err != nil {
			logger.Fatalf("", "Put request failed with error: %s", err.Error())
		}
		if resp.StatusCode >= 400 {
			logger.Fatalf("", "Put request failed with status code %v", resp.StatusCode)
		}
	}
	logger.Successf("", "Successfully uploaded cascade!")
}
