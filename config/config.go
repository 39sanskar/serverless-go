package config

import (
	"errors"
	"os"
)

// Config holds all application configurations
type Config struct {
	AWSRegion string
	TableName string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		return nil, errors.New("AWS_REGION environment variable not set")
	}

	tableName := os.Getenv("DYNAMODB_TABLE_NAME")
	if tableName == "" {
		return nil, errors.New("DYNAMODB_TABLE_NAME environment variable not set")
	}

	return &Config{
		AWSRegion: region,
		TableName: tableName,
	}, nil
}