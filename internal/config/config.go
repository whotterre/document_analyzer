package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost        string
	DBUser        string
	DBPassword    string
	DBName        string
	DBPort        string
	AWSRegion     string
	S3Bucket      string
	OpenRouterKey string
}

var AppConfig *Config

func LoadConfig() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	AppConfig = &Config{
		DBHost:        os.Getenv("DB_HOST"),
		DBUser:        os.Getenv("DB_USER"),
		DBPassword:    os.Getenv("DB_PASSWORD"),
		DBName:        os.Getenv("DB_NAME"),
		DBPort:        os.Getenv("DB_PORT"),
		AWSRegion:     os.Getenv("AWS_REGION"),
		S3Bucket:      os.Getenv("S3_BUCKET_NAME"),
		OpenRouterKey: os.Getenv("OPENROUTER_API_KEY"),
	}
}
