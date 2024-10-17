package services

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

type SesClient struct {
	svc   *ses.Client
	email string
}

func NewSesClient() (*SesClient, error) {
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

	svc := ses.NewFromConfig(cfg)

	return &SesClient{
		svc:   svc,
		email: os.Getenv("AWS_EMAIL"),
	}, nil
}

func (sesClient *SesClient) SendVerificationEmail(recipient, subject, body string) error {

	input := &ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{recipient},
		},
		Message: &types.Message{
			Body: &types.Body{
				Text: &types.Content{
					Data: &body,
				},
			},
			Subject: &types.Content{
				Data: &subject,
			},
		},
		Source: aws.String(sesClient.email),
	}

	_, emailErr := sesClient.svc.SendEmail(context.TODO(), input)

	if emailErr != nil {
		return emailErr
	}

	return nil
}
