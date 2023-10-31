package filestore

import (
	"fmt"
	"os"
	"scaffold/server/config"
	"scaffold/server/logger"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var S3Config *aws.Config

type ObjectMetadata struct {
	Name     string `json:"name" bson:"name"`
	Modified string `json:"modified" bson:"modified"`
	Cascade  string `json:"cascade" bson:"cascade"`
}

func InitBucket() {
	bucket := aws.String(config.Config.FileStore.Bucket)

	// Configure to use MinIO Server
	var S3Config *aws.Config
	if config.Config.FileStore.AccessKey != "" && config.Config.FileStore.SecretKey != "" {
		S3Config = &aws.Config{
			Credentials:      credentials.NewStaticCredentials(config.Config.FileStore.AccessKey, config.Config.FileStore.SecretKey, ""),
			Endpoint:         aws.String(fmt.Sprintf("%s://%s:%d", config.Config.FileStore.Protocol, config.Config.FileStore.Host, config.Config.FileStore.Port)),
			Region:           aws.String(config.Config.FileStore.Region),
			DisableSSL:       aws.Bool(false),
			S3ForcePathStyle: aws.Bool(true),
		}
	} else {
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

func GetFile(inputPath, outputPath string) error {
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

func UploadFile(inputPath, outputPath string) error {
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

func ListObjects() (map[string]ObjectMetadata, error) {
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
		output[*item.Key] = ObjectMetadata{
			Name:     *item.Key,
			Modified: (*item.LastModified).Format("2006-01-02T15:04:05Z"),
		}
	}
	return output, nil
}
