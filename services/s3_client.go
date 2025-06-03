package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Client struct {
	client *s3.Client
	bucket string
}

func NewS3Client() (*S3Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_KEY"),
			os.Getenv("AWS_SESSION_TOKEN"),
		),
		),
		config.WithRegion(os.Getenv("AWS_REGION")),
	)

	if err != nil {
		fmt.Println("Couldn't load default configuration. Have you set up your AWS account?")
		fmt.Println(err)
	}

	client := s3.NewFromConfig(cfg)

	return &S3Client{
		client: client,
		bucket: os.Getenv("AWS_BUCKET_NAME"),
	}, nil
}

func (s *S3Client) UploadImage(file multipart.File, fileName string) (string, error) {
	uploader := manager.NewUploader(s.client)
	result, uploadErr := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fileName),
		Body:   file,
	})

	if uploadErr != nil {
		return "", uploadErr
	}

	return result.Location, uploadErr
}

func (s *S3Client) DeleteImage(fileName string) error {
	var objectIds []types.ObjectIdentifier
	objectIds = append(objectIds, types.ObjectIdentifier{Key: aws.String(fileName)})

	deleter := manager.DeleteObjectsAPIClient(s.client)
	_, deleteErr := deleter.DeleteObjects(context.TODO(), &s3.DeleteObjectsInput{
		Bucket: aws.String(s.bucket),
		Delete: &types.Delete{Objects: objectIds},
	})

	return deleteErr
}
