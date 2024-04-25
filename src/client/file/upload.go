package file

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"scaffold/client/auth"
	"scaffold/client/logger"
)

func DoUpload(profile, cascade, inPath string) {
	p := auth.ReadProfile(profile)
	uri := fmt.Sprintf("%s://%s:%s", p.Protocol, p.Host, p.Port)

	file, _ := os.Open(inPath)
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filepath.Base(file.Name()))
	io.Copy(part, file)
	writer.Close()

	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("%s/api/v1/file/%s", uri, cascade)
	req, _ := http.NewRequest("POST", requestURL, body)
	req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", p.APIToken))
	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Fatalf("", "Encountered error downloading file: %v", err)
	}
	if resp.StatusCode >= 400 {
		logger.Fatalf("", "Error, got status code %d", resp.StatusCode)
	}

	logger.Successf("", "Successfully uploaded %s to %s filestore", inPath, cascade)

}
