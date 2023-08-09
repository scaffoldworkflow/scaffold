package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"scaffold/client/auth"
	"scaffold/client/logger"
)

var Token = ""

type TokenResponse struct {
	Token string `json:"token"`
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func DoConfig(host, port, protocol, wsPort, profile, username, password string) {
	uri := fmt.Sprintf("%s://%s:%s/auth/token/%s/client", protocol, host, port, username)

	var obj TokenResponse

	httpClient := &http.Client{}
	req, _ := http.NewRequest("POST", uri, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", basicAuth(username, password)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Fatalf("", "Token request failed with error %s", err.Error())
	}
	if resp.StatusCode >= 400 {
		logger.Fatalf("", "Request failed with status code %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Fatalf("", "Encountered error reading body: %s", err.Error())
	}
	err = json.Unmarshal([]byte(body), &obj)
	if err != nil {
		logger.Fatalf("", "Encountered error unmarshalling token JSON: %s", err.Error())
	}

	auth.WriteProfile(profile, protocol, host, port, wsPort, obj.Token)

	logger.Successf("", "Successfully configured profile '%s'", profile)
}
