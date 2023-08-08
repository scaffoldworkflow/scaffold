package cascade

import (
	"fmt"
	"os"
	"scaffold/client/logger"
	"scaffold/client/utils"

	"gopkg.in/yaml.v3"
)

func DoApply(host, port, fileName string) {
	uri := fmt.Sprintf("http://%s:%s", host, port)

	var yamlData map[string]interface{}

	fileData, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(fileData, &yamlData)
	if err != nil {
		panic(err)
	}

	exists := false

	// Check to see if a cascade already exists
	// requestURL := fmt.Sprintf("%s/api/v1/cascade/%s", uri, yamlData["name"].(string))
	// logger.Debugf("", "RequestURL: %s", requestURL)
	// resp, err := http.Get(requestURL)
	// if err != nil {
	// 	logger.Fatalf("", "Encountered error checking for existing cascade: %s", err.Error())
	// }
	// logger.Debugf("", "Status code: %d", resp.StatusCode)
	// if resp.StatusCode == 200 {
	// 	exists = true
	// }

	out, err := utils.SendGet(uri, fmt.Sprintf("api/v1/cascade/%s", yamlData["name"].(string)))
	logger.Debugf("", "response contents: %v", out)
	logger.Debugf("", "response error: %v", err)
	if err == nil {
		exists = true
	}

	if exists {
		logger.Debugf("", "Already exists!")
		contents, err := utils.SendPut(uri, fmt.Sprintf("api/v1/cascade/%s", yamlData["name"].(string)), yamlData)
		if err != nil {
			logger.Fatalf("", "Encountered error %s with contents %v\n", err.Error(), contents)
		}
		logger.Debugf("", "Response contents: %v", contents)
	} else {
		logger.Debugf("", "Doesn't exist")
		contents, err := utils.SendPost(uri, "api/v1/cascade", yamlData)
		if err != nil {
			logger.Fatalf("", "Encountered error %s with contents %v\n", err.Error(), contents)
		}
	}
	logger.Successf("", "Successfully uploaded cascade!")
}
