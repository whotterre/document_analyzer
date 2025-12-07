package initializers

import (
	"context"
	"log"
	"whotterre/doculyzer/internal/config"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var S3Client *s3.Client

func ConnectToS3() {
	cfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(config.AppConfig.AWSRegion),
	)
	if err != nil {
		log.Fatal("Failed to load AWS config")
	}

	S3Client = s3.NewFromConfig(cfg)
}
