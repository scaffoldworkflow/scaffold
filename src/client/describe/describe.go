package describe

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"scaffold/client/auth"
	"scaffold/client/constants"
	"scaffold/client/logger"
	"scaffold/client/utils"
	"strings"

	"gopkg.in/yaml.v2"
)

func DoDescribe(profile, object, context, format string) {
	if context == constants.ALL_CONTEXTS {
		logger.Fatalf("", "%s is not allowed for describe actions", constants.ALL_CONTEXTS)
	}

	logger.Debugf("", "Getting objects")
	p := auth.ReadProfile(profile)
	uri := fmt.Sprintf("%s://%s:%s", p.Protocol, p.Host, p.Port)

	logger.Debugf("", "Checking if object is valid")
	objects := []string{"workflow", "datastore", "state", "task", "file", "user", "input"}

	parts := strings.Split(object, "/")

	if !utils.Contains(objects, parts[0]) {
		logger.Fatalf("", "Invalid object type passed: '%s'. Valid object types are %v", object, objects)
	}

	if len(parts) == 1 {
		logger.Fatalf("", "Object passed in need to be of format '<object type>/<object name>")
	}

	logger.Debugf("", "Getting context")
	if parts[0] != "workflow" && parts[0] != "datastore" && parts[0] != "user" {
		if context == "" {
			context = p.Workflow
		}
		object = fmt.Sprintf("%s/%s/%s", parts[0], context, parts[1])
	}

	data := getJSON(p, uri, object)

	doDescribe(data, format)
}

func getJSON(p auth.ProfileObj, uri, object string) []byte {
	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("%s/api/v1/%s", uri, object)
	req, _ := http.NewRequest("GET", requestURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", p.APIToken))
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Fatalf("", "Encountered error: %s", err.Error())
	}
	if resp.StatusCode >= 400 {
		logger.Fatalf("", "Got status code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Fatalf("", "Error reading body: %s", err.Error())
	}
	resp.Body.Close()

	return body
}

func doDescribe(data []byte, format string) {
	var o map[string]interface{}

	err := json.Unmarshal(data, &o)
	if err != nil {
		logger.Fatalf("", "Unable to marshal object JSON: %s", err.Error())
	}

	if format == "yaml" {
		output, _ := yaml.Marshal(o)
		fmt.Println(string(output))
	} else {
		output, _ := json.MarshalIndent(o, "", "    ")
		fmt.Println(string(output))
	}
}
