package filestore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"scaffold/server/config"
	"scaffold/server/constants"
	"scaffold/server/workflow"
	"strings"

	logger "github.com/jfcarter2358/go-logger"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var S3Config *aws.Config

type ObjectMetadata struct {
	Name     string `json:"name" bson:"name" yaml:"name"`
	Modified string `json:"modified" bson:"modified" yaml:"modified"`
	Workflow string `json:"workflow" bson:"workflow" yaml:"workflow"`
}

func InitBucket() {
	if config.Config.FileStore.Type == constants.FILESTORE_TYPE_S3 {
		bucket := aws.String(config.Config.FileStore.Bucket)

		// Configure to use MinIO Server
		if config.Config.FileStore.AccessKey != "" && config.Config.FileStore.SecretKey != "" {
			logger.Tracef("", "logging in with access and secret keys")
			S3Config = &aws.Config{
				Credentials:      credentials.NewStaticCredentials(config.Config.FileStore.AccessKey, config.Config.FileStore.SecretKey, ""),
				Endpoint:         aws.String(fmt.Sprintf("%s://%s:%d", config.Config.FileStore.Protocol, config.Config.FileStore.Host, config.Config.FileStore.Port)),
				Region:           aws.String(config.Config.FileStore.Region),
				DisableSSL:       aws.Bool(false),
				S3ForcePathStyle: aws.Bool(true),
			}
		} else {
			logger.Tracef("", "logging in with AWS credentials")
			S3Config = &aws.Config{
				Endpoint:         aws.String(fmt.Sprintf("%s://%s:%d", config.Config.FileStore.Protocol, config.Config.FileStore.Host, config.Config.FileStore.Port)),
				Region:           aws.String(config.Config.FileStore.Region),
				DisableSSL:       aws.Bool(false),
				S3ForcePathStyle: aws.Bool(true),
			}
		}
		session, err := session.NewSession(S3Config)
		if err != nil {
			panic(err)
		}

		client := s3.New(session)

		cparams := &s3.CreateBucketInput{
			Bucket: bucket, // Required
		}

		buckets, err := client.ListBuckets(nil)
		if err != nil {
			panic(err)
		}

		alreadyExists := false
		for _, bucket := range buckets.Buckets {
			if *bucket.Name == config.Config.FileStore.Bucket {
				alreadyExists = true
				break
			}
		}

		if !alreadyExists {
			// Create a new bucket using the CreateBucket call.
			_, err := client.CreateBucket(cparams)
			if err != nil {
				logger.Errorf("", "Encountered error with creating bucket: %v", err)
				// panic(err)
			}
		}
	}
}

func GetFile(inputPath, outputPath string) error {
	switch config.Config.FileStore.Type {
	case constants.FILESTORE_TYPE_S3:
		return doS3Download(inputPath, outputPath)
	case constants.FILESTORE_TYPE_ARTIFACTORY:
		return doArtifactoryDownload(inputPath, outputPath)
	}
	return fmt.Errorf("invalid filestore type: %s", config.Config.FileStore.Type)
}

func UploadFile(inputPath, outputPath string) error {
	switch config.Config.FileStore.Type {
	case constants.FILESTORE_TYPE_S3:
		return doS3Upload(inputPath, outputPath)
	case constants.FILESTORE_TYPE_ARTIFACTORY:
		return doArtifactoryUpload(inputPath, outputPath)
	}
	return fmt.Errorf("invalid filestore type: %s", config.Config.FileStore.Type)
}

func ListObjects() (map[string]ObjectMetadata, error) {
	switch config.Config.FileStore.Type {
	case constants.FILESTORE_TYPE_S3:
		return doS3List()
	case constants.FILESTORE_TYPE_ARTIFACTORY:
		return doArtifactoryList()
	}
	return map[string]ObjectMetadata{}, fmt.Errorf("invalid filestore type: %s", config.Config.FileStore.Type)
}

func doArtifactoryDownload(inputPath, outputPath string) error {
	uri := fmt.Sprintf("%s://%s:%d/artifactory/%s", config.Config.FileStore.Protocol, config.Config.FileStore.Host, config.Config.FileStore.Port, config.Config.FileStore.Bucket)

	// Create the file
	out, err := os.Create(outputPath)
	if err != nil {
		return nil
	}
	defer out.Close()

	// Get the data
	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("%s/%s", uri, inputPath)
	req, _ := http.NewRequest("GET", requestURL, nil)
	req.SetBasicAuth(config.Config.FileStore.AccessKey, config.Config.FileStore.SecretKey)
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("got status code %d on file download", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Fatalf("", "Encountered error reading body: %s", err.Error())
	}
	strBody := string(body)
	lines := strings.Split(strBody, "\n")
	lines = lines[4 : len(lines)-2]
	data := []byte(strings.Join(lines, "\n"))

	// Writer the body to file
	// _, err = io.Copy(out, resp.Body)
	_, err = out.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func doArtifactoryList() (map[string]ObjectMetadata, error) {
	output := make(map[string]ObjectMetadata)
	workflows, _ := workflow.GetAllWorkflows()
	uri := fmt.Sprintf("%s://%s:%d/artifactory/%s", config.Config.FileStore.Protocol, config.Config.FileStore.Host, config.Config.FileStore.Port, config.Config.FileStore.Bucket)
	for _, c := range workflows {
		// Get the data
		httpClient := &http.Client{}
		requestURL := fmt.Sprintf("%s/%s", uri, c.Name)
		req, _ := http.NewRequest("GET", requestURL, nil)
		req.SetBasicAuth(config.Config.FileStore.AccessKey, config.Config.FileStore.SecretKey)
		resp, err := httpClient.Do(req)
		if err != nil {
			return map[string]ObjectMetadata{}, err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			return map[string]ObjectMetadata{}, fmt.Errorf("got status code %d on file list", resp.StatusCode)
		}

		var data map[string]interface{}
		//Read the response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return map[string]ObjectMetadata{}, err
		}
		json.Unmarshal(body, &data)
		resp.Body.Close()

		logger.Debugf("", "Artifactory list response: %v", data)

		if data["children"] != nil {
			for _, item := range data["children"].([]interface{}) {
				itemMap := item.(map[string]interface{})
				path := itemMap["uri"].(string)
				name := path[1:]
				lastModified, err := getArtifactoryLastModified(fmt.Sprintf("%s/%s", requestURL, name))
				if err != nil {
					return map[string]ObjectMetadata{}, err
				}
				output[name] = ObjectMetadata{
					Name:     fmt.Sprintf("%s/%s", c.Name, name),
					Modified: lastModified,
					Workflow: c.Name,
				}
			}
		}
	}
	return output, nil
}

func getArtifactoryLastModified(url string) (string, error) {
	httpClient := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(config.Config.FileStore.AccessKey, config.Config.FileStore.SecretKey)
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("got status code %d on file last modified get", resp.StatusCode)
	}

	var data map[string]interface{}
	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	json.Unmarshal(body, &data)
	resp.Body.Close()

	return data["lastModified"].(string), nil
}

func doArtifactoryUpload(inputPath, outputPath string) error {
	uri := fmt.Sprintf("%s://%s:%d/artifactory/%s", config.Config.FileStore.Protocol, config.Config.FileStore.Host, config.Config.FileStore.Port, config.Config.FileStore.Bucket)

	file, _ := os.Open(inputPath)
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filepath.Base(file.Name()))
	io.Copy(part, file)
	writer.Close()

	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("%s/%s", uri, outputPath)
	req, _ := http.NewRequest("PUT", requestURL, body)
	req.SetBasicAuth(config.Config.FileStore.AccessKey, config.Config.FileStore.SecretKey)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("got status code %d on file upload", resp.StatusCode)
	}

	return nil
}

func doS3Download(inputPath, outputPath string) error {
	session, err := session.NewSession(S3Config)
	if err != nil {
		panic(err)
	}
	downloader := s3manager.NewDownloader(session)
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = downloader.Download(
		file,
		&s3.GetObjectInput{
			Bucket: aws.String(config.Config.FileStore.Bucket),
			Key:    aws.String(inputPath),
		},
	)

	return err
}

func doS3List() (map[string]ObjectMetadata, error) {
	session, err := session.NewSession(S3Config)
	if err != nil {
		panic(err)
	}
	svc := s3.New(session)

	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(config.Config.FileStore.Bucket)})
	if err != nil {
		return map[string]ObjectMetadata{}, err
	}

	output := make(map[string]ObjectMetadata)

	for _, item := range resp.Contents {
		workflow := strings.Split(*item.Key, "/")[0]
		output[*item.Key] = ObjectMetadata{
			Name:     *item.Key,
			Modified: (*item.LastModified).Format("2006-01-02T15:04:05Z"),
			Workflow: workflow,
		}
	}
	return output, nil
}

func doS3Upload(inputPath, outputPath string) error {
	session, err := session.NewSession(S3Config)
	if err != nil {
		panic(err)
	}
	uploader := s3manager.NewUploader(session)
	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(config.Config.FileStore.Bucket),
		Key:    aws.String(outputPath),
		Body:   file,
	})

	return err
}
