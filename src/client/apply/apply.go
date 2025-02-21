package apply

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"scaffold/client/auth"
	"strings"

	"gopkg.in/yaml.v3"

	logger "github.com/jfcarter2358/go-logger"
)

// var ValidObjects = []string{"alert", "release", "project", "runbook", "user"}

func DoApply(profile, fileName string) {
	logger.Debugf("", "Applying object")
	p := auth.ReadProfile(profile)
	uri := fmt.Sprintf("%s://%s:%s", p.Protocol, p.Host, p.Port)

	logger.Debugf("", "Checking if object is valid")

	// if !utils.Contains(ValidObjects, object) {
	// 	logger.Fatalf("", "Invalid object type passed: '%s'. Valid object types are '%s'", object, strings.Join(ValidObjects, "', '"))
	// }

	doApply(profile, fileName, uri)
}

func doUpdate(p auth.ProfileObj, uri, object string, data map[string]interface{}) {
	postBody, _ := json.Marshal(data)
	postBodyBuffer := bytes.NewBuffer(postBody)

	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("%s/api/v1/%s", uri, object)
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

func doApply(profile, fileName, uri string) {
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

	// TODO: Check if version field is present and error out if not
	version := yamlData["version"].(string)
	versionParts := strings.Split(version, "/")
	// TODO: Check length of version parts and error out if not equal to 2
	objectType := versionParts[1]

	// if objType == "runbook" || objType == "monitor" || objType == "alert" {
	// 	name = yamlData["id"].(string)
	// }

	// if objType != "workflow" && objType != "datastore" && objType != "user" && objType != "runbook" && objType != "monitor" && objType != "alert" {
	// 	yamlData["workflow"] = context
	// 	name = fmt.Sprintf("%s/%s", context, name)
	// }

	doUpdate(p, uri, objectType, yamlData)
	logger.Successf("", "Manifest successfully applied")
}
