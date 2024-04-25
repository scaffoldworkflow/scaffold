package apply

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"scaffold/client/auth"
	"scaffold/client/constants"
	"scaffold/client/logger"
	"scaffold/client/utils"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

func DoApply(profile, object, context, fileName string) {
	if context == constants.ALL_CONTEXTS {
		logger.Fatalf("", "%s is not allowed for delete actions", constants.ALL_CONTEXTS)
	}

	logger.Debugf("", "Applying object")
	p := auth.ReadProfile(profile)
	uri := fmt.Sprintf("%s://%s:%s", p.Protocol, p.Host, p.Port)

	logger.Debugf("", "Checking if object is valid")
	objects := []string{"cascade", "datastore", "state", "task", "file", "user", "input"}

	if !utils.Contains(objects, object) {
		logger.Fatalf("", "Invalid object type passed: '%s'. Valid object types are %v", object, objects)
	}

	logger.Debugf("", "Getting context")
	if context == "" {
		context = p.Cascade
	}

	doApply(profile, fileName, context, uri, object)
}

func checkExists(p auth.ProfileObj, uri, name, object string) (bool, error) {
	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("%s/api/v1/%s/%s", uri, object, name)
	req, _ := http.NewRequest("GET", requestURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", p.APIToken))
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Errorf("", "Encountered error checking for %s existence: %s", object, err.Error())
		return false, err
	}
	if resp.StatusCode == http.StatusNotFound {
		logger.Debugf("", "Got status code %d, doesn't exist", resp.StatusCode)
		return false, nil
	}

	return true, nil
}

func doUpdate(p auth.ProfileObj, uri, object, name string, data map[string]interface{}) {
	postBody, _ := json.Marshal(data)
	postBodyBuffer := bytes.NewBuffer(postBody)

	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("%s/api/v1/%s/%s", uri, object, name)
	req, _ := http.NewRequest("PUT", requestURL, postBodyBuffer)
	req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", p.APIToken))
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Fatalf("", "PUT request failed with error: %s", err.Error())
	}
	if resp.StatusCode >= 400 {
		logger.Fatalf("", "PUT request failed with status code %v", resp.StatusCode)
	}
}

func doCreate(p auth.ProfileObj, uri, object string, data map[string]interface{}) {
	postBody, _ := json.Marshal(data)
	postBodyBuffer := bytes.NewBuffer(postBody)

	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("%s/api/v1/%s", uri, object)
	req, _ := http.NewRequest("POST", requestURL, postBodyBuffer)
	req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", p.APIToken))
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Fatalf("", "POST request failed with error: %s", err.Error())
	}
	if resp.StatusCode >= 400 {
		logger.Fatalf("", "POST request failed with status code %v", resp.StatusCode)
	}
}

func doApply(profile, fileName, context, uri, objType string) {
	p := auth.ReadProfile(profile)

	var yamlData map[string]interface{}

	fileData, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(fileData, &yamlData)
	if err != nil {
		panic(err)
	}

	name := yamlData["name"].(string)

	if objType != "cascade" && objType != "datastore" && objType != "user" {
		yamlData["cascade"] = context
		name = fmt.Sprintf("%s/%s", context, name)
	}

	exists, err := checkExists(p, uri, name, objType)
	if err != nil {
		logger.Fatalf("", "Unable to check existence of %s %s", objType, name)
	}

	if exists {
		doUpdate(p, uri, objType, name, yamlData)
		logger.Successf("", "%s %s successfully updated", cases.Title(language.AmericanEnglish, cases.Compact).String(objType), name)
		return
	}
	doCreate(p, uri, objType, yamlData)
	logger.Successf("", "Successfully created %s %s", objType, name)
}
