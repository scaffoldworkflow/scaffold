package describe

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"scaffold/client/auth"
	"scaffold/client/logger"
	"scaffold/client/objects"
	"scaffold/client/utils"
	"strings"

	"gopkg.in/yaml.v2"
)

func DoDescribe(profile, object, context, format string) {
	logger.Debugf("", "Getting objects")
	p := auth.ReadProfile(profile)
	uri := fmt.Sprintf("%s://%s:%s", p.Protocol, p.Host, p.Port)

	logger.Debugf("", "Checking if object is valid")
	objects := []string{"cascade", "datastore", "state", "task"}

	parts := strings.Split(object, "/")

	if !utils.Contains(objects, parts[0]) {
		logger.Fatalf("Invalid object type passed: '%s'. Valid object types are 'cascade', 'datastore', 'state', 'task'", object)
	}

	if len(parts) == 1 {
		logger.Fatalf("", "Object passed in need to be of format '<object type>/<object name>")
	}

	logger.Debugf("", "Getting context")
	if parts[0] != "cascade" && parts[0] != "datastore" {
		if context == "" {
			context = p.Cascade
		}
		object = fmt.Sprintf("%s/%s/%s", parts[0], context, parts[1])
	}

	data := getJSON(p, uri, object)

	switch parts[0] {
	case "cascade":
		describeCascade(data, format)
	case "state":
		describeState(data, context, format)
	case "task":
		describeTask(data, context, format)
	case "datastore":
		describeDataStore(data, format)
	}
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

func describeCascade(data []byte, format string) {
	var c objects.Cascade

	err := json.Unmarshal(data, &c)
	if err != nil {
		logger.Fatalf("", "Unable to marshal datastores JSON: %s", err.Error())
	}

	if format == "yaml" {
		output, _ := yaml.Marshal(c)
		fmt.Println(string(output))
	} else {
		output, _ := json.MarshalIndent(c, "", "    ")
		fmt.Println(string(output))
	}
}

func describeDataStore(data []byte, format string) {
	var d objects.DataStore

	err := json.Unmarshal(data, &d)
	if err != nil {
		logger.Fatalf("", "Unable to marshal datastores JSON: %s", err.Error())
	}

	if format == "yaml" {
		output, _ := yaml.Marshal(d)
		fmt.Println(string(output))
	} else {
		output, _ := json.MarshalIndent(d, "", "    ")
		fmt.Println(string(output))
	}
}

func describeState(data []byte, context, format string) {
	var s objects.State

	err := json.Unmarshal(data, &s)
	if err != nil {
		logger.Fatalf("", "Unable to marshal states JSON: %s", err.Error())
	}

	if format == "yaml" {
		output, _ := yaml.Marshal(s)
		fmt.Println(string(output))
	} else {
		output, _ := json.MarshalIndent(s, "", "    ")
		fmt.Println(string(output))
	}
}

func describeTask(data []byte, context, format string) {
	var t objects.Task

	err := json.Unmarshal(data, &t)
	if err != nil {
		logger.Fatalf("", "Unable to marshal tasks JSON: %s", err.Error())
	}

	if format == "yaml" {
		output, _ := yaml.Marshal(t)
		fmt.Println(string(output))
	} else {
		output, _ := json.MarshalIndent(t, "", "    ")
		fmt.Println(string(output))
	}
}
