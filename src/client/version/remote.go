package version

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"scaffold/client/auth"
	"scaffold/client/logger"
)

func DoRemote(profile string) {
	logger.Debugf("", "Getting objects")
	p := auth.ReadProfile(profile)
	uri := fmt.Sprintf("%s://%s:%s", p.Protocol, p.Host, p.Port)

	var result map[string]string
	data := getJSON(p, uri)
	err := json.Unmarshal(data, &result)
	if err != nil {
		logger.Fatalf("", "Unable to unmarshal version JSON: %s", err.Error())
	}

	fmt.Printf("Scaffold Remote Version: %s\n", result["version"])
}

func getJSON(p auth.ProfileObj, uri string) []byte {
	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("%s/health/healthy", uri)
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
