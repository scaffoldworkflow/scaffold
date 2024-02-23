package file

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"scaffold/client/auth"
	"scaffold/client/logger"
)

func DoDownload(profile, cascade, name, outPath string) {
	p := auth.ReadProfile(profile)
	uri := fmt.Sprintf("%s://%s:%s", p.Protocol, p.Host, p.Port)

	// Create the file
	out, err := os.Create(outPath)
	if err != nil {
		logger.Fatalf("", "Error creating output file path: %s", err.Error())
	}
	defer out.Close()

	// Get the data
	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("%s/api/v1/file/%s/%s/download", uri, cascade, name)
	req, _ := http.NewRequest("GET", requestURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", p.APIToken))
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Fatalf("", "Encountered error downloading file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		logger.Fatalf("", "Error, got status code %d", resp.StatusCode)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		logger.Fatalf("", "Error writing file: %s", err.Error())
	}

	logger.Successf("", "Successfully downloaded %s from %s filestore to %s", name, cascade, outPath)
}
