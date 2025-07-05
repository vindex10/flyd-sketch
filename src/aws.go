package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	log "github.com/sirupsen/logrus"
)

var s3Client *s3.Client

func initS3Client() *s3.Client {
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(CFG.awsRegion), config.WithCredentialsProvider(aws.AnonymousCredentials{}))
	if err != nil {
		log.Fatal(err)
	}

	// Create an Amazon S3 service client
	s3Client = s3.NewFromConfig(cfg)
	return s3Client
}
