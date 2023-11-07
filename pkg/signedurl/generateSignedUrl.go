package signedurl

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/zone/IStyle/config"
)

func GetSignedUrl(key string) (string, error) {

	env, err := config.LoadConfig()

	if err != nil {
		return "", err
	}

	awsSession, err := session.NewSession(&aws.Config{
		Region:      aws.String("eu-north-1"),
		Credentials: credentials.NewStaticCredentials(env.S3_ACCESS_KEY, env.S3_SECRET_KEY, ""),
	})

	if err != nil {
		return "", err
	}
	// Create S3 service client
	svc := s3.New(awsSession)
	req, _ := svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(env.S3_BUCKET),
		Key:    aws.String(key),
	})

	str, err := req.Presign(15 * time.Minute)

	if err != nil {
		return "", err
	}

	return str, nil
}
