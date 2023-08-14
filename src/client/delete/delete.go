package delete

import (
	"fmt"
	"net/http"
	"scaffold/client/auth"
	"scaffold/client/logger"
)

func DoDelete(profile, cascadeName string) {
	p := auth.ReadProfile(profile)

	uri := fmt.Sprintf("%s://%s:%s", p.Protocol, p.Host, p.Port)

	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("%s/api/v1/cascade/%s", uri, cascadeName)
	req, _ := http.NewRequest("DELETE", requestURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", p.APIToken))
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Fatalf("", "Delete request failed with error: %s", err.Error)
	}
	if resp.StatusCode >= 400 {
		logger.Fatalf("", "Delete request failed with status code %v", resp.StatusCode)
	}

	logger.Successf("", "Successfully deleted cascade!")
}
